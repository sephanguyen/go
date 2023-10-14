package sqlparser

import (
	"errors"
	"strings"
)

func NewSqlQueryCreate() *sqlQueryCreate {
	return new(sqlQueryCreate)
}

type sqlQueryCreate struct {
}

func (p *sqlQueryCreate) CanProcess(query string) bool {
	return strings.Index(query, "create ") == 0
}

func (p *sqlQueryCreate) GetParser() SqlQueryParser {
	return p
}

func (p *sqlQueryCreate) Build(text string) (*SqlQuery, error) {
	splitedStrs := strings.Split(text, " ")
	numberStrs := len(splitedStrs)
	if numberStrs < 3 {
		return nil, errors.New("invalid query")
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
		Action: CreateAction,
		Entity: entityType,
		Name:   tableName,
	}, nil
}

func (p *sqlQueryCreate) parseQuery(text string, strs []string) (entityType EntityType, tableName string, err error) {
	numberStrs := len(strs)
	entityType = ParseEntity(strs[1])
	tableName = strings.ToLower(strs[2]) // CREATE SCHEME/TABLE/INDEX {DATABASE_NAME}/{TABLE_NAME}/{INDEX_NAME}

	if entityType == IndexEntity {
		if tableName == "if" {
			if numberStrs < 8 {
				err = errors.New("invalid query: " + text)
				return
			}
			tableName = strs[7] // CREATE INDEX IF NOT EXISTS {INDEX_NAME} ON {TABLE_NAME}
		} else {
			if numberStrs < 5 {
				err = errors.New("invalid query: " + text)
				return
			}
			tableName = strs[4] // CREATE INDEX {INDEX_NAME} ON {TABLE_NAME}
		}
	} else if tableName == "if" {
		if numberStrs < 6 {
			err = errors.New("invalid query: " + text)
			return
		}
		tableName = strs[5] // CREATE TABLE IF NOT EXISTS {TABLE_NAME}
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
