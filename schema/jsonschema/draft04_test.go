package jsonschema_test

import (
	"encoding/json"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDraft04RejectsLaterValidationVocabulary(t *testing.T) {
	for _, keyword := range []string{"$defs", "$anchor", "$dynamicAnchor", "$dynamicRef", "$recursiveAnchor", "$recursiveRef", "$vocabulary", "const", "contains", "minContains", "maxContains", "propertyNames", "if", "then", "else", "dependentRequired", "dependentSchemas", "unevaluatedItems", "unevaluatedProperties", "prefixItems", "contentEncoding", "contentMediaType", "contentSchema"} {
		t.Run(keyword, func(t *testing.T) {
			_, err := (jsonschema.ResourceCompiler{}).Compile(jsonschema.ResourceRequest{Schema: map[string]any{keyword: map[string]any{}}, Base: "pkg:root.json", Prefix: "#/components/schemas/"})
			if err == nil || !strings.Contains(err.Error(), keyword) {
				t.Fatal(keyword, err)
			}
		})
	}
}
func TestDraft04ChecksEverySchemaPositionIncludingUnusedDefinitions(t *testing.T) {
	newer := map[string]any{"contains": map[string]any{"type": "string"}}
	for name, document := range map[string]map[string]any{
		"property": {"properties": map[string]any{"items": newer}}, "pattern": {"patternProperties": map[string]any{".*": newer}},
		"unused definition": {"definitions": map[string]any{"unused": newer}}, "dependency": {"dependencies": map[string]any{"trigger": newer}},
		"item": {"items": newer}, "tuple": {"items": []any{newer}}, "additional items": {"additionalItems": newer}, "additional properties": {"additionalProperties": newer},
		"not": {"not": newer}, "all": {"allOf": []any{newer}}, "any": {"anyOf": []any{newer}}, "one": {"oneOf": []any{newer}},
		"reference sibling": {"$ref": "defs.json", "if": map[string]any{"type": "object"}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := (jsonschema.ResourceCompiler{}).Compile(jsonschema.ResourceRequest{Schema: document, Base: "pkg:root.json", Prefix: "#/definitions/"})
			if err == nil {
				t.Fatal("newer keyword silently ignored")
			}
		})
	}
	raw, _ := json.Marshal(map[string]any{"definitions": map[string]any{"unused": newer}, "type": "object"})
	_, err := (jsonschema.ResourceCompiler{FS: fstest.MapFS{"pkg:ref.json": {Data: raw}}}).Compile(jsonschema.ResourceRequest{Schema: map[string]any{"$ref": "ref.json"}, Base: "pkg:root.json", Prefix: "#/definitions/"})
	if err == nil {
		t.Fatal("referenced document bypassed dialect guard")
	}
}
func TestDraft04DoesNotInterpretPropertyNamesOrLiteralDataAsKeywords(t *testing.T) {
	source := map[string]any{"type": "object", "properties": map[string]any{"contains": map[string]any{"type": "string"}, "if": map[string]any{"type": "boolean"}}, "dependencies": map[string]any{"contains": []any{"if"}}, "default": map[string]any{"if": true, "contains": "literal"}, "example": map[string]any{"dependentSchemas": "literal"}, "nullable": true, "readOnly": true, "x-label": map[string]any{"const": "extension data"}}
	if _, err := (jsonschema.ResourceCompiler{}).Compile(jsonschema.ResourceRequest{Schema: source, Base: "pkg:root.json", Prefix: "#/definitions/"}); err != nil {
		t.Fatal(err)
	}
}

func TestDraft04CanonicalizesTypedSchemaContainers(t *testing.T) {
	source := map[string]any{"$ref": "ref.json", "definitions": map[string]map[string]any{"unused": {"contains": map[string]any{"type": "string"}}}}
	_, err := (jsonschema.ResourceCompiler{FS: fstest.MapFS{"pkg:ref.json": {Data: []byte(`{"type":"object"}`)}}}).Compile(jsonschema.ResourceRequest{Schema: source, Base: "pkg:root.json", Prefix: "#/definitions/"})
	if err == nil || !strings.Contains(err.Error(), "contains") {
		t.Fatal("typed container bypassed dialect check", err)
	}
}
