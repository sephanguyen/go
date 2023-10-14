package http

import (
	"fmt"
	"net/http"

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
	if err.RawError == nil {
		return "internal error but raw error has not been set"
	}
	return fmt.Sprintf(`interal error: '%s'`, err.RawError.Error())
}
func (err InternalError) DomainCode() int {
	return errcode.DomainCodeInternal
}

func DomainErrorToHTTPResponseStatusCode(domainError errcode.DomainError) int {
	if domainError == nil {
		return http.StatusOK
	}

	switch domainError.DomainCode() {
	case errcode.DomainCodeInvalid:
		return http.StatusBadRequest
	case errcode.DomainCodeNotFound:
		return http.StatusNotFound
	case errcode.DomainCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
