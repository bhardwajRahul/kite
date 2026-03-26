package ai

import (
	"reflect"
	"testing"
)

func TestBuildToolCallEventData(t *testing.T) {
	got := buildToolCallEventData(
		streamedToolCall{ID: "call-1", Name: "get_resource"},
		map[string]interface{}{"kind": "Pod", "name": "nginx"},
	)

	want := map[string]interface{}{
		"tool":         "get_resource",
		"tool_call_id": "call-1",
		"args":         map[string]interface{}{"kind": "Pod", "name": "nginx"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected tool call event data:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestBuildToolResultEventData(t *testing.T) {
	got := buildToolResultEventData("call-1", "get_resource", "ok", false)
	want := map[string]interface{}{
		"tool":         "get_resource",
		"tool_call_id": "call-1",
		"result":       "ok",
		"is_error":     false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected tool result event data:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestBuildActionRequiredEventData(t *testing.T) {
	got := buildActionRequiredEventData(
		streamedToolCall{ID: "call-1", Name: "create_resource"},
		"session-1",
		map[string]interface{}{"yaml": "apiVersion: v1"},
	)

	want := map[string]interface{}{
		"tool":         "create_resource",
		"tool_call_id": "call-1",
		"session_id":   "session-1",
		"args":         map[string]interface{}{"yaml": "apiVersion: v1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected action required event data:\nwant: %#v\ngot:  %#v", want, got)
	}
}
