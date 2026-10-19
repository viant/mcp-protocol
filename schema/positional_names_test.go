package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-protocol/schema/fieldname"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"reflect"
	"testing"
)

type namedZone struct {
	IanaTimezoneStr *string `json:"ianaTimezoneStr" mcp:"IanaTimezoneStr"`
}
type namedAdvertiser struct {
	AccountingId *int       `json:"accountingId,omitempty" mcp:"AccountingId"`
	TimeZone     *namedZone `json:"timeZone"`
	Explicit     string     `json:"custom"`
	Hidden       string     `json:"-" mcp:"resurrect"`
	MCPHidden    string     `json:"mcpHidden" mcp:"-"`
	Has          bool       `json:"Has" setMarker:"true"`
}
type defaultAdvertiser struct {
	AccountingId *int         `json:"accountingId,omitempty"`
	TimeZone     *defaultZone `json:"timeZone"`
	Explicit     string       `json:"custom"`
	Hidden       string       `json:"-"`
	MCPHidden    string       `json:"mcpHidden" mcp:"-"`
	Has          bool         `json:"Has" setMarker:"true"`
}
type defaultZone struct {
	IanaTimezoneStr *string `json:"ianaTimezoneStr"`
}

func compile(t *testing.T, v any) map[string]any {
	r, e := (jsonschema.Reflector{ExcludeInternal: true, ExcludeField: func(f reflect.StructField) bool { return f.Tag.Get("mcp") == "-" }}).Compile(jsonschema.Request{Type: reflect.TypeOf(v)})
	require.NoError(t, e)
	return r
}
func TestPositionalTwoFieldsAndDefaults(t *testing.T) {
	before := compile(t, defaultAdvertiser{})
	after := compile(t, namedAdvertiser{})
	p := after["properties"].(map[string]any)
	p["accountingId"] = p["AccountingId"]
	delete(p, "AccountingId")
	z := p["timeZone"].(map[string]any)["properties"].(map[string]any)
	z["ianaTimezoneStr"] = z["IanaTimezoneStr"]
	delete(z, "IanaTimezoneStr")
	require.Equal(t, before, after)
	legacy, req := schema.StructToProperties(reflect.TypeFor[namedZone]())
	require.Contains(t, legacy, "IanaTimezoneStr")
	require.Empty(t, req)
	type requiredName struct {
		Value string `json:"value" mcp:"Value"`
	}
	_, required := schema.StructToProperties(reflect.TypeFor[requiredName]())
	require.Equal(t, []string{"Value"}, required)
	require.NotContains(t, req, "ianaTimezoneStr")
	n := 1
	s := "UTC"
	a := namedAdvertiser{AccountingId: &n, TimeZone: &namedZone{&s}}
	b := defaultAdvertiser{AccountingId: &n, TimeZone: &defaultZone{&s}}
	ja, e := json.Marshal(a)
	require.NoError(t, e)
	jb, e := json.Marshal(b)
	require.NoError(t, e)
	require.Equal(t, string(jb), string(ja))
	require.NotContains(t, compile(t, namedAdvertiser{})["properties"], "AccountingIdNotPresent")
}
func TestOrdinaryNestedDecoder(t *testing.T) {
	for _, raw := range []string{`{}`, `{"AccountingId":null,"timeZone":null}`, `{"AccountingId":0,"timeZone":{"IanaTimezoneStr":""}}`, `{"accountingId":2,"timeZone":{"ianaTimezoneStr":"UTC"}}`} {
		var m any
		require.NoError(t, json.Unmarshal([]byte(raw), &m))
		normalized, e := fieldname.Normalize(reflect.TypeFor[namedAdvertiser](), m)
		require.NoError(t, e)
		data, e := json.Marshal(normalized)
		require.NoError(t, e)
		var value namedAdvertiser
		require.NoError(t, json.Unmarshal(data, &value))
		if raw == `{}` {
			require.Nil(t, value.AccountingId)
			require.Nil(t, value.TimeZone)
		}
		if raw == `{"AccountingId":0,"timeZone":{"IanaTimezoneStr":""}}` {
			require.NotNil(t, value.AccountingId)
			require.Equal(t, 0, *value.AccountingId)
			require.NotNil(t, value.TimeZone.IanaTimezoneStr)
			require.Empty(t, *value.TimeZone.IanaTimezoneStr)
		}
	}
	for _, raw := range []string{`{"AccountingId":1,"accountingId":1}`, `{"timeZone":{"IanaTimezoneStr":"UTC","ianaTimezoneStr":"UTC"}}`} {
		var m any
		require.NoError(t, json.Unmarshal([]byte(raw), &m))
		_, e := fieldname.Normalize(reflect.TypeFor[namedAdvertiser](), m)
		require.ErrorContains(t, e, "conflicting MCP names")
	}
	var m any
	require.NoError(t, json.Unmarshal([]byte(`{"AccountingId":"bad"}`), &m))
	n, e := fieldname.Normalize(reflect.TypeFor[namedAdvertiser](), m)
	require.NoError(t, e)
	b, e := json.Marshal(n)
	require.NoError(t, e)
	var target namedAdvertiser
	require.Error(t, json.Unmarshal(b, &target))
}
func TestCollisionFallbackAndLoad(t *testing.T) {
	type collision struct {
		A string `json:"a" mcp:"b"`
		B string `json:"b"`
	}
	p, req := schema.StructToProperties(reflect.TypeFor[collision]())
	require.Contains(t, p, "a")
	require.Contains(t, p, "b")
	require.ElementsMatch(t, []string{"a", "b"}, req)
	_, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[collision]()})
	require.ErrorContains(t, e, "conflicts")
	var s schema.ToolInputSchema
	require.ErrorContains(t, s.Load(collision{}), "conflicts")
}

