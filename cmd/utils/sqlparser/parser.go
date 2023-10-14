package sqlparser

import (
	"bufio"
	"os"
	"strings"
)

func NewSqlParser() *SqlParser {
	return &SqlParser{
		validators: []SqlQueryProcessor{
			NewSqlQueryAlter(),
			NewSqlQueryCreate(),
			NewSqlQueryDrop(),
		},
	}
}

type SqlParser struct {
	validators []SqlQueryProcessor
}

func (p *SqlParser) ParseFromFile(filePath string) ([]*SqlQuery, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var (
		line      string
		validator SqlQueryProcessor
		queries   []*SqlQuery
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = strings.ToLower(strings.Trim(scanner.Text(), " "))
		if validator == nil {
			for _, val := range p.validators {
				if val.CanProcess(line) {
					validator = val
					break
				}
			}
			if validator == nil {
				continue
			}
		}

		parser := validator.GetParser()
		query, err := parser.Build(line)
		if err != nil {
			return nil, err
		}

		if query == nil {
			continue
		}

		if query.Entity != UnknownEntity {
			queries = append(queries, query)
		}
		validator = nil
	}

	return queries, scanner.Err()
}
