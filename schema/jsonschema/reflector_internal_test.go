package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/viant/mcp-protocol/schema/jsonschema"
)

type InternalFields struct {
	Public      string                 `json:"public"`
	Secret      string                 `json:"secret" internal:"true"`
	Pointer     *string                `json:"pointer" internal:"true"`
	Structured  *InternalFields        `json:"structured" internal:"true"`
	NotInternal string                 `json:"notInternal" internal:"false"`
	Literal     string                 `json:"literal" format:"-"`
	Has         *struct{ Public bool } `setMarker:"true"`
	Hidden      string                 `json:"-"`
}

func TestReflectorInternalModeCompatibility(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(map[bool]string{false: "default", true: "publicInput"}[exclude], func(t *testing.T) {
			annotated := map[string]bool{}
			result, err := (jsonschema.Reflector{ExcludeInternal: exclude, Annotate: func(path string, _ reflect.StructField) (string, any) { annotated[path] = true; return path, nil }}).Compile(jsonschema.Request{Type: reflect.TypeFor[InternalFields]()})
			if err != nil {
				t.Fatal(err)
			}
			properties := result["properties"].(map[string]any)
			for _, name := range []string{"public", "notInternal", "literal"} {
				if properties[name] == nil {
					t.Errorf("missing public property %s", name)
				}
			}
			for _, name := range []string{"secret", "pointer", "structured"} {
				if (properties[name] != nil) == exclude {
					t.Errorf("property %s, exclude=%v: %v", name, exclude, properties)
				}
			}
			for _, name := range []string{"Has", "Hidden"} {
				if properties[name] != nil {
					t.Errorf("hidden property %s", name)
				}
			}
			if exclude {
				for _, name := range []string{"Secret", "Pointer", "Structured"} {
					if annotated[name] {
						t.Errorf("internal annotation called for %s", name)
					}
				}
			}
		})
	}
}

type InternalEmbedded struct {
	Promoted string `json:"promoted"`
}
type PublicEmbedded struct {
	Shared string `json:"shared"`
}
type InternalEmbedding struct {
	*InternalEmbedded `internal:"true"`
	PublicEmbedded
	Shared  string            `json:"shared" internal:"true"`
	Tagged  *InternalEmbedded `json:"tagged" internal:"true"`
	Visible string            `json:"visible"`
}

func TestReflectorInternalEmbeddingAndDominance(t *testing.T) {
	result, err := (jsonschema.Reflector{ExcludeInternal: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[InternalEmbedding]()})
	if err != nil {
		t.Fatal(err)
	}
	properties := result["properties"].(map[string]any)
	if len(properties) != 1 || properties["visible"] == nil {
		t.Fatalf("embedding or shadowed public field leaked: %v", properties)
	}
	ordinary, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[InternalEmbedding]()})
	if err != nil {
		t.Fatal(err)
	}
	if len(ordinary["properties"].(map[string]any)) != 4 {
		t.Fatalf("default projection changed: %v", ordinary)
	}
}

type InternalRecursive struct {
	Value  string                     `json:"value"`
	Secret string                     `json:"secret" internal:"true"`
	Next   *InternalRecursive         `json:"next"`
	List   []InternalFields           `json:"list"`
	Map    map[string]*InternalFields `json:"map"`
}

func TestReflectorInternalNestedAndRecursiveSchemas(t *testing.T) {
	result, err := (jsonschema.Reflector{ExcludeInternal: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[InternalRecursive](), Property: "payload/~"})
	if err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(value any) {
		switch node := value.(type) {
		case map[string]any:
			if properties, ok := node["properties"].(map[string]any); ok {
				for _, name := range []string{"secret", "pointer", "structured"} {
					if properties[name] != nil {
						t.Errorf("nested internal property %s", name)
					}
				}
			}
			for _, child := range node {
				walk(child)
			}
		case []any:
			for _, child := range node {
				walk(child)
			}
		}
	}
	walk(result)
	if len(result["$defs"].(map[string]any)) == 0 {
		t.Fatal("recursive definitions missing")
	}
}

type OpaqueInternal struct{}

func (OpaqueInternal) MarshalJSON() ([]byte, error) { return json.Marshal("opaque") }

type OpaqueContainer struct {
	Public string         `json:"public"`
	Opaque OpaqueInternal `json:"opaque" internal:"true"`
}

func TestReflectorInternalExcludedBeforeCustomAuthority(t *testing.T) {
	_, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[OpaqueContainer]()})
	if err == nil {
		t.Fatal("ordinary custom authority error changed")
	}
	result, err := (jsonschema.Reflector{ExcludeInternal: true}).Compile(jsonschema.Request{Type: reflect.TypeFor[OpaqueContainer]()})
	if err != nil {
		t.Fatal(err)
	}
	if len(result["properties"].(map[string]any)) != 1 {
		t.Fatal(result)
	}
}
