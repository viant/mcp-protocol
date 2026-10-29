package schema

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const exactToolNumbers = `{"small":7,"large":9007199254740993,"maxUint64":18446744073709551615,"negative":-9007199254740993,"fraction":0.12345678901234567890123456789,"exponent":1.234567890123456789e+100,"zero":-0,"null":null,"nested":[9007199254740993,null,[18446744073709551615,1e-400]],"text":"9007199254740993","bool":true}`

func TestToolCallNumericJSONRoundTrip(t *testing.T) {
	params := `{"_meta":{"io.modelcontextprotocol/protocolVersion":"2025-06-18"},"name":"exact","arguments":{"request":{"values":` + exactToolNumbers + `}},"inputResponses":` + exactToolNumbers + `}`
	result := `{"content":[],"resultType":"complete","structuredContent":` + exactToolNumbers + `}`
	inputRequired := `{"resultType":"input_required","requestState":"next","inputRequests":` + exactToolNumbers + `,"structuredContent":` + exactToolNumbers + `}`
	for _, tc := range []struct {
		name, raw string
		target    func() json.Unmarshaler
		paths     [][]string
	}{
		{"params", params, func() json.Unmarshaler { return &CallToolRequestParams{} }, [][]string{{"arguments", "request", "values"}, {"inputResponses"}}},
		{"request", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + params + `}`, func() json.Unmarshaler { return &CallToolRequest{} }, [][]string{{"params", "arguments", "request", "values"}, {"params", "inputResponses"}}},
		{"result", result, func() json.Unmarshaler { return &CallToolResult{} }, [][]string{{"structuredContent"}}},
		{"input required", inputRequired, func() json.Unmarshaler { return &CallToolResult{} }, [][]string{{"structuredContent"}, {"inputRequests"}}},
		{"response result", result, func() json.Unmarshaler { return &CallToolResultResponseResult{} }, [][]string{{"structuredContent"}}},
		{"response input required", inputRequired, func() json.Unmarshaler { return &CallToolResultResponseResult{} }, [][]string{{"structuredContent"}, {"inputRequests"}}},
		{"response", `{"jsonrpc":"2.0","id":1,"result":` + result + `}`, func() json.Unmarshaler { return &CallToolResultResponse{} }, [][]string{{"result", "structuredContent"}}},
	} {
		for shape, input := range map[string]string{"finite": tc.raw, "extreme": strings.ReplaceAll(tc.raw, exactToolNumbers, strings.TrimSuffix(exactToolNumbers, "}")+`,"overflow":1e400}`)} {
			t.Run(tc.name+"/"+shape, func(t *testing.T) {
				target := tc.target()
				if err := json.Unmarshal([]byte(input), target); err != nil {
					t.Fatal(err)
				}
				// Typed envelope IDs keep their generated integer type; only interface
				// payload leaves preserve decimal precision, preserving request correlation.
				switch value := target.(type) {
				case *CallToolRequest:
					if value.Id != RequestId(1) {
						t.Fatalf("request ID changed: %#v", value.Id)
					}
				case *CallToolResultResponse:
					if value.Id != RequestId(1) {
						t.Fatalf("response ID changed: %#v", value.Id)
					}
				}
				wire, err := json.Marshal(target)
				if err != nil {
					t.Fatal(err)
				}
				for _, path := range tc.paths {
					raw := rawToolField(t, wire, path...)
					if !reflect.DeepEqual(toolNumberOracle(t, rawToolField(t, []byte(input), path...)), toolNumberOracle(t, raw)) {
						t.Fatalf("numeric tokens changed at %v: %s", path, wire)
					}
					if string(rawToolField(t, raw, "large")) != "9007199254740993" || string(rawToolField(t, raw, "maxUint64")) != "18446744073709551615" {
						t.Fatalf("integers rounded or became strings: %s", wire)
					}
				}
				// Direct generated-method calls must retain json.Unmarshal's whole-value validation.
				for _, suffix := range []string{" null", " {}", " junk"} {
					if err := tc.target().UnmarshalJSON([]byte(tc.raw + suffix)); err == nil {
						t.Fatalf("trailing data accepted: %q", suffix)
					}
				}
			})
		}
	}
}

