package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const SkillsExtension = "io.modelcontextprotocol/skills"
const MethodSkillsList = "skills/list"
const MethodSkillsGet = "skills/get"

// Skill is SEP-2640 metadata, not file content or an execution permission.
type Skill struct {
	Uri         string                 `json:"uri"`
	Frontmatter map[string]interface{} `json:"frontmatter"`
	Resources   SkillResources         `json:"resources"`
}

func (s *Skill) UnmarshalJSON(raw []byte) error {
	type plain Skill
	var required map[string]json.RawMessage
	if err := json.Unmarshal(raw, &required); err != nil {
		return err
	}
	for _, key := range []string{"uri", "frontmatter", "resources"} {
		if len(required[key]) == 0 || bytes.Equal(required[key], []byte("null")) {
			return fmt.Errorf("skill %s is required", key)
		}
	}
	var decoded plain
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if decoded.Uri == "" || decoded.Frontmatter == nil {
		return fmt.Errorf("skill URI and frontmatter are required")
	}
	*s = Skill(decoded)
	return nil
}

type SkillResource struct {
	Uri    string `json:"uri"`
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

func (s *SkillResource) UnmarshalJSON(raw []byte) error {
	type plain SkillResource
	var required map[string]json.RawMessage
	if err := json.Unmarshal(raw, &required); err != nil {
		return err
	}
	for _, key := range []string{"uri", "digest", "size"} {
		if len(required[key]) == 0 || bytes.Equal(required[key], []byte("null")) {
			return fmt.Errorf("skill resource %s is required", key)
		}
	}
	var decoded plain
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	if decoded.Uri == "" || decoded.Digest == "" || decoded.Size < 0 {
		return fmt.Errorf("invalid skill resource")
	}
	*s = SkillResource(decoded)
	return nil
}

// SkillResources represents the extension's array-or-"dynamic" union.
type SkillResources struct {
	Files   []SkillResource
	Dynamic bool
}

func (r SkillResources) MarshalJSON() ([]byte, error) {
	if r.Dynamic {
		if len(r.Files) != 0 {
			return nil, fmt.Errorf("dynamic skill cannot have a static inventory")
		}
		return []byte(`"dynamic"`), nil
	}
	if r.Files == nil {
		return nil, fmt.Errorf("skill resources inventory is required")
	}
	return json.Marshal(r.Files)
}

func (r *SkillResources) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte(`"dynamic"`)) {
		*r = SkillResources{Dynamic: true}
		return nil
	}
	if len(data) == 0 || data[0] != '[' {
		return fmt.Errorf("skill resources must be an array or dynamic")
	}
	var files []SkillResource
	if err := json.Unmarshal(data, &files); err != nil {
		return err
	}
	*r = SkillResources{Files: files}
	return nil
}

type ListSkillsRequest struct {
	Method string                  `json:"method"`
	Params ListSkillsRequestParams `json:"params"`
}
type ListSkillsRequestParams struct {
	Cursor *string                `json:"cursor,omitempty"`
	Meta   map[string]interface{} `json:"_meta,omitempty"`
}

func (p *ListSkillsRequestParams) UnmarshalJSON(raw []byte) error {
	type plain ListSkillsRequestParams
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if cursor, ok := fields["cursor"]; ok && bytes.Equal(bytes.TrimSpace(cursor), []byte("null")) {
		return fmt.Errorf("cursor must be a string")
	}
	var decoded plain
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	*p = ListSkillsRequestParams(decoded)
	return nil
}

type GetSkillRequest struct {
	Method string                `json:"method"`
	Params GetSkillRequestParams `json:"params"`
}
type GetSkillRequestParams struct {
	Uri  string                 `json:"uri"`
	Meta map[string]interface{} `json:"_meta,omitempty"`
}
type ListSkillsResult struct {
	Meta       *ResultMetaObject         `json:"_meta,omitempty"`
	ResultType string                    `json:"resultType,omitempty"`
	Skills     []Skill                   `json:"skills"`
	NextCursor *string                   `json:"nextCursor,omitempty"`
	TtlMs      *int                      `json:"ttlMs,omitempty"`
	CacheScope CacheableResultCacheScope `json:"cacheScope,omitempty"`
}
type GetSkillResult struct {
	TtlMs      *int                      `json:"ttlMs,omitempty"`
	CacheScope CacheableResultCacheScope `json:"cacheScope,omitempty"`
	Meta       *ResultMetaObject         `json:"_meta,omitempty"`
	ResultType string                    `json:"resultType,omitempty"`
	Skill      Skill                     `json:"skill"`
}
