package fieldname

import (
	"reflect"
	"strings"
)

// Required reports the legacy MCP schema requiredness of one selected JSON
// field. It is schema metadata only; runtime validation is owned elsewhere.
func Required(field reflect.StructField) bool {
	tag := string(field.Tag)
	if strings.Contains(tag, "required=false") || strings.Contains(tag, "optional") {
		return false
	}
	_, omitempty := jsonOption(field, "omitempty")
	value, explicit := field.Tag.Lookup("required")
	explicit = value == "true" || (value == "" && explicit)
	return (field.Type.Kind() != reflect.Pointer && !omitempty) || explicit
}

// OptionalOwner reports whether a legacy inline embedding suppresses the
// required members it contributes. Legacy behavior intentionally does not let
// required tags on the owner override pointer, omitempty, or optional markers.
func OptionalOwner(field reflect.StructField) bool {
	if field.Type.Kind() == reflect.Pointer {
		return true
	}
	if _, ok := jsonOption(field, "omitempty"); ok {
		return true
	}
	tag := string(field.Tag)
	return strings.Contains(tag, "required=false") || strings.Contains(tag, "optional")
}

func jsonOption(field reflect.StructField, want string) (string, bool) {
	for _, option := range strings.Split(field.Tag.Get("json"), ",")[1:] {
		if option == want {
			return option, true
		}
	}
	return "", false
}
