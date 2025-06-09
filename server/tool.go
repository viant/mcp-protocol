package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
)

// ToolHandlerFunc defines a function to handle a tool call.
type ToolHandlerFunc func(ctx context.Context, request *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error)

// ToolEntry holds a handler together with its public metadata.
type ToolEntry struct {
	Handler  ToolHandlerFunc
	Metadata schema.Tool
}

// RegisterToolWithSchema registers a tool with an explicit JSON schema.
func (d *Registry) RegisterToolWithSchema(name, description string, inputSchema schema.ToolInputSchema, outputSchema *schema.ToolOutputSchema, handler ToolHandlerFunc) {
	d.RegisterTool(&ToolEntry{
		Handler: handler,
		Metadata: schema.Tool{
			Name:         name,
			Description:  &description,
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
		},
	})
}

// RegisterTool adds a prepared ToolEntry to the registry.
func (d *Registry) RegisterTool(entry *ToolEntry) {
	d.Methods.Put(schema.MethodToolsList, true)
	d.Methods.Put(schema.MethodToolsCall, true)
	d.ToolRegistry.Put(entry.Metadata.Name, entry)
}

// ListRegisteredTools returns metadata for all registered tools.
func (d *Registry) ListRegisteredTools() []schema.Tool {
	var tools []schema.Tool
	d.ToolRegistry.Range(func(_ string, entry *ToolEntry) bool {
		tools = append(tools, entry.Metadata)
		return true
	})
	return tools
}

// getToolHandler retrieves the handler for a registered tool.
func (d *Registry) getToolHandler(name string) (ToolHandlerFunc, bool) {
	if entry, ok := d.ToolRegistry.Get(name); ok {
		return entry.Handler, true
	}
	return nil, false
}

// RegisterTool derives JSON schemas from the generic I/O types and registers the tool.
func RegisterTool[I any, O any](registry *Registry, name, description string, handler func(ctx context.Context, input I) (*schema.CallToolResult, *jsonrpc.Error)) error {
	var (
		inVar     I
		outVar    O
		inSchema  schema.ToolInputSchema
		outSchema schema.ToolOutputSchema
	)

	// Build input schema
	sampleType := reflect.TypeOf(inVar)
	if sampleType.Kind() == reflect.Pointer {
		if err := inSchema.Load(inVar); err != nil {
			return fmt.Errorf("failed to derive input schema for tool %s: %w", name, err)
		}
	} else {
		if err := inSchema.Load(&inVar); err != nil {
			return fmt.Errorf("failed to derive input schema for tool %s: %w", name, err)
		}
	}

	// Build output schema
	outputType := reflect.TypeOf(outVar)
	if outputType.Kind() == reflect.Pointer {
		if err := outSchema.Load(outVar); err != nil {
			return fmt.Errorf("failed to derive output schema for tool %s: %w", name, err)
		}
	} else {
		if err := outSchema.Load(&outVar); err != nil {
			return fmt.Errorf("failed to derive output schema for tool %s: %w", name, err)
		}
	}

	// Wrap the typed handler so it matches ToolHandlerFunc.
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

	registry.RegisterToolWithSchema(name, description, inSchema, &outSchema, wrapped)
	return nil
}
