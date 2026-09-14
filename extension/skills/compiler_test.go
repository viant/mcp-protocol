package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/mcp-protocol/schema"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSkillFormatAndInventory(t *testing.T) {
	for _, prefix := range []string{"skill://team/demo", "docs://team/domain/demo"} {
		files := fstest.MapFS{
			"SKILL.md":              {Data: []byte("---\nname: demo\ndescription: Try a demo.\nmetadata:\n  version: '1'\nallowed-tools: Bash\nfuture: {enabled: true, values: [1, 2]}\n---\nBody\n")},
			"nested/child/SKILL.md": {Data: []byte("---\nname: child\ndescription: Nested.\n---\nChild")},
			"binary.bin":            {Data: []byte{0, 1, 255}},
		}
		compiled, err := (Compiler{Source: files}).Compile(context.Background(), prefix+"/SKILL.md")
		if err != nil {
			t.Fatal(err)
		}
		entry := compiled.Metadata()
		if err = Validate(entry); err != nil {
			t.Fatal(err)
		}
		if len(entry.Resources.Files) != 3 || entry.Frontmatter["allowed-tools"] != "Bash" || entry.Frontmatter["future"] == nil {
			t.Fatalf("lost fields: %+v", entry)
		}
		copy, err := Clone(entry)
		if err != nil {
			t.Fatal(err)
		}
		copy.Frontmatter["future"].(map[string]interface{})["enabled"] = false
		if entry.Frontmatter["future"].(map[string]interface{})["enabled"] != true {
			t.Fatal("shallow clone")
		}
	}
}

func TestSkillInvalidFormatAndLimits(t *testing.T) {
	for _, front := range []string{"name: Demo\ndescription: demo", "name: de--mo\ndescription: demo", "name: demo", "name: demo\ndescription: 42", "name: demo\ndescription: demo\nmetadata: {version: 1}", "name: demo\ndescription: demo\nname: another", "name: demo\ndescription: demo\nfuture: {1: value}"} {
		if _, err := Frontmatter([]byte("---\n" + front + "\n---\nbody")); err == nil {
			t.Fatalf("accepted %s", front)
		}
	}
	for _, uri := range []string{"skill://other/SKILL.md", "skill://demo", "skill://demo/SKILL.md?", "skill://demo/SKILL.md?q=x", "skill://demo/../demo/SKILL.md", "skill://demo/%53KILL.md", "skill://user@demo/SKILL.md"} {
		if _, err := Root(uri, "demo"); err == nil {
			t.Fatalf("accepted %s", uri)
		}
	}
	for _, raw := range []string{`null`, `{}`, `"static"`, `42`} {
		var r schema.SkillResources
		if json.Unmarshal([]byte(raw), &r) == nil {
			t.Fatalf("accepted resources %s", raw)
		}
	}
	files := fstest.MapFS{"SKILL.md": {Data: []byte("---\nname: demo\ndescription: demo\n---\n")}}
	for i := 0; i < 512; i++ {
		files[fmt.Sprintf("file%d", i)] = &fstest.MapFile{Data: []byte("x")}
	}
	if _, err := (Compiler{Source: files}).Compile(context.Background(), "skill://demo/SKILL.md"); err == nil {
		t.Fatal("accepted 513 files")
	}
	files = fstest.MapFS{"SKILL.md": {Data: []byte("---\nname: demo\ndescription: demo\n---\n")}, "huge": {Data: []byte(strings.Repeat("x", 16*1024*1024))}}
	if _, err := (Compiler{Source: files}).Compile(context.Background(), "skill://demo/SKILL.md"); err == nil {
		t.Fatal("accepted oversized skill")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Compiler{Source: files}).Compile(ctx, "skill://demo/SKILL.md"); err == nil {
		t.Fatal("ignored cancellation")
	}
}
