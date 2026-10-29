package schema_test

import (
	"encoding/json"
	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-protocol/schema/jsonschema"
	"reflect"
	"testing"
)

type stringWireChoice int

func (stringWireChoice) MarshalJSON() ([]byte, error) { return []byte(`"three"`), nil }

type textMarshaledNumber int64

func (textMarshaledNumber) MarshalText() ([]byte, error) { return []byte("x"), nil }
func TestChoiceCustomTextScalarCannotChangeEnumWireType(t *testing.T) {
	typ := reflect.TypeOf(struct {
		Number textMarshaledNumber `choice:"3" choice:"1" choice:"2"`
	}{})
	legacy, _ := schema.StructToProperties(typ)
	b, _ := json.Marshal(legacy["Number"]["enum"])
	if string(b) != "[3,1,2]" {
		t.Fatal(string(b))
	}
	ordinary, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ})
	if e != nil {
		t.Fatal(e)
	}
	plain := ordinary["properties"].(map[string]any)["Number"].(map[string]any)
	b, _ = json.Marshal(plain["enum"])
	if string(b) != "[3,1,2]" {
		t.Fatal(string(b))
	}
	modern, e := (jsonschema.Reflector{WireTypes: map[reflect.Type]reflect.Type{reflect.TypeFor[textMarshaledNumber](): reflect.TypeFor[int64]()}}).Compile(jsonschema.Request{Type: typ})
	if e != nil {
		t.Fatal(e)
	}
	p := modern["properties"].(map[string]any)["Number"].(map[string]any)
	b, _ = json.Marshal(p["enum"])
	if string(b) != "[3,1,2]" {
		t.Fatal(string(b))
	}
}
func TestChoiceRejectsCustomAndQuotedWireConflicts(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(struct {
		Mode stringWireChoice `wireSchema:"{\"type\":\"string\"}" choice:"3"`
	}{}), reflect.TypeOf(struct {
		Mode int `json:",string" choice:"3"`
	}{})} {
		if _, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ}); e == nil {
			t.Fatal("contradictory choice wire schema accepted")
		}
	}
}
func TestChoicePointerAndDeclaredWireType(t *testing.T) {
	typ := reflect.TypeOf(struct {
		Mode *stringWireChoice `choice:"z" choice:"a" choice:"m"`
	}{})
	r, e := (jsonschema.Reflector{WireTypes: map[reflect.Type]reflect.Type{reflect.TypeFor[stringWireChoice](): reflect.TypeFor[string]()}}).Compile(jsonschema.Request{Type: typ})
	if e != nil {
		t.Fatal(e)
	}
	p := r["properties"].(map[string]any)["Mode"].(map[string]any)
	b, _ := json.Marshal(p["enum"])
	if string(b) != `["z","a","m",null]` {
		t.Fatal(string(b))
	}
}

func TestChoiceNullablePointerBothAdapters(t *testing.T) {
	typ := reflect.TypeOf(struct {
		Mode *int  `json:"mode,omitempty" choice:"3" choice:"1" choice:"2" default:"3"`
		Flag *bool `json:"flag,omitempty" choice:"false" choice:"true" default:"false"`
	}{})
	legacy, _ := schema.StructToProperties(typ)
	modern, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ})
	if e != nil {
		t.Fatal(e)
	}
	for _, props := range []map[string]map[string]any{legacy, func() map[string]map[string]any {
		p := map[string]map[string]any{}
		for k, v := range modern["properties"].(map[string]any) {
			p[k] = v.(map[string]any)
		}
		return p
	}()} {
		for key, want := range map[string]string{"mode": "[3,1,2,null]", "flag": "[false,true,null]"} {
			b, _ := json.Marshal(props[key]["enum"])
			if string(b) != want {
				t.Fatal(string(b))
			}
		}
		if _, found := props["mode"]["default"]; found {
			t.Fatal("choice metadata introduced a new pointer default policy")
		}
	}
}

type orderedChoiceInput struct {
	Number      int    `json:"number" choice:"3" choice:"1" choice:"2"`
	Text        string `json:"text" choice:"z" choice:"a" choice:"m"`
	NumericText string `json:"numericText" choice:"3" choice:"1" choice:"2"`
	Flag        bool   `json:"flag" choice:"false" choice:"true"`
	Big         int64  `json:"big" choice:"9007199254740993" choice:"9007199254740994"`
	Ordinary    string `json:"ordinary"`
}

func TestChoiceAdaptersPreserveOrderedWireTypes(t *testing.T) {
	legacy, _ := schema.StructToProperties(reflect.TypeFor[orderedChoiceInput]())
	modern, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: reflect.TypeFor[orderedChoiceInput]()})
	if e != nil {
		t.Fatal(e)
	}
	for _, props := range []map[string]map[string]interface{}{legacy, func() map[string]map[string]interface{} {
		r := map[string]map[string]interface{}{}
		for k, v := range modern["properties"].(map[string]any) {
			r[k] = v.(map[string]any)
		}
		return r
	}()} {
		for k, want := range map[string]string{"number": "[3,1,2]", "text": `["z","a","m"]`, "numericText": `["3","1","2"]`, "flag": "[false,true]", "big": "[9007199254740993,9007199254740994]"} {
			b, _ := json.Marshal(props[k]["enum"])
			if string(b) != want {
				t.Fatalf("%s enum=%s want=%s", k, b, want)
			}
		}
		if _, ok := props["ordinary"]["enum"]; ok {
			t.Fatal("ordinary metadata changed")
		}
	}
}
func TestChoiceAdaptersFailClosed(t *testing.T) {
	typ := reflect.TypeOf(struct {
		Mode int `choice:"1" choice:"1"`
	}{})
	legacy, _ := schema.StructToProperties(typ)
	if legacy["Mode"]["not"] == nil {
		t.Fatal("legacy invalid choice unconstrained")
	}
	if _, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: typ}); e == nil {
		t.Fatal("modern invalid choice accepted")
	}
}

func TestChoiceMetadataUnsupportedShapesAndMalformedSiblingTags(t *testing.T) {
	for _, tag := range []reflect.StructTag{`choice:"3" description:bad`, `choice:"3"`} {
		typ := reflect.TypeFor[int]()
		if string(tag) == `choice:"3"` {
			typ = reflect.TypeFor[[]int]()
		}
		shape := reflect.StructOf([]reflect.StructField{{Name: "Mode", Type: typ, Tag: tag}})
		legacy, _ := schema.StructToProperties(shape)
		if legacy["Mode"]["not"] == nil {
			t.Fatal("invalid metadata became unconstrained")
		}
		if _, e := (jsonschema.Reflector{}).Compile(jsonschema.Request{Type: shape}); e == nil {
			t.Fatal("invalid metadata accepted")
		}
	}
}
