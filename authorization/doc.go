// Package authorization defines data structures that describe an MCP server’s
// authorization requirements as well as helper types for carrying OAuth2 / OIDC
// credentials.
//
// The package can be used by both servers and clients:
//   - Servers declare fine-grained policies (per-tool or per-resource) or a
//     single global policy using Policy and Authorization types.
//   - Clients can attach a bearer or ID token to a request via Token and pass
//     it through context using the TokenKey constant.
//
// These types mirror the corresponding sections of the Model Context Protocol
// specification, allowing implementations to share a common, strongly-typed
// representation.
package authorization
