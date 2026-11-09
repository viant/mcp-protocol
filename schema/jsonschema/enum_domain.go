package jsonschema

import "github.com/viant/tagly/choice"

// enumDomainSchema lowers the domain owned by the same enum used for binding.
func enumDomainSchema(domain *choice.InputDomain, nullable bool) map[string]any {
	branches := []any{}
	if domain.OpenIntegers {
		branches = append(branches, map[string]any{"type": "integer", "minimum": domain.Minimum, "maximum": domain.Maximum})
	} else if len(domain.Integers) > 0 {
		branches = append(branches, map[string]any{"type": "integer", "enum": domain.Integers})
	}
	if domain.OpenStrings {
		branches = append(branches, map[string]any{"type": "string"})
	} else {
		if len(domain.Strings) > 0 {
			branches = append(branches, map[string]any{"type": "string", "enum": domain.Strings})
		}
		for _, alternative := range domain.StringAlternatives {
			branch := map[string]any{"type": "string", "pattern": alternative.Pattern, "minLength": alternative.Length}
			if !alternative.VariableLength {
				branch["maxLength"] = alternative.Length
			}
			branches = append(branches, branch)
		}
	}
	if nullable || domain.Nullable {
		branches = append(branches, map[string]any{"type": "null"})
	}
	return map[string]any{"anyOf": branches}
}
