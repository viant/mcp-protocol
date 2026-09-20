package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/viant/mcp-protocol/schema/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

func TestReflectorRawMessage(t *testing.T) {
	type request struct {
		ProposedValue json.RawMessage `json:"proposed_value"`
	}
	type rawAlias = json.RawMessage
	for _, typ := range []reflect.Type{
		reflect.TypeFor[json.RawMessage](),
		reflect.TypeFor[rawAlias](),
		reflect.TypeFor[*json.RawMessage](),
		reflect.TypeFor[request](),
		reflect.TypeFor[[]json.RawMessage](),
		reflect.TypeFor[map[string]json.RawMessage](),
	} {
		t.Run(typ.String(), func(t *testing.T) {
			schema, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ})
			if err != nil {
				t.Fatal(err)
			}
			if typ == reflect.TypeFor[json.RawMessage]() && !reflect.DeepEqual(schema, map[string]any{}) {
				t.Fatalf("expected empty schema, got %#v", schema)
			}
			if typ == reflect.TypeFor[request]() {
				property := schema["properties"].(map[string]any)["proposed_value"]
				if !reflect.DeepEqual(property, map[string]any{}) {
					t.Fatalf("expected empty property schema, got %#v", property)
				}
			}
			validator, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range []string{`{"fields":{"freq_capping":3}}`, `[1,"two",null]`, `"text"`, `42`, `1.5`, `true`, `false`, `null`} {
				t.Run(value, func(t *testing.T) {
					document := value
					switch typ.Kind() {
					case reflect.Struct:
						document = `{"proposed_value":` + value + `}`
					case reflect.Map:
						document = `{"key":` + value + `}`
					case reflect.Slice:
						if typ != reflect.TypeFor[json.RawMessage]() {
							document = `[` + value + `]`
						}
					}
					result, err := validator.Validate(gojsonschema.NewStringLoader(document))
					if err != nil {
						t.Fatal(err)
					}
					if !result.Valid() {
						t.Fatal(result.Errors())
					}
					if err := json.Unmarshal([]byte(document), reflect.New(typ).Interface()); err != nil {
						t.Fatalf("binding JSON value: %v", err)
					}
				})
			}
		})
	}
}
