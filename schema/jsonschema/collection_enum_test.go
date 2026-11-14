package jsonschema

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/viant/tagly/choice"
	"reflect"
	"testing"
)

func TestMCPCollectionEnumItems(t *testing.T) {
	type fields struct {
		Channels []string  `json:"channels" mcpEnum:"true" choice:"DISPLAY" choice:"VIDEO"`
		Optional *[]string `json:"optional" mcpEnum:"true" choice:"DISPLAY" choice:"VIDEO"`
		Scalar   string    `json:"scalar" mcpEnum:"true" choice:"1:ACTIVE" enumAccept:"scalar" enumStyle:"literal"`
		NoNull   []string  `json:"noNull" mcpEnum:"true" choice:"DISPLAY" format:"nullable=false"`
		Ordinary []string  `json:"ordinary"`
	}
	for _, output := range []bool{false, true} {
		document, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeOf(fields{})})
		require.NoError(t, err)
		properties := document["properties"].(map[string]any)
		for _, name := range []string{"channels", "optional"} {
			property := properties[name].(map[string]any)
			require.Equal(t, []string{"array", "null"}, property["type"])
			require.NotContains(t, property, "enum")
			items := property["items"].(map[string]any)
			require.Equal(t, "string", items["type"])
			require.Equal(t, []any{"DISPLAY", "VIDEO"}, items["enum"])
		}
		require.Equal(t, "array", properties["noNull"].(map[string]any)["type"])
		require.Equal(t, "array", properties["ordinary"].(map[string]any)["type"])
	}
}

func TestExplicitNestedCollectionEnumKeepsItems(t *testing.T) {
	type deal struct {
		Channels []string `json:"channels" mcpEnum:"true" enumAccept:"literal" choice:"DISPLAY" choice:"VIDEO"`
	}
	type input struct {
		Deals []deal `json:"deals"`
	}
	for _, output := range []bool{false, true} {
		document, err := (Reflector{Output: output}).Compile(Request{Type: reflect.TypeFor[input]()})
		require.NoError(t, err)
		channels := document["properties"].(map[string]any)["deals"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)["channels"].(map[string]any)
		require.Equal(t, []string{"array", "null"}, channels["type"])
		require.Equal(t, []any{"DISPLAY", "VIDEO"}, channels["items"].(map[string]any)["enum"])
	}
}

func TestNumericCollectionEnumUsesScalarElementDomain(t *testing.T) {
	type fields struct {
		Codes []int `json:"codes,omitempty" mcpEnum:"true" choice:"3:THREE" choice:"7:SEVEN" enumAccept:"literal" enumStyle:"ordinal"`
	}
	type scalar struct {
		Code int `json:"code" choice:"3:THREE" choice:"7:SEVEN" enumAccept:"literal" enumStyle:"ordinal"`
	}
	for _, output := range []bool{false, true} {
		reflector := Reflector{Output: output}
		document, err := reflector.Compile(Request{Type: reflect.TypeFor[fields]()})
		require.NoError(t, err)
		scalarDocument, err := reflector.Compile(Request{Type: reflect.TypeFor[scalar]()})
		require.NoError(t, err)
		property := document["properties"].(map[string]any)["codes"].(map[string]any)
		require.NotContains(t, property, "enum")
		require.Equal(t, scalarDocument["properties"].(map[string]any)["code"], property["items"])
		require.Empty(t, document["required"])
	}
}

func TestNumericCollectionEnumInputOutputPoliciesMatchScalar(t *testing.T) {
	for _, policy := range []string{
		`enumAccept:"scalar" enumStyle:"ordinal" enumInteger:"passthrough"`,
		`enumAccept:"scalar" enumStyle:"ordinal" enumAlias:"3:Third" enumCase:"fold"`,
		`enumAccept:"ordinal" enumStyle:"ordinal"`,
	} {
		tag := reflect.StructTag(`json:"value" choice:"3:THREE" choice:"7:SEVEN" ` + policy)
		scalarType := reflect.StructOf([]reflect.StructField{{Name: "Value", Type: reflect.TypeFor[int](), Tag: tag}})
		collectionType := reflect.StructOf([]reflect.StructField{{Name: "Value", Type: reflect.TypeFor[[]int](), Tag: reflect.StructTag(string(tag) + ` mcpEnum:"true"`)}})
		for _, output := range []bool{false, true} {
			reflector := Reflector{Output: output}
			scalarDocument, err := reflector.Compile(Request{Type: scalarType})
			require.NoError(t, err, "policy=%s output=%t", policy, output)
			collectionDocument, err := reflector.Compile(Request{Type: collectionType})
			require.NoError(t, err, "policy=%s output=%t", policy, output)
			property := collectionDocument["properties"].(map[string]any)["value"].(map[string]any)
			require.Equal(t, scalarDocument["properties"].(map[string]any)["value"], property["items"], "policy=%s output=%t", policy, output)
		}
	}
}

func TestNumericCollectionEnumOptInAndNullability(t *testing.T) {
	type fields struct {
		Codes    []int `json:"codes,omitempty" mcpEnum:"true" choice:"3:THREE" enumAccept:"literal" enumStyle:"literal" enumNull:"reject"`
		Ordinary []int `json:"ordinary,omitempty"`
	}
	document, err := (Reflector{}).Compile(Request{Type: reflect.TypeFor[fields]()})
	require.NoError(t, err)
	properties := document["properties"].(map[string]any)
	codes := properties["codes"].(map[string]any)
	require.Equal(t, "array", codes["type"])
	require.Equal(t, "string", codes["items"].(map[string]any)["type"])
	require.Equal(t, "integer", properties["ordinary"].(map[string]any)["items"].(map[string]any)["type"])
	require.Empty(t, document["required"])
	output, err := (Reflector{Output: true}).Compile(Request{Type: reflect.TypeFor[fields]()})
	require.NoError(t, err)
	require.Equal(t, []string{"array", "null"}, output["properties"].(map[string]any)["codes"].(map[string]any)["type"])
}

func TestNumericCollectionSchemaMatchesNativeCodec(t *testing.T) {
	type fields struct {
		Codes []int `json:"codes,omitempty" mcpEnum:"true" choice:"3:THREE" choice:"7:SEVEN" enumAccept:"literal" enumStyle:"literal" enumNull:"reject"`
	}
	field := reflect.TypeFor[fields]().Field(0)
	codec, err := choice.ParseEnum(field.Tag, field.Type)
	require.NoError(t, err)
	require.NotNil(t, codec)
	decoded, err := codec.DecodeJSON([]byte(`["SEVEN","THREE"]`))
	require.NoError(t, err)
	var codes []int
	require.NoError(t, json.Unmarshal(decoded, &codes))
	require.Equal(t, []int{7, 3}, codes)
	// Storage stays numeric; only the opted-in codec exposes names.
	stored, err := json.Marshal(codes)
	require.NoError(t, err)
	require.JSONEq(t, `[7,3]`, string(stored))
	encoded, err := codec.Encode(codes)
	require.NoError(t, err)
	wire, err := json.Marshal(encoded)
	require.NoError(t, err)
	require.JSONEq(t, `["SEVEN","THREE"]`, string(wire))
	for _, raw := range []string{`[null]`, `[3]`, `["UNKNOWN"]`, `["THREE"] trailing`, `[9999999999999999999999999999]`} {
		_, err := codec.DecodeJSON([]byte(raw))
		require.Error(t, err, raw)
	}
	for _, raw := range []string{`[]`} {
		_, err := codec.DecodeJSON([]byte(raw))
		require.NoError(t, err, raw)
	}
}
