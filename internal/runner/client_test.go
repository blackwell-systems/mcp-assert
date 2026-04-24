package runner

import "testing"

func TestExpandEnvVars_DollarBrace(t *testing.T) {
	t.Setenv("TEST_TOKEN", "secret123")
	got := expandEnvVars("${TEST_TOKEN}")
	if got != "secret123" {
		t.Errorf("expandEnvVars(${TEST_TOKEN}) = %q, want %q", got, "secret123")
	}
}

func TestExpandEnvVars_DollarNoBrace(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")
	got := expandEnvVars("$TEST_VAR")
	if got != "hello" {
		t.Errorf("expandEnvVars($TEST_VAR) = %q, want %q", got, "hello")
	}
}

func TestExpandEnvVars_UnsetVar(t *testing.T) {
	got := expandEnvVars("${UNSET_VAR_12345}")
	if got != "" {
		t.Errorf("expandEnvVars(${UNSET_VAR_12345}) = %q, want %q", got, "")
	}
}

func TestExpandEnvVars_NoVars(t *testing.T) {
	got := expandEnvVars("literal-value")
	if got != "literal-value" {
		t.Errorf("expandEnvVars(literal-value) = %q, want %q", got, "literal-value")
	}
}

func TestExpandEnvVars_MixedContent(t *testing.T) {
	t.Setenv("TEST_HOST", "example.com")
	got := expandEnvVars("https://${TEST_HOST}/api")
	if got != "https://example.com/api" {
		t.Errorf("expandEnvVars(https://${TEST_HOST}/api) = %q, want %q", got, "https://example.com/api")
	}
}

func TestExpandEnvVars_MultipleVars(t *testing.T) {
	t.Setenv("TEST_A", "foo")
	t.Setenv("TEST_B", "bar")
	got := expandEnvVars("${TEST_A}:${TEST_B}")
	if got != "foo:bar" {
		t.Errorf("expandEnvVars(${TEST_A}:${TEST_B}) = %q, want %q", got, "foo:bar")
	}
}
