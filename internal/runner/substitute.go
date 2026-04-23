package runner

import (
	"encoding/json"
	"fmt"
	"strings"
)

// substituteAll replaces {{fixture}} and any captured variables in args.
func substituteAll(args map[string]any, fixture string, captured map[string]string) map[string]any {
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = substituteValue(v, fixture, captured)
	}
	return out
}

// substituteFixture replaces {{fixture}} only (backward compat for callers without captures).
func substituteFixture(args map[string]any, fixture string) map[string]any {
	return substituteAll(args, fixture, nil)
}

func substituteValue(v any, fixture string, captured map[string]string) any {
	switch val := v.(type) {
	case string:
		s := val
		if fixture != "" {
			s = strings.ReplaceAll(s, "{{fixture}}", fixture)
		}
		for name, value := range captured {
			s = strings.ReplaceAll(s, "{{"+name+"}}", value)
		}
		return s
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = substituteValue(item, fixture, captured)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = substituteValue(item, fixture, captured)
		}
		return out
	default:
		return v
	}
}

// extractJSONPath extracts a value from JSON text using a simple dot-notation path.
// Reuses the jsonPathLookup logic from the assertion checker.
func extractJSONPath(jsonText, path string) (string, error) {
	path = strings.TrimPrefix(path, "$.")
	if path == "" || path == "$" {
		return jsonText, nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return "", fmt.Errorf("response is not valid JSON: %w", err)
	}

	parts := strings.Split(path, ".")
	current := parsed
	for _, part := range parts {
		// Handle array index: "field[0]"
		if idx := strings.Index(part, "["); idx >= 0 {
			field := part[:idx]
			indexStr := strings.TrimSuffix(part[idx+1:], "]")

			obj, ok := current.(map[string]any)
			if !ok {
				return "", fmt.Errorf("expected object at %q, got %T", field, current)
			}
			arr, ok := obj[field].([]any)
			if !ok {
				return "", fmt.Errorf("expected array at %q", field)
			}
			var i int
			if _, err := fmt.Sscanf(indexStr, "%d", &i); err != nil {
				return "", fmt.Errorf("invalid array index %q", indexStr)
			}
			if i < 0 || i >= len(arr) {
				return "", fmt.Errorf("index %d out of range (len=%d)", i, len(arr))
			}
			current = arr[i]
			continue
		}

		obj, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("expected object at %q, got %T", part, current)
		}
		v, ok := obj[part]
		if !ok {
			return "", fmt.Errorf("field %q not found", part)
		}
		current = v
	}

	// Convert to string.
	switch val := current.(type) {
	case string:
		return val, nil
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val)), nil
		}
		return fmt.Sprintf("%g", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	default:
		// For objects/arrays, marshal back to JSON.
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val), nil
		}
		return string(data), nil
	}
}
