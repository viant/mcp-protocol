package schema

import (
	"reflect"
	"testing"
)

// Helper: get property as map (inner type is already map[string]interface{}).
func prop(schemaProps ToolInputSchemaProperties, key string) map[string]interface{} {
	v, ok := schemaProps[key]
	if !ok {
		return nil
	}
	return v
}

func containsAll(slice []string, want ...string) bool {
	set := map[string]bool{}
	for _, s := range slice {
		set[s] = true
	}
	for _, w := range want {
		if !set[w] {
			return false
		}
	}
	return true
}

// Test: anonymous embedded struct is flattened, requireds propagate when non-pointer.
func TestStructToProperties_EmbeddedAnonymous(t *testing.T) {
	type A struct {
		X string `json:"x"`
	}
	type B struct {
		A     // anonymous embed
		Y int `json:"y"`
	}

	props, req := StructToProperties(reflect.TypeOf(B{}))

	if prop(props, "x") == nil || prop(props, "y") == nil {
		t.Fatalf("expected flattened properties x and y, got: %#v", props)
	}
	if containsAll(req, "x", "y") == false {
		t.Fatalf("expected required to include x and y, got: %v", req)
	}
}

// Test: tagged inline embedding flattens fields.
func TestStructToProperties_EmbeddedInlineTag(t *testing.T) {
	type A struct {
		X string `json:"x"`
	}
	type B struct {
		A `json:",inline"`
		Z string `json:"z"`
	}

	props, req := StructToProperties(reflect.TypeOf(B{}))

	if prop(props, "x") == nil || prop(props, "z") == nil {
		t.Fatalf("expected flattened properties x and z, got: %#v", props)
	}
	if !containsAll(req, "x", "z") {
		t.Fatalf("expected required to include x and z, got: %v", req)
	}
}

// Test: pointer embedded struct flattens but does not propagate required.
func TestStructToProperties_EmbeddedPointerOptional(t *testing.T) {
	type A struct {
		X string `json:"x"`
	}
	type B struct {
		*A        // pointer embed
		Z  string `json:"z"`
	}
	props, req := StructToProperties(reflect.TypeOf(B{}))
	if prop(props, "x") == nil || prop(props, "z") == nil {
		t.Fatalf("expected flattened properties x and z, got: %#v", props)
	}
	// z is required (non-pointer), x should not be propagated as required via pointer embedding
	if containsAll(req, "z") == false || containsAll(req, "x") {
		t.Fatalf("expected required to include only z (not x), got: %v", req)
	}
}

// Test: enum, description, default, and format tags.
func TestStructToProperties_FieldMetadata(t *testing.T) {
	type S struct {
		State  string `json:"state,omitempty" choice:"open" choice:"closed" choice:"all" description:"issue state filter"`
		Domain string `json:"domain,omitempty" default:"github.com" description:"GitHub host"`
		URL    string `json:"url,omitempty" format:"uri"`
	}
	props, _ := StructToProperties(reflect.TypeOf(S{}))
	// Enum
	st := prop(props, "state")
	if st == nil {
		t.Fatalf("missing state prop")
	}
	if en, ok := st["enum"].([]string); !ok || len(en) != 3 {
		// Depending on conversions, the enum might be []interface{}
		if en2, ok2 := st["enum"].([]interface{}); !ok2 || len(en2) != 3 {
			t.Fatalf("expected enum with 3 values, got: %T %#v", st["enum"], st["enum"])
		}
	}
	if d := st["description"]; d != "issue state filter" {
		t.Fatalf("unexpected description: %v", d)
	}

	dm := prop(props, "domain")
	if dm["default"] != "github.com" {
		t.Fatalf("expected default github.com, got: %v", dm["default"])
	}

	u := prop(props, "url")
	if u["format"] != "uri" {
		t.Fatalf("expected format uri, got: %v", u["format"])
	}
}

// Test: required vs omitempty and explicit required tags.
func TestStructToProperties_RequiredLogic(t *testing.T) {
	type S struct {
		A string  `json:"a"`
		B *string `json:"b,omitempty"`
		C string  `json:"c" required:"false"`
		D string  `json:"d" required:"true"`
		E string  `json:"e" optional`
	}
	props, req := StructToProperties(reflect.TypeOf(S{}))
	if prop(props, "a") == nil || prop(props, "b") == nil {
		t.Fatalf("missing properties a or b")
	}
	// a and d should be required; c and e should not; b is pointer+omitempty so not required
	wantReq := []string{"a", "d"}
	if !containsAll(req, wantReq...) || containsAll(req, "b", "c", "e") {
		t.Fatalf("unexpected required set: %v", req)
	}
}

// Test: internal tag skips fields.
func TestStructToProperties_InternalSkip(t *testing.T) {
	type S struct {
		Visible string `json:"visible"`
		Hidden  string `json:"hidden" internal:"true"`
	}
	props, _ := StructToProperties(reflect.TypeOf(S{}))
	if prop(props, "visible") == nil {
		t.Fatalf("expected visible prop present")
	}
	if prop(props, "hidden") != nil {
		t.Fatalf("expected hidden prop to be skipped")
	}
}

// Test: pointer nullability not set for slice items; set for top-level pointer fields.
func TestBuildJSONSchema_NullablePointers(t *testing.T) {
	type S struct {
		P  *string   `json:"p"`
		PS []*string `json:"ps"`
	}
	props, _ := StructToProperties(reflect.TypeOf(S{}))

	p := prop(props, "p")
	if p == nil || p["nullable"] != true {
		t.Fatalf("expected p to be nullable=true, got: %#v", p)
	}

	ps := prop(props, "ps")
	if ps == nil {
		t.Fatalf("expected ps prop present")
	}
	items, _ := ps["items"].(map[string]interface{})
	if items == nil {
		t.Fatalf("expected ps.items present")
	}
	if _, has := items["nullable"]; has {
		t.Fatalf("did not expect nullable on slice items, got: %#v", items)
	}
}

// Test: inline merge does not overwrite existing parent properties with same key.
func TestStructToProperties_InlineNoOverwrite(t *testing.T) {
	type A struct {
		X string `json:"x"`
	}
	type B struct {
		A     // anonymous embed provides x
		X int `json:"x"` // parent also defines x
	}
	props, _ := StructToProperties(reflect.TypeOf(B{}))
	x := prop(props, "x")
	if x == nil {
		t.Fatalf("missing x prop")
	}
	// Expect type to reflect parent's field (int), not embedded (string)
	if x["type"] != "integer" {
		t.Fatalf("expected x type integer from parent, got: %v", x["type"])
	}
}
