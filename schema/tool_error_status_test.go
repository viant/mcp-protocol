package schema

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	legacy "github.com/viant/mcp-protocol/schema/2025-11-25"
)

func TestToolErrorHTTPStatusIsOptionalTypedAndErrorOnly(t *testing.T) {
	failed := true
	success := false
	for _, value := range []any{400, int32(401), uint(403), int64(422), float64(500), json.Number("599"), json.Number("403.0")} {
		result := &CallToolResult{IsError: &failed, Meta: &ResultMetaObject{AdditionalProperties: map[string]any{ToolErrorHTTPStatusMetaKey: value}}}
		status, ok := ToolErrorHTTPStatus(result)
		require.True(t, ok)
		require.GreaterOrEqual(t, status, 400)
		require.LessOrEqual(t, status, 599)
	}
	for _, value := range []any{nil, "403", true, 399, 600, 403.25, math.NaN(), math.Inf(1), json.Number("invalid"), json.Number("401.5"), map[string]any{"status": 403}} {
		result := &CallToolResult{IsError: &failed, Meta: &ResultMetaObject{AdditionalProperties: map[string]any{ToolErrorHTTPStatusMetaKey: value}}}
		_, ok := ToolErrorHTTPStatus(result)
		require.False(t, ok, "%v", value)
	}
	for _, result := range []*CallToolResult{nil, {}, {IsError: &success, Meta: &ResultMetaObject{AdditionalProperties: map[string]any{ToolErrorHTTPStatusMetaKey: 403}}}} {
		_, ok := ToolErrorHTTPStatus(result)
		require.False(t, ok)
		require.False(t, SetToolErrorHTTPStatus(result, 403))
	}
	result := &CallToolResult{IsError: &failed}
	require.False(t, SetToolErrorHTTPStatus(result, 200))
	require.False(t, SetToolErrorHTTPStatus(result, 600))
	require.Nil(t, result.Meta)
}
func TestToolErrorHTTPStatusPreservesBodyMetadataAndLegacyWire(t *testing.T) {
	failed := true
	original := &ResultMetaObject{IoModelcontextprotocolServerInfo: &Implementation{Name: "fixture", Version: "1"}, AdditionalProperties: map[string]any{"example/flag": true}}
	result := &CallToolResult{IsError: &failed, Meta: original, Content: []CallToolResultContentElem{TextContent{Type: "text", Text: `{"status":"error","nullable":null}`}}, StructuredContent: map[string]any{"status": "error", "nullable": nil}}
	contentBefore, err := json.Marshal(result.Content)
	require.NoError(t, err)
	bodyBefore, err := json.Marshal(result.StructuredContent)
	require.NoError(t, err)
	require.True(t, SetToolErrorHTTPStatus(result, 403))
	require.NotSame(t, original, result.Meta)
	require.NotContains(t, original.AdditionalProperties.(map[string]any), ToolErrorHTTPStatusMetaKey)
	wire, err := json.Marshal(result)
	require.NoError(t, err)
	var decoded CallToolResult
	require.NoError(t, json.Unmarshal(wire, &decoded))
	status, ok := ToolErrorHTTPStatus(&decoded)
	require.True(t, ok)
	require.Equal(t, 403, status)
	contentAfter, err := json.Marshal(decoded.Content)
	require.NoError(t, err)
	require.JSONEq(t, string(contentBefore), string(contentAfter))
	bodyAfter, err := json.Marshal(decoded.StructuredContent)
	require.NoError(t, err)
	require.Equal(t, string(bodyBefore), string(bodyAfter))
	require.Equal(t, original.IoModelcontextprotocolServerInfo, decoded.Meta.IoModelcontextprotocolServerInfo)
	require.Equal(t, true, decoded.Meta.AdditionalProperties.(map[string]any)["example/flag"])
	var old legacy.CallToolResult
	require.NoError(t, json.Unmarshal(wire, &old))
	require.Equal(t, float64(403), old.Meta[ToolErrorHTTPStatusMetaKey])
	require.JSONEq(t, string(bodyBefore), string(mustStatusJSON(t, old.StructuredContent)))
}
func mustStatusJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return raw
}

