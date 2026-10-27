package fieldname

import "reflect"

// Visibility carries the existing schema owner's selection policy to discovery
// and invocation. Its zero value preserves ordinary native JSON visibility.
type Visibility struct {
	// FallbackName optionally maps the selected JSON fallback before explicit
	// MCP names are resolved. Nil preserves existing field names.
	FallbackName func(reflect.StructField, string) string

	ExcludeInternal bool
	ExcludeField    func(reflect.StructField) bool
}

// Excluded follows the exact native selected index, including promoted owners.
// Winner selection happens before this check, so excluded winners never expose
// shadowed fields. Only internal:"true" is hidden by ExcludeInternal.
func (v Visibility) Excluded(owner reflect.Type, index []int) bool {
	for _, position := range index {
		for owner.Kind() == reflect.Pointer {
			owner = owner.Elem()
		}
		field := owner.Field(position)
		if v.ExcludeInternal && field.Tag.Get("internal") == "true" || v.ExcludeField != nil && v.ExcludeField(field) {
			return true
		}
		owner = field.Type
	}
	return false
}
func selectedVisibility(options []Visibility) Visibility {
	if len(options) > 0 {
		return options[0]
	}
	return Visibility{}
}

// Resolve applies a caller's fallback policy before explicitly authored MCP names.
func (v Visibility) Resolve(field reflect.StructField, fallback string) (Name, error) {
	if v.FallbackName != nil {
		fallback = v.FallbackName(field, fallback)
	}
	return Resolve(field, fallback)
}
