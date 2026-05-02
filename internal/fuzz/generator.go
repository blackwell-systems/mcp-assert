// Package fuzz generates adversarial inputs for MCP tools from their JSON Schema.
//
// Unlike audit mode (which generates one happy-path input per tool), fuzz mode
// generates many inputs per tool designed to trigger crashes, hangs, and
// protocol violations. Inputs are category-based (not purely random) so that
// they are more likely to slip past schema validation and reach the tool's
// handler logic.
package fuzz

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// InputCase is a single fuzz input with a human-readable label describing
// what category of adversarial input it represents.
type InputCase struct {
	Label string
	Args  map[string]any
}

// GenerateInputs produces a set of adversarial inputs for a tool based on
// its JSON Schema. The seed controls randomness for reproducibility.
func GenerateInputs(schema mcp.ToolInputSchema, runs int, seed int64) []InputCase {
	rng := rand.New(rand.NewSource(seed))

	// Always include these structural cases.
	cases := []InputCase{
		{Label: "empty object", Args: map[string]any{}},
		{Label: "null args", Args: nil},
	}

	// Omit each required field one at a time.
	for _, req := range schema.Required {
		partial := generateValidArgs(schema)
		delete(partial, req)
		cases = append(cases, InputCase{
			Label: fmt.Sprintf("missing required: %s", req),
			Args:  partial,
		})
	}

	// Per-property type-specific cases.
	for name, prop := range schema.Properties {
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}
		typ, _ := propMap["type"].(string)
		cases = append(cases, generatePropertyCases(name, typ, schema, rng)...)
	}

	// Fill remaining slots with random mutations.
	for len(cases) < runs {
		cases = append(cases, generateRandomMutation(schema, rng))
	}

	// Trim to requested count.
	if len(cases) > runs {
		cases = cases[:runs]
	}

	return cases
}

// generateValidArgs creates a minimal valid input matching the schema.
func generateValidArgs(schema mcp.ToolInputSchema) map[string]any {
	args := make(map[string]any)
	required := make(map[string]bool)
	for _, r := range schema.Required {
		required[r] = true
	}

	for name, prop := range schema.Properties {
		if !required[name] {
			continue
		}
		propMap, ok := prop.(map[string]any)
		if !ok {
			args[name] = "test"
			continue
		}
		typ, _ := propMap["type"].(string)
		switch typ {
		case "string":
			args[name] = "test"
		case "integer":
			args[name] = 1
		case "number":
			args[name] = 1.0
		case "boolean":
			args[name] = true
		case "array":
			args[name] = []any{"test"}
		case "object":
			args[name] = map[string]any{"key": "value"}
		default:
			args[name] = "test"
		}
	}
	return args
}

// generatePropertyCases produces adversarial inputs for a single property.
func generatePropertyCases(name, typ string, schema mcp.ToolInputSchema, rng *rand.Rand) []InputCase {
	var cases []InputCase

	// Wrong type: send the opposite of what's expected.
	wrongType := generateWrongType(typ, rng)
	base := generateValidArgs(schema)
	base[name] = wrongType
	cases = append(cases, InputCase{
		Label: fmt.Sprintf("wrong type for %s (sent %T)", name, wrongType),
		Args:  base,
	})

	switch typ {
	case "string":
		for _, sc := range stringCases(name) {
			base := generateValidArgs(schema)
			base[name] = sc.value
			cases = append(cases, InputCase{
				Label: fmt.Sprintf("%s: %s", name, sc.label),
				Args:  base,
			})
		}

	case "integer", "number":
		for _, nc := range numberCases(name, typ) {
			base := generateValidArgs(schema)
			base[name] = nc.value
			cases = append(cases, InputCase{
				Label: fmt.Sprintf("%s: %s", name, nc.label),
				Args:  base,
			})
		}

	case "boolean":
		// Send truthy/falsy non-boolean values.
		for _, val := range []struct {
			label string
			value any
		}{
			{"null boolean", nil},
			{"int as boolean (0)", 0},
			{"int as boolean (1)", 1},
			{"string as boolean", "true"},
		} {
			base := generateValidArgs(schema)
			base[name] = val.value
			cases = append(cases, InputCase{
				Label: fmt.Sprintf("%s: %s", name, val.label),
				Args:  base,
			})
		}

	case "array":
		for _, ac := range arrayCases(name) {
			base := generateValidArgs(schema)
			base[name] = ac.value
			cases = append(cases, InputCase{
				Label: fmt.Sprintf("%s: %s", name, ac.label),
				Args:  base,
			})
		}

	case "object":
		for _, oc := range objectCases(name) {
			base := generateValidArgs(schema)
			base[name] = oc.value
			cases = append(cases, InputCase{
				Label: fmt.Sprintf("%s: %s", name, oc.label),
				Args:  base,
			})
		}
	}

	return cases
}

