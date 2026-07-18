package main

import (
	"context"
	"encoding/json"
)

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "rekindled_show",
			"title":       "Show text",
			"description": "Place a titled, deliberately paginated document on the Kindle. Each array item is one physical page. ReKindled preflights every page before replacing the visible document; overflow is returned as a tool error and the prior display is retained. An optional preset id selects type at the same time.",
			"inputSchema": objectSchema(map[string]any{
				"title": map[string]any{"type": "string", "maxLength": 120, "description": "Optional restrained running header; no line breaks."},
				"pages": map[string]any{
					"type":        "array",
					"minItems":    1,
					"maxItems":    64,
					"description": "UTF-8 page bodies in reading order. End each at a semantic boundary; do not include page separators.",
					"items":       map[string]any{"type": "string", "minLength": 1, "maxLength": 24576},
				},
				"preset": map[string]any{"type": "string", "pattern": "^[a-z0-9-]+$", "description": "Optional exact id from rekindled_presets, for example rekindled-mono-air."},
			}, []string{"pages"}),
			"annotations": mutationAnnotations("Show text", true),
		},
		{
			"name":        "rekindled_navigate",
			"title":       "Turn page",
			"description": "Move the live reader exactly one page. At the first or last page the request is harmless and the display stays where it is.",
			"inputSchema": objectSchema(map[string]any{
				"direction": map[string]any{"type": "string", "enum": []string{"next", "previous"}, "description": "The page-turn direction."},
			}, []string{"direction"}),
			"annotations": mutationAnnotations("Turn page", false),
		},
		{
			"name":        "rekindled_set_type",
			"title":       "Choose type",
			"description": "Select an installed typography preset and redraw the current document. The candidate layout is checked first; if the current pages do not fit, the active preset and display remain unchanged.",
			"inputSchema": objectSchema(map[string]any{
				"preset": map[string]any{"type": "string", "pattern": "^[a-z0-9-]+$", "description": "Exact id from rekindled_presets."},
			}, []string{"preset"}),
			"annotations": mutationAnnotations("Choose type", true),
		},
		{
			"name":        "rekindled_status",
			"title":       "Inspect display",
			"description": "Read cable connectivity, reader state, active preset, device-side footprint, recent log lines, and this MCP binary's footprint. This does not change the Kindle.",
			"inputSchema": objectSchema(map[string]any{}, nil),
			"annotations": readOnlyAnnotations("Inspect display"),
		},
		{
			"name":        "rekindled_presets",
			"title":       "List typography",
			"description": "List the prepared local font presets with face category, physical target, and FBInk body size. Use an exact id with rekindled_set_type or rekindled_show.",
			"inputSchema": objectSchema(map[string]any{}, nil),
			"annotations": readOnlyAnnotations("List typography"),
		},
		{
			"name":        "rekindled_type_lab",
			"title":       "Tune typography",
			"description": "Calculate a non-destructive candidate for body size, angular cap height, and absolute line pitch from an installed preset. This changes no files and does not touch the Kindle. Use either body_size_change_percent or target_cap_height_degrees, not both. line_pitch_change_percent is measured against the current physical baseline pitch, so simultaneous body-size changes do not hide the requested spacing change.",
			"inputSchema": objectSchema(map[string]any{
				"preset":                    map[string]any{"type": "string", "pattern": "^[a-z0-9-]+$", "description": "Exact id from rekindled_presets."},
				"body_size_change_percent":  map[string]any{"type": "number", "minimum": -40, "maximum": 100, "default": 0, "description": "Relative change to FBInk body pixels; -5 means five percent smaller."},
				"target_cap_height_degrees": map[string]any{"type": "number", "minimum": 0.15, "maximum": 1.5, "description": "Optional angular target that determines body pixels at the selected distance. Mutually exclusive with a non-zero body-size change."},
				"line_pitch_change_percent": map[string]any{"type": "number", "minimum": -20, "maximum": 100, "default": 0, "description": "Change in absolute baseline-to-baseline pixels relative to the current preset; 10 means ten percent more open."},
				"viewing_distance_inches":   map[string]any{"type": "number", "minimum": 6, "maximum": 60, "description": "Optional viewing distance for angular calculations; defaults to the preset calibration."},
			}, []string{"preset"}),
			"annotations": readOnlyAnnotations("Tune typography"),
		},
		{
			"name":        "rekindled_help",
			"title":       "Read ReKindled help",
			"description": "Return concise operating guidance for the overview, showing text, choosing type, Kindle gestures, or recovery. The complete manual is also available at rekindled://help.",
			"inputSchema": objectSchema(map[string]any{
				"topic": map[string]any{"type": "string", "enum": []string{"overview", "show", "type", "typography", "gestures", "recovery"}, "default": "overview", "description": "Help topic."},
			}, nil),
			"annotations": readOnlyAnnotations("Read ReKindled help"),
		},
	}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func mutationAnnotations(title string, idempotent bool) map[string]any {
	return map[string]any{
		"title":           title,
		"readOnlyHint":    false,
		"destructiveHint": false,
		"idempotentHint":  idempotent,
		"openWorldHint":   false,
	}
}

