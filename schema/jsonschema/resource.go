package jsonschema

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonpointer"
	"github.com/xeipuuv/gojsonschema"
	"io/fs"
	"net/url"
	"sort"
	"strings"
)

// ResourceCompiler validates and bundles authored JSON Schema documents. All
// external references pass through FS; there is no HTTP or OS fallback loader.
type ResourceCompiler struct{ FS fs.FS }
type ResourceRequest struct {
	Schema map[string]any
	Base   string
	Prefix string
}
type Bundle struct {
	Schema      map[string]any
	Definitions map[string]map[string]any
	Resources   []string
}

func (c ResourceCompiler) Compile(request ResourceRequest) (*Bundle, error) {
	if request.Schema == nil {
		return nil, fmt.Errorf("authored schema is required")
	}
	// Validate the JSON document actually consumed by the schema owner, including
	// typed Go maps/slices supplied through the public request API.
	encodedRoot, err := json.Marshal(request.Schema)
	if err != nil {
		return nil, err
	}
	var source map[string]any
	if err = json.Unmarshal(encodedRoot, &source); err != nil {
		return nil, err
	}
	if err := (draft04Dialect{}).validate(source, "#"); err != nil {
		return nil, err
	}
	base, err := (resourceLocation{}).resolve("", request.Base)
	if err != nil {
		return nil, err
	}
	factory := &resourceFactory{fs: c.FS, documents: map[string]any{base: source}}
	compiler := bundleCompiler{scope: string(encodedRoot), factory: factory, prefix: request.Prefix, definitions: map[string]map[string]any{}, references: map[string]string{}}
	result, err := compiler.node(source, base)
	if err != nil {
		return nil, err
	}
	// Validate authored documents with the native built-in metaschema. Bundling
	// resolves every external reference through FS before the validator sees it.
	meta, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader("http://json-schema.org/draft-04/schema#"))
	if err != nil {
		return nil, err
	}
	locations := make([]string, 0, len(factory.documents))
	for location := range factory.documents {
		locations = append(locations, location)
	}
	sort.Strings(locations)
	for _, location := range locations {
		checked, err := meta.Validate(gojsonschema.NewGoLoader(factory.documents[location]))
		if err != nil {
			return nil, err
		}
		if !checked.Valid() {
			return nil, fmt.Errorf("invalid schema %s: %v", location, checked.Errors())
		}
	}
	validation := map[string]any{}
	for key, value := range result {
		validation[key] = value
	}
	switch request.Prefix {
	case "#/components/schemas/":
		validation["components"] = map[string]any{"schemas": compiler.definitions}
	case "#/definitions/":
		validation["definitions"] = compiler.definitions
	default:
		return nil, fmt.Errorf("unsupported schema reference mount %q", request.Prefix)
	}
	validator := gojsonschema.NewSchemaLoader()
	validator.AutoDetect = false
	validator.Draft = gojsonschema.Draft4
	if _, err = validator.Compile(gojsonschema.NewGoLoader(validation)); err != nil {
		return nil, fmt.Errorf("schema %s: %w", request.Base, err)
	}
	references := make([]string, 0, len(factory.documents))
	for location := range factory.documents {
		if location != base {
			references = append(references, (resourceLocation{}).reference(location))
		}
	}
	sort.Strings(references)
	return &Bundle{Schema: result, Definitions: compiler.definitions, Resources: references}, nil
}

// resourceFactory caches parsed schema documents for one compilation, not a
// parallel filesystem authority. Both validation and bundling see identical bytes.
type resourceFactory struct {
	fs        fs.FS
	documents map[string]any
}

func (f *resourceFactory) load(uri string) (any, error) {
	location, err := (resourceLocation{}).resolve("", uri)
	if err != nil {
		return nil, err
	}
	parsed, _ := url.Parse(location)
	parsed.Fragment = ""
	location = parsed.String()
	if value, ok := f.documents[location]; ok {
		return value, nil
	}
	if f.fs == nil {
		return nil, fmt.Errorf("schema resource filesystem required for %s", uri)
	}
	data, err := fs.ReadFile(f.fs, (resourceLocation{}).reference(location))
	if err != nil {
		return nil, err
	}
	var value any
	if err = json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("schema resource %s: %w", uri, err)
	}
	if object, ok := value.(map[string]any); ok {
		if err := (draft04Dialect{}).validate(object, location); err != nil {
			return nil, err
		}
	}
	f.documents[location] = value
	return value, nil
}

type resourceLocation struct{}

