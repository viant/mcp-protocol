package server

import (
	"context"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

// PromptHandlerFunc defines a function to handle a prompt call.
type PromptHandlerFunc func(ctx context.Context, request *schema.GetPromptRequestParams) (*schema.GetPromptResult, *jsonrpc.Error)

// PromptEntry holds a handler with its metadata.
type PromptEntry struct {
	Handler PromptHandlerFunc
	Prompt  *schema.Prompt
}

// Prompts is a collection of PromptEntry.
type Prompts []*PromptEntry

// RegisterPrompts registers a prompt on this handler.
func (d *Registry) RegisterPrompts(prompt *schema.Prompt, handler PromptHandlerFunc) {
	d.Methods.Put(schema.MethodPromptsList, true)
	d.Methods.Put(schema.MethodPromptsGet, true)
	d.Prompts.Put(prompt.Name, &PromptEntry{Prompt: prompt, Handler: handler})
}
