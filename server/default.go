package server

import (
	"context"
	"fmt"
	"github.com/viant/jsonrpc"
	"github.com/viant/jsonrpc/transport"
	"github.com/viant/mcp-protocol/client"
	"github.com/viant/mcp-protocol/logger"
	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-protocol/syncmap"
)

// DefaultServer provides default implementations for server-side methods.
// You can embed this in your own Server and register tools/resources via its helper methods.
type DefaultServer struct {
	Notifier           transport.Notifier
	Logger             logger.Logger
	Client             client.Operations
	ClientInitialize   *schema.InitializeRequestParams
	Subscription       *syncmap.Map[string, bool]
	ServerCapabilities *schema.ServerCapabilities
	*Registry
}

// Initialize stores the initialization parameters.
func (d *DefaultServer) Initialize(ctx context.Context, init *schema.InitializeRequestParams, result *schema.InitializeResult) {
	d.ClientInitialize = init
	if d.ServerCapabilities != nil {
		result.Capabilities = *d.ServerCapabilities
	}
	if d.ToolRegistry.Size() > 0 {
		result.Capabilities.Tools = &schema.ServerCapabilitiesTools{}
	}
	if d.ResourceRegistry.Size() > 0 {
		result.Capabilities.Resources = &schema.ServerCapabilitiesResources{}
	}
	if d.Prompts.Size() > 0 {
		result.Capabilities.Prompts = &schema.ServerCapabilitiesPrompts{}
	}

	d.Client.Init(ctx, &d.ClientInitialize.Capabilities)

}

// ListResources returns method-not-found by default.
func (d *DefaultServer) ListResources(ctx context.Context, request *schema.ListResourcesRequest) (*schema.ListResourcesResult, *jsonrpc.Error) {
	// Return list of registered resources
	resources := d.ListRegisteredResources()
	return &schema.ListResourcesResult{
		Resources: resources,
	}, nil
}

// ListResourceTemplates returns method-not-found by default.
func (d *DefaultServer) ListResourceTemplates(ctx context.Context, request *schema.ListResourceTemplatesRequest) (*schema.ListResourceTemplatesResult, *jsonrpc.Error) {
	// Return list of registered resource templates
	templates := d.ListRegisteredResourceTemplates()
	return &schema.ListResourceTemplatesResult{
		ResourceTemplates: templates,
	}, nil
}

// ReadResource returns method-not-found by default.
func (d *DefaultServer) ReadResource(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
	// Delegate to registered resource handler
	handler, ok := d.getResourceHandler(request.Params.Uri)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("resource %v not found", request.Params.Uri), nil)
	}
	return handler(ctx, request)
}

// Subscribe adds the URI to the subscription map.
func (d *DefaultServer) Subscribe(ctx context.Context, request *schema.SubscribeRequest) (*schema.SubscribeResult, *jsonrpc.Error) {
	d.Subscription.Put(request.Params.Uri, true)
	return &schema.SubscribeResult{}, nil
}

// Unsubscribe removes the URI from the subscription map.
func (d *DefaultServer) Unsubscribe(ctx context.Context, request *schema.UnsubscribeRequest) (*schema.UnsubscribeResult, *jsonrpc.Error) {
	d.Subscription.Delete(request.Params.Uri)
	return &schema.UnsubscribeResult{}, nil
}

// ListTools returns method-not-found by default.
func (d *DefaultServer) ListTools(ctx context.Context, request *schema.ListToolsRequest) (*schema.ListToolsResult, *jsonrpc.Error) {
	// Return the list of registered tools
	tools := d.ListRegisteredTools()
	return &schema.ListToolsResult{
		Tools: tools,
	}, nil
}

// CallTool returns method-not-found by default.
func (d *DefaultServer) CallTool(ctx context.Context, request *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error) {
	// Delegate to the registered tool handler
	handler, ok := d.getToolHandler(request.Params.Name)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("tool %v not found", request.Params.Name), nil)
	}
	return handler(ctx, request)
}

// Complete returns method-not-found by default.
func (d *DefaultServer) Complete(ctx context.Context, request *schema.CompleteRequest) (*schema.CompleteResult, *jsonrpc.Error) {
	return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("method %v not found", request.Method), nil)
}

// OnNotification is a no-op by default.
func (d *DefaultServer) OnNotification(ctx context.Context, notification *jsonrpc.Notification) {
}

// ListPrompts lists all registered prompts on this DefaultServer.
func (d *DefaultServer) ListPrompts(ctx context.Context, request *schema.ListPromptsRequest) (*schema.ListPromptsResult, *jsonrpc.Error) {
	result := &schema.ListPromptsResult{}
	for _, entry := range d.Prompts.Values() {
		result.Prompts = append(result.Prompts, *entry.Prompt)
	}
	return result, nil
}

// GetPrompt returns the result of a prompt call.
func (d *DefaultServer) GetPrompt(ctx context.Context, request *schema.GetPromptRequest) (*schema.GetPromptResult, *jsonrpc.Error) {
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

// Implements returns true for supported methods.
func (d *DefaultServer) Implements(method string) bool {
	has, _ := d.Methods.Get(method)
	return has
}

// NewDefaultServer creates a new DefaultServer with initialized registries.
// You can then call RegisterResource, RegisterTool, etc., on it before running the server.
func NewDefaultServer(notifier transport.Notifier, logger logger.Logger, client client.Operations) *DefaultServer {
	return &DefaultServer{
		Notifier:     notifier,
		Logger:       logger,
		Client:       client,
		Subscription: syncmap.NewMap[string, bool](),
		Registry:     NewRegistry(),
	}
}

func WithDefaultServer(ctx context.Context, options ...Option) NewServer {
	return func(ctx context.Context, notifier transport.Notifier, logger logger.Logger, client client.Operations) (Server, error) {
		implementer := NewDefaultServer(notifier, logger, client)
		for _, option := range options {
			if err := option(implementer); err != nil {
				return nil, err
			}
		}
		return implementer, nil
	}
}
