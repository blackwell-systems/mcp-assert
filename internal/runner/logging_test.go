package runner

import (
	"testing"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

func intPtr(i int) *int { return &i }

func TestCheckLogging_MinMessages(t *testing.T) {
	msgs := []assertion.LogMessage{{Level: "info", Data: "hello"}}
	err := assertion.CheckLogging(assertion.LoggingExpect{MinMessages: intPtr(1)}, msgs)
	if err != nil {
		t.Errorf("expected pass, got: %v", err)
	}
	err = assertion.CheckLogging(assertion.LoggingExpect{MinMessages: intPtr(5)}, msgs)
	if err == nil {
		t.Error("expected fail for min_messages=5 with 1 message")
	}
}

func TestCheckLogging_ContainsLevel(t *testing.T) {
	msgs := []assertion.LogMessage{
		{Level: "info", Data: "msg1"},
		{Level: "debug", Data: "msg2"},
	}
	err := assertion.CheckLogging(assertion.LoggingExpect{
		ContainsLevel: []string{"info"},
	}, msgs)
	if err != nil {
		t.Errorf("expected pass: %v", err)
	}
	err = assertion.CheckLogging(assertion.LoggingExpect{
		ContainsLevel: []string{"error"},
	}, msgs)
	if err == nil {
		t.Error("expected fail for missing error level")
	}
}

func TestCheckLogging_ContainsData(t *testing.T) {
	msgs := []assertion.LogMessage{{Level: "info", Data: "hello world"}}
	err := assertion.CheckLogging(assertion.LoggingExpect{
		ContainsData: []string{"hello"},
	}, msgs)
	if err != nil {
		t.Errorf("expected pass: %v", err)
	}
}

func TestCheckLogging_Empty(t *testing.T) {
	err := assertion.CheckLogging(assertion.LoggingExpect{}, nil)
	if err != nil {
		t.Errorf("expected pass for empty expectations: %v", err)
	}
}
