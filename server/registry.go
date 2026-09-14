package server

import "github.com/viant/mcp-protocol/syncmap"
import "github.com/viant/mcp-protocol/schema"
import skillformat "github.com/viant/mcp-protocol/extension/skills"
import "sync"

// Registry holds registered tools, resources, prompts, etc. for a handler instance.
type Registry struct {
	staticMu                 sync.RWMutex
	staticSkills             map[string]*skillformat.Static
	staticResources          map[string]schema.Resource
	skills                   *syncmap.Map[string, schema.Skill]
	ToolRegistry             *syncmap.Map[string, *ToolEntry]
	ResourceRegistry         *syncmap.Map[string, *ResourceEntry]
	ResourceTemplateRegistry *syncmap.Map[string, *ResourceTemplateEntry]
	Prompts                  *syncmap.Map[string, *PromptEntry]
	Methods                  *syncmap.Map[string, bool]
}

// NewRegistry creates and initialises an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		staticSkills:             map[string]*skillformat.Static{},
		staticResources:          map[string]schema.Resource{},
		skills:                   syncmap.NewMap[string, schema.Skill](),
		ToolRegistry:             syncmap.NewMap[string, *ToolEntry](),
		ResourceRegistry:         syncmap.NewMap[string, *ResourceEntry](),
		ResourceTemplateRegistry: syncmap.NewMap[string, *ResourceTemplateEntry](),
		Prompts:                  syncmap.NewMap[string, *PromptEntry](),
		Methods:                  syncmap.NewMap[string, bool](),
	}
}
