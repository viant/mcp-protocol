package fieldname

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"strings"
	"testing"
)

type formattedEmbedded struct{ Id int }
type formattedRow struct {
	formattedEmbedded
	Id *int `json:"id"`
}

func lowerFallback(field reflect.StructField, fallback string) string {
	if field.Tag.Get("json") != "" {
		return fallback
	}
	return strings.ToLower(fallback)
}

func TestFormattedEmbeddedDominance(t *testing.T) {
	policy := Visibility{FallbackName: lowerFallback}
	fields, err := Fields(reflect.TypeFor[formattedRow](), policy)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	require.Equal(t, []int{1}, fields[0].Field.Index)
	normalized, err := Normalize(reflect.TypeFor[formattedRow](), map[string]any{"id": float64(7)}, policy)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"id": float64(7)}, normalized)
	type nested struct{ Rows []formattedRow }
	_, err = Normalize(reflect.TypeFor[nested](), map[string]any{"rows": []any{map[string]any{"id": float64(7)}}}, policy)
	require.NoError(t, err)
	policy.ExcludeField = func(field reflect.StructField) bool { return field.Type.Kind() == reflect.Pointer }
	fields, err = Fields(reflect.TypeFor[formattedRow](), policy)
	require.NoError(t, err)
	require.Empty(t, fields, "an excluded winner must not reveal its shadowed embedded field")
}

func TestFormattedDominancePreservesConflicts(t *testing.T) {
	policy := Visibility{FallbackName: lowerFallback}
	type sameDepth struct {
		Id    int
		Other int `json:"id"`
	}
	fields, err := Fields(reflect.TypeFor[sameDepth](), policy)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	require.Equal(t, "Other", fields[0].Field.Name)
	type ambiguous struct {
		Id int
		ID int
	}
	_, err = Fields(reflect.TypeFor[ambiguous](), policy)
	require.Error(t, err)
	type explicitEmbedded struct {
		Value int `mcp:"id"`
	}
	type explicitRow struct {
		explicitEmbedded
		Other int `mcp:"id"`
	}
	_, err = Fields(reflect.TypeFor[explicitRow](), policy)
	require.Error(t, err, "explicit MCP names must not silently shadow one another")
	type aliases struct {
		Id    int
		Other int `mcp:"other,aliases=id"`
	}
	_, err = Fields(reflect.TypeFor[aliases](), policy)
	require.Error(t, err)
}
