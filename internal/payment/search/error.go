package search

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

type ElasticError struct {
	ErrType    string
	Reason     string
	rootCauses string
}

func (es ElasticError) Error() string {
	return fmt.Sprintf("errtype: %s\nreason: %s\nrootCauses: %s\n", es.ErrType, es.Reason, es.rootCauses)
}

func NewElasticError(e *elastic.ErrorDetails) error {
	rootCauses := ""
	for _, smallCause := range e.RootCause {
		rootCauses += NewElasticError(smallCause).Error()
	}
	return ElasticError{
		ErrType:    e.Type,
		Reason:     e.Reason,
		rootCauses: rootCauses,
	}
}
