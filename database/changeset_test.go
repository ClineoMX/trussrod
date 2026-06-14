package database

import (
	"reflect"
	"testing"
	"time"
)

func TestNewChangeset(t *testing.T) {
	cs := NewChangeset("users")
	if cs.table != "users" {
		t.Fatalf("expected table users, got %q", cs.table)
	}
	if cs.Index != 1 {
		t.Fatalf("expected Index 1, got %d", cs.Index)
	}
}

func TestChangeset_Set(t *testing.T) {
	cs := NewChangeset("users").
		Set("name", "alice").
		Set("email", "alice@example.com")

	query, args := cs.Build()

	wantQuery := "UPDATE users SET name = $1, email = $2"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice", "alice@example.com"}) {
		t.Fatalf("unexpected args: %v", args)
	}
	if cs.Index != 3 {
		t.Fatalf("expected Index 3 after two Set calls, got %d", cs.Index)
	}
}

func TestChangeset_SetStringIfNotNil(t *testing.T) {
	name := "alice"
	cs := NewChangeset("users").
		SetStringIfNotNil("name", &name).
		SetStringIfNotNil("nickname", nil)

	query, args := cs.Build()

	wantQuery := "UPDATE users SET name = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestChangeset_SetBytesIfNotNil(t *testing.T) {
	data := []byte("payload")
	cs := NewChangeset("documents").
		SetBytesIfNotNil("content", &data).
		SetBytesIfNotNil("thumbnail", nil)

	query, args := cs.Build()

	wantQuery := "UPDATE documents SET content = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{data}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestChangeset_SetTimeIfNotNil(t *testing.T) {
	ts := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	cs := NewChangeset("events").
		SetTimeIfNotNil("starts_at", &ts).
		SetTimeIfNotNil("ends_at", nil)

	query, args := cs.Build()

	wantQuery := "UPDATE events SET starts_at = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{ts}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestChangeset_Where(t *testing.T) {
	cs := NewChangeset("users").
		Set("active", true).
		Where("id", "user-1")

	query, args := cs.Build()

	wantQuery := "UPDATE users SET active = $1 WHERE id = $2"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{true, "user-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestChangeset_WhereMultiple(t *testing.T) {
	cs := NewChangeset("users").
		Set("active", false).
		Where("tenant_id", "tenant-1").
		Where("id", "user-1")

	query, args := cs.Build()

	wantQuery := "UPDATE users SET active = $1 WHERE tenant_id = $2 AND id = $3"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{false, "tenant-1", "user-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestChangeset_OrderByLimitOffsetReturning(t *testing.T) {
	cs := NewChangeset("posts").
		Set("published", true).
		Where("author_id", "author-1").
		OrderBy("created_at", "DESC").
		Limit(10).
		Offset(20).
		Returning("id", "updated_at")

	query, args := cs.Build()

	wantQuery := "UPDATE posts SET published = $1 WHERE author_id = $2 ORDER BY created_at DESC LIMIT 10 OFFSET 20 RETURNING id, updated_at"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{true, "author-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestChangeset_LimitOffsetZeroOmitted(t *testing.T) {
	cs := NewChangeset("users").
		Set("name", "bob").
		Limit(0).
		Offset(0)

	query, _ := cs.Build()

	wantQuery := "UPDATE users SET name = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
}

func TestChangeset_BuildFull(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *Changeset
		wantQuery string
		wantArgs  []any
	}{
		{
			name: "set only",
			build: func() *Changeset {
				return NewChangeset("items").Set("status", "active")
			},
			wantQuery: "UPDATE items SET status = $1",
			wantArgs:  []any{"active"},
		},
		{
			name: "set and where",
			build: func() *Changeset {
				return NewChangeset("items").
					Set("status", "archived").
					Where("id", 42)
			},
			wantQuery: "UPDATE items SET status = $1 WHERE id = $2",
			wantArgs:  []any{"archived", 42},
		},
		{
			name: "returning without where",
			build: func() *Changeset {
				return NewChangeset("items").
					Set("count", 1).
					Returning("count")
			},
			wantQuery: "UPDATE items SET count = $1 RETURNING count",
			wantArgs:  []any{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := tt.build().Build()
			if query != tt.wantQuery {
				t.Fatalf("expected query %q, got %q", tt.wantQuery, query)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Fatalf("expected args %v, got %v", tt.wantArgs, args)
			}
		})
	}
}

func TestChangeset_ChainingReturnsSameBuilder(t *testing.T) {
	cs := NewChangeset("users")
	if cs.Set("name", "alice") != cs {
		t.Fatal("Set should return the same changeset for chaining")
	}
	if cs.Where("id", 1) != cs {
		t.Fatal("Where should return the same changeset for chaining")
	}
	if cs.OrderBy("name", "ASC") != cs {
		t.Fatal("OrderBy should return the same changeset for chaining")
	}
	if cs.Limit(1) != cs {
		t.Fatal("Limit should return the same changeset for chaining")
	}
	if cs.Offset(1) != cs {
		t.Fatal("Offset should return the same changeset for chaining")
	}
	if cs.Returning("id") != cs {
		t.Fatal("Returning should return the same changeset for chaining")
	}
}