type recursiveNamed struct {
	Value string          `json:"value" mcp:"Value"`
	Next  *recursiveNamed `json:"next"`
}

func TestRecursiveAndSharedDoNotContaminate(t *testing.T) {
	a := compile(t, recursiveNamed{})
	b := compile(t, recursiveNamed{})
	require.Equal(t, a, b)
	defs := a["$defs"].(map[string]any)
	require.Len(t, defs, 1)
	for _, d := range defs {
		require.Contains(t, d.(map[string]any)["properties"], "Value")
	}
	type shared struct {
		A *namedZone `json:"a"`
		B *namedZone `json:"b"`
	}
	s := compile(t, shared{})
	p := s["properties"].(map[string]any)
	require.Contains(t, p["a"].(map[string]any)["properties"], "IanaTimezoneStr")
	require.Contains(t, p["b"].(map[string]any)["properties"], "IanaTimezoneStr")
}

type customNamed struct {
	Value string `json:"value" mcp:"Value"`
}

func (*customNamed) UnmarshalJSON([]byte) error { return nil }
func TestCustomDecoderNotAdvertised(t *testing.T) {
	_, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[customNamed]()})
	require.ErrorContains(t, e, "require decoding authority")
	_, e = fieldname.Normalize(reflect.TypeFor[customNamed](), map[string]any{"Value": "x"})
	require.ErrorContains(t, e, "require decoding authority")
}
func TestExistingNameAndAliases(t *testing.T) {
	type in struct {
		Value string `json:"value" mcp:"name=Public,aliases=old|older"`
	}
	p := compile(t, in{})
	require.Contains(t, p["properties"], "Public")
	v, e := fieldname.Normalize(reflect.TypeFor[in](), map[string]any{"old": "x"})
	require.NoError(t, e)
	require.Equal(t, map[string]any{"value": "x"}, v)
	_, e = fieldname.Normalize(reflect.TypeFor[in](), map[string]any{"old": "x", "Public": "x"})
	require.Error(t, e)
}
func TestCustomLegacyAndAliasOnlyDefault(t *testing.T) {
	p, _ := schema.StructToProperties(reflect.TypeFor[customNamed]())
	require.Contains(t, p, "value")
	require.NotContains(t, p, "Value")
	var s schema.ToolInputSchema
	require.ErrorContains(t, s.Load(customNamed{}), "require decoding authority")
	type aliasOnly struct {
		Value string `json:"value" mcp:",aliases=old"`
	}
	p2 := compile(t, aliasOnly{})
	require.Contains(t, p2["properties"], "value")
}
func TestAtomicLegacyAliasCollisionAndSwap(t *testing.T) {
	type collision struct {
		A string `json:"a" mcp:"same"`
		B string `json:"b" mcp:"same"`
		C string `json:"c" mcp:"C"`
	}
	p, req := schema.StructToProperties(reflect.TypeFor[collision]())
	require.Contains(t, p, "a")
	require.Contains(t, p, "b")
	require.Contains(t, p, "c")
	require.NotContains(t, p, "C")
	require.ElementsMatch(t, []string{"a", "b", "c"}, req)
	type swap struct {
		A string `json:"a" mcp:"b"`
		B int    `json:"b" mcp:"a"`
	}
	p, _ = schema.StructToProperties(reflect.TypeFor[swap]())
	require.Equal(t, "integer", p["b"]["type"])
	require.Equal(t, "string", p["a"]["type"])
}
func TestReviewPromotedMapAndSpellingCollisions(t *testing.T) {
	type Embedded struct {
		Value string `json:"value" mcp:"Value"`
	}
	type input struct {
		Embedded
		ByName map[string]Embedded `json:"byName"`
	}
	raw := map[string]any{"Value": "promoted", "byName": map[string]any{"one": map[string]any{"Value": "mapped"}}}
	n, e := fieldname.Normalize(reflect.TypeFor[input](), raw)
	require.NoError(t, e)
	b, e := json.Marshal(n)
	require.NoError(t, e)
	var got input
	require.NoError(t, json.Unmarshal(b, &got))
	require.Equal(t, "promoted", got.Value)
	require.Equal(t, "mapped", got.ByName["one"].Value)
	require.Contains(t, raw, "Value")
	type ambiguous struct {
		A string `json:"a" mcp:",aliases=b"`
		B string `json:"b"`
	}
	_, e = fieldname.Normalize(reflect.TypeFor[ambiguous](), map[string]any{"b": "x"})
	require.ErrorContains(t, e, "conflicts")
	_, e = (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[ambiguous]()})
	require.ErrorContains(t, e, "conflicts")
	type child struct {
		A string `json:"a" mcp:"same"`
		B string `json:"b" mcp:"same"`
	}
	type parent struct {
		Child *child `json:"child"`
	}
	var s schema.ToolInputSchema
	require.ErrorContains(t, s.Load(parent{}), "conflicts")
}
func TestUnrelatedOptionsRemainIgnored(t *testing.T) {
	type ordinary struct {
		Value string `json:"value" mcp:",unrelated=option"`
	}
	p := compile(t, ordinary{})
	require.Contains(t, p["properties"], "value")
	var s schema.ToolInputSchema
	require.NoError(t, s.Load(ordinary{}))
}

