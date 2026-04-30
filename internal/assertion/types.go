// Package assertion defines the core types and validation logic for mcp-assert.
//
// The package is organized into focused files:
//   - types.go       — Suite, Assertion, Expect, Result, and all block types
//   - loader.go      — YAML file loading with subdirectory recursion
//   - checker.go     — 18 assertion type implementations (contains, json_path, etc.)
//   - trajectory.go  — 4 trajectory assertion types (order, presence, absence, args_contain)
//   - logging*.go    — Logging assertion block and checker
//   - sampling*.go   — Sampling assertion block type
package assertion

import (
	"encoding/json"
	"fmt"
	"time"
)

// Suite is a collection of assertion files loaded from a directory.
// The loader recursively discovers all .yaml files under Dir and parses
// each one into an Assertion. One Suite maps to one --suite CLI argument.
type Suite struct {
	Assertions []Assertion // parsed assertions from all YAML files in the directory tree
	Dir        string      // root directory that was scanned
}

// Assertion is the top-level type deserialized from a single YAML file.
// Each file defines exactly one assertion. The assertion type is determined
// by which block is set: Assert (tools), AssertResources, AssertPrompts,
// AssertCompletion, AssertSampling, AssertLogging, or Trajectory.
// Only one block should be set per file; the runner uses the first non-nil
// block to determine execution mode.
type Assertion struct {
	Name   string       `yaml:"name"`   // human-readable name, shown in test output and JUnit reports
	Server ServerConfig `yaml:"server"` // how to connect to the MCP server under test
	Setup  []ToolCall   `yaml:"setup"`  // tool calls to execute before the assertion (e.g., create test data)

	// Assertion blocks: set exactly one per YAML file.
	Assert           AssertBlock            `yaml:"assert"`                      // tool assertion (call a tool, check the result)
	AssertResources  *ResourceAssertBlock   `yaml:"assert_resources,omitempty"`  // resource assertion (list or read resources)
	AssertPrompts    *PromptAssertBlock     `yaml:"assert_prompts,omitempty"`    // prompt assertion (list or get prompts)
	AssertCompletion *CompletionAssertBlock `yaml:"assert_completion,omitempty"` // completion assertion (test autocomplete)
	AssertSampling   *SamplingAssertBlock   `yaml:"assert_sampling,omitempty"`   // sampling assertion (test server-initiated LLM requests)
	AssertLogging    *LoggingAssertBlock    `yaml:"assert_logging,omitempty"`    // logging assertion (test log message notifications)

	// Execution control.
	Timeout       string `yaml:"timeout"`                    // per-assertion timeout (e.g., "30s", "1m"); overrides CLI --timeout
	Skip          bool   `yaml:"skip,omitempty"`             // unconditionally skip this assertion
	SkipUnlessEnv string `yaml:"skip_unless_env,omitempty"`  // skip if this environment variable is not set (for auth-gated tests)

	// Trajectory mode: instead of calling a single tool, replay a recorded
	// sequence and assert properties of the call sequence.
	Trace      []TraceEntry          `yaml:"trace,omitempty"`      // inline tool call sequence to replay
	AuditLog   string                `yaml:"audit_log,omitempty"`  // path to an agent-lsp JSONL audit log to replay instead of inline trace
	Trajectory []TrajectoryAssertion `yaml:"trajectory,omitempty"` // assertions on the replayed sequence (order, presence, absence, args_contain)
}

// ResourceAssertBlock tests MCP resource operations (resources/list or resources/read).
// Set exactly one of List or Read to choose the operation. Subscribe/Unsubscribe
// test the resource subscription lifecycle; ExpectNotification verifies that a
// resource update notification is (or isn't) received within the assertion timeout.
type ResourceAssertBlock struct {
	List               *ResourceListArgs `yaml:"list,omitempty"`                // call resources/list; nil means skip listing
	Read               string            `yaml:"read,omitempty"`               // URI to pass to resources/read; empty means skip reading
	Expect             Expect            `yaml:"expect"`                       // assertions applied to the list or read result
	Subscribe          string            `yaml:"subscribe,omitempty"`           // URI to subscribe to before the assertion
	Unsubscribe        string            `yaml:"unsubscribe,omitempty"`         // URI to unsubscribe from after the assertion
	ExpectNotification *bool             `yaml:"expect_notification,omitempty"` // if true, assert a resource update notification arrives; if false, assert none arrives
}

