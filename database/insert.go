package database

import (
	"fmt"
	"strings"
	"time"
)

type Insert struct {
	table      string
	columns    []string
	values     []any
	returning  string
	onConflict *ConflictClause
	Index      int
}

func NewInsert(table string) *Insert {
	return &Insert{
		table: table,
		Index: 1,
	}
}

func (i *Insert) Columns(columns ...string) *Insert {
	i.columns = append(i.columns, columns...)
	return i
}

func (i *Insert) Values(values ...any) *Insert {
	i.values = append(i.values, values...)
	return i
}

func (i *Insert) Set(column string, value any) *Insert {
	i.columns = append(i.columns, column)
	i.values = append(i.values, value)
	i.Index++
	return i
}

func (i *Insert) SetStringIfNotNil(column string, value *string) *Insert {
	if value != nil {
		i.Set(column, *value)
	}
	return i
}

func (i *Insert) SetBytesIfNotNil(column string, value *[]byte) *Insert {
	if value != nil {
		i.Set(column, *value)
	}
	return i
}

func (i *Insert) SetTimeIfNotNil(column string, value *time.Time) *Insert {
	if value != nil {
		i.Set(column, *value)
	}
	return i
}

func (i *Insert) Returning(columns ...string) *Insert {
	i.returning = strings.Join(columns, ", ")
	return i
}

// ConflictClause builds the ON CONFLICT portion of an INSERT statement.
// It is created via Insert.OnConflict or Insert.OnConflictConstraint.
type ConflictClause struct {
	insert     *Insert
	target     []string
	constraint string
	doNothing  bool
	sets       []conflictAssignment
	conditions []conflictAssignment
}

// conflictAssignment represents a single SET assignment or WHERE predicate
// within an ON CONFLICT ... DO UPDATE clause. When bound is true the clause is
// emitted as "<clause> = $n" and value is appended to the args; otherwise the
// clause is emitted verbatim.
type conflictAssignment struct {
	clause string
	value  any
	bound  bool
}

// OnConflict begins an ON CONFLICT clause targeting the given columns.
// Pass no columns for a bare ON CONFLICT (e.g. ON CONFLICT DO NOTHING).
func (i *Insert) OnConflict(target ...string) *ConflictClause {
	c := &ConflictClause{insert: i, target: target}
	i.onConflict = c
	return c
}

// OnConflictConstraint begins an ON CONFLICT ON CONSTRAINT <name> clause.
func (i *Insert) OnConflictConstraint(name string) *ConflictClause {
	c := &ConflictClause{insert: i, constraint: name}
	i.onConflict = c
	return c
}

// DoNothing resolves the conflict with DO NOTHING and returns to the insert
// builder for further chaining (e.g. Returning).
func (c *ConflictClause) DoNothing() *Insert {
	c.doNothing = true
	return c.insert
}

// DoUpdateSet adds a "column = $n" assignment with a bound value to the
// DO UPDATE clause.
func (c *ConflictClause) DoUpdateSet(column string, value any) *ConflictClause {
	c.sets = append(c.sets, conflictAssignment{clause: column, value: value, bound: true})
	return c
}

// DoUpdateSetExcluded adds "column = EXCLUDED.column" assignments for each
// column. This is the common upsert pattern of overwriting the existing row
// with the values that were proposed for insertion.
func (c *ConflictClause) DoUpdateSetExcluded(columns ...string) *ConflictClause {
	for _, col := range columns {
		c.sets = append(c.sets, conflictAssignment{clause: fmt.Sprintf("%s = EXCLUDED.%s", col, col)})
	}
	return c
}

// DoUpdateSetRaw adds a raw assignment expression (e.g. "count = items.count + 1")
// to the DO UPDATE clause.
func (c *ConflictClause) DoUpdateSetRaw(expr string) *ConflictClause {
	c.sets = append(c.sets, conflictAssignment{clause: expr})
	return c
}

// Where adds a "column = $n" predicate to the DO UPDATE clause. Multiple
// predicates are joined with AND.
func (c *ConflictClause) Where(column string, value any) *ConflictClause {
	c.conditions = append(c.conditions, conflictAssignment{clause: column, value: value, bound: true})
	return c
}

// WhereRaw adds a raw predicate to the DO UPDATE clause. Multiple predicates
// are joined with AND.
func (c *ConflictClause) WhereRaw(expr string) *ConflictClause {
	c.conditions = append(c.conditions, conflictAssignment{clause: expr})
	return c
}

// Insert returns the underlying insert builder to continue chaining after a
// DO UPDATE clause (e.g. to add Returning).
func (c *ConflictClause) Insert() *Insert {
	return c.insert
}

// build renders the ON CONFLICT clause, assigning placeholders starting at
// startIndex, and returns the SQL fragment along with any bound args.
func (c *ConflictClause) build(startIndex int) (string, []any) {
	var sb strings.Builder
	sb.WriteString(" ON CONFLICT")

	if c.constraint != "" {
		sb.WriteString(" ON CONSTRAINT " + c.constraint)
	} else if len(c.target) > 0 {
		sb.WriteString(" (" + strings.Join(c.target, ", ") + ")")
	}

	// A bare OnConflict with no DO UPDATE assignments defaults to DO NOTHING.
	if c.doNothing || len(c.sets) == 0 {
		sb.WriteString(" DO NOTHING")
		return sb.String(), nil
	}

	var args []any
	index := startIndex

	setParts := make([]string, len(c.sets))
	for j, a := range c.sets {
		if a.bound {
			setParts[j] = fmt.Sprintf("%s = $%d", a.clause, index)
			args = append(args, a.value)
			index++
			continue
		}
		setParts[j] = a.clause
	}
	sb.WriteString(" DO UPDATE SET " + strings.Join(setParts, ", "))

	if len(c.conditions) > 0 {
		whereParts := make([]string, len(c.conditions))
		for j, w := range c.conditions {
			if w.bound {
				whereParts[j] = fmt.Sprintf("%s = $%d", w.clause, index)
				args = append(args, w.value)
				index++
				continue
			}
			whereParts[j] = w.clause
		}
		sb.WriteString(" WHERE " + strings.Join(whereParts, " AND "))
	}

	return sb.String(), args
}

func (i *Insert) Build() (string, []any) {
	placeholders := make([]string, len(i.values))
	for j := range i.values {
		placeholders[j] = fmt.Sprintf("$%d", j+1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		i.table,
		strings.Join(i.columns, ", "),
		strings.Join(placeholders, ", "),
	)

	args := i.values
	if i.onConflict != nil {
		conflictSQL, conflictArgs := i.onConflict.build(len(i.values) + 1)
		query += conflictSQL
		if len(conflictArgs) > 0 {
			args = append(append([]any{}, i.values...), conflictArgs...)
		}
	}

	if i.returning != "" {
		query += " RETURNING " + i.returning
	}

	if len(args) == 0 {
		return query, nil
	}

	return query, args
}
