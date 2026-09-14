package jsonschema_test

import (
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"testing"
	"testing/fstest"
)

func TestResourceSchemaValidationAndBundling(t *testing.T) {
	files := fstest.MapFS{"pkg:schemas/node.json": {Data: []byte(`{"type":"object","properties":{"id":{"type":"integer"},"next":{"$ref":"#"}}}`)}, "pkg:bad.json": {Data: []byte(`{"type":"wrong"}`)}}
	tests := []struct {
		name   string
		schema map[string]any
		valid  bool
	}{
		{"inline", map[string]any{"type": "string"}, true},
		{"recursive relative", map[string]any{"$ref": "schemas/node.json"}, true},
		{"namespaced", map[string]any{"$ref": "pkg:schemas/node.json"}, true},
		{"pointer", map[string]any{"$ref": "schemas/node.json#/properties/id"}, true},
		{"missing", map[string]any{"$ref": "missing.json"}, false},
		{"dangling", map[string]any{"$ref": "schemas/node.json#/absent"}, false},
		{"bad type", map[string]any{"type": "invalid"}, false},
		{"bad referenced", map[string]any{"$ref": "bad.json"}, false},
		{"invalid keyword", map[string]any{"type": "string", "minLength": -1}, false},
		{"no network", map[string]any{"$ref": "https://example.com/schema.json"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle, err := (jsonschema.ResourceCompiler{FS: files}).Compile(jsonschema.ResourceRequest{Schema: test.schema, Base: "pkg:docs.yaml", Prefix: "#/components/schemas/"})
			if (err == nil) != test.valid {
				t.Fatalf("bundle=%+v err=%v", bundle, err)
			}
			if test.valid && test.name == "recursive relative" && len(bundle.Definitions) == 0 {
				t.Fatal("missing recursive definitions")
			}
		})
	}
}
