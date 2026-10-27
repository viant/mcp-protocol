package jsonschema

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/mcp-protocol/schema/fieldname"
)

type fallbackChild struct{ MetricID int }
type fallbackBody struct {
	RecordID int
	Children []fallbackChild
	Explicit string `json:"Exact_Name"`
	Named    int    `mcp:"AccountingId"`
}

func fallbackLower(field reflect.StructField, fallback string) string {
	if name, _, _ := strings.Cut(field.Tag.Get("json"), ","); name != "" {
		return fallback
	}
	return strings.ToLower(fallback)
}
func TestFallbackNameSchemaAndBinding(t *testing.T) {
	policy := fieldname.Visibility{FallbackName: fallbackLower}
	schema, err := (Reflector{FallbackName: fallbackLower}).Compile(Request{Type: reflect.TypeFor[fallbackBody]()})
	require.NoError(t, err)
	props := schema["properties"].(map[string]any)
	for _, key := range []string{"recordid", "children", "Exact_Name", "AccountingId"} {
		require.Contains(t, props, key)
	}
	require.Contains(t, schema["required"], "recordid")
	child := props["children"].(map[string]any)["items"].(map[string]any)
	require.Contains(t, child["properties"], "metricid")
	raw := map[string]any{"recordid": float64(7), "children": []any{map[string]any{"metricid": float64(8)}}, "Exact_Name": "fixed", "AccountingId": float64(9)}
	got, err := fieldname.Normalize(reflect.TypeFor[fallbackBody](), raw, policy)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"RecordID": float64(7), "Children": []any{map[string]any{"MetricID": float64(8)}}, "Exact_Name": "fixed", "Named": float64(9)}, got)
	require.Contains(t, raw, "recordid", "normalization mutated caller input")
	_, err = fieldname.Normalize(reflect.TypeFor[fallbackBody](), map[string]any{"recordid": 7, "RecordID": 7}, policy)
	require.Error(t, err)
}
func TestFallbackNameCollision(t *testing.T) {
	type row struct {
		UserID int
		UserId int
	}
	_, err := (Reflector{}).Compile(Request{Type: reflect.TypeFor[row]()})
	require.NoError(t, err)
	_, err = (Reflector{FallbackName: fallbackLower}).Compile(Request{Type: reflect.TypeFor[row]()})
	require.Error(t, err)
}

type fallbackOpaque struct{ RecordID int }

func (*fallbackOpaque) UnmarshalJSON([]byte) error { panic("reflection must not execute decoder") }
func TestFallbackNamePreservesCustomDecoderAuthority(t *testing.T) {
	policy := fieldname.Visibility{FallbackName: fallbackLower}
	typ := reflect.TypeFor[fallbackOpaque]()
	expected, err := (Reflector{}).Compile(Request{Type: typ})
	require.NoError(t, err)
	actual, err := (Reflector{FallbackName: fallbackLower}).Compile(Request{Type: typ})
	require.NoError(t, err)
	require.Equal(t, expected, actual)
	raw := map[string]any{"RecordID": float64(7)}
	actualValue, err := fieldname.Normalize(typ, raw, policy)
	require.NoError(t, err)
	require.Equal(t, raw, actualValue)
}

type fallbackCustomParent struct {
	Child struct {
		UserID int
		UserId int
	}
}

func (*fallbackCustomParent) UnmarshalJSON([]byte) error { panic("must not execute decoder") }
func TestFallbackNameCustomSubtreeStaysUnchanged(t *testing.T) {
	typ := reflect.TypeFor[fallbackCustomParent]()
	policy := fieldname.Visibility{FallbackName: fallbackLower}
	names, err := fieldname.HasNames(typ, policy)
	require.NoError(t, err)
	require.False(t, names)
	raw := map[string]any{"Child": map[string]any{"UserID": 1, "UserId": 2}}
	actual, err := fieldname.Normalize(typ, raw, policy)
	require.NoError(t, err)
	require.Equal(t, raw, actual)
}
