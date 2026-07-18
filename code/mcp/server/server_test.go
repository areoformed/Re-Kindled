package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type fakeDevice struct{}

func (fakeDevice) Show(_ context.Context, input ShowInput) (map[string]any, error) {
	return map[string]any{"ok": true, "pages": len(input.Pages)}, nil
}

func (fakeDevice) Navigate(_ context.Context, direction string) (map[string]any, error) {
	return map[string]any{"ok": true, "direction": direction}, nil
}

func (fakeDevice) SetType(_ context.Context, preset string) (map[string]any, error) {
	return map[string]any{"ok": true, "preset": preset}, nil
}

func (fakeDevice) Status(context.Context) (map[string]any, error) {
	return map[string]any{"connected": true, "reader_running": true, "active_preset": "rekindled-mono-air"}, nil
}

func (fakeDevice) Presets() ([]PresetSummary, error) {
	return []PresetSummary{{
		ID:                    "rekindled-mono-air",
		Label:                 "ReKindled Mono Air",
		BodyFBInkPixels:       49,
		LineGapPixels:         8,
		LineSpacingMultiplier: 1.163265306122449,
	}}, nil
}

func (fakeDevice) TypeLab(input TypeLabInput) (map[string]any, error) {
	return map[string]any{"preset": input.Preset, "ok": true}, nil
}

func TestProtocolDiscoveryAndHelp(t *testing.T) {
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"rekindled://help"}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"rekindled_help","arguments":{"topic":"gestures"}}}`,
	}, "\n") + "\n"
	var output strings.Builder
	if err := NewServer(fakeDevice{}).Serve(context.Background(), strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("got %d response lines, want 4:\n%s", len(lines), output.String())
	}
	responses := make([]map[string]any, len(lines))
	for index, line := range lines {
		if err := json.Unmarshal([]byte(line), &responses[index]); err != nil {
			t.Fatalf("response %d is not one-line JSON: %v\n%s", index+1, err, line)
		}
	}
	initialize := responses[0]["result"].(map[string]any)
	if initialize["protocolVersion"] != "2025-11-25" || initialize["instructions"] == "" {
		t.Fatalf("incomplete initialize result: %#v", initialize)
	}
	tools := responses[1]["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 7 {
		t.Fatalf("got %d tools, want 7", len(tools))
	}
	resource := responses[2]["result"].(map[string]any)["contents"].([]any)[0].(map[string]any)
	if !strings.Contains(resource["text"].(string), "# ReKindled MCP") {
		t.Fatal("manual resource omitted its heading")
	}
	help := responses[3]["result"].(map[string]any)
	if help["isError"] == true {
		t.Fatalf("help failed: %#v", help)
	}
}

func TestEncodeDocument(t *testing.T) {
	document, err := encodeDocument(ShowInput{Title: "A Letter", Pages: []string{"first", "second"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "@title: A Letter\nfirst\n---PAGE---\nsecond\n"
	if document != want {
		t.Fatalf("document = %q, want %q", document, want)
	}
}

func TestEncodeDocumentRejectsReservedSeparator(t *testing.T) {
	_, err := encodeDocument(ShowInput{Pages: []string{"first\n---PAGE---\nsecond"}})
	if err == nil {
		t.Fatal("expected reserved separator rejection")
	}
}

func TestUnsupportedVersionNegotiatesLatest(t *testing.T) {
	result, rpcErr := NewServer(fakeDevice{}).initialize(json.RawMessage(`{"protocolVersion":"1900-01-01"}`))
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	if result.(map[string]any)["protocolVersion"] != latestProtocolVersion {
		t.Fatalf("result = %#v", result)
	}
}
