package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/jsonrpc"
	skillformat "github.com/viant/mcp-protocol/extension/skills"
	"github.com/viant/mcp-protocol/schema"
)

// RegisterStaticSkill installs compiler-owned readers and inventory together.
// It never invokes an existing resource handler to infer bytes or completeness.
func (r *Registry) RegisterStaticSkill(compiled *skillformat.Static) error {
	if err := compiled.Validate(); err != nil {
		return err
	}
	r.staticMu.Lock()
	defer r.staticMu.Unlock()
	entry := compiled.Metadata()
	if _, exists := r.skills.Get(entry.Uri); exists {
		return fmt.Errorf("duplicate skill URI %q", entry.Uri)
	}
	inventory := map[string]schema.SkillResource{}
	for _, file := range entry.Resources.Files {
		inventory[file.Uri] = file
	}
	for _, resource := range r.ResourceRegistry.Values() {
		if compiled.Contains(resource.Metadata.Uri) {
			if _, ok := inventory[resource.Metadata.Uri]; !ok {
				return fmt.Errorf("existing resource outside static inventory: %s", resource.Metadata.Uri)
			}
		}
	}
	for _, other := range r.skills.Values() {
		if other.Resources.Dynamic && compiled.Contains(other.Uri) {
			return fmt.Errorf("static root contains dynamic skill")
		}
	}
	for _, other := range r.staticSkills {
		old := map[string]schema.SkillResource{}
		for _, file := range other.Metadata().Resources.Files {
			old[file.Uri] = file
			expected, exists := inventory[file.Uri]
			if compiled.Contains(file.Uri) && !exists || exists && (file.Digest != expected.Digest || file.Size != expected.Size) {
				return fmt.Errorf("conflicting static skill inventories")
			}
		}
		for uri := range inventory {
			if other.Contains(uri) {
				if _, ok := old[uri]; !ok {
					return fmt.Errorf("static skill adds unlisted nested content")
				}
			}
		}
	}
	metadata := compiled.Resources()
	for i, resource := range metadata {
		if existing, ok := r.ResourceRegistry.Get(resource.Uri); ok {
			copy, err := cloneStaticResource(existing.Metadata)
			if err != nil {
				return err
			}
			copy.Uri, copy.Size, copy.MimeType = resource.Uri, resource.Size, resource.MimeType
			if resource.Uri == entry.Uri {
				copy.Title, copy.Description = resource.Title, resource.Description
			}
			metadata[i] = copy
		}
	}
	r.staticSkills[entry.Uri] = compiled
	r.skills.Put(entry.Uri, entry)
	for _, resource := range metadata {
		r.staticResources[resource.Uri] = resource
		r.ResourceRegistry.Put(resource.Uri, &ResourceEntry{Metadata: resource, Handler: compiled.ReadResource})
	}
	for _, method := range []string{schema.MethodResourcesList, schema.MethodResourcesRead, schema.MethodSkillsList, schema.MethodSkillsGet} {
		r.Methods.Put(method, true)
	}
	return nil
}

func cloneStaticResource(resource schema.Resource) (schema.Resource, error) {
	data, err := json.Marshal(resource)
	if err != nil {
		return schema.Resource{}, err
	}
	var result schema.Resource
	err = json.Unmarshal(data, &result)
	return result, err
}

// ReadStaticSkillResource claims the entire sealed root, including misses. This
// prevents an ordinary resource/template from adding unlisted files afterward.
func (r *Registry) ReadStaticSkillResource(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error, bool) {
	if request == nil {
		return nil, nil, false
	}
	r.staticMu.RLock()
	defer r.staticMu.RUnlock()
	for _, compiled := range r.staticSkills {
		if compiled.Contains(request.Params.Uri) {
			result, err := compiled.ReadResource(ctx, request)
			return result, err, true
		}
	}
	return nil, nil, false
}
