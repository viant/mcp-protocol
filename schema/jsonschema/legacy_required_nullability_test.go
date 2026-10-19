package jsonschema

import (
	"reflect"
	"testing"

	legacy "github.com/viant/mcp-protocol/schema"
)

type flightID struct {
	IDs []int
}
type flightDelete struct {
	ID *flightID
}

type requiredControls struct {
	Plain    string  `json:"plain"`
	Pointer  *string `json:"pointer"`
	Omitted  string  `json:"omitted,omitempty"`
	Optional string  `json:"optional" optional`
	NotReq   string  `json:"notReq" required:"false"`
	Forced   *string `json:"forced,omitempty" required:"true"`
}

type recursiveRequired struct {
	IDs  []int              `json:"ids"`
	Next *recursiveRequired `json:"next,omitempty"`
}

type requiredEmbed struct {
	ID int `json:"id" mcp:"identifier"`
}
type aliasesAndVisibility struct {
	requiredEmbed `json:",inline"`
	Hidden        string `json:"hidden" internal:"true"`
}

type optionalRequiredEmbed struct {
	*requiredEmbed `json:",inline"`
}

type itemNullability struct {
	Items   []*int           `json:"items"`
	Objects []nullableObject `json:"objects"`
	Field   *int             `json:"field"`
	Lookup  map[string]*int  `json:"lookup"`
}
type nullableObject struct {
	Field *int `json:"field"`
}

func TestLegacyRequiredFieldsAndNestedReference(t *testing.T) {
	got := mustCompileSchema(t, reflect.TypeFor[flightDelete]())
	id := got["properties"].(map[string]any)["ID"].(map[string]any)
	flight := id
	if union, ok := id["type"].([]string); ok && len(union) > 0 {
		flight = map[string]any{}
		for key, value := range id {
			flight[key] = value
		}
		flight["type"] = "object"
	}
	if !reflect.DeepEqual(flight["required"], []string{"IDs"}) {
		t.Fatalf("FlightId required = %#v; want [IDs]", flight["required"])
	}
}

func TestLegacyRequiredTagControlsAndAliasedVisibility(t *testing.T) {
	got := mustCompileSchema(t, reflect.TypeFor[requiredControls]())
	_, legacyRequired := legacy.StructToProperties(reflect.TypeFor[requiredControls]())
	if !reflect.DeepEqual(got["required"], legacyRequired) {
		t.Fatalf("modern required = %#v; legacy required = %#v", got["required"], legacyRequired)
	}
	if !reflect.DeepEqual(got["required"], []string{"plain", "notReq", "forced"}) {
		t.Fatalf("required = %#v; want legacy parity [plain notReq forced]", got["required"])
	}
	aliased, err := (Reflector{ExcludeInternal: true}).Compile(Request{Type: reflect.TypeFor[aliasesAndVisibility]()})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(aliased["required"], []string{"identifier"}) {
		t.Fatalf("aliased required = %#v; want [identifier]", aliased["required"])
	}
	properties := aliased["properties"].(map[string]any)
	if _, ok := properties["hidden"]; ok {
		t.Fatalf("hidden property was exposed: %#v", properties)
	}
}

func TestLegacyRequirednessPropagatesThroughRequiredInlineOwnerOnly(t *testing.T) {
	required := mustCompileSchema(t, reflect.TypeFor[requiredEmbed]())
	if !reflect.DeepEqual(required["required"], []string{"identifier"}) {
		t.Fatalf("required inline owner list = %#v; want [identifier]", required["required"])
	}
	optional := mustCompileSchema(t, reflect.TypeFor[optionalRequiredEmbed]())
	if _, ok := optional["required"]; ok {
		t.Fatalf("optional inline owner leaked required members: %#v", optional["required"])
	}
}

func TestLegacyRecursiveObjectKeepsRequiredInDefinition(t *testing.T) {
	got := mustCompileSchema(t, reflect.TypeFor[recursiveRequired]())
	if !reflect.DeepEqual(got["required"], []string{"ids"}) {
		t.Fatalf("root required = %#v; want [ids]", got["required"])
	}
	defs := got["$defs"].(map[string]any)
	if len(defs) != 1 {
		t.Fatalf("definitions = %#v", defs)
	}
	for _, definition := range defs {
		if !reflect.DeepEqual(definition.(map[string]any)["required"], []string{"ids"}) {
			t.Fatalf("recursive definition required = %#v", definition.(map[string]any)["required"])
		}
	}
}

func TestLegacySuppressesNullabilityOnlyForSliceItems(t *testing.T) {
	got := mustCompileSchema(t, reflect.TypeFor[itemNullability]())
	properties := got["properties"].(map[string]any)
	items := properties["items"].(map[string]any)["items"].(map[string]any)
	if items["type"] != "integer" {
		t.Fatalf("slice item type = %#v; want integer without null", items["type"])
	}
	objectItem := properties["objects"].(map[string]any)["items"].(map[string]any)
	objectField := objectItem["properties"].(map[string]any)["field"].(map[string]any)
	if !reflect.DeepEqual(objectField["type"], []string{"integer", "null"}) {
		t.Fatalf("object member under slice type = %#v; want integer|null", objectField["type"])
	}
	field := properties["field"].(map[string]any)
	if !reflect.DeepEqual(field["type"], []string{"integer", "null"}) {
		t.Fatalf("pointer member type = %#v; want integer|null", field["type"])
	}
	lookup := properties["lookup"].(map[string]any)["additionalProperties"].(map[string]any)
	if !reflect.DeepEqual(lookup["type"], []string{"integer", "null"}) {
		t.Fatalf("map value type = %#v; want integer|null", lookup["type"])
	}
}

func mustCompileSchema(t *testing.T, typ reflect.Type) map[string]any {
	t.Helper()
	got, err := (Reflector{}).Compile(Request{Type: typ})
	if err != nil {
		t.Fatal(err)
	}
	return got
}
