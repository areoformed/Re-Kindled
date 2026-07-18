package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

const latestProtocolVersion = "2025-11-25"

var supportedProtocolVersions = map[string]bool{
	"2025-11-25": true,
	"2025-06-18": true,
}

type Device interface {
	Show(context.Context, ShowInput) (map[string]any, error)
	Navigate(context.Context, string) (map[string]any, error)
	SetType(context.Context, string) (map[string]any, error)
	Status(context.Context) (map[string]any, error)
	Presets() ([]PresetSummary, error)
	TypeLab(TypeLabInput) (map[string]any, error)
}

type Server struct {
	device Device
}

func NewServer(device Device) *Server {
	return &Server{device: device}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content           []toolContent `json:"content"`
	StructuredContent any           `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
}

func (server *Server) Serve(ctx context.Context, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			if err := encoder.Encode(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{Code: -32700, Message: "Parse error"}}); err != nil {
				return err
			}
			continue
		}
		if req.JSONRPC != "2.0" || req.Method == "" {
			if len(req.ID) > 0 {
				if err := encoder.Encode(response{JSONRPC: "2.0", ID: responseID(req.ID), Error: &rpcError{Code: -32600, Message: "Invalid Request"}}); err != nil {
					return err
				}
			}
			continue
		}
		if len(req.ID) == 0 {
			server.handleNotification(req)
			continue
		}

		result, rpcErr := server.handleRequest(ctx, req)
		out := response{JSONRPC: "2.0", ID: responseID(req.ID), Result: result, Error: rpcErr}
		if rpcErr != nil {
			out.Result = nil
		}
		if err := encoder.Encode(out); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read MCP stream: %w", err)
	}
	return nil
}

func responseID(id json.RawMessage) json.RawMessage {
	if len(id) == 0 {
		return json.RawMessage("null")
	}
	return id
}

func (server *Server) handleNotification(req request) {
	switch req.Method {
	case "notifications/initialized", "notifications/cancelled", "notifications/progress":
		return
	default:
		log.Printf("ignored notification %q", req.Method)
	}
}

func (server *Server) handleRequest(ctx context.Context, req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return server.initialize(req.Params)
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": toolDefinitions()}, nil
	case "tools/call":
		return server.callTool(ctx, req.Params)
	case "resources/list":
		return map[string]any{"resources": resourceDefinitions()}, nil
	case "resources/read":
		return server.readResource(ctx, req.Params)
	case "resources/templates/list":
		return map[string]any{"resourceTemplates": []any{}}, nil
	default:
		return nil, &rpcError{Code: -32601, Message: "Method not found"}
	}
}

func (server *Server) initialize(raw json.RawMessage) (any, *rpcError) {
	var params struct {
		ProtocolVersion string          `json:"protocolVersion"`
		Capabilities    json.RawMessage `json:"capabilities"`
		ClientInfo      json.RawMessage `json:"clientInfo"`
	}
	if err := decodeObject(raw, &params); err != nil {
		return nil, invalidParams(err)
	}
	if params.ProtocolVersion == "" {
		return nil, invalidParams(errors.New("protocolVersion is required"))
	}
	version := latestProtocolVersion
	if supportedProtocolVersions[params.ProtocolVersion] {
		version = params.ProtocolVersion
	}
	return map[string]any{
		"protocolVersion": version,
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    serverName,
			"title":   serverTitle,
			"version": serverVersion,
		},
		"instructions": instructions,
	}, nil
}

func (server *Server) callTool(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := decodeObject(raw, &params); err != nil {
		return nil, invalidParams(err)
	}
	if params.Name == "" {
		return nil, invalidParams(errors.New("name is required"))
	}

	var result map[string]any
	var text string
	var err error
	switch params.Name {
	case "rekindled_show":
		var input ShowInput
		if decodeErr := decodeArguments(params.Arguments, &input); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		result, err = server.device.Show(ctx, input)
		text = fmt.Sprintf("Displayed %d page(s) on ReKindled.", len(input.Pages))
	case "rekindled_navigate":
		var input struct {
			Direction string `json:"direction"`
		}
		if decodeErr := decodeArguments(params.Arguments, &input); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		result, err = server.device.Navigate(ctx, input.Direction)
		text = "Moved " + input.Direction + "."
	case "rekindled_set_type":
		var input struct {
			Preset string `json:"preset"`
		}
		if decodeErr := decodeArguments(params.Arguments, &input); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		result, err = server.device.SetType(ctx, input.Preset)
		text = "Selected typography preset " + input.Preset + "."
	case "rekindled_status":
		if decodeErr := requireNoArguments(params.Arguments); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		result, err = server.device.Status(ctx)
		text = statusText(result)
	case "rekindled_presets":
		if decodeErr := requireNoArguments(params.Arguments); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		var presets []PresetSummary
		presets, err = server.device.Presets()
		result = map[string]any{"presets": presets, "recommended": "rekindled-mono-air"}
		text = fmt.Sprintf("Found %d typography presets. The compact everyday default is rekindled-mono-air.", len(presets))
	case "rekindled_type_lab":
		var input TypeLabInput
		if decodeErr := decodeArguments(params.Arguments, &input); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		result, err = server.device.TypeLab(input)
		text = "Calculated a non-destructive typography recipe. Materialize it, preflight it on FBInk, and judge it on the physical panel before making it a default."
	case "rekindled_help":
		var input struct {
			Topic string `json:"topic"`
		}
		if decodeErr := decodeArguments(params.Arguments, &input); decodeErr != nil {
			return toolFailure(decodeErr), nil
		}
		if input.Topic == "" {
			input.Topic = "overview"
		}
		var found bool
		text, found = helpTopics[input.Topic]
		if !found {
			err = fmt.Errorf("unknown help topic %q", input.Topic)
		} else {
			result = map[string]any{"topic": input.Topic, "text": text, "manual_uri": "rekindled://help"}
		}
	default:
		return nil, &rpcError{Code: -32602, Message: "Unknown tool", Data: map[string]any{"name": params.Name}}
	}
	if err != nil {
		return toolFailure(err), nil
	}
	return toolResult{Content: []toolContent{{Type: "text", Text: text}}, StructuredContent: result}, nil
}

func toolFailure(err error) toolResult {
	return toolResult{
		Content: []toolContent{{Type: "text", Text: err.Error()}},
		StructuredContent: map[string]any{
			"ok":    false,
			"error": err.Error(),
		},
		IsError: true,
	}
}

func decodeObject(raw json.RawMessage, target any) error {
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage("{}")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func decodeArguments(raw json.RawMessage, target any) error {
	return decodeObject(raw, target)
}

func requireNoArguments(raw json.RawMessage) error {
	var arguments map[string]any
	if err := decodeObject(raw, &arguments); err != nil {
		return err
	}
	if len(arguments) != 0 {
		return errors.New("this tool takes no arguments")
	}
	return nil
}

func invalidParams(err error) *rpcError {
	return &rpcError{Code: -32602, Message: "Invalid params", Data: err.Error()}
}

func statusText(status map[string]any) string {
	if connected, ok := status["connected"].(bool); !ok || !connected {
		return "ReKindled is not connected."
	}
	if running, ok := status["reader_running"].(bool); ok && running {
		return fmt.Sprintf("ReKindled is connected; the reader is running with %v.", status["active_preset"])
	}
	return "ReKindled is connected; the reader is stopped. A show call will start it."
}
