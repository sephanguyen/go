package sqlparser

import (
	"errors"
	"fmt"
	"strings"
)

func NewSqlQueryDrop() *sqlQueryDrop {
	return new(sqlQueryDrop)
}

type sqlQueryDrop struct {
}

func (p *sqlQueryDrop) CanProcess(query string) bool {
	return strings.Index(query, "drop ") == 0
}

func (p *sqlQueryDrop) GetParser() SqlQueryParser {
	return p
}

func (p *sqlQueryDrop) Build(text string) (*SqlQuery, error) {
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

func (p *sqlQueryDrop) parseQuery(text string, strs []string) (entityType EntityType, tableName string, err error) {
	numberStrs := len(strs)
	entityType = ParseEntity(strs[1])
	tableName = strings.ToLower(strs[2]) // DROP TABLE/INDEX {TABLE_NAME}/{INDEX_NAME}

	if entityType == IndexEntity {
		if tableName == "if" {
			if numberStrs < 7 {
				err = errors.New("invalid query: " + text)
				return
			}
			tableName = strs[6] // DROP INDEX IF EXISTS {INDEX_NAME} ON {TABLE_NAME}
		} else {
			if numberStrs < 5 {
				err = errors.New("invalid query: " + text)
				return
			}
			tableName = strs[4] // DROP INDEX {INDEX_NAME} ON {TABLE_NAME}
		}
	} else if tableName == "if" {
		if numberStrs < 5 {
			err = errors.New("invalid query: " + text)
			return
		}
		tableName = strs[4] // DROP TABLE IF EXISTS {TABLE_NAME}
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
