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

func TestInsert_OnConflictDoNothing(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *Insert
		wantQuery string
		wantArgs  []any
	}{
		{
			name: "with target columns",
			build: func() *Insert {
				return NewInsert("users").
					Set("email", "alice@example.com").
					OnConflict("email").
					DoNothing()
			},
			wantQuery: "INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING",
			wantArgs:  []any{"alice@example.com"},
		},
		{
			name: "with multiple target columns",
			build: func() *Insert {
				return NewInsert("memberships").
					Columns("org_id", "user_id").
					Values(1, 2).
					OnConflict("org_id", "user_id").
					DoNothing()
			},
			wantQuery: "INSERT INTO memberships (org_id, user_id) VALUES ($1, $2) ON CONFLICT (org_id, user_id) DO NOTHING",
			wantArgs:  []any{1, 2},
		},
		{
			name: "bare on conflict without target",
			build: func() *Insert {
				return NewInsert("users").
					Set("email", "alice@example.com").
					OnConflict().
					DoNothing()
			},
			wantQuery: "INSERT INTO users (email) VALUES ($1) ON CONFLICT DO NOTHING",
			wantArgs:  []any{"alice@example.com"},
		},
		{
			name: "on constraint",
			build: func() *Insert {
				return NewInsert("users").
					Set("email", "alice@example.com").
					OnConflictConstraint("users_email_key").
					DoNothing()
			},
			wantQuery: "INSERT INTO users (email) VALUES ($1) ON CONFLICT ON CONSTRAINT users_email_key DO NOTHING",
			wantArgs:  []any{"alice@example.com"},
		},
		{
			name: "bare on conflict defaults to do nothing without an action",
			build: func() *Insert {
				i := NewInsert("users").Set("email", "alice@example.com")
				i.OnConflict("email")
				return i
			},
			wantQuery: "INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING",
			wantArgs:  []any{"alice@example.com"},
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

func TestInsert_OnConflictDoUpdate(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *Insert
		wantQuery string
		wantArgs  []any
	}{
		{
			name: "set excluded columns",
			build: func() *Insert {
				return NewInsert("users").
					Columns("email", "name").
					Values("alice@example.com", "Alice").
					OnConflict("email").
					DoUpdateSetExcluded("name").
					Insert()
			},
			wantQuery: "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name",
			wantArgs:  []any{"alice@example.com", "Alice"},
		},
		{
			name: "set excluded multiple columns",
			build: func() *Insert {
				return NewInsert("users").
					Columns("email", "name", "updated_at").
					Values("alice@example.com", "Alice", "now").
					OnConflict("email").
					DoUpdateSetExcluded("name", "updated_at").
					Insert()
			},
			wantQuery: "INSERT INTO users (email, name, updated_at) VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at",
			wantArgs:  []any{"alice@example.com", "Alice", "now"},
		},
		{
			name: "set bound value appends placeholder after values",
			build: func() *Insert {
				return NewInsert("counters").
					Columns("key", "count").
					Values("hits", 1).
					OnConflict("key").
					DoUpdateSet("count", 99).
					Insert()
			},
			wantQuery: "INSERT INTO counters (key, count) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET count = $3",
			wantArgs:  []any{"hits", 1, 99},
		},
		{
			name: "set raw expression",
			build: func() *Insert {
				return NewInsert("counters").
					Columns("key", "count").
					Values("hits", 1).
					OnConflict("key").
					DoUpdateSetRaw("count = counters.count + 1").
					Insert()
			},
			wantQuery: "INSERT INTO counters (key, count) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET count = counters.count + 1",
			wantArgs:  []any{"hits", 1},
		},
		{
			name: "mixed excluded, bound and raw assignments keep order and indexing",
			build: func() *Insert {
				return NewInsert("users").
					Columns("email", "name").
					Values("alice@example.com", "Alice").
					OnConflict("email").
					DoUpdateSetExcluded("name").
					DoUpdateSet("login_count", 5).
					DoUpdateSetRaw("updated_at = now()").
					Insert()
			},
			wantQuery: "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, login_count = $3, updated_at = now()",
			wantArgs:  []any{"alice@example.com", "Alice", 5},
		},
		{
			name: "do update with where predicates",
			build: func() *Insert {
				return NewInsert("users").
					Columns("email", "name").
					Values("alice@example.com", "Alice").
					OnConflict("email").
					DoUpdateSetExcluded("name").
					Where("active", true).
					WhereRaw("users.deleted_at IS NULL").
					Insert()
			},
			wantQuery: "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name WHERE active = $3 AND users.deleted_at IS NULL",
			wantArgs:  []any{"alice@example.com", "Alice", true},
		},
		{
			name: "do update with returning",
			build: func() *Insert {
				return NewInsert("users").
					Columns("email", "name").
					Values("alice@example.com", "Alice").
					OnConflict("email").
					DoUpdateSetExcluded("name").
					Insert().
					Returning("id")
			},
			wantQuery: "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name RETURNING id",
			wantArgs:  []any{"alice@example.com", "Alice"},
		},
		{
			name: "set bound value with where bound value indexes sequentially",
			build: func() *Insert {
				return NewInsert("counters").
					Columns("key", "count").
					Values("hits", 1).
					OnConflict("key").
					DoUpdateSet("count", 99).
					Where("locked", false).
					Insert()
			},
			wantQuery: "INSERT INTO counters (key, count) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET count = $3 WHERE locked = $4",
			wantArgs:  []any{"hits", 1, 99, false},
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

func TestInsert_OnConflictDoNothingWithReturning(t *testing.T) {
	query, args := NewInsert("users").
		Set("email", "alice@example.com").
		OnConflict("email").
		DoNothing().
		Returning("id").
		Build()

	wantQuery := "INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING RETURNING id"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"alice@example.com"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestInsert_OnConflictChainingReturnsSameBuilders(t *testing.T) {
	i := NewInsert("users").Set("email", "alice@example.com")

	c := i.OnConflict("email")
	if c.insert != i {
		t.Fatal("OnConflict should reference the originating insert")
	}
	if c.DoUpdateSetExcluded("name") != c {
		t.Fatal("DoUpdateSetExcluded should return the same conflict clause")
	}
	if c.DoUpdateSet("count", 1) != c {
		t.Fatal("DoUpdateSet should return the same conflict clause")
	}
	if c.DoUpdateSetRaw("updated_at = now()") != c {
		t.Fatal("DoUpdateSetRaw should return the same conflict clause")
	}
	if c.Where("active", true) != c {
		t.Fatal("Where should return the same conflict clause")
	}
	if c.WhereRaw("deleted_at IS NULL") != c {
		t.Fatal("WhereRaw should return the same conflict clause")
	}
	if c.Insert() != i {
		t.Fatal("Insert should return the originating insert")
	}
	if i.OnConflict("email").DoNothing() != i {
		t.Fatal("DoNothing should return the originating insert")
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
