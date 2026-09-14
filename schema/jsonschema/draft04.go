package jsonschema

import (
	"fmt"
	"sort"
)

// draft04Dialect guards recognized later-draft vocabulary that draft-04's
// permissive metaschema would otherwise accept and silently ignore. OpenAPI
// annotations and extension data remain allowed; they do not impose constraints.
type draft04Dialect struct{}

func (d draft04Dialect) validate(schema map[string]any, path string) error {
	keys := make([]string, 0, len(schema))
	for key := range schema {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		switch key {
		case "$defs", "$anchor", "$dynamicAnchor", "$dynamicRef", "$recursiveAnchor", "$recursiveRef", "$vocabulary",
			"const", "contains", "minContains", "maxContains", "propertyNames", "if", "then", "else",
			"dependentRequired", "dependentSchemas", "unevaluatedItems", "unevaluatedProperties", "prefixItems",
			"contentEncoding", "contentMediaType", "contentSchema":
			return fmt.Errorf("schema %s/%s is unsupported in draft-04", path, key)
		}
		value := schema[key]
		switch key {
		case "properties", "patternProperties", "definitions", "dependencies":
			if entries, ok := value.(map[string]any); ok {
				names := make([]string, 0, len(entries))
				for name := range entries {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					if child, ok := entries[name].(map[string]any); ok {
						if err := d.validate(child, path+"/"+key+"/"+name); err != nil {
							return err
						}
					}
				}
			}
		case "items", "additionalItems", "additionalProperties", "not", "allOf", "anyOf", "oneOf":
			if child, ok := value.(map[string]any); ok {
				if err := d.validate(child, path+"/"+key); err != nil {
					return err
				}
			}
			if children, ok := value.([]any); ok {
				for index, item := range children {
					if child, ok := item.(map[string]any); ok {
						if err := d.validate(child, fmt.Sprintf("%s/%s/%d", path, key, index)); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
