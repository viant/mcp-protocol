// Package schema hosts the Go types that mirror the *Model Context Protocol*
// (MCP) JSON schema.
//
// All request / response / notification payloads, helper enumerations and
// method constants are generated from the authoritative protocol definition
// and live in this package so that servers and clients can share a single,
// strongly-typed contract layer.
//
// The version of the JSON schema that the generated files correspond to is
// exposed via the LatestProtocolVersion constant.
package schema
