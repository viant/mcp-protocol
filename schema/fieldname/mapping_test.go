package fieldname

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePreservesListNullAndEmpty(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[[]string](), reflect.TypeFor[[1]string]()} {
		for _, tc := range []struct {
			name  string
			value any
			wire  string
		}{
			{"nil interface", nil, "null"},
			{"typed nil list", []any(nil), "null"},
			{"empty list", []any{}, "[]"},
		} {
			t.Run(typ.String()+"/"+tc.name, func(t *testing.T) {
				got, err := Normalize(typ, tc.value)
				require.NoError(t, err)
				require.Equal(t, tc.value, got)
				wire, err := json.Marshal(got)
				require.NoError(t, err)
				require.Equal(t, tc.wire, string(wire))
			})
		}
	}
}

func TestNormalizeRenamedNestedListNullAndEmpty(t *testing.T) {
	type row struct {
		Categories []string `json:"iab_category" mcp:"iabCategory,aliases=categories"`
	}
	type body struct {
		Rows []row `json:"rows" mcp:"records"`
	}
	for _, spelling := range []string{"iab_category", "iabCategory", "categories"} {
		for _, tc := range []struct {
			name  string
			value any
			wire  string
		}{
			{"null", nil, `{"rows":[{"iab_category":null}]}`},
			{"typed nil", []any(nil), `{"rows":[{"iab_category":null}]}`},
			{"empty", []any{}, `{"rows":[{"iab_category":[]}]}`},
			{"populated", []any{"arts"}, `{"rows":[{"iab_category":["arts"]}]}`},
		} {
			t.Run(spelling+"/"+tc.name, func(t *testing.T) {
				got, err := Normalize(reflect.TypeFor[body](), map[string]any{
					"records": []any{map[string]any{spelling: tc.value}},
				})
				require.NoError(t, err)
				wire, err := json.Marshal(got)
				require.NoError(t, err)
				require.Equal(t, tc.wire, string(wire))
			})
		}
	}
}
