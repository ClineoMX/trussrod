package database

import (
	"reflect"
	"testing"
)

func TestNewSelect(t *testing.T) {
	s := NewSelect("users")
	if s.table != "users" {
		t.Fatalf("expected table users, got %q", s.table)
	}
	if s.Index != 1 {
		t.Fatalf("expected Index 1, got %d", s.Index)
	}
	if len(s.columns) != 0 {
		t.Fatalf("expected empty columns, got %v", s.columns)
	}
}

func TestSelect_Column(t *testing.T) {
	s := NewSelect("users").
		Columns("id", "name", "email")

	query, args := s.Build()

	wantQuery := "SELECT id, name, email FROM users"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if args != nil && len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestSelect_ColumnSingle(t *testing.T) {
	s := NewSelect("users").Columns("count(*)")

	query, _ := s.Build()

	wantQuery := "SELECT count(*) FROM users"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
}

func TestSelect_Where(t *testing.T) {
	s := NewSelect("users").
		Column("id", "name").
		Where("id", "user-1")

	query, args := s.Build()

	wantQuery := "SELECT id, name FROM users WHERE id = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"user-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
	if s.Index != 2 {
		t.Fatalf("expected Index 2 after one Where call, got %d", s.Index)
	}
}

func TestSelect_WhereMultiple(t *testing.T) {
	s := NewSelect("users").
		Column("id").
		Where("tenant_id", "tenant-1").
		Where("active", true)

	query, args := s.Build()

	wantQuery := "SELECT id FROM users WHERE tenant_id = $1 AND active = $2"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"tenant-1", true}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelect_OrderByLimitOffset(t *testing.T) {
	s := NewSelect("posts").
		Column("id", "title").
		Where("author_id", "author-1").
		OrderBy("created_at", "DESC").
		Limit(10).
		Offset(20)

	query, args := s.Build()

	wantQuery := "SELECT id, title FROM posts WHERE author_id = $1 ORDER BY created_at DESC LIMIT 10 OFFSET 20"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"author-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelect_LimitOffsetZeroOmitted(t *testing.T) {
	s := NewSelect("users").
		Column("id").
		Limit(0).
		Offset(0)

	query, _ := s.Build()

	wantQuery := "SELECT id FROM users"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
}

func TestSelect_BuildFull(t *testing.T) {
	tests := []struct {
		name      string
		build     func() *Select
		wantQuery string
		wantArgs  []any
	}{
		{
			name: "columns only",
			build: func() *Select {
				return NewSelect("items").Columns("id", "status")
			},
			wantQuery: "SELECT id, status FROM items",
			wantArgs:  nil,
		},
		{
			name: "columns and where",
			build: func() *Select {
				return NewSelect("items").
					Column("id").
					Where("status", "active")
			},
			wantQuery: "SELECT id FROM items WHERE status = $1",
			wantArgs:  []any{"active"},
		},
		{
			name: "order by without where",
			build: func() *Select {
				return NewSelect("items").
					Column("id").
					OrderBy("created_at", "ASC")
			},
			wantQuery: "SELECT id FROM items ORDER BY created_at ASC",
			wantArgs:  nil,
		},
		{
			name: "limit without offset",
			build: func() *Select {
				return NewSelect("items").
					Column("id").
					Limit(5)
			},
			wantQuery: "SELECT id FROM items LIMIT 5",
			wantArgs:  nil,
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

func TestSelect_InnerJoin(t *testing.T) {
	s := NewSelect("users").
		Column("users.id", "orders.total").
		InnerJoin("orders", "orders.user_id = users.id")

	query, args := s.Build()

	wantQuery := "SELECT users.id, orders.total FROM users INNER JOIN orders ON orders.user_id = users.id"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if args != nil && len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestSelect_LeftJoinEq(t *testing.T) {
	s := NewSelect("users").
		Column("users.id", "profiles.bio").
		LeftJoinEq("profiles", "profiles.user_id", "users.id")

	query, _ := s.Build()

	wantQuery := "SELECT users.id, profiles.bio FROM users LEFT JOIN profiles ON profiles.user_id = users.id"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
}

func TestSelect_MultipleJoins(t *testing.T) {
	s := NewSelect("users").
		Column("users.id").
		InnerJoinEq("orders", "orders.user_id", "users.id").
		LeftJoin("order_items", "order_items.order_id = orders.id")

	query, _ := s.Build()

	wantQuery := "SELECT users.id FROM users INNER JOIN orders ON orders.user_id = users.id LEFT JOIN order_items ON order_items.order_id = orders.id"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
}

func TestSelect_JoinWithWhere(t *testing.T) {
	s := NewSelect("users").
		Column("users.id", "orders.total").
		InnerJoinEq("orders", "orders.user_id", "users.id").
		Where("users.active", true)

	query, args := s.Build()

	wantQuery := "SELECT users.id, orders.total FROM users INNER JOIN orders ON orders.user_id = users.id WHERE users.active = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{true}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelect_JoinOnWithWhere(t *testing.T) {
	s := NewSelect("users").
		Column("users.id").
		Where("users.active", true).
		InnerJoinOn("orders", "orders.status", "pending")

	query, args := s.Build()

	wantQuery := "SELECT users.id FROM users INNER JOIN orders ON orders.status = $1 WHERE users.active = $2"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"pending", true}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelect_RightAndFullJoin(t *testing.T) {
	s := NewSelect("users").
		Column("users.id").
		RightJoin("orders", "orders.user_id = users.id").
		FullJoin("archives", "archives.user_id = users.id")

	query, _ := s.Build()

	wantQuery := "SELECT users.id FROM users RIGHT JOIN orders ON orders.user_id = users.id FULL JOIN archives ON archives.user_id = users.id"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
}

func TestSelect_WhereIfNotNil_SkipsNil(t *testing.T) {
	s := NewSelect("users").
		Columns("id").
		Where("patient_id", "patient-1").
		WhereIfNotNil("status", nil)

	query, args := s.Build()

	wantQuery := "SELECT id FROM users WHERE patient_id = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"patient-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
	if s.Index != 2 {
		t.Fatalf("expected Index 2 after skipped nil, got %d", s.Index)
	}
}

func TestSelect_WhereIfNotNil_IncludesValue(t *testing.T) {
	status := "active"
	s := NewSelect("users").
		Columns("id").
		WhereIfNotNil("status", status)

	query, args := s.Build()

	wantQuery := "SELECT id FROM users WHERE status = $1"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"active"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelect_WhereIfNotNil_MixedWithWhere(t *testing.T) {
	s := NewSelect("notes n").
		Columns("n.id").
		Where("n.patient_id", "patient-1").
		WhereIfNotNil("n.status", nil).
		Where("n.author", "author-1")

	query, args := s.Build()

	wantQuery := "SELECT n.id FROM notes n WHERE n.patient_id = $1 AND n.author = $2"
	if query != wantQuery {
		t.Fatalf("expected query %q, got %q", wantQuery, query)
	}
	if !reflect.DeepEqual(args, []any{"patient-1", "author-1"}) {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelect_ChainingReturnsSameBuilder(t *testing.T) {
	s := NewSelect("users")
	if s.Columns("id") != s {
		t.Fatal("Columns should return the same select for chaining")
	}
	if s.InnerJoin("orders", "orders.user_id = users.id") != s {
		t.Fatal("InnerJoin should return the same select for chaining")
	}
	if s.LeftJoinEq("profiles", "profiles.user_id", "users.id") != s {
		t.Fatal("LeftJoinEq should return the same select for chaining")
	}
	if s.Where("id", 1) != s {
		t.Fatal("Where should return the same select for chaining")
	}
	if s.WhereIfNotNil("status", nil) != s {
		t.Fatal("WhereIfNotNil should return the same select for chaining")
	}
	if s.OrderBy("name", "ASC") != s {
		t.Fatal("OrderBy should return the same select for chaining")
	}
	if s.Limit(1) != s {
		t.Fatal("Limit should return the same select for chaining")
	}
	if s.Offset(1) != s {
		t.Fatal("Offset should return the same select for chaining")
	}
}
