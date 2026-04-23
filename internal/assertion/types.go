package assertion

import "time"

// Suite is a collection of assertion files loaded from a directory.
type Suite struct {
	Assertions []Assertion
	Dir        string
}

// Assertion defines a single test: call a tool with known inputs, check the output.
type Assertion struct {
	Name    string       `yaml:"name"`
	Server  ServerConfig `yaml:"server"`
	Setup   []ToolCall   `yaml:"setup"`
	Assert  AssertBlock  `yaml:"assert"`
	Timeout string       `yaml:"timeout"`
}

// ServerConfig specifies how to start the MCP server under test.
type ServerConfig struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

// ToolCall is a single MCP tool invocation.
type ToolCall struct {
	Tool string         `yaml:"tool"`
	Args map[string]any `yaml:"args"`
}

// AssertBlock defines the tool to call and the expected results.
type AssertBlock struct {
	Tool   string         `yaml:"tool"`
	Args   map[string]any `yaml:"args"`
	Expect Expect         `yaml:"expect"`
}

// Expect defines deterministic assertions on the tool result.
type Expect struct {
	Contains     []string          `yaml:"contains"`
	NotContains  []string          `yaml:"not_contains"`
	Equals       *string           `yaml:"equals"`
	JSONPath     map[string]any    `yaml:"json_path"`
	MinResults   *int              `yaml:"min_results"`
	MaxResults   *int              `yaml:"max_results"`
	NotEmpty      *bool             `yaml:"not_empty"`
	NotError      *bool             `yaml:"not_error"`
	IsError       *bool             `yaml:"is_error"`
	MatchesRegex  []string          `yaml:"matches_regex"`
	FileContains  map[string]string `yaml:"file_contains"`
	FileUnchanged []string          `yaml:"file_unchanged"`
	NetDelta      *int              `yaml:"net_delta"`
	InOrder       []string          `yaml:"in_order"`
}

// Result is the outcome of running a single assertion.
type Result struct {
	Name     string        `json:"name"`
	Status   Status        `json:"status"`
	Detail   string        `json:"detail,omitempty"`
	Duration time.Duration `json:"duration_ms"`
	Language string        `json:"language,omitempty"`
	Trial    int           `json:"trial,omitempty"`
}

// Status is the outcome of an assertion.
type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusSkip Status = "SKIP"
)
