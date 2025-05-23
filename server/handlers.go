package server

import "github.com/viant/mcp-protocol/syncmap"

// Registry holds registered handlers and tools.
type Registry struct {
	// ToolRegistry holds per-instance registered tools and handlers.
	ToolRegistry             *syncmap.Map[string, *ToolEntry]
	ResourceRegistry         *syncmap.Map[string, *ResourceEntry]
	ResourceTemplateRegistry *syncmap.Map[string, *ResourceTemplateEntry]
	Prompts                  *syncmap.Map[string, *PromptEntry]
	Methods                  *syncmap.Map[string, bool]
}

// NewHandlerRegistry creates a new Registry instance.
func NewRegistry() *Registry {
	// constructor for Registry
	return &Registry{
		ToolRegistry:             syncmap.NewMap[string, *ToolEntry](),
		ResourceRegistry:         syncmap.NewMap[string, *ResourceEntry](),
		ResourceTemplateRegistry: syncmap.NewMap[string, *ResourceTemplateEntry](),
		Prompts:                  syncmap.NewMap[string, *PromptEntry](),
		Methods:                  syncmap.NewMap[string, bool](),
	}
}
