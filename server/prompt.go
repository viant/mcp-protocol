package server

import (
	"context"
	"fmt"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

// PromptHandlerFunc defines a function to handle a tool call.
type PromptHandlerFunc func(ctx context.Context, request *schema.GetPromptRequestParams) (*schema.GetPromptResult, *jsonrpc.Error)

// PromptEntry holds a handler with its metadata.
type PromptEntry struct {
	Handler PromptHandlerFunc
	Prompt  *schema.Prompt
}

// RegisterPrompts registers a prompt on this DefaultImplementer.
func (d *DefaultImplementer) RegisterPrompts(prompt *schema.Prompt, handler PromptHandlerFunc) {
	d.Methods.Put(schema.MethodPromptsList, true)
	d.Prompts.Put(prompt.Name, &PromptEntry{Prompt: prompt, Handler: handler})
}

// ListPrompts lists all registered prompts on this DefaultImplementer.
func (d *DefaultImplementer) ListPrompts(ctx context.Context, request *schema.ListPromptsRequest) (*schema.ListPromptsResult, *jsonrpc.Error) {
	result := &schema.ListPromptsResult{}
	for _, entry := range d.Prompts.Values() {
		result.Prompts = append(result.Prompts, *entry.Prompt)
	}
	return result, nil
}

// GetPrompt returns the result of a prompt call.
func (d *DefaultImplementer) GetPrompt(ctx context.Context, request *schema.GetPromptRequest) (*schema.GetPromptResult, *jsonrpc.Error) {
	promptEntry, ok := d.Prompts.Get(request.Params.Name)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(
			fmt.Sprintf("prompt %q not found", request.Params.Name), nil)
	}
	prompt := promptEntry.Prompt
	for _, arg := range prompt.Arguments {
		if arg.Required != nil && *arg.Required {
			if _, ok := request.Params.Arguments[arg.Name]; !ok {
				return nil, jsonrpc.NewInvalidRequest(fmt.Sprintf("missing required argument %q", arg.Name), nil)
			}
		}
	}
	return promptEntry.Handler(ctx, &request.Params)
}
