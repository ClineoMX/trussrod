package iam

import (
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUIDV4_Format(t *testing.T) {
	id, err := NewUUIDV4()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uuidV4Pattern.MatchString(id) {
		t.Fatalf("expected a valid UUID v4, got %q", id)
	}
}

func TestNewUUIDV4_Unique(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for range n {
		id, err := NewUUIDV4()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("expected unique UUIDs, got duplicate %q", id)
		}
		seen[id] = struct{}{}
	}
}
