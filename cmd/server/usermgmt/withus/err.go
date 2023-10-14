package withus

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
)

type InternalError struct {
	Index  int
	RawErr error
	UserID string
}

func (err InternalError) Error() string {
	return err.RawErr.Error()
}

func (err InternalError) DomainError() string {
	return "internal error"
}
func (err InternalError) DomainCode() int {
	return errcode.InternalError
}

type MissingMandatoryError struct {
	Index      int
	FieldName  string
	EntityName string
}

func (err MissingMandatoryError) Error() string {
	return fmt.Sprintf(`The field '%s' is required in entity '%s at index %d'`, err.FieldName, err.EntityName, err.Index)
}
func (err MissingMandatoryError) DomainError() string {
	return fmt.Sprintf(`%s[%v].%s required`, err.EntityName, err.Index, err.FieldName)
}
func (err MissingMandatoryError) DomainCode() int {
	return errcode.MissingMandatory
}
