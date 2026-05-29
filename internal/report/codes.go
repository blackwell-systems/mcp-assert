// codes.go defines the unified error taxonomy for mcp-assert.
// All commands (lint, audit, run, fuzz, etc.) use these codes.
package report

import "fmt"

// ErrorCode represents a structured error identifier.
type ErrorCode string

// ErrorSeverity indicates the severity level of an issue.
type ErrorSeverity string

const (
	SeverityError   ErrorSeverity = "error"
	SeverityWarning ErrorSeverity = "warning"
	SeverityInfo    ErrorSeverity = "info"
)

// ErrorDefinition holds metadata about a specific error type.
type ErrorDefinition struct {
	Code        ErrorCode
	Name        string
	Description string
	Severity    ErrorSeverity
	Category    string
	Remediation string
}

// Error codes - organized by category with room for expansion
const (
	// E000: Success/Info
	E000 ErrorCode = "E000" // Success

	// E1xx: Schema & Definition Issues (Static Analysis)
	E101 ErrorCode = "E101" // Missing Tool Description
	E102 ErrorCode = "E102" // Missing Parameter Type
	E103 ErrorCode = "E103" // Required Parameter No Description
	E104 ErrorCode = "E104" // Parameter Not In Description (reserved)
	E105 ErrorCode = "E105" // Free Text Propagation
	E106 ErrorCode = "E106" // Output Not Guaranteed (reserved)
	E107 ErrorCode = "E107" // Circular Dependency
	E110 ErrorCode = "E110" // Tool Ambiguity (reserved)
	E112 ErrorCode = "E112" // Sensitive Parameter Exposed
	E113 ErrorCode = "E113" // Duplicate Tool Names

	// E2xx: Runtime Errors
	E201 ErrorCode = "E201" // Server Panic
	E202 ErrorCode = "E202" // Timeout
	E203 ErrorCode = "E203" // Connection Failed

	// E3xx: Output Issues
	E301 ErrorCode = "E301" // Output Explosion
	E302 ErrorCode = "E302" // Malformed JSON (reserved)

	// E4xx: Assertion Failures
	E401 ErrorCode = "E401" // Assertion Failed (reserved)
	E402 ErrorCode = "E402" // Snapshot Mismatch (reserved)

	// E5xx: Side Effects & Behavioral (reserved for future)
	E500 ErrorCode = "E500" // Unexpected Side Effect (reserved)
	E501 ErrorCode = "E501" // Non-Idempotent Operation (reserved)

	// W1xx: Schema & Definition Warnings
	W101 ErrorCode = "W101" // Generic/Vague Description
	W102 ErrorCode = "W102" // Optional Parameter No Description
	W103 ErrorCode = "W103" // String Without Constraints
	W104 ErrorCode = "W104" // Generic Parameter Names
	W105 ErrorCode = "W105" // Tool Similarity
	W106 ErrorCode = "W106" // Schema Bloat
	W107 ErrorCode = "W107" // Non-Deterministic Output

	// W3xx: Output Warnings
	W301 ErrorCode = "W301" // Large Output Warning (reserved)
)

