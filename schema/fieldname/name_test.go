package fieldname

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestResolve(t *testing.T) {
	for _, c := range []struct {
		tag, want   string
		hidden, err bool
	}{{"", "jsonName", false, false}, {`mcp:",aliases=old|other;old"`, "jsonName", false, false}, {`mcp:"AccountingId,aliases=old"`, "AccountingId", false, false}, {`mcp:"name=AccountingId,aliases=old"`, "AccountingId", false, false}, {`mcp:"AccountingId,name=AccountingId"`, "AccountingId", false, false}, {`mcp:"AccountingId,name=other"`, "", false, true}, {`mcp:"AccountingId,bogus=x"`, "AccountingId", false, false}, {`mcp:"-"`, "jsonName", true, false}} {
		n, e := Resolve(reflect.StructField{Name: "Field", Tag: reflect.StructTag(c.tag)}, "jsonName")
		if c.err {
			require.Error(t, e)
			continue
		}
		require.NoError(t, e)
		require.Equal(t, c.want, n.Public)
		require.Equal(t, c.hidden, n.Hidden)
	}
}
