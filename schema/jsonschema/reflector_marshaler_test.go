package jsonschema_test

import (
	"encoding/json"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"reflect"
	"testing"
	"time"
)

type ptrJSON struct{ Value string }

func (*ptrJSON) MarshalJSON() ([]byte, error) { panic("schema generation executed MarshalJSON") }

type ptrText struct{ Value string }

func (*ptrText) MarshalText() ([]byte, error) { panic("schema generation executed MarshalText") }

type valueJSON struct{ Value string }

func (valueJSON) MarshalJSON() ([]byte, error) { panic("schema generation executed MarshalJSON") }

type valueText struct{ Value string }

func (valueText) MarshalText() ([]byte, error) { panic("schema generation executed MarshalText") }

type ptrDecoder struct{ Value string }

func (*ptrDecoder) UnmarshalJSON([]byte) error { panic("schema generation executed UnmarshalJSON") }

type opaqueRaw json.RawMessage

func (opaqueRaw) MarshalJSON() ([]byte, error) { panic("schema generation executed MarshalJSON") }

func TestReflectorRejectsOpaqueMethodSets(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[ptrJSON](), reflect.TypeFor[*ptrJSON](), reflect.TypeFor[[]ptrJSON](), reflect.TypeFor[map[string]ptrJSON](), reflect.TypeFor[struct{ Child ptrJSON }](), reflect.TypeFor[ptrText](), reflect.TypeFor[*ptrText](), reflect.TypeFor[valueJSON](), reflect.TypeFor[valueText](), reflect.TypeFor[ptrDecoder](), reflect.TypeFor[opaqueRaw](), reflect.TypeFor[*opaqueRaw]()} {
		t.Run(typ.String(), func(t *testing.T) {
			if _, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ}); err == nil {
				t.Fatal("opaque source accepted", typ)
			}
		})
	}
	for _, typ := range []reflect.Type{reflect.TypeFor[struct{ Value string }](), reflect.TypeFor[time.Time](), reflect.TypeFor[*time.Time]()} {
		if _, err := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ}); err != nil {
			t.Fatal(typ, err)
		}
	}
}