type customParent struct {
	Child namedZone `json:"child"`
}

func (*customParent) UnmarshalJSON([]byte) error { return nil }

type customRows []namedZone

func (*customRows) UnmarshalJSON([]byte) error { return nil }
func TestWholeCustomRepresentationBoundary(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[customParent](), reflect.TypeFor[customRows]()} {
		_, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ})
		require.ErrorContains(t, e, "require decoding authority")
		_, e = fieldname.Normalize(typ, map[string]any{"child": map[string]any{"IanaTimezoneStr": "UTC"}})
		require.ErrorContains(t, e, "require decoding authority")
	}
	p, _ := schema.StructToProperties(reflect.TypeFor[customParent]())
	require.Contains(t, p["child"]["properties"], "ianaTimezoneStr")
	require.NotContains(t, p["child"]["properties"], "IanaTimezoneStr")
}
func TestNilTypesAndHiddenValidation(t *testing.T) {
	_, e := fieldname.Fields(nil)
	require.Error(t, e)
	_, e = fieldname.Normalize(nil, map[string]any{})
	require.Error(t, e)
	require.False(t, fieldname.CustomJSON(nil))
	type bad struct {
		A string `json:"a" mcp:"same"`
		B string `json:"b" mcp:"same"`
	}
	type in struct {
		Hidden bad    `json:"hidden"`
		Value  string `json:"value"`
	}
	var s schema.ToolInputSchema
	require.NoError(t, s.Load(in{}, schema.WithSkipFieldHook(func(f reflect.StructField) bool { return f.Name == "Hidden" })))
}
func TestPromotedSameLeafNamesHaveDistinctOwners(t *testing.T) {
	type Left struct {
		Value string `json:"left" mcp:"same"`
	}
	type Right struct {
		Value string `json:"right" mcp:"same"`
	}
	type input struct {
		Left
		Right
	}
	_, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[input]()})
	require.ErrorContains(t, e, "conflicts")
	_, e = fieldname.Normalize(reflect.TypeFor[input](), map[string]any{"same": "x"})
	require.ErrorContains(t, e, "conflicts")
}
func TestRecursiveContainerNameCheckTerminates(t *testing.T) {
	type recursiveMap map[string]recursiveMap
	type recursiveSlice []recursiveSlice
	var legacy schema.ToolInputSchema
	require.NoError(t, legacy.Load(struct{ Data recursiveMap }{}))
	require.NoError(t, legacy.Load(struct{ Data recursiveSlice }{}))
	for _, typ := range []reflect.Type{reflect.TypeFor[recursiveMap](), reflect.TypeFor[recursiveSlice]()} {
		yes, e := fieldname.HasNames(typ)
		require.NoError(t, e)
		require.False(t, yes)
	}
}

