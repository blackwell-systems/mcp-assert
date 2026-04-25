package assertion

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// checkFunc evaluates a single expectation type against the tool result.
// Returns nil if the check passes (or does not apply), or an error on failure.
type checkFunc func(expect Expect, response string, isError bool) error

// checkEntry pairs a name with its check function for ordered evaluation.
type checkEntry struct {
	name  string
	check checkFunc
}

// checkRegistry defines the ordered list of expectation checks.
// Evaluation order is preserved from the original if/else chain;
// the first failing check short-circuits and returns its error.
var checkRegistry = []checkEntry{
	{"not_error", checkNotError},
	{"is_error", checkIsError},
	{"not_empty", checkNotEmpty},
	{"equals", checkEquals},
	{"contains", checkContains},
	{"contains_any", checkContainsAny},
	{"not_contains", checkNotContains},
	{"matches_regex", checkMatchesRegex},
	{"json_path", checkJSONPath},
	{"min_max_results", checkMinMaxResults},
	{"net_delta", checkNetDelta},
	{"file_contains", checkFileContains},
	{"file_not_contains", checkFileNotContains},
	{"file_not_exists", checkFileNotExists},
	{"in_order", checkInOrder},
}

// Check evaluates all expectations against the tool result text.
// Returns nil if all assertions pass, or an error describing the first failure.
func Check(expect Expect, resultText string, isError bool) error {
	for _, entry := range checkRegistry {
		if err := entry.check(expect, resultText, isError); err != nil {
			return err
		}
	}
	return nil
}

func checkNotError(expect Expect, _ string, isError bool) error {
	if expect.NotError != nil && *expect.NotError && isError {
		return fmt.Errorf("expected no error but tool returned isError=true")
	}
	return nil
}

func checkIsError(expect Expect, _ string, isError bool) error {
	if expect.IsError != nil && *expect.IsError && !isError {
		return fmt.Errorf("expected error but tool returned isError=false")
	}
	return nil
}

func checkNotEmpty(expect Expect, response string, _ bool) error {
	if expect.NotEmpty != nil && *expect.NotEmpty {
		trimmed := strings.TrimSpace(response)
		if trimmed == "" || trimmed == "null" || trimmed == "[]" || trimmed == "{}" {
			return fmt.Errorf("expected non-empty result, got: %q", trimmed)
		}
	}
	return nil
}

func checkEquals(expect Expect, response string, _ bool) error {
	if expect.Equals != nil {
		if strings.TrimSpace(response) != strings.TrimSpace(*expect.Equals) {
			return fmt.Errorf("expected exact match:\n  want: %s\n  got:  %s", *expect.Equals, response)
		}
	}
	return nil
}

func checkContains(expect Expect, response string, _ bool) error {
	for _, s := range expect.Contains {
		if !strings.Contains(response, s) {
			return fmt.Errorf("expected result to contain %q, got: %.200s", s, response)
		}
	}
	return nil
}

func checkContainsAny(expect Expect, response string, _ bool) error {
	if len(expect.ContainsAny) > 0 {
		found := false
		for _, s := range expect.ContainsAny {
			if strings.Contains(response, s) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected result to contain at least one of %q, got: %.200s", expect.ContainsAny, response)
		}
	}
	return nil
}

func checkNotContains(expect Expect, response string, _ bool) error {
	for _, s := range expect.NotContains {
		if strings.Contains(response, s) {
			return fmt.Errorf("expected result NOT to contain %q", s)
		}
	}
	return nil
}

func checkMatchesRegex(expect Expect, response string, _ bool) error {
	for _, pattern := range expect.MatchesRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("matches_regex: invalid pattern %q: %w", pattern, err)
		}
		if !re.MatchString(response) {
			return fmt.Errorf("expected result to match regex %q, got: %.200s", pattern, response)
		}
	}
	return nil
}

func checkJSONPath(expect Expect, response string, _ bool) error {
	if len(expect.JSONPath) > 0 {
		var parsed any
		if err := json.Unmarshal([]byte(response), &parsed); err != nil {
			return fmt.Errorf("json_path: result is not valid JSON: %w", err)
		}
		for path, expected := range expect.JSONPath {
			actual, err := jsonPathLookup(parsed, path)
			if err != nil {
				return fmt.Errorf("json_path %q: %w", path, err)
			}
			expectedStr := fmt.Sprintf("%v", expected)
			actualStr := fmt.Sprintf("%v", actual)
			if actualStr != expectedStr {
				return fmt.Errorf("json_path %q: want %v, got %v", path, expected, actual)
			}
		}
	}
	return nil
}

