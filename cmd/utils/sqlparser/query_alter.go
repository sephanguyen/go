package sqlparser

import (
	"errors"
	"fmt"
	"strings"
)

func NewSqlQueryAlter() *sqlQueryAlter {
	return new(sqlQueryAlter)
}

type sqlQueryAlter struct {
}

func (p *sqlQueryAlter) CanProcess(query string) bool {
	return strings.Index(query, "alter ") == 0
}

func (p *sqlQueryAlter) GetParser() SqlQueryParser {
	return p
}

func (p *sqlQueryAlter) Build(text string) (*SqlQuery, error) {
	splitedStrs := strings.Split(text, " ")
	numberStrs := len(splitedStrs)
	if numberStrs < 3 {
		return nil, fmt.Errorf("invalid query: %s", text)
	}

	var strs []string
	for _, str := range splitedStrs {
		s := strings.Trim(str, " ")
		if len(s) == 0 {
			continue
		}
		strs = append(strs, s)
	}

	entityType, tableName, err := p.parseQuery(text, strs)
	if err != nil {
		return nil, err
	}
	return &SqlQuery{
		Action: AlterAction,
		Entity: entityType,
		Name:   tableName,
	}, nil
}

func (p *sqlQueryAlter) parseQuery(text string, strs []string) (entityType EntityType, tableName string, err error) {
	numberStrs := len(strs)
	tableName = strings.ToLower(strs[2]) // ALTER TABLE {TABLE_NAME}
	entityType = ParseEntity(strs[1])

	if tableName == "only" {
		if numberStrs < 4 {
			err = errors.New("invalid query: " + text)
			return
		}
		tableName = strs[3] // ALTER TABLE ONLY {TABLE_NAME}
	} else if tableName == "if" {
		if numberStrs < 5 {
			err = errors.New("invalid query: " + text)
			return
		}
		tableName = strs[4] // ALTER TABLE IF EXISTS {TABLE_NAME}
	}

	idx := strings.Index(tableName, ";")
	if idx != -1 {
		tableName = tableName[0:idx]
	}
	idx = strings.Index(tableName, "(")
	if idx != -1 {
		tableName = tableName[0:idx]
	}
	return
}
