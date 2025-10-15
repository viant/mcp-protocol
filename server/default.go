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

// DefaultHandler provides default implementations for server-side methods.
// You can embed this in your own Handler and register tools/resources via its helper methods.
type DefaultHandler struct {
	Notifier           transport.Notifier
	Logger             logger.Logger
	Client             client.Operations
	ClientInitialize   *schema.InitializeRequestParams
	Subscription       *syncmap.Map[string, bool]
	ServerCapabilities *schema.ServerCapabilities
	*Registry
}

// Initialize stores the initialization parameters.
func (d *DefaultHandler) Initialize(ctx context.Context, init *schema.InitializeRequestParams, result *schema.InitializeResult) {
	d.ClientInitialize = init
	d.Client.Init(ctx, &d.ClientInitialize.Capabilities)
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
func (d *DefaultHandler) ListResources(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListResourcesRequest]) (*schema.ListResourcesResult, *jsonrpc.Error) {
	// Return list of registered resources
	resources := d.ListRegisteredResources()
	return &schema.ListResourcesResult{
		Resources: resources,
	}, nil
}

// ListResourceTemplates returns method-not-found by default.
func (d *DefaultHandler) ListResourceTemplates(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListResourceTemplatesRequest]) (*schema.ListResourceTemplatesResult, *jsonrpc.Error) {
	// Return list of registered resource templates
	templates := d.ListRegisteredResourceTemplates()
	return &schema.ListResourceTemplatesResult{
		ResourceTemplates: templates,
	}, nil
}

// ReadResource returns method-not-found by default.
func (d *DefaultHandler) ReadResource(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.ReadResourceRequest]) (*schema.ReadResourceResult, *jsonrpc.Error) {
	request := jRequest.Request
	// Delegate to registered resource handler
	handler, ok := d.getResourceHandler(request.Params.Uri)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("resource %v not found", request.Params.Uri), nil)
	}
	return handler(ctx, request)
}

// Subscribe adds the URI to the subscription map.
func (d *DefaultHandler) Subscribe(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.SubscribeRequest]) (*schema.SubscribeResult, *jsonrpc.Error) {
	request := jRequest.Request

	d.Subscription.Put(request.Params.Uri, true)
	return &schema.SubscribeResult{}, nil
}

// Unsubscribe removes the URI from the subscription map.
func (d *DefaultHandler) Unsubscribe(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.UnsubscribeRequest]) (*schema.UnsubscribeResult, *jsonrpc.Error) {
	request := jRequest.Request
	d.Subscription.Delete(request.Params.Uri)
	return &schema.UnsubscribeResult{}, nil
}

// ListTools returns method-not-found by default.
func (d *DefaultHandler) ListTools(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.ListToolsRequest]) (*schema.ListToolsResult, *jsonrpc.Error) {
	// Return the list of registered tools
	tools := d.ListRegisteredTools()
	if d.ClientInitialize == nil {
		return nil, &jsonrpc.Error{Code: jsonrpc.InternalError, Message: "uninilalized"}
	}
	if !schema.IsProtocolNewer(d.ClientInitialize.ProtocolVersion, "2025-03-26") {
		//needs to clean output schema, it was introduced after version "2025-03-26"
		for i := range tools {
			tool := &tools[i]
			tool.OutputSchema = nil
		}
	}
	return &schema.ListToolsResult{
		Tools: tools,
	}, nil
}

// CallTool returns method-not-found by default.
func (d *DefaultHandler) CallTool(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.CallToolRequest]) (*schema.CallToolResult, *jsonrpc.Error) {
	// Delegate to the registered tool handler
	request := jRequest.Request
	handler, ok := d.getToolHandler(request.Params.Name)
	if !ok {
		return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("tool %v not found", request.Params.Name), nil)
	}
	return handler(ctx, request)
}

// Complete returns method-not-found by default.
func (d *DefaultHandler) Complete(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.CompleteRequest]) (*schema.CompleteResult, *jsonrpc.Error) {
	request := jRequest.Request
	return nil, jsonrpc.NewMethodNotFound(fmt.Sprintf("method %v not found", request.Method), nil)
}

// OnNotification is a no-op by default.
func (d *DefaultHandler) OnNotification(ctx context.Context, notification *jsonrpc.Notification) {
}

// ListPrompts lists all registered prompts on this DefaultHandler.
func (d *DefaultHandler) ListPrompts(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.ListPromptsRequest]) (*schema.ListPromptsResult, *jsonrpc.Error) {
	result := &schema.ListPromptsResult{}
	for _, entry := range d.Prompts.Values() {
		result.Prompts = append(result.Prompts, *entry.Prompt)
	}
	return result, nil
}

// GetPrompt returns the result of a prompt call.
func (d *DefaultHandler) GetPrompt(ctx context.Context, jRequest *jsonrpc.TypedRequest[*schema.GetPromptRequest]) (*schema.GetPromptResult, *jsonrpc.Error) {
	request := jRequest.Request
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
func (d *DefaultHandler) Implements(method string) bool {
	has, _ := d.Methods.Get(method)
	return has
}

// NewDefaultHandler creates a new DefaultHandler with initialized registries.
// You can then call RegisterResource, RegisterTool, etc., on it before running the server.
func NewDefaultHandler(notifier transport.Notifier, logger logger.Logger, client client.Operations) *DefaultHandler {
	return &DefaultHandler{
		Notifier:     notifier,
		Logger:       logger,
		Client:       client,
		Subscription: syncmap.NewMap[string, bool](),
		Registry:     NewRegistry(),
	}
}

func WithDefaultHandler(ctx context.Context, options ...Option) NewHandler {
	return func(ctx context.Context, notifier transport.Notifier, logger logger.Logger, client client.Operations) (Handler, error) {
		implementer := NewDefaultHandler(notifier, logger, client)
		for _, option := range options {
			if err := option(implementer); err != nil {
				return nil, err
			}
		}
		return implementer, nil
	}
}
