package runner

import (
	"testing"
	"time"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

func TestResolveTimeout_FromAssertion(t *testing.T) {
	a := assertion.Assertion{Timeout: "5s"}
	got := resolveTimeout(a, 30*time.Second)
	if got != 5*time.Second {
		t.Errorf("resolveTimeout with YAML timeout = %v, want 5s", got)
	}
}

func TestResolveTimeout_FallbackToCLI(t *testing.T) {
	a := assertion.Assertion{} // no timeout set
	got := resolveTimeout(a, 30*time.Second)
	if got != 30*time.Second {
		t.Errorf("resolveTimeout without YAML timeout = %v, want 30s", got)
	}
}

func TestResolveTimeout_InvalidDuration(t *testing.T) {
	a := assertion.Assertion{Timeout: "not-a-duration"}
	got := resolveTimeout(a, 15*time.Second)
	if got != 15*time.Second {
		t.Errorf("resolveTimeout with invalid duration = %v, want fallback 15s", got)
	}
}

func TestPassResult(t *testing.T) {
	start := time.Now()
	r := passResult("test_assertion", start)
	if r.Name != "test_assertion" {
		t.Errorf("name = %q, want %q", r.Name, "test_assertion")
	}
	if r.Status != assertion.StatusPass {
		t.Errorf("status = %q, want PASS", r.Status)
	}
	if r.Detail != "" {
		t.Errorf("detail should be empty, got %q", r.Detail)
	}
}

func TestFailResult(t *testing.T) {
	start := time.Now()
	r := failResult("bad_test", start, "connection refused")
	if r.Name != "bad_test" {
		t.Errorf("name = %q, want %q", r.Name, "bad_test")
	}
	if r.Status != assertion.StatusFail {
		t.Errorf("status = %q, want FAIL", r.Status)
	}
	if r.Detail != "connection refused" {
		t.Errorf("detail = %q, want %q", r.Detail, "connection refused")
	}
}

func TestSkipResult(t *testing.T) {
	start := time.Now()
	r := skipResult("skipped_test", start, "env var not set")
	if r.Status != assertion.StatusSkip {
		t.Errorf("status = %q, want SKIP", r.Status)
	}
	if r.Detail != "env var not set" {
		t.Errorf("detail = %q, want %q", r.Detail, "env var not set")
	}
}

func TestRunAssertion_SkipFlag(t *testing.T) {
	a := assertion.Assertion{
		Name: "skipped",
		Skip: true,
	}
	r := runAssertion(a, "", 2*time.Second, "")
	if r.Status != assertion.StatusSkip {
		t.Errorf("expected SKIP for skip=true, got %s", r.Status)
	}
}

func TestRunAssertion_SkipUnlessEnv(t *testing.T) {
	a := assertion.Assertion{
		Name:         "env-gated",
		SkipUnlessEnv: "MCP_ASSERT_NONEXISTENT_TEST_VAR_12345",
	}
	r := runAssertion(a, "", 2*time.Second, "")
	if r.Status != assertion.StatusSkip {
		t.Errorf("expected SKIP for unset env var, got %s", r.Status)
	}
}
