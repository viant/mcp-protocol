package fieldconstraints

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
)

// Apply adds authored bounds to the compiled wire schema.
// JSON Schema evaluates each bound only for the applicable wire type.
func Apply(property map[string]any, tag reflect.StructTag) error {
	for _, pair := range [...]struct {
		kind string
		min  string
		max  string
	}{
		{"array", "minItems", "maxItems"},
		{"string", "minLength", "maxLength"},
		{"number", "minimum", "maximum"},
	} {
		min, hasMin := tag.Lookup(pair.min)
		max, hasMax := tag.Lookup(pair.max)
		if !hasMin && !hasMax {
			continue
		}
		if !hasSchemaType(property, pair.kind) {
			return fmt.Errorf("%s and %s require wire type %s", pair.min, pair.max, pair.kind)
		}
		if pair.kind == "number" {
			if err := applyNumberBounds(property, min, max, hasMin, hasMax); err != nil {
				return err
			}
			continue
		}
		if err := applySizeBounds(property, pair.min, pair.max, min, max, hasMin, hasMax); err != nil {
			return err
		}
	}
	if pattern, ok := tag.Lookup("pattern"); ok {
		if !hasSchemaType(property, "string") {
			return fmt.Errorf("pattern requires a string wire type")
		}
		property["pattern"] = pattern
	}
	return nil
}

func applySizeBounds(property map[string]any, minKey, maxKey, min, max string, hasMin, hasMax bool) error {
	var values [2]int
	for i, bound := range [...]struct {
		key   string
		raw   string
		found bool
	}{
		{minKey, min, hasMin},
		{maxKey, max, hasMax},
	} {
		if !bound.found {
			continue
		}
		value, err := strconv.Atoi(bound.raw)
		if err != nil || value < 0 {
			return fmt.Errorf("%s must be a nonnegative integer", bound.key)
		}
		values[i] = value
		property[bound.key] = value
	}
	if hasMin && hasMax && values[0] > values[1] {
		return fmt.Errorf("%s cannot exceed %s", minKey, maxKey)
	}
	return nil
}

func applyNumberBounds(property map[string]any, min, max string, hasMin, hasMax bool) error {
	var values [2]*big.Float
	for i, bound := range [...]struct {
		key   string
		raw   string
		found bool
	}{
		{"minimum", min, hasMin},
		{"maximum", max, hasMax},
	} {
		if !bound.found {
			continue
		}
		// Preserve the exact JSON number in discovery. Precision for comparison
		// grows with the authored literal, including integers above 2^53.
		value, _, err := big.ParseFloat(bound.raw, 10, uint(4*len(bound.raw)+64), big.ToNearestEven)
		if err != nil || !json.Valid([]byte(bound.raw)) {
			return fmt.Errorf("%s must be a JSON number", bound.key)
		}
		values[i] = value
		property[bound.key] = json.Number(bound.raw)
	}
	if hasMin && hasMax && values[0].Cmp(values[1]) > 0 {
		return fmt.Errorf("minimum cannot exceed maximum")
	}
	return nil
}

func hasSchemaType(property map[string]any, want string) bool {
	matches := func(kind string) bool {
		return kind == want || want == "number" && kind == "integer"
	}
	switch types := property["type"].(type) {
	case string:
		return matches(types)
	case []string:
		for _, kind := range types {
			if matches(kind) {
				return true
			}
		}
		return false
	}
	for _, keyword := range [...]string{"anyOf", "oneOf", "allOf"} {
		branches, _ := property[keyword].([]any)
		for _, branch := range branches {
			if child, ok := branch.(map[string]any); ok && hasSchemaType(child, want) {
				return true
			}
		}
	}
	return false
}
