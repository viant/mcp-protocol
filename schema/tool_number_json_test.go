package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestToolCallsUseLegacyJSONNumbers(t *testing.T) {
	const values = `{"small":7,"large":9007199254740993,"maxUint64":18446744073709551615,"negative":-9007199254740993,"fraction":0.12345678901234567890123456789,"exponent":1.234567890123456789e100,"tiny":1e-400,"nested":[9007199254740993,null],"null":null,"text":"7","bool":true}`
	params := `{"_meta":null,"name":"legacy","arguments":` + values + `,"inputResponses":` + values + `,"unknown":7}`
	result := `{"content":[],"resultType":"complete","structuredContent":` + values + `,"inputRequests":` + values + `,"isError":true}`
	var want interface{}
	if err := json.Unmarshal([]byte(values), &want); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, raw string
		target    func() json.Unmarshaler
		payload   func(json.Unmarshaler) interface{}
	}{
		{"params", params, func() json.Unmarshaler { return &CallToolRequestParams{} }, func(value json.Unmarshaler) interface{} { return value.(*CallToolRequestParams).Arguments }},
		{"request", `{"jsonrpc":"2.0","id":9007199254740993,"method":"tools/call","params":` + params + `}`, func() json.Unmarshaler { return &CallToolRequest{} }, func(value json.Unmarshaler) interface{} { return value.(*CallToolRequest).Params.Arguments }},
		{"result", result, func() json.Unmarshaler { return &CallToolResult{} }, func(value json.Unmarshaler) interface{} { return value.(*CallToolResult).StructuredContent }},
		{"response result", result, func() json.Unmarshaler { return &CallToolResultResponseResult{} }, func(value json.Unmarshaler) interface{} {
			return value.(*CallToolResultResponseResult).StructuredContent
		}},
		{"response", `{"jsonrpc":"2.0","id":9007199254740993,"result":` + result + `}`, func() json.Unmarshaler { return &CallToolResultResponse{} }, func(value json.Unmarshaler) interface{} {
			return value.(*CallToolResultResponse).Result.StructuredContent
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := tc.target()
			if err := json.Unmarshal([]byte(tc.raw), target); err != nil {
				t.Fatal(err)
			}
			got := tc.payload(target)
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("legacy decoding changed: %#v", got)
			}
			object := got.(map[string]interface{})
			if object["small"].(float64) != 7 || object["large"].(float64) != 9007199254740992 || object["tiny"].(float64) != 0 {
				t.Fatal("legacy extraction/rounding changed")
			}
			switch value := target.(type) {
			case *CallToolRequest:
				if value.Id != RequestId(9007199254740993) {
					t.Fatal("typed ID rounded")
				}
			case *CallToolResultResponse:
				if value.Id != RequestId(9007199254740993) {
					t.Fatal("typed ID rounded")
				}
			}
			for _, suffix := range []string{" null", " {}", " junk"} {
				if err := tc.target().UnmarshalJSON([]byte(tc.raw + suffix)); err == nil {
					t.Fatalf("trailing data accepted: %q", suffix)
				}
			}
		})
	}
}

func TestToolCallLegacyOverflowAndNull(t *testing.T) {
	for _, raw := range []string{`{"_meta":null,"name":"legacy","arguments":{"n":1e400}}`, `{"_meta":null,"name":"legacy","arguments":{"n":01}}`} {
		if err := json.Unmarshal([]byte(raw), &CallToolRequestParams{}); err == nil {
			t.Fatalf("invalid/overflow accepted: %s", raw)
		}
	}
	for _, suffix := range []string{"", `,"arguments":null`, `,"arguments":{}`} {
		var params CallToolRequestParams
		if err := json.Unmarshal([]byte(`{"_meta":null,"name":"legacy"`+suffix+`}`), &params); err != nil {
			t.Fatal(err)
		}
		if len(params.Arguments) != 0 {
			t.Fatal("null/missing changed")
		}
	}
}