// ResourceListArgs holds parameters for the resources/list call.
type ResourceListArgs struct {
	Cursor string `yaml:"cursor,omitempty"` // pagination cursor; empty for the first page
}

// PromptAssertBlock tests MCP prompt operations (prompts/list or prompts/get).
// Set exactly one of List or Get. The Expect block is applied to whichever
// operation's result is returned.
type PromptAssertBlock struct {
	List   *PromptListArgs `yaml:"list,omitempty"` // call prompts/list; nil means skip listing
	Get    *PromptGetArgs  `yaml:"get,omitempty"`  // call prompts/get; nil means skip getting
	Expect Expect          `yaml:"expect"`         // assertions applied to the list or get result
}

// PromptListArgs holds parameters for the prompts/list call.
type PromptListArgs struct {
	Cursor string `yaml:"cursor,omitempty"` // pagination cursor; empty for the first page
}

// PromptGetArgs holds parameters for the prompts/get call.
type PromptGetArgs struct {
	Name      string            `yaml:"name"`                // prompt name (required); must match a name from prompts/list
	Arguments map[string]string `yaml:"arguments,omitempty"` // key-value arguments to interpolate into the prompt template
}

// CompletionAssertBlock tests MCP completion/complete, which provides
// autocomplete suggestions for prompt arguments or resource URIs.
type CompletionAssertBlock struct {
	Ref      CompletionRef `yaml:"ref"`      // identifies what to complete against (a prompt or resource)
	Argument CompletionArg `yaml:"argument"` // the argument being completed with a partial value
	Expect   Expect        `yaml:"expect"`   // assertions on the completion suggestions
}

// CompletionRef identifies the prompt or resource to complete against.
type CompletionRef struct {
	Type string `yaml:"type"` // "ref/prompt" or "ref/resource"
	Name string `yaml:"name"` // prompt name (for ref/prompt) or resource URI (for ref/resource)
}

// CompletionArg specifies which argument is being completed and the partial input.
type CompletionArg struct {
	Name  string `yaml:"name"`  // argument name (e.g., "language" for a prompt with a language parameter)
	Value string `yaml:"value"` // partial value typed so far (e.g., "py" to get suggestions starting with "py")
}

// TraceEntry is a single tool call in a recorded sequence, used for trajectory
// assertions. A list of TraceEntry values represents the agent's tool call
// history, either inline in the YAML or loaded from an agent-lsp audit log.
type TraceEntry struct {
	Tool string         `yaml:"tool" json:"tool"`               // tool name that was called
	Args map[string]any `yaml:"args,omitempty" json:"args,omitempty"` // arguments passed to the tool
}

// TrajectoryAssertion checks a property of a tool call sequence.
// Four assertion types are supported:
//   - "order": the listed tools must appear in this order (not necessarily contiguous)
//   - "presence": every listed tool must appear at least once
//   - "absence": none of the listed tools may appear
//   - "args_contain": the named tool must have been called with args matching the partial map
type TrajectoryAssertion struct {
	Type  string         `yaml:"type"`                 // one of: "order", "presence", "absence", "args_contain"
	Tools []string       `yaml:"tools,omitempty"`      // tool names for order, presence, and absence checks
	Tool  string         `yaml:"tool,omitempty"`        // single tool name for args_contain
	Args  map[string]any `yaml:"args,omitempty"`        // partial argument map for args_contain (all keys must match)
}

