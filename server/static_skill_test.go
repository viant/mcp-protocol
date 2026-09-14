package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/viant/jsonrpc"
	skillformat "github.com/viant/mcp-protocol/extension/skills"
	"github.com/viant/mcp-protocol/schema"
	"strings"
	"testing"
	"testing/fstest"
)

func TestStaticSkillRejectsUnprovenManifests(t *testing.T) {
	ctx := context.Background()
	files := fstest.MapFS{"SKILL.md": {Data: []byte("---\nname: demo\ndescription: Actual description.\n---\nActual body")}, "secret.txt": {Data: []byte("actual")}}
	compiled, err := (skillformat.Compiler{Source: files}).Compile(ctx, "skill://demo/SKILL.md")
	require.NoError(t, err)
	for _, mutation := range []func(*schema.Skill){
		func(s *schema.Skill) {},
		func(s *schema.Skill) { s.Resources.Files = s.Resources.Files[:1] },
		func(s *schema.Skill) { s.Resources.Files[0].Size = 0 },
		func(s *schema.Skill) { s.Resources.Files[0].Digest = "sha256:" + strings.Repeat("0", 64) },
		func(s *schema.Skill) { s.Frontmatter["description"] = "Forged" },
	} {
		d := NewDefaultHandler(nil, nil, nil)
		called := false
		for _, file := range compiled.Resources() {
			d.RegisterResource(file, func(context.Context, *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
				called = true
				return nil, nil
			})
		}
		metadata := compiled.Metadata()
		mutation(&metadata)
		require.Error(t, d.RegisterSkill(metadata))
		require.False(t, called)
		require.False(t, d.ImplementsSkills())
		require.Error(t, d.RegisterStaticSkill(&skillformat.Static{}))
	}
}

func TestStaticSkillByteAuthoritySurvivesHostileOverrides(t *testing.T) {
	ctx := context.Background()
	source := fstest.MapFS{"SKILL.md": {Data: []byte("---\nname: demo\ndescription: demo\n---\nBody")}, "empty.txt": {}, "binary.txt": {Data: []byte{0xff, 0}}}
	compiled, err := (skillformat.Compiler{Source: source}).Compile(ctx, "docs://team/demo/SKILL.md")
	require.NoError(t, err)
	d := NewDefaultHandler(nil, nil, nil)
	called := false
	arbitrary := func(context.Context, *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
		called = true
		return nil, nil
	}
	d.RegisterResource(schema.Resource{Uri: "docs://team/demo/SKILL.md", Name: "fake"}, arbitrary)
	require.NoError(t, d.RegisterStaticSkill(compiled))
	require.False(t, called)
	original := compiled.Metadata()
	source["SKILL.md"].Data = []byte("changed source")
	source["new.txt"] = &fstest.MapFile{Data: []byte("new")}
	mutated := compiled.Metadata()
	mutated.Resources.Files[0].Size = 999
	mutated.Frontmatter["description"] = "changed"
	require.Error(t, d.RegisterResource(schema.Resource{Uri: "docs://team/demo/new.txt"}, arbitrary))
	// The existing public ordinary registry cannot replace the sealed read path.
	for _, uri := range []string{"docs://team/demo/SKILL.md", "docs://team/demo/new.txt"} {
		d.ResourceRegistry.Put(uri, &ResourceEntry{Metadata: schema.Resource{Uri: uri, Name: "fake"}, Handler: arbitrary})
	}
	require.Len(t, d.ListRegisteredResources(), 3)
	require.Equal(t, original, d.ListRegisteredSkills()[0])
	for _, file := range original.Resources.Files {
		read, e := d.ReadResource(ctx, &jsonrpc.TypedRequest[*schema.ReadResourceRequest]{Request: &schema.ReadResourceRequest{Params: schema.ReadResourceRequestParams{Uri: file.Uri}}})
		require.Nil(t, e)
		require.Len(t, read.Contents, 1)
		data := []byte(read.Contents[0].Text)
		if read.Contents[0].Blob != "" {
			data, err = base64.StdEncoding.DecodeString(read.Contents[0].Blob)
			require.NoError(t, err)
		}
		require.Equal(t, file.Size, int64(len(data)))
	}
	_, e := d.ReadResource(ctx, &jsonrpc.TypedRequest[*schema.ReadResourceRequest]{Request: &schema.ReadResourceRequest{Params: schema.ReadResourceRequestParams{Uri: "docs://team/demo/new.txt"}}})
	require.NotNil(t, e)
	require.False(t, called)
}

func TestStaticSkillNestedConflictsAndExtraResources(t *testing.T) {
	ctx := context.Background()
	makeSkill := func(root, body string) *skillformat.Static {
		s, err := (skillformat.Compiler{Source: fstest.MapFS{"SKILL.md": {Data: []byte(fmt.Sprintf("---\nname: demo\ndescription: demo\n---\n%s", body))}}}).Compile(ctx, root+"/SKILL.md")
		require.NoError(t, err)
		return s
	}
	root := makeSkill("skill://demo", "parent")
	d := NewDefaultHandler(nil, nil, nil)
	d.RegisterResource(schema.Resource{Uri: "skill://demo/unlisted.txt"}, nil)
	require.Error(t, d.RegisterStaticSkill(root))
	require.False(t, d.ImplementsSkills())
	d = NewDefaultHandler(nil, nil, nil)
	require.NoError(t, d.RegisterStaticSkill(root))
	require.Error(t, d.RegisterStaticSkill(makeSkill("skill://demo/nested/demo", "child")))
}
