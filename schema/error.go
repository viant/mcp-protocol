package schema

import "github.com/viant/jsonrpc"

const (
   // Unauthorized indicates authentication is required
   Unauthorized = -32001
   // ResourceNotFound indicates the requested resource was not found
   ResourceNotFound = -32002
)

// NewInvalidPromptName creates a new invalid prompt name
func NewInvalidPromptName(name string) *jsonrpc.Error {
	return jsonrpc.NewError(ResourceNotFound, "Invalid prompt name: "+name, nil)
}

// NewResourceNotFound creates a new resource not found
func NewResourceNotFound(uri string) *jsonrpc.Error {
	return jsonrpc.NewError(ResourceNotFound, "Resource not found", map[string]interface{}{"uri": uri})
}

func NewUnknownTool(toolName string) *jsonrpc.Error {
	return jsonrpc.NewError(jsonrpc.InvalidParams, "Unknown tool:"+toolName, nil)
}