// ServerConfig specifies how to connect to the MCP server under test.
// Three transport modes are supported:
//   - stdio (default): launch the server as a subprocess via Command/Args
//   - sse: connect to an existing server at URL using Server-Sent Events
//   - http: connect to an existing server at URL using Streamable HTTP
//
// For stdio, the server process is started fresh for each assertion and
// killed when the assertion completes. Docker isolation wraps the subprocess
// in a fresh container per assertion.
type ServerConfig struct {
	Command            string             `yaml:"command"`                      // executable to run (e.g., "npx", "uvx", "node", "go")
	Args               []string           `yaml:"args"`                         // arguments passed to the command
	Env                map[string]string  `yaml:"env"`                          // environment variables set on the subprocess (supports ${VAR} expansion)
	Transport          string             `yaml:"transport,omitempty"`          // "stdio" (default), "sse", or "http"
	URL                string             `yaml:"url,omitempty"`                // server endpoint URL; required for sse and http transports
	Headers            map[string]string  `yaml:"headers,omitempty"`            // custom HTTP headers for sse/http transports (supports ${VAR} expansion)
	Docker             string             `yaml:"docker,omitempty"`             // Docker image name; when set, each assertion runs in a fresh container
	ClientCapabilities ClientCapabilities `yaml:"client_capabilities,omitempty"` // capabilities the mcp-assert client advertises to the server
}

// ClientCapabilities declares what the mcp-assert client supports.
// Some MCP servers issue requests back to the client (roots/list, sampling,
// elicitation). These fields configure mock responses so assertions can
// exercise servers that depend on client capabilities.
type ClientCapabilities struct {
	Roots       []string       `yaml:"roots,omitempty"`       // file/dir paths returned for roots/list requests
	Sampling    *SamplingConfig `yaml:"sampling,omitempty"`   // mock LLM response for createMessage sampling requests
	Elicitation map[string]any `yaml:"elicitation,omitempty"` // preset key-value responses for elicitation requests
}

// SamplingConfig provides a mock LLM response for servers that use MCP sampling
// (server-initiated LLM requests via createMessage). The server sends a prompt;
// mcp-assert replies with this canned response instead of calling a real LLM.
type SamplingConfig struct {
	Text       string `yaml:"text"`                  // response text content returned to the server
	Model      string `yaml:"model,omitempty"`       // model name to report (default: "mock")
	StopReason string `yaml:"stop_reason,omitempty"` // stop reason (default: "end_turn")
}

// ToolCall is a single MCP tool invocation, used in the Setup sequence.
// Setup calls execute before the main assertion and can capture values from
// responses for use in subsequent calls (e.g., capture a session ID from a
// create call and use it in the assert call).
type ToolCall struct {
	Tool    string            `yaml:"tool"`                  // tool name to call
	Args    map[string]any    `yaml:"args"`                  // arguments to pass
	Capture map[string]string `yaml:"capture,omitempty"`     // variable_name -> JSONPath expression (e.g., "session_id": "$.id")
}

// AssertBlock defines the primary tool assertion: call a single tool with
// known arguments and check the result against the Expect block. This is the
// most common assertion type, used for testing individual tool correctness.
type AssertBlock struct {
	Tool            string         `yaml:"tool"`                          // tool name to call
	Args            map[string]any `yaml:"args"`                          // arguments to pass to the tool
	Expect          Expect         `yaml:"expect"`                        // assertions on the tool result
	CaptureProgress bool           `yaml:"capture_progress,omitempty"`    // if true, collect progress notifications during execution for MinProgress checks
}

