// Package namespaces defines optional catalog tools shared by MCP applications.
// These are ordinary registered tools, not additions to the MCP wire schema.
package namespaces

const (
	ListToolName        = "namespaces/list"
	GetToolName         = "namespaces/get"
	// ResourceGetToolName is a legacy ordinary tool name. New publishers and
	// clients use the standard MCP resources/read method instead.
	ResourceGetToolName = "resource/get"
)

type Namespace struct {
	URI         string `json:"uri" yaml:"uri"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type Resource struct {
	URI         string `json:"uri" yaml:"uri"`
	Kind        string `json:"kind" yaml:"kind"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty" yaml:"mimeType,omitempty"`
}

type ListRequest struct {
	Cursor *string `json:"cursor,omitempty" yaml:"cursor,omitempty"`
}

type ListResult struct {
	Namespaces []Namespace `json:"namespaces" yaml:"namespaces"`
	NextCursor *string     `json:"nextCursor,omitempty" yaml:"nextCursor,omitempty"`
}

type GetRequest struct {
	URI    string  `json:"uri" yaml:"uri"`
	Kind   string  `json:"kind,omitempty" yaml:"kind,omitempty"`
	Cursor *string `json:"cursor,omitempty" yaml:"cursor,omitempty"`
}

// GetResult contains resource summaries, not full definitions. Absence of
// NextCursor marks the end of this caller's accessible, filtered collection.
type GetResult struct {
	Resources  []Resource `json:"resources" yaml:"resources"`
	NextCursor *string    `json:"nextCursor,omitempty" yaml:"nextCursor,omitempty"`
	Namespace  *Namespace `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ResourceGetRequest is retained for legacy ordinary tool adapters. Standard
// MCP clients select resource contents with resources/read and params.uri.
type ResourceGetRequest struct {
	URI string `json:"uri" yaml:"uri"`
}
