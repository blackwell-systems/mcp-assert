package runner

import (
	"testing"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/report"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestAuditResult_ErrorCodes(t *testing.T) {
	tool := mcp.Tool{Name: "test_tool"}

	tests := []struct {
		name      string
		status    AuditStatus
		wantCode  report.ErrorCode
	}{
		{"healthy maps to E000", AuditHealthy, report.E000},
		{"crash maps to E201", AuditCrash, report.E201},
		{"timeout maps to E202", AuditTimeout, report.E202},
		{"skipped has no code", AuditSkipped, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditResult(tool, tt.status, "test detail", false, true, time.Second)
			if result.ErrorCode != tt.wantCode {
				t.Errorf("ErrorCode = %q, want %q", result.ErrorCode, tt.wantCode)
			}
		})
	}
}

func TestAuditResult_IncludesErrorCode(t *testing.T) {
	tool := mcp.Tool{Name: "test_tool"}
	result := auditResult(tool, AuditCrash, "panic: nil pointer", false, false, 100*time.Millisecond)

	if result.ErrorCode != report.E201 {
		t.Errorf("ErrorCode = %q, want E201", result.ErrorCode)
	}
	if result.Status != AuditCrash {
		t.Errorf("Status = %q, want crash", result.Status)
	}
	if result.Detail != "panic: nil pointer" {
		t.Errorf("Detail = %q, want 'panic: nil pointer'", result.Detail)
	}
}
