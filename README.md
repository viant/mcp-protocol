# MCP (Model Context Protocol) for Go

MCP-Protocol is a Go implementation of the Model Context Protocol — a standardized way for applications to communicate with AI models. It allows developers to seamlessly bridge applications and AI models using a lightweight, JSON-RPC–based protocol.

This repository contains the shared protocol definitions and schemas for MCP. It is used by github.com/viant/mcp, which provides the actual implementation of the MCP server and client framework.

[Official Model Context Protocol Specification](https://modelcontextprotocol.io/introduction)

## Overview

MCP (Model Context Protocol) is designed to provide a standardized communication layer between applications and AI models. The protocol simplifies the integration of AI capabilities into applications by offering a consistent interface for resource access, prompt management, model interaction, and tool invocation.

Key features:
- JSON-RPC 2.0–based communication
- Support for multiple transport protocols (HTTP/SSE, stdio)
- Server-side features:
  - Resource management
  - Model prompting and completion
  - Tool invocation
  - Subscriptions for resource updates
  - Logging
  - Progress reporting
  - Request cancellation
- Client-side features:
  - Roots
  - Sampling

## Module Structure

`github.com/viant/mcp-protocol` is the Go module containing the shared Model Context Protocol (MCP) contracts:

- **schema**: JSON-RPC request, result, and notification types generated from the MCP JSON schema.
- **server**: `server.Operations`, `server.Implementer` interfaces and
  `server.DefaultImplementer` default implementer with no-op stubs.
- **client**: `client.Operations`, `client.Implementer` interfaces for MCP clients.
- **logger**: logging interface (`Logger`) for implementers to emit JSON-RPC notifications.
- **oauth2**: defines meta information for OAuth2 authorization and authentication flows.
- **authorization**: authentication definition for global and fine grain resource/tool level authorization

## Quick Start

### Creating MCP server

```go
package example

import (
  "context"
  "fmt"
  "github.com/viant/jsonrpc"
  "github.com/viant/mcp-protocol/schema"
  serverproto "github.com/viant/mcp-protocol/server"
  "github.com/viant/mcp/server"
  "log"
)

func Usage_Example() {

  newImplementer := serverproto.WithDefaultImplementer(context.Background(), func(implementer *serverproto.DefaultImplementer) {
    // Register a simple resource
    implementer.RegisterResource(schema.Resource{Name: "hello", Uri: "/hello"},
      func(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
        return &schema.ReadResourceResult{Contents: []schema.ReadResourceResultContentsElem{{Text: "Hello, world!"}}}, nil
      })

    type Addition struct {
      A int `json:"a"`
      B int `json:"b"`
    }
    // Register a simple calculator tool: adds two integers
    if err := serverproto.RegisterTool[*Addition](implementer, "add", "Add two integers", func(ctx context.Context, input *Addition) (*schema.CallToolResult, *jsonrpc.Error) {
      sum := input.A + input.B
      return &schema.CallToolResult{Content: []schema.CallToolResultContentElem{{Text: fmt.Sprintf("%d", sum)}}}, nil
    }); err != nil {
      panic(err)
    }
  })

  srv, err := server.New(
    server.WithNewImplementer(newImplementer),
    server.WithImplementation(schema.Implementation{"default", "1.0"}),
  )
  if err != nil {
    log.Fatalf("Failed to create server: %v", err)
  }

  log.Fatal(srv.HTTP(context.Background(), ":4981").ListenAndServe())
}
```

### Creating Custom Implementer

You can create custom implementer to extend the default behavior:

```go
package myimpl

import (
  "context"
  "github.com/viant/jsonrpc"
  "github.com/viant/jsonrpc/transport"
  "github.com/viant/mcp-protocol/client"
  "github.com/viant/mcp-protocol/logger"
  "github.com/viant/mcp-protocol/schema"
  "github.com/viant/mcp-protocol/server"
)

// MyImplementer is a sample MCP implementer embedding the default Base.
type MyImplementer struct {
  *server.DefaultImplementer
}

// ListResources implements the resources/list method.
func (i *MyImplementer) ListResources(
        ctx context.Context,
        req *schema.ListResourcesRequest,
) (*schema.ListResourcesResult, *jsonrpc.Error) {
  // Custom implementation
  return &schema.ListResourcesResult{}, nil
}

// Implements indicates which methods this implementer supports.
func (i *MyImplementer) Implements(method string) bool {
  switch method {
  case schema.MethodResourcesList, schema.MethodResourcesRead:
    return true
  }
  return i.DefaultImplementer.Implements(method)
}

func (i *MyImplementer) ReadResource(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
  // Custom implementation
  return &schema.ReadResourceResult{}, nil
}

// NewMyImplementer returns a factory for MyImplementer.
func NewMyImplementer() server.NewImplementer {
  return func(
          ctx context.Context,
          notifier transport.Notifier,
          log logger.Logger,
          aClient client.Operations,
  ) server.Implementer {
    base := server.NewDefaultImplementer(notifier, log, aClient)
    return &MyImplementer{DefaultImplementer: base}
  }
}

```

## Example: Comprehensive Custom Implementer
Use the [example/custom](https://github.com/viant/mcp/example/custom) package for a more advanced implementer with polling, notifications, and resource watching:
```go
package main

import (
    "context"
    "embed"
    "log"

    "github.com/viant/afs/storage"
    "github.com/viant/mcp/example/custom"
    "github.com/viant/mcp/server"
    "github.com/viant/mcp-protocol/schema"
)

//go:embed data/*
var embedFS embed.FS

func main() {
    config := &custom.Config{
        BaseURL: "embed://data",
        Options: []storage.Option{embedFS},
    }
    newImplementer := custom.New(config)
    srv, err := server.New(
        server.WithNewImplementer(newImplementer),
        server.WithImplementation(schema.Implementation{"custom", "1.0"})
    )
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }
    log.Fatal(srv.HTTP(context.Background(), ":4981").ListenAndServe())
}
```


## Key Components

### JSON-RPC Methods

The protocol defines standard methods for communication:

```go
// JSON-RPC method names for MCP protocol
const (
    MethodInitialize                  = "initialize"
    MethodPing                        = "ping"
    MethodResourcesList               = "resources/list"
    MethodResourcesTemplatesList      = "resources/templates/list"
    MethodResourcesRead               = "resources/read"
    MethodSubscribe                   = "resources/subscribe"
    MethodUnsubscribe                 = "resources/unsubscribe"
    MethodPromptsList                 = "prompts/list"
    MethodPromptsGet                  = "prompts/get"
    MethodToolsList                   = "tools/list"
    MethodToolsCall                   = "tools/call"
    MethodComplete                    = "complete"
    MethodLoggingSetLevel             = "logging/setLevel"
    MethodNotificationInitialized     = "notification/initialized"
    MethodNotificationMessage         = "notification/message"
    MethodNotificationCancel          = "notification/cancel"
    MethodNotificationResourceUpdated = "notifications/resources/updated"
    MethodRootsList                   = "roots/list"
    MethodSamplingCreateMessage       = "sampling/createMessage"
)
```

### Authorization

The protocol supports OAuth2 (official MCP spect) and fine-grained authorization (experimental):

```go
// Config holds OAuth2/OIDC configuration for fine-grained control.
type Config struct {
    // Global resource protection metadata
    Global *Authorization `json:"global,omitempty"`
    // ExcludeURI skips middleware on matching paths
    ExcludeURI string `json:"excludeURI,omitempty"`
    // Per-tool authorization metadata
    Tools map[string]*Authorization `json:"tools,omitempty"`
    // Per-tenant authorization metadata
    Tenants map[string]*Authorization `json:"tenants,omitempty"`
}
```

## Contributing

Contributions welcome—please fork and submit a Pull Request.

## Credits

Author: Adrian Witas

This project is maintained by [Viant](https://github.com/viant).

## License

Apache License 2.0. See the LICENSE file in the project root.
