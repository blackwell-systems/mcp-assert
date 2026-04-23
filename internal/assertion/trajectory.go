package assertion

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// CheckTrajectory verifies trajectory assertions against a tool call trace.
// Returns nil if all checks pass, or an error describing the first failure.
func CheckTrajectory(trajectory []TrajectoryAssertion, trace []TraceEntry) error {
	for _, ta := range trajectory {
		switch ta.Type {
		case "order":
			if err := checkOrder(ta.Tools, trace); err != nil {
				return err
			}
		case "presence":
			if err := checkPresence(ta.Tools, trace); err != nil {
				return err
			}
		case "absence":
			if err := checkAbsence(ta.Tools, trace); err != nil {
				return err
			}
		case "args_contain":
			if err := checkArgsContain(ta.Tool, ta.Args, trace); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown trajectory assertion type: %q", ta.Type)
		}
	}
	return nil
}

// checkOrder verifies that the listed tools appear in this order in the trace.
// Tools don't need to be adjacent — just in sequence.
func checkOrder(tools []string, trace []TraceEntry) error {
	if len(tools) == 0 {
		return nil
	}

	pos := 0
	for _, entry := range trace {
		if entry.Tool == tools[pos] {
			pos++
			if pos == len(tools) {
				return nil
			}
		}
	}

	// Find where the sequence broke.
	if pos == 0 {
		return fmt.Errorf("order: %q not found in trace", tools[0])
	}
	return fmt.Errorf("order: found %q but %q did not follow (got: %s)",
		tools[pos-1], tools[pos], traceToolNames(trace))
}

// checkPresence verifies all listed tools appear at least once in the trace.
func checkPresence(tools []string, trace []TraceEntry) error {
	seen := make(map[string]bool)
	for _, entry := range trace {
		seen[entry.Tool] = true
	}

	for _, tool := range tools {
		if !seen[tool] {
			return fmt.Errorf("presence: %q not found in trace (got: %s)",
				tool, traceToolNames(trace))
		}
	}
	return nil
}

// checkAbsence verifies none of the listed tools appear in the trace.
func checkAbsence(tools []string, trace []TraceEntry) error {
	forbidden := make(map[string]bool)
	for _, t := range tools {
		forbidden[t] = true
	}

	for _, entry := range trace {
		if forbidden[entry.Tool] {
			return fmt.Errorf("absence: %q appeared in trace but should not have",
				entry.Tool)
		}
	}
	return nil
}

// checkArgsContain verifies a tool was called with args that contain the expected values.
func checkArgsContain(tool string, expectedArgs map[string]any, trace []TraceEntry) error {
	for _, entry := range trace {
		if entry.Tool != tool {
			continue
		}
		// Check each expected arg key/value.
		for k, expected := range expectedArgs {
			actual, ok := entry.Args[k]
			if !ok {
				return fmt.Errorf("args_contain: %q called without arg %q", tool, k)
			}
			expectedStr := fmt.Sprintf("%v", expected)
			actualStr := fmt.Sprintf("%v", actual)
			if actualStr != expectedStr {
				return fmt.Errorf("args_contain: %q arg %q: want %v, got %v",
					tool, k, expected, actual)
			}
		}
		return nil // First matching call passed.
	}

	return fmt.Errorf("args_contain: %q not found in trace (got: %s)",
		tool, traceToolNames(trace))
}

// traceToolNames returns a comma-separated list of tool names for error messages.
func traceToolNames(trace []TraceEntry) string {
	names := make([]string, len(trace))
	for i, e := range trace {
		names[i] = e.Tool
	}
	return "[" + strings.Join(names, ", ") + "]"
}

// LoadAuditLog parses an agent-lsp JSONL audit log into a trace.
// Each line is a JSON object with a "tool" field (and optionally "args").
func LoadAuditLog(path string) ([]TraceEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading audit log %s: %w", path, err)
	}

	var entries []TraceEntry
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("audit log line %d: invalid JSON: %w", i+1, err)
		}

		// Support agent-lsp audit format: {"action":"tool_call","tool":"...","args":{...}}
		// Also support simple format: {"tool":"...","args":{...}}
		tool, _ := raw["tool"].(string)
		if tool == "" {
			continue // Skip non-tool-call entries.
		}

		entry := TraceEntry{Tool: tool}
		if args, ok := raw["args"].(map[string]any); ok {
			entry.Args = args
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
