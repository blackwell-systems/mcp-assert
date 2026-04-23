package assertion

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Check evaluates all expectations against the tool result text.
// Returns nil if all assertions pass, or an error describing the first failure.
func Check(expect Expect, resultText string, isError bool) error {
	// not_error
	if expect.NotError != nil && *expect.NotError && isError {
		return fmt.Errorf("expected no error but tool returned isError=true")
	}

	// is_error
	if expect.IsError != nil && *expect.IsError && !isError {
		return fmt.Errorf("expected error but tool returned isError=false")
	}

	// not_empty
	if expect.NotEmpty != nil && *expect.NotEmpty {
		trimmed := strings.TrimSpace(resultText)
		if trimmed == "" || trimmed == "null" || trimmed == "[]" || trimmed == "{}" {
			return fmt.Errorf("expected non-empty result, got: %q", trimmed)
		}
	}

	// equals
	if expect.Equals != nil {
		if strings.TrimSpace(resultText) != strings.TrimSpace(*expect.Equals) {
			return fmt.Errorf("expected exact match:\n  want: %s\n  got:  %s", *expect.Equals, resultText)
		}
	}

	// contains
	for _, s := range expect.Contains {
		if !strings.Contains(resultText, s) {
			return fmt.Errorf("expected result to contain %q, got: %.200s", s, resultText)
		}
	}

	// not_contains
	for _, s := range expect.NotContains {
		if strings.Contains(resultText, s) {
			return fmt.Errorf("expected result NOT to contain %q", s)
		}
	}

	// matches_regex
	for _, pattern := range expect.MatchesRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("matches_regex: invalid pattern %q: %w", pattern, err)
		}
		if !re.MatchString(resultText) {
			return fmt.Errorf("expected result to match regex %q, got: %.200s", pattern, resultText)
		}
	}

	// json_path
	if len(expect.JSONPath) > 0 {
		var parsed any
		if err := json.Unmarshal([]byte(resultText), &parsed); err != nil {
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

	// min_results / max_results
	if expect.MinResults != nil || expect.MaxResults != nil {
		var arr []any
		if err := json.Unmarshal([]byte(resultText), &arr); err != nil {
			// Try object with common array fields.
			var obj map[string]any
			if err2 := json.Unmarshal([]byte(resultText), &obj); err2 == nil {
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

	// net_delta
	if expect.NetDelta != nil {
		var obj map[string]any
		if err := json.Unmarshal([]byte(resultText), &obj); err != nil {
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

	// file_contains
	for path, expected := range expect.FileContains {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("file_contains: cannot read %s: %w", path, err)
		}
		if !strings.Contains(string(content), expected) {
			return fmt.Errorf("file_contains: %s does not contain %q", path, expected)
		}
	}

	// file_unchanged — caller passes snapshots via CheckWithSnapshots; standalone Check skips this.

	// in_order — verify substrings appear in the given order within the result.
	if len(expect.InOrder) > 0 {
		searchFrom := 0
		for _, s := range expect.InOrder {
			idx := strings.Index(resultText[searchFrom:], s)
			if idx < 0 {
				return fmt.Errorf("in_order: %q not found after position %d in result", s, searchFrom)
			}
			searchFrom += idx + len(s)
		}
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