type hiddenCustomProof struct {
	Secret string `json:"secret" internal:"true" mcp:"SecretAlias"`
	Value  string `json:"value"`
}

func (p *hiddenCustomProof) UnmarshalJSON(b []byte) error {
	type plain hiddenCustomProof
	return json.Unmarshal(b, (*plain)(p))
}
func TestSharedVisibilityCustomBoundary(t *testing.T) {
	policy := fieldname.Visibility{ExcludeInternal: true}
	r, e := (jsonschema.Reflector{ExcludeInternal: policy.ExcludeInternal, ExcludeField: policy.ExcludeField}).Compile(jsonschema.Request{Type: reflect.TypeFor[hiddenCustomProof]()})
	require.NoError(t, e)
	require.NotContains(t, r["properties"], "SecretAlias")
	yes, e := fieldname.HasNames(reflect.TypeFor[hiddenCustomProof](), policy)
	require.NoError(t, e)
	require.False(t, yes)
	n, e := fieldname.Normalize(reflect.TypeFor[hiddenCustomProof](), map[string]any{"value": "x"}, policy)
	require.NoError(t, e)
	b, e := json.Marshal(n)
	require.NoError(t, e)
	var v hiddenCustomProof
	require.NoError(t, json.Unmarshal(b, &v))
	require.Equal(t, "x", v.Value)
	type notHidden struct {
		Value string `json:"value" internal:"false" mcp:"Value"`
	}
	r, e = (jsonschema.Reflector{ExcludeInternal: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[notHidden]()})
	require.NoError(t, e)
	require.Contains(t, r["properties"], "Value")
	n, e = fieldname.Normalize(reflect.TypeFor[notHidden](), map[string]any{"Value": "x"}, policy)
	require.NoError(t, e)
	require.Equal(t, map[string]any{"value": "x"}, n)
	type HiddenOwner struct {
		Value string `json:"value" mcp:"Public"`
	}
	type promoted struct {
		HiddenOwner `internal:"true"`
		Visible     string `json:"visible"`
	}
	yes, e = fieldname.HasNames(reflect.TypeFor[promoted](), policy)
	require.NoError(t, e)
	require.False(t, yes)
}
