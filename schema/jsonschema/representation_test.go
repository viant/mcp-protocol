package jsonschema

import (
	"reflect"
	"testing"

	"github.com/viant/mcp-protocol/schema/fieldname"
	"github.com/viant/x/shape"
)

type MetadataEmbedded struct {
	Label string `json:"label"`
}

func (*MetadataEmbedded) Linked() bool { return true }

type metadataMember struct {
	*MetadataEmbedded
	Value  int                           `json:"value"`
	Secret string                        `json:"secret"`
	Has    *struct{ Value, Secret bool } `json:"-" setMarker:"true"`
}
type metadataEnvelope struct {
	Rows []*metadataMember `json:"rows"`
}

func TestRepresentationSchemaAndNameNormalizationAgree(t *testing.T) {
	original := reflect.TypeFor[metadataEnvelope]()
	walker, _ := shape.CompileRewrite(original)
	metadata, err := walker.RewriteRepresentation(func(_ string, parent reflect.Type, f *shape.FieldMetadata) (*shape.FieldMetadata, error) {
		if parent == reflect.TypeFor[metadataMember]() {
			switch f.Name {
			case "Secret":
				return &shape.FieldMetadata{Name: f.Name, Tag: `json:"-" internal:"true"`}, nil
			case "Value":
				return &shape.FieldMetadata{Name: "Amount", Tag: `json:"amount" mcp:"name=Amount,aliases=amountAlias" choice:"1" choice:"2" required:"true"`}, nil
			}
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	reflector := Reflector{ExcludeInternal: true}
	selected, err := reflector.Compile(Request{Type: original, Representation: metadata})
	if err != nil {
		t.Fatal(err)
	}
	row := selected["properties"].(map[string]any)["rows"].(map[string]any)["items"].(map[string]any)
	properties := row["properties"].(map[string]any)
	if properties["secret"] != nil || properties["Has"] != nil || properties["Amount"] == nil {
		t.Fatal(properties)
	}
	if !reflect.DeepEqual(row["required"], []string{"Amount"}) || !reflect.DeepEqual(properties["Amount"].(map[string]any)["enum"], []any{int64(1), int64(2)}) {
		t.Fatal(row)
	}
	policy := fieldname.Visibility{Representation: metadata, ExcludeInternal: true}
	if named, err := fieldname.HasNames(original, policy); err != nil || !named {
		t.Fatal("nested names not found", err)
	}
	input := map[string]any{"rows": []any{map[string]any{"amountAlias": 2, "unknown": "retained"}}}
	normalized, err := fieldname.Normalize(original, input, policy)
	if err != nil {
		t.Fatal(err)
	}
	result := normalized.(map[string]any)["rows"].([]any)[0].(map[string]any)
	if result["amount"] != 2 || result["amountAlias"] != nil || result["unknown"] != "retained" {
		t.Fatal(result)
	}
	if input["rows"].([]any)[0].(map[string]any)["amountAlias"] != 2 {
		t.Fatal("caller input mutated")
	}
	linked, err := reflector.Compile(Request{Type: original})
	if err != nil {
		t.Fatal(err)
	}
	base := linked["properties"].(map[string]any)["rows"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	if base["value"] == nil || base["secret"] == nil {
		t.Fatal("base schema contaminated")
	}
	member := metadata.FieldShape(0).Elem().Elem()
	fields, err := member.JSONFields()
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range fields {
		if field.Name != "amount" {
			continue
		}
		flat, err := reflector.Compile(Request{Type: field.Field.ReflectedType, Field: &field, Representation: member.FieldShapeByIndex(field.Field.Index)})
		if err != nil || !reflect.DeepEqual(flat["enum"], []any{int64(1), int64(2)}) {
			t.Fatal("flattened member lost choice metadata", err)
		}
	}
}
