# Namespace catalog extension

These are optional ordinary MCP tools, not changes to standard MCP methods.

| Tool | Request | Response |
| --- | --- | --- |
| `namespaces/list` | `cursor?` | `namespaces: [{uri,name,description?}], nextCursor?` |
| `namespaces/get` | `uri: ns://<id>, kind?, cursor?` | `resources: [{uri,kind,name,description?,mimeType?}], nextCursor?, namespace?` |
| `resource/get` | `uri` | The selected resource's JSON definition |

Omitting `kind` lists all accessible resource kinds in the namespace. The owning
MCP applies authorization and kind filtering **before** pagination. It sorts
summaries consistently (normally name, then URI). Clients treat cursors as opaque
and stop when `nextCursor` is absent. Root discovery servers can implement only
the two namespace tools; no `windows/list` or `windows/get` prerequisite exists.

Namespace IDs are opaque. `URI` and `ParseURI` preserve them without imposing a
hexadecimal, UUID or DNS-name format. Collection HTTP paths such as
`/namespaces/<id>/<kind>` are optional host adapters, not namespace references.

`Paginate` is an optional generic helper matching the existing skills catalog's
snapshot/offset convention. Its input must already be the caller's accessible,
filtered, ordered summaries. It adds no authorization or credential policy.
Servers may supply their own opaque cursor implementation instead.

Resource schemes identify resource-defined contracts, for example
`skill://<namespace>/<path>` and `mcp://<namespace>/<name>`. An MCP configuration
resource can return JSON compatible with its consumer's configuration model;
that model is deliberately not imported here. Publish authored credential
references, never resolved tokens, passwords or key values.
