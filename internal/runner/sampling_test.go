package runner

import (
	"testing"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
	"gopkg.in/yaml.v3"
)

func TestSamplingAssertBlock_YAML(t *testing.T) {
	input := `
tool: ask_llm
args:
  question: "What is 2+2?"
mock_text: "The answer is 4."
mock_model: "test-model"
expect:
  not_error: true
  contains: ["4"]
`
	var sb assertion.SamplingAssertBlock
	if err := yaml.Unmarshal([]byte(input), &sb); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if sb.Tool != "ask_llm" {
		t.Errorf("expected tool ask_llm, got %s", sb.Tool)
	}
	if sb.MockText != "The answer is 4." {
		t.Errorf("expected mock text, got %s", sb.MockText)
	}
	if sb.MockModel != "test-model" {
		t.Errorf("expected test-model, got %s", sb.MockModel)
	}
}

func TestSamplingConfig_Defaults(t *testing.T) {
	sb := assertion.SamplingAssertBlock{
		Tool:     "ask_llm",
		MockText: "response",
	}
	if sb.MockModel != "" {
		t.Errorf("expected empty model default, got %s", sb.MockModel)
	}
	// The runner fills in "mock" as default at runtime
}
