package repository

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"

	"github.com/pkg/errors"
)

var ErrNoRowAffected = InternalError{
	RawError: errors.New("no rows affected"),
}

type InternalError struct {
	RawError error
}

func (err InternalError) Error() string {
	return fmt.Sprintf(`In repository layer, internal error: '%v'`, err.RawError)
}

func (err InternalError) DomainError() string {
	return "internal error"
}
func (err InternalError) DomainCode() int {
	return errcode.InternalError
}
