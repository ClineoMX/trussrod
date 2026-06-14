package database

import (
	"fmt"
	"strconv"
	"strings"
)

type Select struct {
	table       string
	columns     []string
	where       string
	whereArgs   []any
	orderBy     string
	orderByArgs []any
	limit       int
	offset      int
	Index       int
}

func NewSelect(table string) *Select {
	return &Select{
		table:   table,
		columns: []string{},
		limit:   0,
		offset:  0,
		Index:   1,
	}
}

func (s *Select) Column(columns ...string) *Select {
	s.columns = append(s.columns, columns...)
	return s
}

func (s *Select) Where(column string, value any) *Select {
	if s.where != "" {
		s.where += " AND "
	}
	s.where += fmt.Sprintf("%s = $%d", column, s.Index)
	s.whereArgs = append(s.whereArgs, value)
	s.Index++
	return s
}

func (s *Select) OrderBy(column string, direction string) *Select {
	s.orderBy = fmt.Sprintf("%s %s", column, direction)
	return s
}

func (s *Select) Limit(limit int) *Select {
	s.limit = limit
	return s
}

func (s *Select) Offset(offset int) *Select {
	s.offset = offset
	return s
}

func (s *Select) Build() (string, []any) {
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(s.columns, ", "),
		s.table,
	)
	if s.where != "" {
		query += " WHERE " + s.where
	}
	if s.orderBy != "" {
		query += " ORDER BY " + s.orderBy
	}
	if s.limit > 0 {
		query += " LIMIT " + strconv.Itoa(s.limit)
	}
	if s.offset > 0 {
		query += " OFFSET " + strconv.Itoa(s.offset)
	}
	return query, s.whereArgs
}