// Expect defines deterministic assertions on the tool result.
// Each field is an independent check; all specified checks must pass for the
// assertion to succeed. Unset fields (nil/empty) are skipped. This struct is
// shared across tool, resource, prompt, and completion assertion blocks.
type Expect struct {
	// Text content checks: applied against the concatenated text of all
	// content blocks in the tool result.
	Contains    []string `yaml:"contains"`              // every string must appear in the result text
	ContainsAny []string `yaml:"contains_any,omitempty"` // at least one string must appear
	NotContains []string `yaml:"not_contains"`           // none of these strings may appear
	Equals      *string  `yaml:"equals"`                 // result text must match exactly (pointer so unset is distinguishable from empty string)

	// Structured content checks: applied when the result text is valid JSON.
	JSONPath map[string]any `yaml:"json_path"` // JSONPath expression -> expected value (e.g. "$.name": "alice")

	// Result shape checks: applied to the list of content blocks.
	MinResults *int  `yaml:"min_results"` // minimum number of content blocks returned
	MaxResults *int  `yaml:"max_results"` // maximum number of content blocks returned
	NotEmpty   *bool `yaml:"not_empty"`   // result must contain at least one non-empty content block

	// MCP error semantics: checks on the isError field of the tool result.
	// NotError asserts the tool did NOT return isError:true (healthy response).
	// IsError asserts the tool DID return isError:true (graceful error handling).
	NotError *bool `yaml:"not_error"` // true -> assert isError is absent or false
	IsError  *bool `yaml:"is_error"`  // true -> assert isError is true

	// Pattern matching: applied against the concatenated result text.
	MatchesRegex []string `yaml:"matches_regex"` // each regex must match somewhere in the result text

	// File system checks: verify side effects by reading files after the tool call.
	// Keys are file paths (relative to fixture dir), values are expected substrings.
	FileContains    map[string]string `yaml:"file_contains"`               // file must contain the substring
	FileNotContains map[string]string `yaml:"file_not_contains,omitempty"` // file must NOT contain the substring
	FileUnchanged   []string          `yaml:"file_unchanged"`              // file content must be identical to before the tool call
	FileNotExists   []string          `yaml:"file_not_exists,omitempty"`   // file must not exist after the tool call

	// Speculative execution checks: for agent-lsp simulate_edit_atomic results.
	NetDelta *int `yaml:"net_delta"` // expected net diagnostic delta (0 = safe to apply)

	// Ordering checks: applied against the concatenated result text.
	InOrder []string `yaml:"in_order"` // strings must appear in this order within the result

	// Progress/notification checks: applied to captured notifications during tool execution.
	MinProgress *int `yaml:"min_progress,omitempty"` // minimum number of progress notifications received
}

// Result is the outcome of running a single assertion. It is serialized to
// JSON for --json output, JUnit XML reports, and the pytest/vitest plugin bridge.
// The runner produces one Result per assertion (or per trial when --trials > 1).
type Result struct {
	Name     string        `json:"name"`               // assertion name from the YAML file
	Status   Status        `json:"status"`             // PASS, FAIL, or SKIP
	Detail   string        `json:"detail,omitempty"`   // failure message or skip reason; empty on PASS
	Duration DurationMS    `json:"duration_ms"`        // wall time for this assertion in milliseconds
	Language string        `json:"language,omitempty"` // language server identifier (populated in matrix mode)
	Trial    int           `json:"trial,omitempty"`    // 1-indexed trial number (populated when --trials > 1)
}

// DurationMS wraps time.Duration and serializes to JSON as an integer
// number of milliseconds, matching the "duration_ms" field name.
type DurationMS time.Duration

func (d DurationMS) MarshalJSON() ([]byte, error) {
	ms := time.Duration(d).Milliseconds()
	return []byte(fmt.Sprintf("%d", ms)), nil
}

// Seconds returns the duration as a floating-point number of seconds.
func (d DurationMS) Seconds() float64 {
	return time.Duration(d).Seconds()
}

// Milliseconds returns the duration as an integer number of milliseconds.
func (d DurationMS) Milliseconds() int64 {
	return time.Duration(d).Milliseconds()
}

func (d *DurationMS) UnmarshalJSON(data []byte) error {
	var ms int64
	if err := json.Unmarshal(data, &ms); err != nil {
		return err
	}
	*d = DurationMS(time.Duration(ms) * time.Millisecond)
	return nil
}

// Status is the outcome of an assertion: PASS (all checks succeeded),
// FAIL (at least one check failed or the server crashed), or SKIP
// (assertion was skipped via skip: true, skip_unless_env, or other condition).
type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusSkip Status = "SKIP"
)
