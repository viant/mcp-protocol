// Package server defines the server-side abstractions of the Model Context
// Protocol (MCP).
//
// At its core sits the Operations interface that mirrors the set of JSON-RPC
// methods mandated by the specification (initialize, resources/list, tools/call
// …).  Concrete servers implement Operations and may embed DefaultServer to
// inherit no-op stubs for every method, only overriding what they actually
// support.
//
// Additional helpers make it easy to register new resources or tools and to
// spin up servers over different transports without coupling user code to a
// particular implementation.
package server