func TestToolCallGoNumericLeavesRetainLegacyFloat64(t *testing.T) {
	var params CallToolRequestParams
	if err := json.Unmarshal([]byte(`{"_meta":{"io.modelcontextprotocol/protocolVersion":"2025-06-18"},"name":"exact","arguments":{"n":7}}`), &params); err != nil {
		t.Fatal(err)
	}
	if params.Arguments["n"] != float64(7) {
		t.Fatalf("Go argument value = %T (%v)", params.Arguments["n"], params.Arguments["n"])
	}
	var result CallToolResult
	if err := json.Unmarshal([]byte(`{"content":[],"resultType":"complete","structuredContent":{"n":7}}`), &result); err != nil {
		t.Fatal(err)
	}
	object := result.StructuredContent.(map[string]interface{})
	if object["n"] != float64(7) {
		t.Fatalf("Go result value = %T (%v)", object["n"], object["n"])
	}
}

func rawToolField(t *testing.T, raw []byte, path ...string) []byte {
	t.Helper()
	for _, name := range path {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		raw = fields[name]
		if len(raw) == 0 {
			t.Fatalf("missing field %s", name)
		}
	}
	return raw
}

func toolNumberOracle(t *testing.T, raw []byte) interface{} {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestToolCallLegacyNumbersAndUnknownFields(t *testing.T) {
	var params CallToolRequestParams
	if err := json.Unmarshal([]byte(`{"_meta":null,"name":"legacy","unknown":9007199254740993,"arguments":{"small":7,"fraction":0.1,"exponent":1e3,"large":9007199254740993,"nested":[null,1.25]}}`), &params); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]float64{"small": 7, "fraction": 0.1, "exponent": 1000} {
		if got := params.Arguments[key].(float64); got != want {
			t.Fatalf("%s = %v", key, got)
		}
	}
	if params.Arguments["large"] != json.Number("9007199254740993") {
		t.Fatal("large rounded")
	}
	var result CallToolResult
	if err := json.Unmarshal([]byte(`{"content":[{"type":"text","text":"error","extension":7}],"resultType":"complete","isError":true,"structuredContent":{"small":7,"fraction":0.1,"exponent":1e3,"large":9007199254740993,"null":null}}`), &result); err != nil {
		t.Fatal(err)
	}
	object := result.StructuredContent.(map[string]interface{})
	if object["small"].(float64) != 7 || object["fraction"].(float64) != 0.1 || object["exponent"].(float64) != 1000 || object["large"] != json.Number("9007199254740993") || object["null"] != nil || result.IsError == nil || !*result.IsError {
		t.Fatalf("result changed: %#v", result)
	}
}

func TestToolCallIntegralValuesOutsideSafeRangeRemainJSONNumber(t *testing.T) {
	var params CallToolRequestParams
	if err := json.Unmarshal([]byte(`{"_meta":null,"name":"exact","arguments":{"safe":9007199254740991,"safeNegative":-9007199254740991,"boundary":9007199254740992,"negativeBoundary":-9007199254740992,"roundedGoInteger":1000000000000000100,"equivalentExponent":1.0000000000000001e18,"maxUint64":18446744073709551615}}`), &params); err != nil {
		t.Fatal(err)
	}
	if params.Arguments["safe"].(float64) != 9007199254740991 || params.Arguments["safeNegative"].(float64) != -9007199254740991 {
		t.Fatal("safe legacy boundaries changed")
	}
	for key, token := range map[string]string{"boundary": "9007199254740992", "negativeBoundary": "-9007199254740992", "roundedGoInteger": "1000000000000000100", "equivalentExponent": "1.0000000000000001e18", "maxUint64": "18446744073709551615"} {
		if params.Arguments[key] != json.Number(token) {
			t.Fatalf("%s changed: %T %v", key, params.Arguments[key], params.Arguments[key])
		}
	}
}
