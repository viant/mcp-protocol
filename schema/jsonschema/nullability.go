package jsonschema

// fieldNullability applies authored format metadata to this field only. It
// preserves nested fields, item policies and reference targets. Compilation
// owns these maps; no invocation values or runtime validation are involved.
func fieldNullability(property map[string]any, nullable bool) map[string]any {
	result := make(map[string]any, len(property)+1)
	for key, value := range property {
		result[key] = value
	}
	if nullable {
		// A composition or reference can still exclude null independently of type.
		// Preserve that branch verbatim and add the explicitly permitted null branch.
		for _, key := range []string{"$ref", "anyOf", "oneOf", "allOf", "not"} {
			if _, ok := property[key]; ok {
				return map[string]any{"anyOf": []any{property, map[string]any{"type": "null"}}}
			}
		}
	}
	var types []string
	switch value := property["type"].(type) {
	case string:
		types = []string{value}
	case []string:
		types = value
	}
	if len(types) > 0 {
		selected := make([]string, 0, len(types)+1)
		for _, kind := range types {
			if kind != "null" {
				selected = append(selected, kind)
			}
		}
		if nullable {
			selected = append(selected, "null")
		}
		switch len(selected) {
		case 0:
			return map[string]any{"not": map[string]any{}}
		case 1:
			result["type"] = selected[0]
		default:
			result["type"] = selected
		}
	}
	if values, ok := property["enum"].([]any); ok {
		selected := make([]any, 0, len(values)+1)
		for _, value := range values {
			if value != nil {
				selected = append(selected, value)
			}
		}
		if nullable {
			selected = append(selected, nil)
		}
		result["enum"] = selected
	}
	if !nullable {
		// Do not rewrite referenced definitions or composed alternatives: they may
		// be shared with nullable fields. A local conjunction excludes only this null.
		if _, ref := property["$ref"]; ref {
			return map[string]any{"allOf": []any{property, map[string]any{"not": map[string]any{"type": "null"}}}}
		}
		if len(types) == 0 {
			if _, hasNot := result["not"]; hasNot {
				return map[string]any{"allOf": []any{result, map[string]any{"not": map[string]any{"type": "null"}}}}
			}
			result["not"] = map[string]any{"type": "null"}
		}
	}
	return result
}
