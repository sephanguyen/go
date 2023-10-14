package http

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
)

type InternalError struct {
	RawErr error
}

func (err InternalError) Error() string {
	return fmt.Sprintf(`In port layer, internal error: '%s'`, err.RawErr.Error())
}

func (err InternalError) DomainError() string {
	return "internal error"
}
func (err InternalError) DomainCode() int {
	return errcode.InternalError
}
