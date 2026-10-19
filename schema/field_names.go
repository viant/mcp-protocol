package schema

import (
	"fmt"
	"github.com/viant/mcp-protocol/schema/fieldname"
	"reflect"
	"strings"
)

// Legacy has no error return. Invalid/colliding explicit names leave this entire
// object's original property names and required entries intact. Load validates
// first and can return an error; direct legacy callers cannot observe that error.
func legacyEntries(t reflect.Type) ([]fieldname.Entry, error) {
	var entries []fieldname.Entry
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || f.Tag.Get("json") == "-" || f.Tag.Get("mcp") == "-" || f.Tag.Get("internal") != "" {
			continue
		}
		j := strings.Split(f.Tag.Get("json"), ",")
		if f.Anonymous || strings.Contains(f.Tag.Get("json"), ",inline") {
			continue
		}
		name := j[0]
		if name == "" {
			name = f.Name
		}
		n, e := fieldname.Resolve(f, name)
		if e != nil {
			return nil, e
		}
		entries = append(entries, fieldname.Entry{Field: f, JSONName: name, Name: n})
	}
	return entries, nil
}
func validateLegacyFieldNames(t reflect.Type, options ...StructToPropertiesOption) error {
	var opt structToPropertiesOptions
	for _, o := range options {
		o(&opt)
	}
	seen := map[reflect.Type]bool{}
	var visit func(reflect.Type) error
	visit = func(t reflect.Type) error {
		if seen[t] {
			return nil
		}
		seen[t] = true
		switch t.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
			return visit(t.Elem())
		}
		if t.Kind() != reflect.Struct {
			return nil
		}

		entries, err := fieldname.Fields(t, legacyVisibility(opt.SkipFieldHook))
		if err != nil {
			return err
		}
		if fieldname.CustomJSON(t) {
			yes, e := fieldname.HasNames(t, legacyVisibility(opt.SkipFieldHook))
			if e != nil {
				return e
			}
			if yes {
				return fmt.Errorf("MCP names on custom JSON type %s require decoding authority", t)
			}
		}
		for _, entry := range entries {
			if fieldname.CustomJSON(t) && (entry.Name.Explicit || len(entry.Name.Aliases) > 0) {
				return fmt.Errorf("MCP names on custom JSON type %s require decoding authority", t)
			}
			if err := visit(entry.Field.Type); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(t)
}
func applyLegacyFieldNames(t reflect.Type, props ToolInputSchemaProperties, required []string) (ToolInputSchemaProperties, []string) {
	entries, e := legacyEntries(t)
	if e != nil || fieldname.CustomJSON(t) || fieldname.Validate(entries) != nil {
		return props, required
	}
	renames := map[string]string{}
	dest := map[string]bool{}
	for key := range props {
		dest[key] = true
	}
	for _, entry := range entries {
		if !entry.Name.Explicit || entry.Name.Public == entry.JSONName {
			continue
		}
		if _, ok := props[entry.JSONName]; !ok {
			continue
		}
		renames[entry.JSONName] = entry.Name.Public
		delete(dest, entry.JSONName)
	}
	for _, name := range renames {
		if dest[name] {
			return props, required
		}
		dest[name] = true
	}
	out := make(ToolInputSchemaProperties, len(props))
	for name, p := range props {
		if v, ok := renames[name]; ok {
			name = v
		}
		out[name] = p
	}
	req := append([]string(nil), required...)
	for i, name := range req {
		if v, ok := renames[name]; ok {
			req[i] = v
		}
	}
	return out, req
}

func customBoundaryOptions(t reflect.Type, opts []StructToPropertiesOption) []StructToPropertiesOption {
	if !fieldname.CustomJSON(t) {
		return opts
	}
	return append(append([]StructToPropertiesOption(nil), opts...), func(o *structToPropertiesOptions) { o.skipMCPNames = true })
}

func legacyVisibility(skip func(reflect.StructField) bool) fieldname.Visibility {
	return fieldname.Visibility{ExcludeField: func(f reflect.StructField) bool {
		return f.Tag.Get("mcp") == "-" || f.Tag.Get("internal") != "" || skip != nil && skip(f)
	}}
}
