// Package client declares the interfaces that make up the *client side* of the
// Model Context Protocol (MCP).
//
// The core element is Operations, a collection of strongly-typed methods that
// mirror the JSON-RPC requests defined by the MCP specification (e.g. roots
// listing, sampling, user interaction, etc.).  Implementations embed or
// implement Operations to gain compile-time safety when calling an MCP server.
//
// In addition, the Handler interface extends Operations with the ability to
// receive asynchronous JSON-RPC notifications via the OnNotification hook.
//
// The package contains *interfaces only* and purposefully holds no concrete
// implementation so that different transports (HTTP, stdio, WebSockets, …) can
// provide their own clients while sharing the same contract.
package client
