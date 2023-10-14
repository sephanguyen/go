package errcode

import (
	"github.com/pkg/errors"
)

var (
	ErrUserAddressTypeIsEmpty   = errors.New("user address type is empty")
	ErrUserAddressTypeIsInvalid = errors.New("user address type is invalid")
)
