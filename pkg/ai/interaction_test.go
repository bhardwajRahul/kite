package ai

import (
	"encoding/json"
	"testing"
)

func TestParseInteractionRequestChoice(t *testing.T) {
	request, err := parseInteractionRequest(requestChoiceTool, map[string]interface{}{
		"name":        "resourceType",
		"title":       "  Pick a resource  ",
		"description": "  Choose one  ",
		"options": []interface{}{
			map[string]interface{}{"label": "Pod", "value": "pod", "description": "Workload"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.Kind != interactionKindChoice || request.Name != "resourceType" || request.Title != "Pick a resource" {
		t.Fatalf("unexpected request: %#v", request)
	}
	if len(request.Options) != 1 || request.Options[0].Value != "pod" {
		t.Fatalf("unexpected options: %#v", request.Options)
	}
}

func TestParseInteractionRequestForm(t *testing.T) {
	request, err := parseInteractionRequest(requestFormTool, map[string]interface{}{
		"title":        "  Create deployment  ",
		"description":  "  Fill the required fields  ",
		"submit_label": "  Create  ",
		"fields": []interface{}{
			map[string]interface{}{
				"name":          "image",
				"label":         "Image",
				"type":          "text",
				"required":      true,
				"placeholder":   " nginx:1.27 ",
				"description":   " Container image ",
				"default_value": " nginx:latest ",
			},
			map[string]interface{}{
				"name":     "replicas",
				"label":    "Replicas",
				"type":     "number",
				"required": true,
			},
			map[string]interface{}{
				"name":     "public",
				"label":    "Public",
				"type":     "switch",
				"required": false,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.Kind != interactionKindForm || request.Title != "Create deployment" || request.SubmitLabel != "Create" {
		t.Fatalf("unexpected request: %#v", request)
	}
	if len(request.Fields) != 3 {
		t.Fatalf("unexpected fields: %#v", request.Fields)
	}
	if request.Fields[0].Placeholder != "nginx:1.27" || request.Fields[0].DefaultValue != "nginx:latest" {
		t.Fatalf("expected trimmed field metadata: %#v", request.Fields[0])
	}
}

func TestBuildInteractionToolResultChoice(t *testing.T) {
	request := interactionRequest{
		Kind:  interactionKindChoice,
		Name:  "resourceType",
		Title: "Pick one",
		Options: []interactionOption{
			{Label: "Pod", Value: "pod"},
		},
	}

	got, err := buildInteractionToolResult(request, map[string]interface{}{"resourceType": "pod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "{\n  \"resourceType\": \"pod\"\n}"
	if got != want {
		t.Fatalf("unexpected result:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestBuildInteractionToolResultForm(t *testing.T) {
	request := interactionRequest{
		Kind:  interactionKindForm,
		Title: "Create deployment",
		Fields: []interactionField{
			{Name: "image", Label: "Image", Type: "text", Required: true},
			{Name: "replicas", Label: "Replicas", Type: "number", Required: true},
			{Name: "public", Label: "Public", Type: "switch"},
			{
				Name:    "mode",
				Label:   "Mode",
				Type:    "select",
				Options: []interactionOption{{Label: "Blue", Value: "blue"}},
			},
		},
	}

	got, err := buildInteractionToolResult(request, map[string]interface{}{
		"image":    " nginx:1.27 ",
		"replicas": "3",
		"public":   "yes",
		"mode":     "blue",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("failed to decode result: %v", err)
	}
	if decoded["image"] != "nginx:1.27" {
		t.Fatalf("unexpected image: %#v", decoded["image"])
	}
	if decoded["replicas"].(float64) != 3 {
		t.Fatalf("unexpected replicas: %#v", decoded["replicas"])
	}
	if decoded["public"].(bool) != true {
		t.Fatalf("unexpected public flag: %#v", decoded["public"])
	}
	if decoded["mode"] != "blue" {
		t.Fatalf("unexpected mode: %#v", decoded["mode"])
	}
}

func TestBuildInteractionEventData(t *testing.T) {
	data := buildInteractionEventData(
		requestChoiceTool,
		"call-1",
		"session-1",
		interactionRequest{
			Kind:        interactionKindChoice,
			Name:        "resourceType",
			Title:       "Pick a resource",
			Description: "Choose one",
			SubmitLabel: "Select",
			Options:     []interactionOption{{Label: "Pod", Value: "pod"}},
		},
	)

	if data["tool"] != requestChoiceTool {
		t.Fatalf("unexpected tool: %#v", data["tool"])
	}
	if data["tool_call_id"] != "call-1" {
		t.Fatalf("unexpected tool_call_id: %#v", data["tool_call_id"])
	}
	if data["session_id"] != "session-1" {
		t.Fatalf("unexpected session_id: %#v", data["session_id"])
	}
	if data["kind"] != interactionKindChoice || data["title"] != "Pick a resource" {
		t.Fatalf("unexpected request metadata: %#v", data)
	}
	if data["name"] != "resourceType" || data["description"] != "Choose one" || data["submit_label"] != "Select" {
		t.Fatalf("unexpected optional fields: %#v", data)
	}
}
