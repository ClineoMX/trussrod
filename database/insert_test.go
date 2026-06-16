package database

import (
	"reflect"
	"testing"
	"time"
)

func TestNewInsert(t *testing.T) {
	i := NewInsert("users")
	if i.table != "users" {
		t.Fatalf("expected table users, got %q", i.table)
	}
	if i.Index != 1 {
		t.Fatalf("expected Index 1, got %d", i.Index)
	}
}

func TestInsert_ColumnsValues(t *testing.T) {
	i := NewInsert("users").
		Columns("name", "email").
		Values("alice", "alice@example.com")

	query, args := i.Build()

	wantQuery := "INSERT INTO users (name, email) VALUES ($1, $2)"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice", "alice@example.com"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestInsert_Set(t *testing.T) {
	i := NewInsert("users").
		Set("name", "alice").
		Set("email", "alice@example.com")

	query, args := i.Build()

	wantQuery := "INSERT INTO users (name, email) VALUES ($1, $2)"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice", "alice@example.com"}) {
		t.Fatalf("unexpected args: %v", args)
	}
	if i.Index != 3 {
		t.Fatalf("expected Index 3 after two Set calls, got %d", i.Index)
	}
}

func TestInsert_SetStringIfNotNil(t *testing.T) {
	name := "alice"
	i := NewInsert("users").
		SetStringIfNotNil("name", &name).
		SetStringIfNotNil("nickname", nil).
		Set("email", "alice@example.com")

	query, args := i.Build()

	wantQuery := "INSERT INTO users (name, email) VALUES ($1, $2)"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice", "alice@example.com"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestInsert_SetBytesIfNotNil(t *testing.T) {
	data := []byte("payload")
	i := NewInsert("documents").
		SetBytesIfNotNil("content", &data).
		SetBytesIfNotNil("thumbnail", nil)

	query, args := i.Build()

	wantQuery := "INSERT INTO documents (content) VALUES ($1)"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{data}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestInsert_SetTimeIfNotNil(t *testing.T) {
	ts := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	i := NewInsert("events").
		SetTimeIfNotNil("starts_at", &ts).
		SetTimeIfNotNil("ends_at", nil)

	query, args := i.Build()

	wantQuery := "INSERT INTO events (starts_at) VALUES ($1)"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{ts}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestInsert_Returning(t *testing.T) {
	i := NewInsert("users").
		Set("name", "alice").
		Returning("id", "created_at")

	query, args := i.Build()

	wantQuery := "INSERT INTO users (name) VALUES ($1) RETURNING id, created_at"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestInsert_BuildFull(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *Insert
		wantQuery string
		wantArgs  []any
	}{
		{
			name: "set only",
			build: func() *Insert {
				return NewInsert("items").Set("status", "active")
			},
			wantQuery: "INSERT INTO items (status) VALUES ($1)",
			wantArgs:  []any{"active"},
		},
		{
			name: "columns and values",
			build: func() *Insert {
				return NewInsert("items").
					Columns("id", "status").
					Values(42, "active")
			},
			wantQuery: "INSERT INTO items (id, status) VALUES ($1, $2)",
			wantArgs:  []any{42, "active"},
		},
		{
			name: "returning without extra clauses",
			build: func() *Insert {
				return NewInsert("items").
					Set("count", 1).
					Returning("count")
			},
			wantQuery: "INSERT INTO items (count) VALUES ($1) RETURNING count",
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

func TestInsert_ChainingReturnsSameBuilder(t *testing.T) {
	i := NewInsert("users")
	if i.Columns("name") != i {
		t.Fatal("Columns should return the same insert for chaining")
	}
	if i.Values("alice") != i {
		t.Fatal("Values should return the same insert for chaining")
	}
	if i.Set("email", "alice@example.com") != i {
		t.Fatal("Set should return the same insert for chaining")
	}
	if i.Returning("id") != i {
		t.Fatal("Returning should return the same insert for chaining")
	}
}
