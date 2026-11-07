package fieldname

import (
	"reflect"
	"strings"
)

// Required reports explicitly authored requiredness. Go value kinds and JSON
// omission flags do not imply required presence. Runtime validation is separate.
func Required(field reflect.StructField) bool {
	// Parameter binding flags belong to their owning runtime, not this schema
	// declaration. In particular, an optional binding can document a required
	// public property without changing runtime validation/error ordering.
	if value, ok := field.Tag.Lookup("optional"); ok && value != "false" {
		return false
	}
	value, explicit := field.Tag.Lookup("required")
	explicit = value == "true" || (value == "" && explicit)
	return explicit
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
	return explicitlyOptional(field.Tag)
}

func explicitlyOptional(tag reflect.StructTag) bool {
	if value, ok := tag.Lookup("optional"); ok && value != "false" {
		return true
	}
	if value, ok := tag.Lookup("required"); ok && value == "false" {
		return true
	}
	return strings.Contains(string(tag), "required=false")
}

func jsonOption(field reflect.StructField, want string) (string, bool) {
	for _, option := range strings.Split(field.Tag.Get("json"), ",")[1:] {
		if option == want {
			return option, true
		}
	}
	return "", false
}