type labeledValue struct {
	label string
	value any
}

func stringCases(name string) []labeledValue {
	return []labeledValue{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"very long string (10KB)", strings.Repeat("a", 10_000)},
		{"null bytes", "test\x00value"},
		{"unicode edge case", "test\u200b\u200b\u200b"},   // zero-width spaces
		{"emoji", "\U0001F4A5\U0001F525\U0001F480"},        // explosion, fire, skull
		{"newlines", "line1\nline2\nline3"},
		{"path traversal", "../../../etc/passwd"},
		{"shell injection", "; rm -rf / #"},
		{"SQL injection", "' OR 1=1 --"},
		{"null value", nil},
	}
}

func numberCases(name, typ string) []labeledValue {
	cases := []labeledValue{
		{"zero", 0},
		{"negative", -1},
		{"negative large", -999999},
		{"max int", math.MaxInt32},
		{"min int", math.MinInt32},
		{"null number", nil},
	}
	if typ == "number" {
		cases = append(cases,
			labeledValue{"float precision", 0.1 + 0.2},
			labeledValue{"very small float", 0.000000001},
			labeledValue{"very large float", 1e308},
			labeledValue{"negative zero", math.Copysign(0, -1)},
			labeledValue{"NaN", math.NaN()},
			labeledValue{"positive infinity", math.Inf(1)},
		)
	}
	return cases
}

func arrayCases(name string) []labeledValue {
	return []labeledValue{
		{"empty array", []any{}},
		{"null array", nil},
		{"single element", []any{"test"}},
		{"many elements (1000)", makeRepeatedArray(1000)},
		{"mixed types", []any{"string", 42, true, nil}},
		{"nested arrays", []any{[]any{[]any{"deep"}}}},
	}
}

func objectCases(name string) []labeledValue {
	return []labeledValue{
		{"empty object", map[string]any{}},
		{"null object", nil},
		{"deeply nested", makeNestedObject(20)},
		{"many keys", makeManyKeys(100)},
	}
}

func generateWrongType(expectedType string, rng *rand.Rand) any {
	switch expectedType {
	case "string":
		return 42
	case "integer", "number":
		return "not a number"
	case "boolean":
		return "not a boolean"
	case "array":
		return "not an array"
	case "object":
		return "not an object"
	default:
		return 42
	}
}

func generateRandomMutation(schema mcp.ToolInputSchema, rng *rand.Rand) InputCase {
	args := generateValidArgs(schema)

	// Pick a random property to mutate.
	props := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		props = append(props, name)
	}

	if len(props) == 0 {
		return InputCase{Label: "random: valid args (no properties)", Args: args}
	}

	target := props[rng.Intn(len(props))]

	// Random mutation: either remove, null, or replace with random value.
	switch rng.Intn(4) {
	case 0:
		delete(args, target)
		return InputCase{Label: fmt.Sprintf("random: omit %s", target), Args: args}
	case 1:
		args[target] = nil
		return InputCase{Label: fmt.Sprintf("random: null %s", target), Args: args}
	case 2:
		args[target] = strings.Repeat("x", rng.Intn(10000))
		return InputCase{Label: fmt.Sprintf("random: long string for %s", target), Args: args}
	default:
		args[target] = rng.Intn(1000000) - 500000
		return InputCase{Label: fmt.Sprintf("random: random int for %s", target), Args: args}
	}
}

func makeRepeatedArray(n int) []any {
	arr := make([]any, n)
	for i := range arr {
		arr[i] = fmt.Sprintf("item_%d", i)
	}
	return arr
}

func makeNestedObject(depth int) map[string]any {
	if depth <= 0 {
		return map[string]any{"leaf": "value"}
	}
	return map[string]any{"nested": makeNestedObject(depth - 1)}
}

func makeManyKeys(n int) map[string]any {
	m := make(map[string]any, n)
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	return m
}