func (resourceLocation) resolve(base, reference string) (string, error) {
	if namespace, file, ok := strings.Cut(reference, ":"); ok && !strings.Contains(reference, "://") && !strings.ContainsAny(namespace, "/\\") {
		reference = "resource://" + namespace + "/" + file
	}
	ref, err := url.Parse(reference)
	if err != nil {
		return "", err
	}
	if ref.Opaque != "" {
		ref = &url.URL{Scheme: "resource", Host: ref.Scheme, Path: "/" + ref.Opaque, Fragment: ref.Fragment}
	}
	if ref.Scheme == "" && base == "" {
		ref = &url.URL{Scheme: "resource", Host: "default", Path: "/" + ref.Path, Fragment: ref.Fragment}
	}
	if base != "" {
		parent, err := url.Parse(base)
		if err != nil {
			return "", err
		}
		ref = parent.ResolveReference(ref)
	}
	if ref.Scheme != "resource" || ref.Host == "" || !fs.ValidPath(strings.TrimPrefix(ref.Path, "/")) || ref.RawQuery != "" || ref.User != nil || ref.Port() != "" {
		return "", fmt.Errorf("unsupported schema resource reference %q", reference)
	}
	return ref.String(), nil
}
func (resourceLocation) reference(uri string) string {
	value, _ := url.Parse(uri)
	path := strings.TrimPrefix(value.Path, "/")
	if value.Host == "default" {
		return path
	}
	return value.Host + ":" + path
}

type bundleCompiler struct {
	scope       string
	factory     *resourceFactory
	prefix      string
	definitions map[string]map[string]any
	references  map[string]string
}

func (c *bundleCompiler) node(source map[string]any, base string) (map[string]any, error) {
	// Resource identity is owned by the explicit filesystem reference. Reject
	// alternate URI scopes instead of silently resolving against the wrong file.
	for _, keyword := range []string{"id", "$id"} {
		if value, ok := source[keyword]; ok {
			return nil, fmt.Errorf("schema %s URI scope %v is unsupported; use a resource reference", keyword, value)
		}
	}
	if value, ok := source["$schema"]; ok && value != "http://json-schema.org/draft-04/schema#" && value != "http://json-schema.org/draft-04/schema" {
		return nil, fmt.Errorf("schema dialect %v is not OpenAPI 3.0-compatible draft-04", value)
	}
	if raw, ok := source["$ref"]; ok {
		ref, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("schema $ref must be string")
		}
		location, err := (resourceLocation{}).resolve(base, ref)
		if err != nil {
			return nil, err
		}
		if name := c.references[location]; name != "" {
			return map[string]any{"$ref": c.prefix + name}, nil
		}
		name := fmt.Sprintf("Doc_%x", sha256.Sum256([]byte(location+":"+c.scope)))
		c.references[location] = name
		c.definitions[name] = map[string]any{}
		value, err := c.factory.load(location)
		if err != nil {
			return nil, err
		}
		parsed, _ := url.Parse(location)
		pointer, err := gojsonpointer.NewJsonPointer(parsed.Fragment)
		if err != nil {
			return nil, err
		}
		value, _, err = pointer.Get(value)
		if err != nil {
			return nil, err
		}
		object, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("schema reference %s does not target an object", location)
		}
		resolved, err := c.node(object, location)
		if err != nil {
			return nil, err
		}
		c.definitions[name] = resolved
		return map[string]any{"$ref": c.prefix + name}, nil
	}
	result := map[string]any{}
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := source[key]
		switch key {
		case "definitions", "$id", "id", "$schema":
			continue
		case "properties", "patternProperties":
			object, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("schema %s must be object", key)
			}
			mapped := map[string]any{}
			names := make([]string, 0, len(object))
			for name := range object {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				child, ok := object[name].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("schema %s.%s must be object", key, name)
				}
				item, err := c.node(child, base)
				if err != nil {
					return nil, err
				}
				mapped[name] = item
			}
			result[key] = mapped
		case "allOf", "anyOf", "oneOf":
			items, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("schema %s must be array", key)
			}
			list := make([]any, len(items))
			for i, item := range items {
				child, ok := item.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("schema %s entry must be object", key)
				}
				mapped, err := c.node(child, base)
				if err != nil {
					return nil, err
				}
				list[i] = mapped
			}
			result[key] = list
		case "items", "additionalProperties", "not":
			if child, ok := value.(map[string]any); ok {
				mapped, err := c.node(child, base)
				if err != nil {
					return nil, err
				}
				result[key] = mapped
			} else {
				result[key] = value
			}
		default:
			result[key] = value
		}
	}
	return result, nil
}
