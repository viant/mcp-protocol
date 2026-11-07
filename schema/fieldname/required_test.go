package fieldname

import (
	"reflect"
	"testing"
)

func TestFieldsAreOptionalUnlessExplicitlyRequired(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  reflect.Type
		tag  reflect.StructTag
		want bool
	}{
		{"scalar default", reflect.TypeOf(""), "", false},
		{"pointer default", reflect.TypeOf((*string)(nil)), "", false},
		{"omitempty", reflect.TypeOf(""), `json:",omitempty"`, false},
		{"explicit required pointer", reflect.TypeOf((*string)(nil)), `required:"true" optional:"false"`, true},
		{"explicit false scalar", reflect.TypeOf(""), `required:"false"`, false},
		{"optional marker", reflect.TypeOf(""), `optional:""`, false},
		{"optional false", reflect.TypeOf(""), `optional:"false"`, false},
		{"optional true wins", reflect.TypeOf(""), `required:"true" optional:"true"`, false},
		{"description not a flag", reflect.TypeOf(""), `description:"an optional feature"`, false},
		{"explicit before parameter", reflect.TypeOf((*[]int)(nil)), `required:"true" parameter:",required=false"`, true},
		{"explicit after parameter", reflect.TypeOf((*[]int)(nil)), `parameter:",required=false" required:"true"`, true},
		{"explicit empty required", reflect.TypeOf(""), `parameter:",required=false" required:""`, true},
		{"explicit optional still wins", reflect.TypeOf(""), `required:"true" parameter:",required=false" optional:"true"`, false},
		{"description cannot override", reflect.TypeOf(""), `required:"true" description:"parameter required=false is supported"`, true},
		{"legacy parameter", reflect.TypeOf(""), `parameter:",required=false"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			field := reflect.StructField{Name: "Value", Type: tc.typ, Tag: tc.tag}
			if got := Required(field); got != tc.want {
				t.Fatalf("required=%v want=%v", got, tc.want)
			}
		})
	}
	if OptionalOwner(reflect.StructField{Name: "Embedded", Type: reflect.TypeOf(struct{}{}), Tag: `optional:"false"`}) {
		t.Fatal("false optional suppressed embedded requirements")
	}
}
