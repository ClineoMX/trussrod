package database

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Changeset struct {
	table         string
	clauses       []string
	args          []any
	Index         int
	where         string
	whereArgs     []any
	returning     string
	returningArgs []any
	orderBy       string
	orderByArgs   []any
	limit         int
	offset        int
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

func (c *Changeset) Where(column string, value any) *Changeset {
	if c.where != "" {
		c.where += " AND "
	}
	c.where += fmt.Sprintf("%s = $%d", column, c.Index)
	c.whereArgs = append(c.whereArgs, value)
	c.Index++
	return c
}

func (c *Changeset) OrderBy(column string, direction string) *Changeset {
	c.orderBy = fmt.Sprintf("%s %s", column, direction)
	return c
}

func (c *Changeset) Limit(limit int) *Changeset {
	c.limit = limit
	return c
}

func (c *Changeset) Offset(offset int) *Changeset {
	c.offset = offset
	return c
}

func (c *Changeset) Returning(columns ...string) *Changeset {
	c.returning = strings.Join(columns, ", ")
	return c
}

func (c *Changeset) Build() (string, []any) {
	query := fmt.Sprintf("UPDATE %s SET %s",
		c.table,
		strings.Join(c.clauses, ", "),
	)

	if c.where != "" {
		query += " WHERE " + c.where
	}

	if c.orderBy != "" {
		query += " ORDER BY " + c.orderBy
	}

	if c.limit > 0 {
		query += " LIMIT " + strconv.Itoa(c.limit)
	}

	if c.offset > 0 {
		query += " OFFSET " + strconv.Itoa(c.offset)
	}

	if c.returning != "" {
		query += " RETURNING " + c.returning
	}

	return query, append(c.args, c.whereArgs...)
}