func checkMinMaxResults(expect Expect, response string, _ bool) error {
	if expect.MinResults != nil || expect.MaxResults != nil {
		var arr []any
		if err := json.Unmarshal([]byte(response), &arr); err != nil {
			// Try object with common array fields.
			var obj map[string]any
			if err2 := json.Unmarshal([]byte(response), &obj); err2 == nil {
				for _, key := range []string{"locations", "items", "results", "references", "symbols"} {
					if v, ok := obj[key].([]any); ok {
						arr = v
						break
					}
				}
			}
			if arr == nil {
				return fmt.Errorf("min/max_results: result is not an array or object with array field: %s", err)
			}
		}
		if expect.MinResults != nil && len(arr) < *expect.MinResults {
			return fmt.Errorf("expected at least %d results, got %d", *expect.MinResults, len(arr))
		}
		if expect.MaxResults != nil && len(arr) > *expect.MaxResults {
			return fmt.Errorf("expected at most %d results, got %d", *expect.MaxResults, len(arr))
		}
	}
	return nil
}

func checkNetDelta(expect Expect, response string, _ bool) error {
	if expect.NetDelta != nil {
		var obj map[string]any
		if err := json.Unmarshal([]byte(response), &obj); err != nil {
			return fmt.Errorf("net_delta: result is not valid JSON: %w", err)
		}
		nd, ok := obj["net_delta"].(float64)
		if !ok {
			return fmt.Errorf("net_delta: field not found or not a number in result")
		}
		if int(nd) != *expect.NetDelta {
			return fmt.Errorf("expected net_delta=%d, got %d", *expect.NetDelta, int(nd))
		}
	}
	return nil
}

func checkFileContains(expect Expect, _ string, _ bool) error {
	for path, expected := range expect.FileContains {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("file_contains: cannot read %s: %w", path, err)
		}
		if !strings.Contains(string(content), expected) {
			return fmt.Errorf("file_contains: %s does not contain %q", path, expected)
		}
	}
	return nil
}

func checkFileNotContains(expect Expect, _ string, _ bool) error {
	for path, unexpected := range expect.FileNotContains {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("file_not_contains: cannot read %s: %w", path, err)
		}
		if strings.Contains(string(content), unexpected) {
			return fmt.Errorf("file_not_contains: %s contains %q", path, unexpected)
		}
	}
	return nil
}

func checkFileNotExists(expect Expect, _ string, _ bool) error {
	for _, path := range expect.FileNotExists {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file_not_exists: expected %s not to exist, but it does", path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("file_not_exists: error checking %s: %w", path, err)
		}
	}
	return nil
}

func checkInOrder(expect Expect, response string, _ bool) error {
	if len(expect.InOrder) > 0 {
		searchFrom := 0
		for _, s := range expect.InOrder {
			idx := strings.Index(response[searchFrom:], s)
			if idx < 0 {
				return fmt.Errorf("in_order: %q not found after position %d in result", s, searchFrom)
			}
			searchFrom += idx + len(s)
		}
	}
	return nil
}

// CheckProgress verifies that the number of notifications/progress received meets the
// min_progress expectation. Call this after CallTool when capture_progress is true.
func CheckProgress(expect Expect, count int) error {
	if expect.MinProgress != nil && count < *expect.MinProgress {
		return fmt.Errorf("expected at least %d progress notification(s), got %d", *expect.MinProgress, count)
	}
	return nil
}

// CheckWithSnapshots evaluates all expectations including file_unchanged.
// snapshots maps file paths to their content before the tool was called.
func CheckWithSnapshots(expect Expect, resultText string, isError bool, snapshots map[string]string) error {
	if err := Check(expect, resultText, isError); err != nil {
		return err
	}
	for _, path := range expect.FileUnchanged {
		before, ok := snapshots[path]
		if !ok {
			return fmt.Errorf("file_unchanged: no snapshot for %s", path)
		}
		after, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("file_unchanged: cannot read %s: %w", path, err)
		}
		if string(after) != before {
			return fmt.Errorf("file_unchanged: %s was modified", path)
		}
	}
	return nil
}

// jsonPathLookup does a simple dot-notation lookup on parsed JSON.
// Supports "$.field.subfield" and "$.field[N]".
func jsonPathLookup(data any, path string) (any, error) {
	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")

	current := data
	for _, part := range parts {
		// Check for array index: "field[0]"
		if idx := strings.Index(part, "["); idx >= 0 {
			field := part[:idx]
			indexStr := strings.TrimSuffix(part[idx+1:], "]")

			obj, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected object at %q, got %T", field, current)
			}
			arr, ok := obj[field].([]any)
			if !ok {
				return nil, fmt.Errorf("expected array at %q", field)
			}
			var i int
			if _, err := fmt.Sscanf(indexStr, "%d", &i); err != nil {
				return nil, fmt.Errorf("invalid array index %q", indexStr)
			}
			if i < 0 || i >= len(arr) {
				return nil, fmt.Errorf("index %d out of range (len=%d)", i, len(arr))
			}
			current = arr[i]
			continue
		}

		obj, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected object at %q, got %T", part, current)
		}
		v, ok := obj[part]
		if !ok {
			return nil, fmt.Errorf("field %q not found", part)
		}
		current = v
	}
	return current, nil
}
