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

// DefaultImplementer provides default implementations for server-side methods.
// You can embed this in your own Implementer and register tools/resources via its helper methods.
type DefaultImplementer struct {
	Notifier         transport.Notifier
	Logger           logger.Logger
	Client           client.Operations
	ClientInitialize *schema.InitializeRequestParams
	Subscription     *syncmap.Map[string, bool]
	// ToolRegistry holds per-instance registered tools and handlers.
	ToolRegistry             *syncmap.Map[string, *ToolEntry]
	ResourceRegistry         *syncmap.Map[string, *ResourceEntry]
	ResourceTemplateRegistry *syncmap.Map[string, *ResourceTemplateEntry]
	Methods                  *syncmap.Map[string, bool]
}

// Initialize stores the initialization parameters.
func (d *DefaultImplementer) Initialize(ctx context.Context, init *schema.InitializeRequestParams, result *schema.InitializeResult) {
	d.ClientInitialize = init
	if d.ToolRegistry.Size() > 0 {
		result.Capabilities.Tools = &schema.ServerCapabilitiesTools{}
	}
	if d.ResourceRegistry.Size() > 0 {
		result.Capabilities.Resources = &schema.ServerCapabilitiesResources{}
	}
}

// ListResources returns method-not-found by default.
func (d *DefaultImplementer) ListResources(ctx context.Context, request *schema.ListResourcesRequest) (*schema.ListResourcesResult, *jsonrpc.Error) {
	// Return list of registered resources
	resources := d.ListRegisteredResources()
	return &schema.ListResourcesResult{
		Resources: resources,
	}, nil
}

// ListResourceTemplates returns method-not-found by default.
func (d *DefaultImplementer) ListResourceTemplates(ctx context.Context, request *schema.ListResourceTemplatesRequest) (*schema.ListResourceTemplatesResult, *jsonrpc.Error) {
	// Return list of registered resource templates
	templates := d.ListRegisteredResourceTemplates()
	return &schema.ListResourceTemplatesResult{
		ResourceTemplates: templates,
	}, nil
}

// ReadResource returns method-not-found by default.
func (d *DefaultImplementer) ReadResource(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
	// Delegate to registered resource handler
	handler, ok := d.getResourceHandler(request.Params.Uri)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("resource %v not found", request.Params.Uri), nil)
	}
	return handler(ctx, request)
}

// Subscribe adds the URI to the subscription map.
func (d *DefaultImplementer) Subscribe(ctx context.Context, request *schema.SubscribeRequest) (*schema.SubscribeResult, *jsonrpc.Error) {
	d.Subscription.Put(request.Params.Uri, true)
	return &schema.SubscribeResult{}, nil
}

// Unsubscribe removes the URI from the subscription map.
func (d *DefaultImplementer) Unsubscribe(ctx context.Context, request *schema.UnsubscribeRequest) (*schema.UnsubscribeResult, *jsonrpc.Error) {
	d.Subscription.Delete(request.Params.Uri)
	return &schema.UnsubscribeResult{}, nil
}

// ListTools returns method-not-found by default.
func (d *DefaultImplementer) ListTools(ctx context.Context, request *schema.ListToolsRequest) (*schema.ListToolsResult, *jsonrpc.Error) {
	// Return the list of registered tools
	tools := d.ListRegisteredTools()
	return &schema.ListToolsResult{
		Tools: tools,
	}, nil
}

// CallTool returns method-not-found by default.
func (d *DefaultImplementer) CallTool(ctx context.Context, request *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error) {
	// Delegate to the registered tool handler
	handler, ok := d.getToolHandler(request.Params.Name)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("tool %v not found", request.Params.Name), nil)
	}
	return handler(ctx, request)
}

// ListPrompts returns method-not-found by default.
func (d *DefaultImplementer) ListPrompts(ctx context.Context, request *schema.ListPromptsRequest) (*schema.ListPromptsResult, *jsonrpc.Error) {
	return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("method %v not found", request.Method), nil)
}

// GetPrompt returns method-not-found by default.
func (d *DefaultImplementer) GetPrompt(ctx context.Context, request *schema.GetPromptRequest) (*schema.GetPromptResult, *jsonrpc.Error) {
	return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("method %v not found", request.Method), nil)
}

// Complete returns method-not-found by default.
func (d *DefaultImplementer) Complete(ctx context.Context, request *schema.CompleteRequest) (*schema.CompleteResult, *jsonrpc.Error) {
	return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("method %v not found", request.Method), nil)
}

// OnNotification is a no-op by default.
func (d *DefaultImplementer) OnNotification(ctx context.Context, notification *jsonrpc.Notification) {
}

// Implements returns true for supported methods.
func (d *DefaultImplementer) Implements(method string) bool {
	has, _ := d.Methods.Get(method)
	return has
}

// NewDefaultImplementer creates a new DefaultImplementer with initialized registries.
// You can then call RegisterResource, RegisterTool, etc., on it before running the server.
func NewDefaultImplementer(notifier transport.Notifier, logger logger.Logger, client client.Operations) *DefaultImplementer {
	return &DefaultImplementer{
		Notifier:                 notifier,
		Logger:                   logger,
		Client:                   client,
		Subscription:             syncmap.NewMap[string, bool](),
		ToolRegistry:             syncmap.NewMap[string, *ToolEntry](),
		ResourceRegistry:         syncmap.NewMap[string, *ResourceEntry](),
		ResourceTemplateRegistry: syncmap.NewMap[string, *ResourceTemplateEntry](),
		Methods:                  syncmap.NewMap[string, bool](),
	}
}

func WithDefaultImplementer(ctx context.Context, options ...Option) NewImplementer {
	return func(ctx context.Context, notifier transport.Notifier, logger logger.Logger, client client.Operations) (Implementer, error) {
		implementer := NewDefaultImplementer(notifier, logger, client)
		for _, option := range options {
			if err := option(implementer); err != nil {
				return nil, err
			}
		}
		return implementer, nil
	}
}
