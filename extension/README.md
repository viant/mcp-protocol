# MCP Protocol Extensions

This package hosts optional/shared helpers that are **not** part of the formal
Model Context Protocol specification but are useful for interoperable behaviors
across clients/servers and tools.

Current extensions:

- `continuation`: a reusable pagination/truncation hint (fields for hasMore,
  remaining/returned, nextRange/pageToken, mode, binary) that tools can attach
  to responses when content is clipped or paged.

Keeping extensions out of `schema/` avoids conflating experimental/optional
fields with the official MCP schema.
