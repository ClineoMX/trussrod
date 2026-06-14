package database

import (
	"fmt"
	"strconv"
	"strings"
)

type whereCondition struct {
	column string
	value  any
}

type joinClause struct {
	joinType string
	table    string
	on       string
	onColumn string
	onValue  any
	hasValue bool
}

type Select struct {
	table           string
	columns         []string
	whereConditions []whereCondition
	joins           []joinClause
	orderBy         string
	orderByArgs     []any
	limit           int
	offset          int
	Index           int
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

func (s *Select) InnerJoin(table, on string) *Select {
	s.joins = append(s.joins, joinClause{joinType: "INNER", table: table, on: on})
	return s
}

func (s *Select) LeftJoin(table, on string) *Select {
	s.joins = append(s.joins, joinClause{joinType: "LEFT", table: table, on: on})
	return s
}

func (s *Select) RightJoin(table, on string) *Select {
	s.joins = append(s.joins, joinClause{joinType: "RIGHT", table: table, on: on})
	return s
}

func (s *Select) FullJoin(table, on string) *Select {
	s.joins = append(s.joins, joinClause{joinType: "FULL", table: table, on: on})
	return s
}

func (s *Select) InnerJoinEq(table, leftColumn, rightColumn string) *Select {
	return s.InnerJoin(table, fmt.Sprintf("%s = %s", leftColumn, rightColumn))
}

func (s *Select) LeftJoinEq(table, leftColumn, rightColumn string) *Select {
	return s.LeftJoin(table, fmt.Sprintf("%s = %s", leftColumn, rightColumn))
}

func (s *Select) InnerJoinOn(table, column string, value any) *Select {
	s.joins = append(s.joins, joinClause{
		joinType: "INNER",
		table:    table,
		onColumn: column,
		onValue:  value,
		hasValue: true,
	})
	s.Index++
	return s
}

func (s *Select) Where(column string, value any) *Select {
	s.whereConditions = append(s.whereConditions, whereCondition{column: column, value: value})
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

func (s *Select) buildJoins(index int, args *[]any) (string, int) {
	if len(s.joins) == 0 {
		return "", index
	}

	clauses := make([]string, 0, len(s.joins))
	for _, join := range s.joins {
		if join.hasValue {
			clauses = append(clauses, fmt.Sprintf(
				"%s JOIN %s ON %s = $%d",
				join.joinType,
				join.table,
				join.onColumn,
				index,
			))
			*args = append(*args, join.onValue)
			index++
			continue
		}

		clauses = append(clauses, fmt.Sprintf("%s JOIN %s ON %s", join.joinType, join.table, join.on))
	}

	return " " + strings.Join(clauses, " "), index
}

func (s *Select) buildWhere(index int, args *[]any) (string, int) {
	if len(s.whereConditions) == 0 {
		return "", index
	}

	clauses := make([]string, 0, len(s.whereConditions))
	for _, condition := range s.whereConditions {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", condition.column, index))
		*args = append(*args, condition.value)
		index++
	}

	return " WHERE " + strings.Join(clauses, " AND "), index
}

func (s *Select) Build() (string, []any) {
	args := make([]any, 0, len(s.joins)+len(s.whereConditions))
	index := 1

	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(s.columns, ", "),
		s.table,
	)

	joinSQL, index := s.buildJoins(index, &args)
	query += joinSQL

	whereSQL, _ := s.buildWhere(index, &args)
	query += whereSQL

	if s.orderBy != "" {
		query += " ORDER BY " + s.orderBy
	}
	if s.limit > 0 {
		query += " LIMIT " + strconv.Itoa(s.limit)
	}
	if s.offset > 0 {
		query += " OFFSET " + strconv.Itoa(s.offset)
	}

	if len(args) == 0 {
		return query, nil
	}

	return query, args
}
