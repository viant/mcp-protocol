// Package fieldname resolves explicitly authored MCP field identities. JSON
// field selection and ordinary decoding remain owned by their existing callers.
package fieldname

import (
	"fmt"
	"reflect"
	"strings"
)

type Name struct {
	Public   string
	Aliases  []string
	Hidden   bool
	Explicit bool
}

func Resolve(field reflect.StructField, fallback string) (Name, error) {
	r := Name{Public: fallback}
	raw := strings.TrimSpace(field.Tag.Get("mcp"))
	if raw == "-" {
		r.Hidden = true
		return r, nil
	}
	var positional, named string
	var havePos, haveName bool
	for i, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			if i != 0 {
				continue
			}
			positional = item
			havePos = true
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "name":
			if haveName && named != value {
				return r, fmt.Errorf("conflicting mcp names on field %s", field.Name)
			}
			named = value
			haveName = true
		case "aliases":
			seen := map[string]bool{}
			for _, alias := range strings.FieldsFunc(value, func(c rune) bool { return c == '|' || c == ';' }) {
				alias = strings.TrimSpace(alias)
				if alias != "" && !seen[alias] {
					r.Aliases = append(r.Aliases, alias)
					seen[alias] = true
				}
			}
		default:
			continue
		}
	}
	if havePos && haveName && positional != named {
		return r, fmt.Errorf("conflicting mcp names on field %s", field.Name)
	}
	if havePos {
		r.Public = positional
		r.Explicit = true
	} else if named != "" {
		r.Public = named
		r.Explicit = true
	}
	return r, nil
}

// Entry maps a selected original JSON field to its public MCP identity.
type Entry struct {
	Field    reflect.StructField
	JSONName string
	Name     Name
}

// Validate detects final public collisions without mutating the caller's object.
func Validate(entries []Entry) error {
	seen := map[string]int{}
	explicit := false
	for _, e := range entries {
		explicit = explicit || e.Name.Explicit || len(e.Name.Aliases) > 0
	}
	if !explicit {
		return nil
	}
	for entryIndex, e := range entries {
		if e.Name.Hidden {
			continue
		}
		own := map[string]bool{}
		spellings := append([]string{e.Name.Public, e.JSONName}, e.Name.Aliases...)
		for _, name := range spellings {
			if own[name] {
				continue
			}
			own[name] = true
			if prior, ok := seen[name]; ok && prior != entryIndex {
				return fmt.Errorf("MCP property spelling %q conflicts between %s and %s", name, entries[prior].Field.Name, e.Field.Name)
			}
			seen[name] = entryIndex
		}
	}
	return nil
}
