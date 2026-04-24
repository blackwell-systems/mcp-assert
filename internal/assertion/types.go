// Package assertion defines the core types and validation logic for mcp-assert.
//
// The package is organized into focused files:
//   - types.go       — Suite, Assertion, Expect, Result, and all block types
//   - loader.go      — YAML file loading with subdirectory recursion
//   - checker.go     — 15 assertion type implementations (contains, json_path, etc.)
//   - trajectory.go  — 4 trajectory assertion types (order, presence, absence, args_contain)
//   - logging*.go    — Logging assertion block and checker
//   - sampling*.go   — Sampling assertion block type
package assertion

import "time"

// Suite is a collection of assertion files loaded from a directory.
type Suite struct {
	Assertions []Assertion
	Dir        string
}

// Assertion defines a single test: call a tool with known inputs, check the output.
// For trajectory assertions, set Trace and Trajectory instead of Assert.
// For resource assertions, set AssertResources instead of Assert.
// For prompt assertions, set AssertPrompts instead of Assert.
type Assertion struct {
	Name            string                `yaml:"name"`
	Server          ServerConfig          `yaml:"server"`
	Setup           []ToolCall            `yaml:"setup"`
	Assert          AssertBlock           `yaml:"assert"`
	AssertResources  *ResourceAssertBlock  `yaml:"assert_resources,omitempty"`
	AssertPrompts    *PromptAssertBlock    `yaml:"assert_prompts,omitempty"`
	AssertCompletion *CompletionAssertBlock `yaml:"assert_completion,omitempty"`
	AssertSampling   *SamplingAssertBlock   `yaml:"assert_sampling,omitempty"`
	AssertLogging    *LoggingAssertBlock    `yaml:"assert_logging,omitempty"`
	Timeout         string                `yaml:"timeout"`
	Skip            bool                  `yaml:"skip,omitempty"`
	Trace           []TraceEntry          `yaml:"trace,omitempty"`      // inline tool call sequence
	AuditLog        string                `yaml:"audit_log,omitempty"`  // path to agent-lsp JSONL audit log
	Trajectory      []TrajectoryAssertion `yaml:"trajectory,omitempty"` // sequence checks
}

// ResourceAssertBlock tests MCP resources (resources/list or resources/read).
// Set exactly one of List or Read. Optionally subscribe/unsubscribe to resource updates.
type ResourceAssertBlock struct {
	List               *ResourceListArgs `yaml:"list,omitempty"`                // call resources/list
	Read               string            `yaml:"read,omitempty"`               // URI to pass to resources/read
	Expect             Expect            `yaml:"expect"`
	Subscribe          string            `yaml:"subscribe,omitempty"`           // URI to subscribe to
	Unsubscribe        string            `yaml:"unsubscribe,omitempty"`         // URI to unsubscribe from
	ExpectNotification *bool             `yaml:"expect_notification,omitempty"` // expect a resource update notification
}

// ResourceListArgs holds parameters for resources/list (cursor for pagination).
type ResourceListArgs struct {
	Cursor string `yaml:"cursor,omitempty"`
}

// PromptAssertBlock tests MCP prompts (prompts/list or prompts/get).
// Set exactly one of List or Get.
type PromptAssertBlock struct {
	List   *PromptListArgs `yaml:"list,omitempty"` // call prompts/list
	Get    *PromptGetArgs  `yaml:"get,omitempty"`  // call prompts/get
	Expect Expect          `yaml:"expect"`
}

// PromptListArgs holds parameters for prompts/list (cursor for pagination).
type PromptListArgs struct {
	Cursor string `yaml:"cursor,omitempty"`
}

// PromptGetArgs holds parameters for prompts/get.
type PromptGetArgs struct {
	Name      string            `yaml:"name"`                // prompt name (required)
	Arguments map[string]string `yaml:"arguments,omitempty"` // prompt arguments
}

// CompletionAssertBlock tests MCP completion/complete.
type CompletionAssertBlock struct {
	Ref      CompletionRef `yaml:"ref"`
	Argument CompletionArg `yaml:"argument"`
	Expect   Expect        `yaml:"expect"`
}

// CompletionRef identifies the prompt or resource to complete against.
type CompletionRef struct {
	Type string `yaml:"type"` // "ref/prompt" or "ref/resource"
	Name string `yaml:"name"` // prompt name or resource URI
}

// CompletionArg specifies the argument to complete.
type CompletionArg struct {
	Name  string `yaml:"name"`  // argument name
	Value string `yaml:"value"` // partial value for completion
}

// TraceEntry is a single tool call in a recorded sequence.
type TraceEntry struct {
	Tool string         `yaml:"tool" json:"tool"`
	Args map[string]any `yaml:"args,omitempty" json:"args,omitempty"`
}

// TrajectoryAssertion checks a property of a tool call sequence.
// Type is one of: "order", "presence", "absence", "args_contain".
type TrajectoryAssertion struct {
	Type  string         `yaml:"type"`
	Tools []string       `yaml:"tools,omitempty"`  // for order, presence, absence
	Tool  string         `yaml:"tool,omitempty"`   // for args_contain
	Args  map[string]any `yaml:"args,omitempty"`   // for args_contain: partial match
}

// ServerConfig specifies how to connect to the MCP server under test.
// For stdio transport (default), Command/Args/Env launch the server as a subprocess.
// For HTTP or SSE transport, URL specifies the server endpoint.
type ServerConfig struct {
	Command            string             `yaml:"command"`
	Args               []string           `yaml:"args"`
	Env                map[string]string  `yaml:"env"`
	Transport          string             `yaml:"transport,omitempty"` // "stdio" (default), "sse", "http"
	URL                string             `yaml:"url,omitempty"`       // Required for sse/http transport
	ClientCapabilities ClientCapabilities `yaml:"client_capabilities,omitempty"`
}

// ClientCapabilities declares what the mcp-assert client supports.
// When set, mcp-assert responds to server-initiated requests.
type ClientCapabilities struct {
	Roots     []string          `yaml:"roots,omitempty"`     // File/dir paths to return for roots/list requests
	Sampling  *SamplingConfig   `yaml:"sampling,omitempty"`  // Mock LLM response for sampling requests
	Elicitation map[string]any  `yaml:"elicitation,omitempty"` // Preset values for elicitation requests
}

// SamplingConfig provides a mock LLM response for servers that use MCP sampling.
type SamplingConfig struct {
	Text      string `yaml:"text"`                 // Response text content
	Model     string `yaml:"model,omitempty"`      // Model name to report (default: "mock")
	StopReason string `yaml:"stop_reason,omitempty"` // Stop reason (default: "end_turn")
}

// ToolCall is a single MCP tool invocation.
type ToolCall struct {
	Tool    string            `yaml:"tool"`
	Args    map[string]any    `yaml:"args"`
	Capture map[string]string `yaml:"capture,omitempty"` // variable_name -> jsonpath (e.g. "session_id": "$.session_id")
}

// AssertBlock defines the tool to call and the expected results.
type AssertBlock struct {
	Tool            string         `yaml:"tool"`
	Args            map[string]any `yaml:"args"`
	Expect          Expect         `yaml:"expect"`
	CaptureProgress bool           `yaml:"capture_progress,omitempty"` // collect notifications/progress during tool execution
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
	MinProgress   *int              `yaml:"min_progress,omitempty"` // minimum number of notifications/progress received
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
