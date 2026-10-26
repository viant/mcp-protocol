package jsonschema_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/viant/mcp-protocol/schema/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

type epochOrText struct{}

func (epochOrText) MarshalJSON() ([]byte, error) { panic("schema reflection invoked serializer") }
func (*epochOrText) UnmarshalJSON([]byte) error  { panic("schema reflection invoked decoder") }

func TestCustomScalarWireUnionMatchesNestedInputs(t *testing.T) {
	type row struct {
		When *epochOrText `json:"when,omitempty" wireSchema:"{\"type\":[\"string\",\"integer\"]}"`
	}
	type input struct {
		Rows []row `json:"rows"`
	}
	schema, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[input]()})
	if err != nil {
		t.Fatal(err)
	}
	validator, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value any
		valid bool
	}{{"2026-10-25T00:00:00Z", true}, {int64(1792886400000), true}, {nil, true}, {map[string]any{}, false}, {[]any{}, false}, {true, false}, {1.5, false}} {
		result, err := validator.Validate(gojsonschema.NewGoLoader(map[string]any{"rows": []any{map[string]any{"when": tc.value}}}))
		if err != nil || result.Valid() != tc.valid {
			t.Fatalf("value=%v valid=%v err=%v schema=%v", tc.value, result.Valid(), err, schema)
		}
	}
}

func TestScalarWireRejectsUnsupportedOrConflictingDeclarations(t *testing.T) {
	tests := []struct {
		typ reflect.Type
		tag string
	}{
		{reflect.TypeFor[int](), `wireSchema:"{\"type\":\"string\"}"`},
		{reflect.TypeFor[time.Time](), `wireSchema:"{\"type\":\"integer\"}"`},
		{reflect.TypeFor[epochOrText](), `wireSchema:""`},
		{reflect.TypeFor[epochOrText](), `wireSchema:"{\"type\":\"object\"}"`},
		{reflect.TypeFor[epochOrText](), `wireSchema:"{\"type\":[\"string\",\"string\"]}"`},
		{reflect.TypeFor[epochOrText](), `wireSchema:"{\"type\":\"string\",\"type\":\"integer\"}"`},
		{reflect.TypeFor[epochOrText](), `wireSchema:"{\"type\":\"string\",\"$ref\":\"https://example.test/schema\"}"`},
		{reflect.TypeFor[epochOrText](), `wireSchema:"{\"type\":\"integer\",\"format\":\"date-time\"}"`},
		{reflect.TypeFor[epochOrText](), `json:",string" wireSchema:"{\"type\":\"string\"}"`},
	}
	for _, tc := range tests {
		if _, found, err := jsonschema.ScalarWireSchema(reflect.StructField{Name: "Value", Type: tc.typ, Tag: reflect.StructTag(tc.tag)}); !found || err == nil {
			t.Fatalf("accepted %s %s", tc.typ, tc.tag)
		}
	}
}

func TestHiddenWireSchemaDoesNotAffectDiscovery(t *testing.T) {
	type input struct {
		Hidden epochOrText `json:"-" wireSchema:"invalid"`
		Value  string      `json:"value"`
	}
	if _, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[input]()}); err != nil {
		t.Fatal(err)
	}
}

func TestScalarWireRejectsCompetingTypeAuthority(t *testing.T) {
	type row struct {
		When epochOrText `json:"when" wireSchema:"{\"type\":\"integer\"}"`
	}
	_, err := (jsonschema.Reflector{WireTypes: map[reflect.Type]reflect.Type{reflect.TypeFor[epochOrText](): reflect.TypeFor[string]()}}).Compile(jsonschema.Request{Type: reflect.TypeFor[row]()})
	if err == nil {
		t.Fatal("accepted competing scalar and type authorities")
	}
}
