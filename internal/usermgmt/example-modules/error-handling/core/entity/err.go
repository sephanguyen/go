package entity

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/errcode"
)

type InvalidFieldError struct {
	EntityName string
	Index      int
	FieldName  string
}

func (err InvalidFieldError) Error() string {
	return fmt.Sprintf(`%s of %s[%v] is invalid`, err.FieldName, err.EntityName, err.Index)
}
func (err InvalidFieldError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s invalid`, err.EntityName, err.Index, err.FieldName)
}
func (err InvalidFieldError) DomainCode() int {
	return errcode.DomainCodeInvalid
}

type NotFoundError struct {
	EntityName         string
	Index              int
	SearchedFieldName  string
	SearchedFieldValue string
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf(`can not find '%s' with '%s' = '%s'`, err.EntityName, err.SearchedFieldName, err.SearchedFieldValue)
}
func (err NotFoundError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s notfound`, err.EntityName, err.Index, err.SearchedFieldName)
}
func (err NotFoundError) DomainCode() int {
	return errcode.DomainCodeNotFound
}
