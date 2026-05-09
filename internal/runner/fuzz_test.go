package runner

import (
	"testing"

	"github.com/blackwell-systems/mcp-assert/internal/report"
)

func TestConvertFuzzFailures(t *testing.T) {
	input := []FuzzInputResult{
		{Label: "empty_args", Status: FuzzCrash, Detail: "server died"},
		{Label: "null_values", Status: FuzzTimeout, Detail: "timed out after 5s"},
	}
	got := convertFuzzFailures(input)

	if len(got) != 2 {
		t.Fatalf("expected 2 failures, got %d", len(got))
	}

	if got[0].Label != "empty_args" {
		t.Errorf("[0].Label = %q, want %q", got[0].Label, "empty_args")
	}
	if got[0].Status != "crash" {
		t.Errorf("[0].Status = %q, want %q", got[0].Status, "crash")
	}
	if got[0].Detail != "server died" {
		t.Errorf("[0].Detail = %q, want %q", got[0].Detail, "server died")
	}

	if got[1].Label != "null_values" {
		t.Errorf("[1].Label = %q, want %q", got[1].Label, "null_values")
	}
	if got[1].Status != "timeout" {
		t.Errorf("[1].Status = %q, want %q", got[1].Status, "timeout")
	}
}

func TestConvertFuzzFailures_Empty(t *testing.T) {
	got := convertFuzzFailures(nil)
	if len(got) != 0 {
		t.Errorf("expected 0 failures for nil input, got %d", len(got))
	}
}

func TestConvertFuzzToolReports(t *testing.T) {
	input := []FuzzToolResult{
		{
			Tool:       "read_file",
			Runs:       10,
			Passed:     8,
			DurationMS: 500,
			Failures: []FuzzInputResult{
				{Label: "empty", Status: FuzzCrash, Detail: "crash"},
				{Label: "huge", Status: FuzzTimeout, Detail: "timeout"},
			},
		},
	}
	got := convertFuzzToolReports(input)

	if len(got) != 1 {
		t.Fatalf("expected 1 tool report, got %d", len(got))
	}
	if got[0].Tool != "read_file" {
		t.Errorf("tool = %q, want %q", got[0].Tool, "read_file")
	}
	if got[0].Runs != 10 {
		t.Errorf("runs = %d, want 10", got[0].Runs)
	}
	if got[0].Passed != 8 {
		t.Errorf("passed = %d, want 8", got[0].Passed)
	}
	if len(got[0].Failures) != 2 {
		t.Errorf("failures = %d, want 2", len(got[0].Failures))
	}
}

func TestFuzzStatusConstants(t *testing.T) {
	// Verify string representations match expected report contract.
	tests := []struct {
		status FuzzStatus
		want   string
	}{
		{FuzzPassed, "passed"},
		{FuzzCrash, "crash"},
		{FuzzTimeout, "timeout"},
		{FuzzProtocol, "protocol"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("FuzzStatus %v = %q, want %q", tt.status, string(tt.status), tt.want)
		}
	}
}

func TestConvertFuzzFailures_PreservesFields(t *testing.T) {
	input := []FuzzInputResult{
		{Label: "test", Status: FuzzProtocol, Detail: "internal error: stack trace..."},
	}
	got := convertFuzzFailures(input)

	want := report.FuzzFailure{
		Label:  "test",
		Status: "protocol",
		Detail: "internal error: stack trace...",
	}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}
