package crypt

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrDataToEncryptIsEmpty = errors.New("data to encrypt to empty")
	ErrDataToDecryptIsEmpty = errors.New("data to decrypt to empty")
	ErrInvalidPadding       = errors.New("invalid padding")
)

type ErrInvalidBlockLength int

func (e ErrInvalidBlockLength) Error() string {
	return fmt.Sprintf("invalid block length %d", e)
}

type ErrInvalidDataLength int

func (e ErrInvalidDataLength) Error() string {
	return fmt.Sprintf("invalid data length %d", e)
}
