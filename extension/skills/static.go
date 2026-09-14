package skills

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"
)

// Static can only be initialized by Compiler. Metadata and readers share these
// sealed bytes; neither the source filesystem nor caller-owned manifests survive.
type Static struct {
	entry     schema.Skill
	files     map[string]string
	root      string
	resources []schema.Resource
	mimes     map[string]string
}

func (s *Static) Metadata() schema.Skill {
	if s == nil || s.files == nil {
		return schema.Skill{}
	}
	result, _ := Clone(s.entry)
	return result
}

func (s *Static) Contains(uri string) bool {
	return s != nil && s.root != "" && (uri == s.root || strings.HasPrefix(uri, s.root+"/"))
}

func (s *Static) initMetadata() {
	result := make([]schema.Resource, 0, len(s.files))
	s.mimes = map[string]string{}
	for _, file := range s.entry.Resources.Files {
		parsed, _ := url.Parse(file.Uri)
		kind := mime.TypeByExtension(path.Ext(parsed.Path))
		if path.Ext(parsed.Path) == ".md" {
			kind = "text/markdown"
		}
		if kind == "" {
			kind = http.DetectContentType([]byte(s.files[file.Uri]))
		}
		size := int(file.Size)
		resource := schema.Resource{Uri: file.Uri, Name: file.Uri, MimeType: &kind, Size: &size}
		if file.Uri == s.entry.Uri {
			name, description := s.entry.Frontmatter["name"].(string), s.entry.Frontmatter["description"].(string)
			resource.Title = &name
			resource.Description = &description
		}
		result = append(result, resource)
		s.mimes[file.Uri] = kind
	}
	s.resources = result
}

func (s *Static) Resources() []schema.Resource {
	if s == nil {
		return nil
	}
	result := make([]schema.Resource, len(s.resources))
	for i, resource := range s.resources {
		result[i] = resource
		if resource.Size != nil {
			size := *resource.Size
			result[i].Size = &size
		}
		for _, field := range []struct {
			source *string
			target **string
		}{{resource.MimeType, &result[i].MimeType}, {resource.Title, &result[i].Title}, {resource.Description, &result[i].Description}} {
			if field.source != nil {
				value := *field.source
				*field.target = &value
			}
		}
	}
	return result
}

func (s *Static) ReadResource(ctx context.Context, request *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
	if ctx == nil || request == nil {
		return nil, jsonrpc.NewInvalidParamsError("resource request and context are required", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, jsonrpc.NewInternalError(err.Error(), nil)
	}
	data, ok := s.files[request.Params.Uri]
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("unknown static skill resource", nil)
	}
	content := schema.ReadResourceResultContentsElem{Uri: request.Params.Uri}
	kind := s.mimes[content.Uri]
	content.MimeType = &kind
	if utf8.ValidString(data) && content.MimeType != nil && (strings.HasPrefix(*content.MimeType, "text/") || strings.Contains(*content.MimeType, "json") || strings.Contains(*content.MimeType, "yaml")) {
		content.Text = data
	} else {
		content.Blob = base64.StdEncoding.EncodeToString([]byte(data))
	}
	return &schema.ReadResourceResult{Contents: []schema.ReadResourceResultContentsElem{content}}, nil
}

func (s *Static) Validate() error {
	if s == nil || s.files == nil || s.root == "" {
		return fmt.Errorf("compiler-backed static skill is required")
	}
	return nil
}