func TestToolErrorHTTPStatusChecksExactJSONNumberBoundaries(t *testing.T) {
	failed := true
	for _, test := range []struct {
		value string
		want  int
		valid bool
	}{
		{"403.00000000000000001", 0, false},
		{"402.99999999999999999", 0, false},
		{"599.00000000000000001", 0, false},
		{"403e0", 403, true},
		{"4.03e2", 403, true},
		{"403.00000000000000000", 403, true},
		{"5.99e2", 599, true},
		{"NaN", 0, false}, {"Inf", 0, false}, {"-Inf", 0, false},
		{"403/1", 0, false}, {`"403"`, 0, false},
	} {
		t.Run(test.value, func(t *testing.T) {
			result := &CallToolResult{IsError: &failed, Meta: &ResultMetaObject{AdditionalProperties: map[string]any{ToolErrorHTTPStatusMetaKey: json.Number(test.value)}}}
			status, valid := ToolErrorHTTPStatus(result)
			require.Equal(t, test.valid, valid)
			require.Equal(t, test.want, status)
		})
	}
}

func TestToolErrorHTTPStatusDoesNotRoundFractionalWireMetadata(t *testing.T) {
	for _, test := range []struct {
		number string
		want   int
		valid  bool
	}{
		{"403", 403, true}, {"403e0", 403, true}, {"4.03e2", 403, true},
		{"403.00000000000000001", 0, false}, {"402.99999999999999999", 0, false},
		{`"403"`, 0, false}, {"403.25", 0, false},
	} {
		t.Run(test.number, func(t *testing.T) {
			raw := []byte(`{"content":[],"isError":true,"resultType":"complete","_meta":{"io.viant/http-status":` + test.number + `,"example/number":7.25,"example/rounded":403.00000000000000001,"io.modelcontextprotocol/serverInfo":{"name":"fixture","version":"1"}}}`)
			var result CallToolResult
			require.NoError(t, json.Unmarshal(raw, &result))
			status, valid := ToolErrorHTTPStatus(&result)
			require.Equal(t, test.valid, valid)
			require.Equal(t, test.want, status)
			extras := result.Meta.AdditionalProperties.(map[string]any)
			require.IsType(t, float64(0), extras["example/number"], "unrelated metadata retains its existing numeric representation")
			require.Equal(t, float64(7.25), extras["example/number"])
			require.Equal(t, float64(403), extras["example/rounded"], "only the new status key changes decode behavior")
			require.Equal(t, &Implementation{Name: "fixture", Version: "1"}, result.Meta.IoModelcontextprotocolServerInfo)
			wire, err := json.Marshal(result)
			require.NoError(t, err)
			var again CallToolResult
			require.NoError(t, json.Unmarshal(wire, &again))
			status, valid = ToolErrorHTTPStatus(&again)
			require.Equal(t, test.valid, valid)
			require.Equal(t, test.want, status)
		})
	}
}

func TestToolErrorHTTPStatusBoundsUntrustedNumberConversion(t *testing.T) {
	failed := true
	for _, number := range []string{"403e999999999", "403e-999999999", "1e1000000", "403." + strings.Repeat("0", 61), "403e" + strings.Repeat("0", 61)} {
		result := &CallToolResult{IsError: &failed, Meta: &ResultMetaObject{AdditionalProperties: map[string]any{ToolErrorHTTPStatusMetaKey: json.Number(number)}}}
		_, valid := ToolErrorHTTPStatus(result)
		require.False(t, valid, "untrusted status metadata must be bounded before exact conversion")
	}
	for _, number := range []string{"403e0", "4.03e2", "403." + strings.Repeat("0", 60)} {
		result := &CallToolResult{IsError: &failed, Meta: &ResultMetaObject{AdditionalProperties: map[string]any{ToolErrorHTTPStatusMetaKey: json.Number(number)}}}
		status, valid := ToolErrorHTTPStatus(result)
		require.True(t, valid)
		require.Equal(t, 403, status)
	}
}
