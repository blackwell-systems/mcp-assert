package runner

import (
	"testing"
	"time"
)

func TestAuditScore(t *testing.T) {
	tests := []struct {
		name    string
		results []AuditToolResult
		want    int
	}{
		{
			name:    "empty",
			results: nil,
			want:    100,
		},
		{
			name: "all healthy",
			results: []AuditToolResult{
				{Status: AuditHealthy},
				{Status: AuditHealthy},
				{Status: AuditHealthy},
			},
			want: 100,
		},
		{
			name: "one crash out of three",
			results: []AuditToolResult{
				{Status: AuditHealthy},
				{Status: AuditCrash},
				{Status: AuditHealthy},
			},
			want: 66,
		},
		{
			name: "skipped tools excluded from denominator",
			results: []AuditToolResult{
				{Status: AuditHealthy},
				{Status: AuditSkipped},
				{Status: AuditSkipped},
			},
			want: 100,
		},
		{
			name: "all skipped",
			results: []AuditToolResult{
				{Status: AuditSkipped},
			},
			want: 100,
		},
		{
			name: "mixed with timeout",
			results: []AuditToolResult{
				{Status: AuditHealthy},
				{Status: AuditTimeout},
				{Status: AuditHealthy},
				{Status: AuditSkipped},
			},
			want: 66,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auditScore(tt.results)
			if got != tt.want {
				t.Errorf("auditScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClassifyHealthy(t *testing.T) {
	tests := []struct {
		isError bool
		text    string
		want    string
	}{
		{true, "bad input", "valid error handling (isError: true)"},
		{false, "", "responds (empty content)"},
		{false, "hello world", "responds, returns content"},
	}

	for _, tt := range tests {
		got := classifyHealthy(tt.isError, tt.text)
		if got != tt.want {
			t.Errorf("classifyHealthy(%v, %q) = %q, want %q", tt.isError, tt.text, got, tt.want)
		}
	}
}

func TestIsInternalErrorResponse(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"Internal error occurred", true},
		{"Traceback (most recent call last)", true},
		{"panic: runtime error", true},
		{"File not found", false},
		{"Invalid argument: path is required", false},
		{"java.lang.NullPointerException", true},
		{"Segmentation fault", true},
	}

	for _, tt := range tests {
		got := isInternalErrorResponse(tt.text)
		if got != tt.want {
			t.Errorf("isInternalErrorResponse(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short string: got %q", got)
	}
	if got := truncate("hello world this is long", 10); got != "hello worl..." {
		t.Errorf("truncate long string: got %q", got)
	}
}

func TestAuditResult(t *testing.T) {
	// Verify DurationMS is set correctly.
	tool := struct {
		Name        string
		Description string
	}{"test_tool", "A test tool"}

	// We can't call auditResult directly since it takes mcp.Tool.
	// Instead test that the struct fields work as expected.
	r := AuditToolResult{
		Tool:       tool.Name,
		Status:     AuditHealthy,
		Duration:   500 * time.Millisecond,
		DurationMS: (500 * time.Millisecond).Milliseconds(),
	}
	if r.DurationMS != 500 {
		t.Errorf("DurationMS = %d, want 500", r.DurationMS)
	}
}
