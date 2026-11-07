package jsonschema

import (
	"fmt"
	"strings"

	"github.com/viant/tagly/format"
)

// projectInputAliases operates inside the owning object projection. It returns
// whether ordinary canonical requiredness should still be emitted.
func projectInputAliases(object map[string]any, canonical string, property map[string]any, policy *format.Tag, required bool) bool {
	if policy.InputNames == "" {
		return required
	}
	properties := object["properties"].(map[string]any)
	conservative := policy.InputNamesCase != "" || policy.InputNamesPriority != ""
	description := fmt.Sprintf("Native input alias for %s. Names=%s; priority=%s; kinds=%s; case=%s; exclusions=%s. The runtime enforces this policy. Folded names and canonical precedence have a conservative discovery schema; required presence and winning-value validation remain runtime responsibilities.", canonical, policy.InputNames, policy.InputNamesPriority, policy.InputNamesKinds, policy.InputNamesCase, policy.InputNamesExclude)
	excluded := map[string]bool{}
	for _, name := range strings.Split(policy.InputNamesExclude, "|") {
		excluded[name] = true
	}
	choices := []any{map[string]any{"required": []string{canonical}}}
	for _, alias := range strings.Split(policy.InputNames, "|") {
		if alias == canonical || excluded[alias] {
			continue
		}
		admitted := property
		aliasSchema := property
		if conservative {
			aliasSchema = map[string]any{"description": description}
		} else if policy.InputNamesKinds != "" {
			kinds := map[string]any{"type": strings.Split(policy.InputNamesKinds, "|")}
			admitted = map[string]any{"allOf": []any{property, kinds}}
			aliasSchema = map[string]any{"anyOf": []any{admitted, map[string]any{"not": kinds}}}
		}
		properties[alias] = aliasSchema
		choices = append(choices, map[string]any{"required": []string{alias}, "properties": map[string]any{alias: admitted}})
	}
	if conservative {
		object["description"] = strings.TrimSpace(fmt.Sprint(objectDescription(object)) + " " + description)
		return false
	}
	if required && len(choices) > 1 {
		all, _ := object["allOf"].([]any)
		object["allOf"] = append(all, map[string]any{"anyOf": choices})
		return false
	}
	return required
}

func objectDescription(object map[string]any) string {
	text, _ := object["description"].(string)
	return text
}
