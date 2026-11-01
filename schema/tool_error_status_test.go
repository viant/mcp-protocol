package schema

import (
	"encoding/json"
	"math"
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
