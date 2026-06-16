package database

import (
	"fmt"
	"strings"
	"time"
)

type Insert struct {
	table     string
	columns   []string
	values    []any
	returning string
	Index     int
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

	if i.returning != "" {
		query += " RETURNING " + i.returning
	}

	if len(i.values) == 0 {
		return query, nil
	}

	return query, i.values
}
