package runner

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestElicitation_DeclineAction(t *testing.T) {
	h := &staticElicitationHandler{values: map[string]any{"action": "decline"}}
	result, err := h.Elicit(context.Background(), mcp.ElicitationRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != mcp.ElicitationResponseActionDecline {
		t.Errorf("expected decline, got %s", result.Action)
	}
}

func TestElicitation_CancelAction(t *testing.T) {
	h := &staticElicitationHandler{values: map[string]any{"action": "cancel"}}
	result, err := h.Elicit(context.Background(), mcp.ElicitationRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != mcp.ElicitationResponseActionCancel {
		t.Errorf("expected cancel, got %s", result.Action)
	}
}

func TestElicitation_DefaultAccept(t *testing.T) {
	h := &staticElicitationHandler{values: map[string]any{
		"content": map[string]any{"name": "test"},
	}}
	result, err := h.Elicit(context.Background(), mcp.ElicitationRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != mcp.ElicitationResponseActionAccept {
		t.Errorf("expected accept, got %s", result.Action)
	}
}

func TestElicitation_ActionFilteredFromContent(t *testing.T) {
	h := &staticElicitationHandler{values: map[string]any{
		"action": "accept",
		"name":   "test",
	}}
	result, err := h.Elicit(context.Background(), mcp.ElicitationRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, ok := result.Content.(map[string]any)
	if !ok {
		t.Fatalf("expected map content, got %T", result.Content)
	}
	if _, hasAction := content["action"]; hasAction {
		t.Error("action key should be filtered from content")
	}
	if _, hasName := content["name"]; !hasName {
		t.Error("name key should be present in content")
	}
}