func readOnlyAnnotations(title string) map[string]any {
	return map[string]any{
		"title":           title,
		"readOnlyHint":    true,
		"destructiveHint": false,
		"idempotentHint":  true,
		"openWorldHint":   false,
	}
}

func resourceDefinitions() []map[string]any {
	return []map[string]any{
		{
			"uri":         "rekindled://typography",
			"name":        "rekindled-typography",
			"title":       "Typography lab guide",
			"description": "Sizing equations, line-pitch semantics, experiment loop, and physical validation rules.",
			"mimeType":    "text/markdown",
		},
		{
			"uri":         "rekindled://help",
			"name":        "rekindled-help",
			"title":       "ReKindled manual",
			"description": "Complete human- and agent-readable operating guide, including examples, gestures, and recovery.",
			"mimeType":    "text/markdown",
		},
		{
			"uri":         "rekindled://presets",
			"name":        "rekindled-presets",
			"title":       "Typography preset catalog",
			"description": "Prepared local type choices and their physical sizing.",
			"mimeType":    "application/json",
		},
		{
			"uri":         "rekindled://status",
			"name":        "rekindled-status",
			"title":       "Live display status",
			"description": "Cable connection, reader state, active type, footprint, and recent diagnostics.",
			"mimeType":    "application/json",
		},
	}
}

func (server *Server) readResource(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := decodeObject(raw, &params); err != nil {
		return nil, invalidParams(err)
	}
	var text, mimeType string
	switch params.URI {
	case "rekindled://help":
		text = manualMarkdown
		mimeType = "text/markdown"
	case "rekindled://presets":
		presets, err := server.device.Presets()
		if err != nil {
			return nil, &rpcError{Code: -32603, Message: "Could not read preset catalog", Data: err.Error()}
		}
		encoded, err := json.MarshalIndent(map[string]any{"presets": presets, "recommended": "rekindled-mono-air"}, "", "  ")
		if err != nil {
			return nil, &rpcError{Code: -32603, Message: "Could not encode preset catalog", Data: err.Error()}
		}
		text = string(encoded)
		mimeType = "application/json"
	case "rekindled://typography":
		text = typographyGuide
		mimeType = "text/markdown"
	case "rekindled://status":
		status, err := server.device.Status(ctx)
		if err != nil {
			return nil, &rpcError{Code: -32603, Message: "Could not read status", Data: err.Error()}
		}
		encoded, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return nil, &rpcError{Code: -32603, Message: "Could not encode status", Data: err.Error()}
		}
		text = string(encoded)
		mimeType = "application/json"
	default:
		return nil, &rpcError{Code: -32002, Message: "Resource not found", Data: map[string]any{"uri": params.URI}}
	}
	return map[string]any{
		"contents": []map[string]any{{
			"uri":      params.URI,
			"mimeType": mimeType,
			"text":     text,
		}},
	}, nil
}
