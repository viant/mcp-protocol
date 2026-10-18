package namespaces

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestPaginateFilteredSnapshot(t *testing.T) {
	// The managing server supplies only its visible, kind-filtered summaries.
	entries := []Resource{{URI: "window://public/a", Kind: "window", Name: "A"}, {URI: "window://public/b", Kind: "window", Name: "B"}, {URI: "window://public/c", Kind: "window", Name: "C"}}
	first, next, err := Paginate(entries, nil, 2)
	if err != nil || next == nil || !reflect.DeepEqual(first, entries[:2]) {
		t.Fatalf("first: %v %v %v", first, next, err)
	}
	first[0].Name = "changed"
	if entries[0].Name != "A" {
		t.Fatal("page aliases original slice")
	}
	last, end, err := Paginate(entries, next, 2)
	if err != nil || end != nil || !reflect.DeepEqual(last, entries[2:]) {
		t.Fatalf("last: %v %v %v", last, end, err)
	}
	changed := append([]Resource{}, entries...)
	changed[0].Name = "new snapshot"
	if _, _, err = Paginate(changed, next, 2); !errors.Is(err, ErrCursor) {
		t.Fatalf("stale cursor accepted: %v", err)
	}
	if _, _, err = Paginate(entries[1:], next, 2); !errors.Is(err, ErrCursor) {
		t.Fatalf("different visible collection accepted: %v", err)
	}
}

func TestPaginateRejectsMalformedAndInvalidOffsets(t *testing.T) {
	entries := []string{"a", "b", "c", "d", "e"}
	_, next, _ := Paginate(entries, nil, 2)
	decoded, _ := base64.RawURLEncoding.DecodeString(*next)
	revision := strings.Split(string(decoded), ":")[0]
	cursors := []string{"", "!", base64.RawURLEncoding.EncodeToString([]byte("wrong:2"))}
	for _, offset := range []string{"0", "-2", "1", "5", "6", "02", "+2", "9999999999999999999999999999", "2:more"} {
		cursors = append(cursors, base64.RawURLEncoding.EncodeToString([]byte(revision+":"+offset)))
	}
	for _, cursor := range cursors {
		if _, _, err := Paginate(entries, &cursor, 2); !errors.Is(err, ErrCursor) {
			t.Errorf("accepted %q: %v", cursor, err)
		}
	}
	for _, size := range []int{0, -1} {
		if _, _, err := Paginate(entries, nil, size); err == nil {
			t.Errorf("accepted page size %d", size)
		}
	}
}

func TestPaginateEmptyAndLargePage(t *testing.T) {
	page, next, err := Paginate([]Resource(nil), nil, 20)
	if err != nil || page == nil || len(page) != 0 || next != nil {
		t.Fatalf("empty: %#v %v %v", page, next, err)
	}
	entries := []int{1, 2, 3}
	pageInts, next, err := Paginate(entries, nil, int(^uint(0)>>1))
	if err != nil || next != nil || !reflect.DeepEqual(pageInts, entries) {
		t.Fatalf("large page: %v %v %v", pageInts, next, err)
	}
	if _, _, err := Paginate([]any{func() {}}, nil, 1); err == nil {
		t.Fatal("unserializable snapshot accepted")
	}
}

func TestToolContractJSON(t *testing.T) {
	encoded, err := json.Marshal(GetRequest{URI: "ns://opaque_ID"})
	if err != nil || string(encoded) != `{"uri":"ns://opaque_ID"}` {
		t.Fatalf("optional filters: %s %v", encoded, err)
	}
	encoded, err = json.Marshal(GetResult{Resources: []Resource{}})
	if err != nil || string(encoded) != `{"resources":[]}` {
		t.Fatalf("final page: %s %v", encoded, err)
	}
	encoded, err = json.Marshal(ResourceGetRequest{URI: "mcp://platform/api"})
	if err != nil || string(encoded) != `{"uri":"mcp://platform/api"}` {
		t.Fatalf("get: %s %v", encoded, err)
	}
}
