// Package server contains the *protocol-level* abstractions that make up the
// “server side” of the Model Context Protocol (MCP).
//
// The MCP specification itself calls this role “server”.  Within the code we
// expose a Handler interface (plus DefaultHandler) to emphasise that this
// component handles already-decoded JSON-RPC requests, while a separate
// transport layer (in another module) is responsible for listening on HTTP,
// stdio, WebSockets, etc.
//
// A Handler typically embeds DefaultHandler and selectively overrides the
// Operations it needs.  The package is intentionally transport-agnostic so it
// can be reused by different listener implementations.
package server
