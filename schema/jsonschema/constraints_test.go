package jsonschema

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/x/shape"
	"github.com/xeipuuv/gojsonschema"
)

func TestFieldConstraintsCompileForInputsAndOutputs(t *testing.T) {
	type record struct {
		Name   string  `json:"name" minLength:"2" maxLength:"4" pattern:"^[A-Z]+$"`
		Amount float64 `json:"amount" minimum:"-1.25" maximum:"2.5"`
		Hidden bool    `json:"-" maxLength:"invalid"`
	}
	type input struct {
		Records *[]record `json:"records,omitempty" minItems:"1" maxItems:"2"`
		Empty   []string  `json:"empty,omitempty" maxItems:"0"`
	}
	for _, output := range []bool{false, true} {
		compiled, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeFor[input]()})
		require.NoError(t, err)
		validator, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(compiled))
		require.NoError(t, err)
		for _, test := range []struct {
			payload string
			valid   bool
		}{
			{`{}`, true},
			{`{"records":null,"empty":[]}`, true},
			{`{"empty":["x"]}`, false},
			{`{"records":[]}`, false},
			{`{"records":[{"name":"AB","amount":2.5},{"name":"ABCD","amount":-1.25}]}`, true},
			{`{"records":[{},{},{}]}`, false},
			{`{"records":[{"name":"A","amount":2.5}]}`, false},
			{`{"records":[{"name":"ABCDE","amount":2.5}]}`, false},
			{`{"records":[{"name":"ab","amount":2.5}]}`, false},
			{`{"records":[{"name":"AB","amount":2.51}]}`, false},
			{`{"records":[{"name":"AB","amount":-1.26}]}`, false},
		} {
			result, err := validator.Validate(gojsonschema.NewStringLoader(test.payload))
			require.NoError(t, err)
			require.Equal(t, test.valid, result.Valid(), "output=%t payload=%s: %v", output, test.payload, result.Errors())
		}
	}
}

func TestFieldConstraintsUseEffectiveMetadataAndAliases(t *testing.T) {
	type input struct {
		Rows []string `json:"rows" maxItems:"9"`
	}
	typ := reflect.TypeFor[input]()
	walker, err := shape.CompileRewrite(typ)
	require.NoError(t, err)
	metadata, err := walker.RewriteRepresentation(func(_ string, _ reflect.Type, f *shape.FieldMetadata) (*shape.FieldMetadata, error) {
		return &shape.FieldMetadata{Name: f.Name, Tag: `json:"selected" format:"inputNames=legacy" minItems:"1" maxItems:"2"`}, nil
	})
	require.NoError(t, err)
	reflector := Reflector{InputAliases: true}
	effective, err := reflector.Compile(Request{Type: typ, Representation: metadata})
	require.NoError(t, err)
	properties := effective["properties"].(map[string]any)
	require.Equal(t, properties["selected"], properties["legacy"])
	require.Equal(t, 2, properties["selected"].(map[string]any)["maxItems"])
	fields, err := metadata.JSONFields()
	require.NoError(t, err)
	require.Len(t, fields, 1)
	field := fields[0]
	flattened, err := reflector.Compile(Request{Type: field.Field.ReflectedType, Field: &field, Representation: metadata.FieldShapeByIndex(field.Field.Index)})
	require.NoError(t, err)
	require.Equal(t, 1, flattened["minItems"])
	require.Equal(t, 2, flattened["maxItems"])
	base, err := reflector.Compile(Request{Type: typ})
	require.NoError(t, err)
	require.Equal(t, 9, base["properties"].(map[string]any)["rows"].(map[string]any)["maxItems"])
}

func TestFieldConstraintsFollowScalarWireType(t *testing.T) {
	type input struct {
		Scalar *testConstrainedScalar `json:"scalar" wireSchema:"{\"type\":[\"string\",\"integer\"]}" maxLength:"4" minimum:"9007199254740993" maximum:"9007199254740994"`
		Quoted int                    `json:"quoted,string" maxLength:"3"`
	}
	compiled, err := (Reflector{}).Compile(Request{Type: reflect.TypeFor[input]()})
	require.NoError(t, err)
	properties := compiled["properties"].(map[string]any)
	scalar := properties["scalar"].(map[string]any)
	require.Equal(t, []string{"string", "integer", "null"}, scalar["type"])
	require.Equal(t, json.Number("9007199254740993"), scalar["minimum"])
	require.Equal(t, json.Number("9007199254740994"), scalar["maximum"])
	require.Equal(t, 4, scalar["maxLength"])
	require.Equal(t, map[string]any{"type": "string", "maxLength": 3}, properties["quoted"])
}

type testConstrainedScalar struct{}

func (testConstrainedScalar) MarshalJSON() ([]byte, error) {
	panic("schema compilation must never serialize a value")
}

func TestFieldConstraintsRejectInvalidDeclarations(t *testing.T) {
	for _, tc := range []struct {
		typ   reflect.Type
		tag   reflect.StructTag
		error string
	}{
		{reflect.TypeFor[[]int](), `minItems:"2" maxItems:"1"`, "minItems cannot exceed maxItems"},
		{reflect.TypeFor[[]int](), `maxItems:"-1"`, "maxItems must be a nonnegative integer"},
		{reflect.TypeFor[string](), `maxLength:"1.5"`, "maxLength must be a nonnegative integer"},
		{reflect.TypeFor[string](), `minItems:"1"`, "require wire type array"},
		{reflect.TypeFor[int](), `pattern:"^[0-9]+$"`, "pattern requires a string wire type"},
		{reflect.TypeFor[float64](), `minimum:"NaN"`, "minimum must be a JSON number"},
		{reflect.TypeFor[float64](), `maximum:"0x10"`, "maximum must be a JSON number"},
		{reflect.TypeFor[uint64](), `minimum:"9007199254740994" maximum:"9007199254740993"`, "minimum cannot exceed maximum"},
	} {
		t.Run(string(tc.tag), func(t *testing.T) {
			typ := reflect.StructOf([]reflect.StructField{{Name: "Value", Type: tc.typ, Tag: tc.tag}})
			_, err := (Reflector{}).Compile(Request{Type: typ})
			require.ErrorContains(t, err, "field Value")
			require.ErrorContains(t, err, tc.error)
		})
	}
}
