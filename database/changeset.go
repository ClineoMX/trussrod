package database

import (
	"fmt"
	"strings"
	"time"
)

type Changeset struct {
	table   string
	clauses []string
	args    []any
	Index   int
}

func NewChangeset(table string) *Changeset {
	return &Changeset{
		table: table,
		Index: 1,
	}
}

func (c *Changeset) Set(column string, value any) *Changeset {
	c.clauses = append(c.clauses, fmt.Sprintf("%s = $%d", column, c.Index))
	c.args = append(c.args, value)
	c.Index++
	return c
}

func (c *Changeset) SetStringIfNotNil(column string, value *string) *Changeset {
	if value != nil {
		c.Set(column, *value)
	}
	return c
}

func (c *Changeset) SetBytesIfNotNil(column string, value *[]byte) *Changeset {
	if value != nil {
		c.Set(column, *value)
	}
	return c
}

func (c *Changeset) SetTimeIfNotNil(column string, value *time.Time) *Changeset {
	if value != nil {
		c.Set(column, *value)
	}
	return c
}

func (c *Changeset) Build(where string, whereArgs ...any) (string, []any, error) {
	if len(c.clauses) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s",
		c.table,
		strings.Join(c.clauses, ", "),
	)

	if where != "" {
		finalWhere := where
		for i, arg := range whereArgs {
			placeholder := fmt.Sprintf("$%d", c.Index+i)
			finalWhere = strings.Replace(finalWhere, "?", placeholder, 1)
			c.args = append(c.args, arg)
		}

		query += " WHERE " + finalWhere
	}

	return query, c.args, nil
}
