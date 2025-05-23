package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
	"reflect"
)

// ToolHandlerFunc defines a function to handle a tool call.
type ToolHandlerFunc func(ctx context.Context, request *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error)

// ToolEntry holds a handler with its metadata.
type ToolEntry struct {
	Handler  ToolHandlerFunc
	Metadata schema.Tool
}

// Tools is a collection of ToolEntry
type Tools []*ToolEntry

// RegisterToolWithSchema registers a tool with name, description, input schema, and handler on this Base.
// The tool will be advertised to clients with the provided metadata.
func (d *Registry) RegisterToolWithSchema(name string, description string, inputSchema schema.ToolInputSchema, handler ToolHandlerFunc) {
	d.RegisterTool(&ToolEntry{
		Handler: handler,
		Metadata: schema.Tool{
			Name:        name,
			Description: &description,
			InputSchema: inputSchema,
		},
	})
}

// RegisterTool registers a tool with name, description, input schema, and handler
func (d *Registry) RegisterTool(entry *ToolEntry) {
	d.Methods.Put(schema.MethodToolsList, true)
	d.Methods.Put(schema.MethodToolsCall, true)
	d.ToolRegistry.Put(entry.Metadata.Name, entry)
}

// ListRegisteredTools returns metadata for all registered tools on this Base.
func (d *Registry) ListRegisteredTools() []schema.Tool {
	var tools []schema.Tool
	d.ToolRegistry.Range(func(_ string, entry *ToolEntry) bool {
		tools = append(tools, entry.Metadata)
		return true
	})
	return tools
}

// getToolHandler retrieves the handler for a registered tool on this Base.
func (d *Registry) getToolHandler(name string) (ToolHandlerFunc, bool) {
	entry, ok := d.ToolRegistry.Get(name)
	if !ok {
		return nil, false
	}
	return entry.Handler, true
}

// RegisterTool registers a tool on this Base by deriving its input schema from a struct type.
// Handler receives a typed input value and returns a CallToolResult.
func RegisterTool[I any](registry *Registry, name string, description string, handler func(ctx context.Context, input I) (*schema.CallToolResult, *jsonrpc.Error)) error {
	// Derive input schema from struct type I

	var sample I
	var inputSchema schema.ToolInputSchema
	sampleType := reflect.TypeOf(sample)
	if sampleType.Kind() == reflect.Pointer {
		if err := inputSchema.Load(sample); err != nil {
			return fmt.Errorf("failed to derive input schema for tool %s: %w", name, err)
		}
	} else {
		if err := inputSchema.Load(&sample); err != nil {
			return fmt.Errorf("failed to derive input schema for tool %s: %w", name, err)
		}
	}

	// Wrap handler to unmarshal arguments into typed struct
	wrapped := func(ctx context.Context, request *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error) {
		var input I
		if args := request.Params.Arguments; args != nil {
			data, err := json.Marshal(args)
			if err != nil {
				return nil, jsonrpc.NewError(jsonrpc.InvalidParams, err.Error(), nil)
			}
			if err := json.Unmarshal(data, &input); err != nil {
				return nil, jsonrpc.NewError(jsonrpc.InvalidParams, err.Error(), nil)
			}
		}
		return handler(ctx, input)
	}
	// Register with metadata and wrapped handler on this Base
	registry.RegisterToolWithSchema(name, description, inputSchema, wrapped)
	return nil
}
