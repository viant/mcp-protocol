package skills

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"sort"
	"strings"

	"github.com/viant/mcp-protocol/schema"
)

// Compiler seals a complete filesystem tree into immutable byte authority.
type Compiler struct{ Source fs.FS }

func (c Compiler) Compile(ctx context.Context, uri string) (*Static, error) {
	if ctx == nil || c.Source == nil {
		return nil, fmt.Errorf("skill context and source are required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	files := map[string]string{}
	var total int64
	err := fs.WalkDir(c.Source, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("skill symlink %q", name)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("skill non-file %q", name)
		}
		data, err := c.readFile(name)
		if err != nil {
			return err
		}
		total += int64(len(data))
		if len(files) >= 512 || total > 16*1024*1024 {
			return fmt.Errorf("skill exceeds SEP-2640 inventory limits")
		}
		files[name] = string(data)
		return nil
	})
	if err != nil {
		return nil, err
	}
	data, ok := files["SKILL.md"]
	if !ok {
		return nil, fmt.Errorf("skill entrypoint missing")
	}
	front, err := Frontmatter([]byte(data))
	if err != nil {
		return nil, err
	}
	root, err := Root(uri, front["name"].(string))
	if err != nil {
		return nil, err
	}
	base, _ := url.Parse(root)
	result := &Static{entry: schema.Skill{Uri: uri, Frontmatter: front, Resources: schema.SkillResources{Files: []schema.SkillResource{}}}, files: map[string]string{}, root: root}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data := files[name]
		u := *base
		u.Path += "/" + name
		u.RawPath = ""
		result.files[u.String()] = data
		result.entry.Resources.Files = append(result.entry.Resources.Files, schema.SkillResource{Uri: u.String(), Digest: fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(data))), Size: int64(len(data))})
	}
	result.initMetadata()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (c Compiler) readFile(name string) ([]byte, error) {
	file, err := c.Source.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16*1024*1024 {
		return nil, fmt.Errorf("invalid or oversized skill file %q", name)
	}
	data, err := io.ReadAll(io.LimitReader(file, 16*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 16*1024*1024 {
		return nil, fmt.Errorf("oversized skill file %q", name)
	}
	return data, nil
}

// Validate checks wire metadata shape. It is not static registration authority;
// only Compiler's sealed Static value can supply that authority.
func Validate(entry schema.Skill) error {
	if err := ValidateFrontmatter(entry.Frontmatter); err != nil {
		return err
	}
	root, err := Root(entry.Uri, entry.Frontmatter["name"].(string))
	if err != nil {
		return err
	}
	if entry.Resources.Dynamic {
		if len(entry.Resources.Files) != 0 {
			return fmt.Errorf("dynamic skill has static files")
		}
		return nil
	}
	if entry.Resources.Files == nil || len(entry.Resources.Files) > 512 {
		return fmt.Errorf("invalid skill inventory")
	}
	seen := map[string]bool{}
	var total int64
	for _, file := range entry.Resources.Files {
		u, err := url.Parse(file.Uri)
		if err != nil || !strings.HasPrefix(file.Uri, root+"/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawPath != "" || strings.ContainsAny(u.Path, "\\%") {
			return fmt.Errorf("file outside skill root")
		}
		rel := strings.TrimPrefix(file.Uri, root+"/")
		decoded, err := url.PathUnescape(rel)
		if err != nil || !fs.ValidPath(decoded) || decoded == "." {
			return fmt.Errorf("invalid skill file path")
		}
		digest, err := hex.DecodeString(strings.TrimPrefix(file.Digest, "sha256:"))
		if err != nil || len(digest) != 32 || file.Digest != "sha256:"+hex.EncodeToString(digest) || file.Size < 0 {
			return fmt.Errorf("invalid skill digest/size")
		}
		if seen[file.Uri] {
			return fmt.Errorf("duplicate skill file")
		}
		seen[file.Uri] = true
		if file.Size > 16*1024*1024-total {
			return fmt.Errorf("skill exceeds size limit")
		}
		total += file.Size
	}
	if !seen[entry.Uri] {
		return fmt.Errorf("skill entrypoint missing from inventory")
	}
	return nil
}

func Clone(entry schema.Skill) (schema.Skill, error) {
	raw, err := json.Marshal(entry)
	if err != nil {
		return schema.Skill{}, err
	}
	var result schema.Skill
	err = json.Unmarshal(raw, &result)
	return result, err
}
