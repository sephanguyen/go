package field

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
)

var _ errcode.DomainError = InvalidDataError{}

type InvalidDataError struct {
	Method       string
	FieldName    string
	Reason       string
	InvalidValue interface{}
}

func (err InvalidDataError) Error() string {
	return fmt.Sprintf(`field %s invalid: %s, invalid value: %s`, err.FieldName, err.Reason, err.InvalidValue)
}
func (err InvalidDataError) DomainError() string {
	return fmt.Sprintf(`%s invalid`, err.FieldName)
}
func (err InvalidDataError) DomainCode() int {
	return errcode.InvalidData
}
