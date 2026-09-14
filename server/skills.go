package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/viant/jsonrpc"
	skillformat "github.com/viant/mcp-protocol/extension/skills"
	"github.com/viant/mcp-protocol/schema"
)

// Skills is optional so existing ordinary-resource handlers remain compatible.
type Skills interface {
	ListSkills(context.Context, *jsonrpc.TypedRequest[*schema.ListSkillsRequest]) (*schema.ListSkillsResult, *jsonrpc.Error)
	GetSkill(context.Context, *jsonrpc.TypedRequest[*schema.GetSkillRequest]) (*schema.GetSkillResult, *jsonrpc.Error)
}

func (r *Registry) ImplementsSkills() bool { return r != nil && r.skills != nil && r.skills.Size() > 0 }

func (r *Registry) RegisterSkill(entry schema.Skill) error {
	if !entry.Resources.Dynamic {
		return fmt.Errorf("static skill registration requires compiler-backed byte authority")
	}
	r.staticMu.Lock()
	defer r.staticMu.Unlock()
	for _, static := range r.staticSkills {
		if static.Contains(entry.Uri) {
			return fmt.Errorf("dynamic skill conflicts with static root")
		}
	}
	if err := skillformat.Validate(entry); err != nil {
		return err
	}
	if resource, ok := r.ResourceRegistry.Get(entry.Uri); !ok || resource.Handler == nil {
		return fmt.Errorf("skill entrypoint is not a registered resource")
	}
	for _, file := range entry.Resources.Files {
		resource, ok := r.ResourceRegistry.Get(file.Uri)
		if !ok || resource.Handler == nil {
			return fmt.Errorf("skill inventory contains unpublished file %q", file.Uri)
		}
	}
	if _, exists := r.skills.Get(entry.Uri); exists {
		return fmt.Errorf("duplicate skill URI %q", entry.Uri)
	}
	copy, err := skillformat.Clone(entry)
	if err != nil {
		return err
	}
	r.skills.Put(entry.Uri, copy)
	r.Methods.Put(schema.MethodSkillsList, true)
	r.Methods.Put(schema.MethodSkillsGet, true)
	return nil
}

// ListRegisteredSkills returns detached metadata only, never fetching content.
func (r *Registry) ListRegisteredSkills() []schema.Skill {
	result := []schema.Skill{}
	if r.skills == nil {
		return result
	}
	for _, entry := range r.skills.Values() {
		copy, _ := skillformat.Clone(entry)
		result = append(result, copy)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Uri < result[j].Uri })
	return result
}

func (d *DefaultHandler) ListSkills(ctx context.Context, request *jsonrpc.TypedRequest[*schema.ListSkillsRequest]) (*schema.ListSkillsResult, *jsonrpc.Error) {
	if err := ctx.Err(); err != nil {
		return nil, jsonrpc.NewInternalError(err.Error(), nil)
	}
	entries := d.ListRegisteredSkills()
	raw, _ := json.Marshal(entries)
	revision := fmt.Sprintf("%x", sha256.Sum256(raw))
	start := 0
	if request != nil && request.Request != nil && request.Request.Params.Cursor != nil {
		decoded, err := base64.RawURLEncoding.DecodeString(*request.Request.Params.Cursor)
		parts := strings.Split(string(decoded), ":")
		if err != nil || len(parts) != 2 || parts[0] != revision {
			return nil, jsonrpc.NewInvalidParamsError("invalid or stale skills cursor", nil)
		}
		start, err = strconv.Atoi(parts[1])
		if err != nil || start <= 0 || start >= len(entries) || start%32 != 0 {
			return nil, jsonrpc.NewInvalidParamsError("invalid skills cursor", nil)
		}
	}
	end := min(start+32, len(entries))
	result := &schema.ListSkillsResult{ResultType: schema.ResultTypeComplete, Skills: entries[start:end]}
	if end < len(entries) {
		cursor := base64.RawURLEncoding.EncodeToString([]byte(revision + ":" + strconv.Itoa(end)))
		result.NextCursor = &cursor
	}
	return result, nil
}

func (d *DefaultHandler) GetSkill(ctx context.Context, request *jsonrpc.TypedRequest[*schema.GetSkillRequest]) (*schema.GetSkillResult, *jsonrpc.Error) {
	if err := ctx.Err(); err != nil {
		return nil, jsonrpc.NewInternalError(err.Error(), nil)
	}
	if request == nil || request.Request == nil || request.Request.Params.Uri == "" {
		return nil, jsonrpc.NewInvalidParamsError("skill URI is required", nil)
	}
	entry, ok := d.skills.Get(request.Request.Params.Uri)
	if !ok {
		return nil, jsonrpc.NewInvalidParamsError("unknown skill URI", nil)
	}
	copy, err := skillformat.Clone(entry)
	if err != nil {
		return nil, jsonrpc.NewInternalError(err.Error(), nil)
	}
	return &schema.GetSkillResult{ResultType: schema.ResultTypeComplete, Skill: copy}, nil
}
