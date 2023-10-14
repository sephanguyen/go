package repository

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/errcode"
)

type InternalError struct {
	RawError error
}

func (err InternalError) Error() string {
	if err.RawError == nil {
		return "internal error but raw error has not been set"
	}
	return fmt.Sprintf(`interal error: '%s'`, err.RawError.Error())
}

func (err InternalError) DomainError() string {
	return "internal error"
}
func (err InternalError) DomainCode() int {
	return errcode.DomainCodeInternal
}