// ErrorRegistry maps error codes to their definitions.
var ErrorRegistry = map[ErrorCode]ErrorDefinition{
	// Success
	E000: {
		Code:        E000,
		Name:        "Success",
		Severity:    SeverityInfo,
		Category:    "execution",
		Description: "Tool executed successfully",
	},

	// E1xx: Schema & Definition Issues
	E101: {
		Code:        E101,
		Name:        "Missing Tool Description",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Tool has no description field or description is empty",
		Remediation: "Add a clear description explaining what this tool does, when to use it, and what it returns",
	},
	E102: {
		Code:        E102,
		Name:        "Missing Parameter Type",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Parameter has no type defined in schema",
		Remediation: "Add 'type' field to parameter schema (string, number, boolean, object, array)",
	},
	E103: {
		Code:        E103,
		Name:        "Required Parameter No Description",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Required parameter has no description",
		Remediation: "Add description explaining what value the parameter expects",
	},

	E105: {
		Code:        E105,
		Name:        "Free Text Propagation",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Unconstrained string output from one tool flows into another tool's input (LLM hallucination risk)",
		Remediation: "Add enum, pattern, or format constraints to the output field, or validate input on the receiving tool",
	},
	E107: {
		Code:        E107,
		Name:        "Circular Dependency",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Tool dependency cycle detected (tool A's output feeds B, B's output feeds A)",
		Remediation: "Break the cycle by restructuring tool responsibilities or removing the circular reference",
	},
	E112: {
		Code:        E112,
		Name:        "Sensitive Parameter Exposed",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Parameter name suggests sensitive data (password, secret, token, api_key) without being marked writeOnly or redacted",
		Remediation: "Mark sensitive parameters with writeOnly:true or use a dedicated secrets mechanism",
	},
	E113: {
		Code:        E113,
		Name:        "Duplicate Tool Names",
		Severity:    SeverityError,
		Category:    "schema",
		Description: "Multiple tools share the same name (exact collision)",
		Remediation: "Rename tools to be unique",
	},

	// E2xx: Runtime Errors
	E201: {
		Code:        E201,
		Name:        "Server Panic",
		Severity:    SeverityError,
		Category:    "runtime",
		Description: "Server crashed or returned internal error (-32603)",
		Remediation: "Fix the server crash by handling edge cases and adding error recovery",
	},
	E202: {
		Code:        E202,
		Name:        "Timeout",
		Severity:    SeverityError,
		Category:    "runtime",
		Description: "Tool did not respond within the timeout period",
		Remediation: "Optimize tool performance or increase timeout threshold",
	},
	E203: {
		Code:        E203,
		Name:        "Connection Failed",
		Severity:    SeverityError,
		Category:    "runtime",
		Description: "Failed to connect to MCP server",
		Remediation: "Verify server is running, check transport configuration, and ensure network connectivity",
	},

	// E3xx: Output Issues
	E301: {
		Code:        E301,
		Name:        "Output Explosion",
		Severity:    SeverityError,
		Category:    "output",
		Description: "Tool output exceeds recommended size limit (context window exhaustion risk)",
		Remediation: "Add pagination, limit array sizes, or add max_items constraint to schema",
	},

	// W1xx: Schema Warnings
	W101: {
		Code:        W101,
		Name:        "Generic/Vague Description",
		Severity:    SeverityWarning,
		Category:    "schema",
		Description: "Tool description is too vague or generic for effective tool selection",
		Remediation: "Be specific about inputs, outputs, use cases, and when to use this tool",
	},
	W102: {
		Code:        W102,
		Name:        "Optional Parameter No Description",
		Severity:    SeverityWarning,
		Category:    "schema",
		Description: "Optional parameter has no description",
		Remediation: "Add description explaining what the parameter does and valid values",
	},
	W103: {
		Code:        W103,
		Name:        "String Without Constraints",
		Severity:    SeverityWarning,
		Category:    "schema",
		Description: "String parameter has no enum, pattern, example, or default",
		Remediation: "Add format, enum, pattern, or examples to constrain valid values",
	},
	W104: {
		Code:        W104,
		Name:        "Generic Parameter Names",
		Severity:    SeverityWarning,
		Category:    "schema",
		Description: "Parameter has generic name (data, value, input, etc.) with no description",
		Remediation: "Use descriptive parameter names or add detailed description",
	},
	W105: {
		Code:        W105,
		Name:        "Tool Similarity",
		Severity:    SeverityWarning,
		Category:    "schema",
		Description: "Two or more tools have similar descriptions (LLM may confuse them)",
		Remediation: "Make tool names and descriptions more distinct, or merge similar tools",
	},
	W106: {
		Code:        W106,
		Name:        "Schema Bloat",
		Severity:    SeverityWarning,
		Category:    "schema",
		Description: "Total tools/list response exceeds 8K tokens (context window concern)",
		Remediation: "Reduce number of tools or simplify schemas",
	},
	W107: {
		Code:        W107,
		Name:        "Non-Deterministic Output",
		Severity:    SeverityWarning,
		Category:    "behavior",
		Description: "Tool produces different outputs for identical inputs across multiple calls",
		Remediation: "Ensure tool is deterministic, or document non-deterministic behavior. Non-deterministic tools are unreliable for agent workflows.",
	},
}

// Get retrieves error definition by code.
func (c ErrorCode) Get() ErrorDefinition {
	if def, ok := ErrorRegistry[c]; ok {
		return def
	}
	return ErrorDefinition{
		Code:     c,
		Name:     string(c),
		Severity: SeverityError,
		Category: "unknown",
	}
}

// String returns the error code as a string.
func (c ErrorCode) String() string {
	return string(c)
}

// IsError returns true if the code represents an error (not warning or info).
func (c ErrorCode) IsError() bool {
	def := c.Get()
	return def.Severity == SeverityError
}

// IsWarning returns true if the code represents a warning.
func (c ErrorCode) IsWarning() bool {
	def := c.Get()
	return def.Severity == SeverityWarning
}

// FormatError formats an error with code, name, and message.
func FormatError(code ErrorCode, message string) string {
	def := code.Get()
	return fmt.Sprintf("[%s] %s: %s", code, def.Name, message)
}

// FormatErrorWithTool formats an error including the tool name.
func FormatErrorWithTool(code ErrorCode, tool, message string) string {
	def := code.Get()
	return fmt.Sprintf("[%s] %s (%s): %s", code, def.Name, tool, message)
}

// CategoryErrors returns all error codes for a given category.
func CategoryErrors(category string) []ErrorCode {
	var codes []ErrorCode
	for code, def := range ErrorRegistry {
		if def.Category == category {
			codes = append(codes, code)
		}
	}
	return codes
}

// AllCategories returns all unique categories.
func AllCategories() []string {
	seen := make(map[string]bool)
	var categories []string
	for _, def := range ErrorRegistry {
		if !seen[def.Category] {
			seen[def.Category] = true
			categories = append(categories, def.Category)
		}
	}
	return categories
}
