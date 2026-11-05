package schema

import (
	"encoding/json"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

type RequiredBase struct {
	Key      string `json:"key" required:"true"`
	Endpoint string `json:"endpoint" required:"true"`
}

type RequiredEmbeddedFirst struct {
	RequiredBase
	Key string `json:"key" required:"true"`
}

type RequiredDirectFirst struct {
	Key string `json:"key" required:"true"`
	RequiredBase
}

type RequiredInline struct {
	Base RequiredBase `json:",inline"`
	Key  string       `json:"key" required:"true"`
}

func TestRequiredFieldsRemainUniqueAcrossEmbedding(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"embedded first", RequiredEmbeddedFirst{}},
		{"direct first", RequiredDirectFirst{}},
		{"explicit inline", RequiredInline{}},
		{"nested object", struct {
			Payload RequiredEmbeddedFirst `json:"payload"`
		}{}},
		{"array element", struct {
			Items []RequiredEmbeddedFirst `json:"items"`
		}{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var generated ToolInputSchema
			if err := generated.Load(tc.value); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(generated)
			if err != nil {
				t.Fatal(err)
			}
			loader := gojsonschema.NewSchemaLoader()
			loader.Validate = true
			if _, err = loader.Compile(gojsonschema.NewBytesLoader(raw)); err != nil {
				t.Fatalf("generated schema fails meta-schema validation: %v", err)
			}
		})
	}
}

func TestRequiredEmbeddingKeepsConstraints(t *testing.T) {
	var generated ToolInputSchema
	if err := generated.Load(RequiredEmbeddedFirst{}); err != nil {
		t.Fatal(err)
	}
	if len(generated.Required) != 2 || generated.Required[0] != "key" || generated.Required[1] != "endpoint" {
		t.Fatalf("required constraints/order changed: %v", generated.Required)
	}
	raw, err := json.Marshal(generated)
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value map[string]any
		valid bool
	}{
		{map[string]any{"key": "item", "endpoint": "local"}, true},
		{map[string]any{"endpoint": "local"}, false},
		{map[string]any{"key": "item"}, false},
	} {
		result, err := compiled.Validate(gojsonschema.NewGoLoader(tc.value))
		if err != nil {
			t.Fatal(err)
		}
		if result.Valid() != tc.valid {
			t.Fatalf("validation=%v want %v for %v", result.Errors(), tc.valid, tc.value)
		}
	}
}
