package user

import (
	"github.com/manabie-com/backend/internal/golibs/errorx"

	"github.com/pkg/errors"
)

var (
	ErrUserIsNil       = errors.New("user is nil")
	ErrUserListEmpty   = errors.New("user list is empty")
	ErrInvalidLangCode = errors.New("invalid lang code")
	ErrTenantIDIsEmpty = errors.New("tenant id is empty")

	ErrUserNotFound   = errors.New("user not found")
	ErrTenantNotFound = errors.New("tenant not found")
)

type userValidationError struct {
	fieldName    UserField
	reason       errorx.InvalidArgumentReason
	errorMessage string
}

func newUserValidationError(fieldName UserField, reason errorx.InvalidArgumentReason, errorMessage string) error {
	e := &userValidationError{
		fieldName:    fieldName,
		reason:       reason,
		errorMessage: errorMessage,
	}
	return e
}

func (e *userValidationError) FieldName() string {
	return string(e.fieldName)
}

func (e *userValidationError) Reason() errorx.InvalidArgumentReason {
	return e.reason
}

func (e *userValidationError) Error() string {
	return e.errorMessage
}

func IsUserValidationErr(err error) bool {
	_, ok := err.(*userValidationError)
	return ok
}

var (
	ErrUserUIDEmpty          = newUserValidationError(UserFieldUID, errorx.InvalidArgumentReasonIsEmpty, "user's uid can not be empty")
	ErrUIDMaxLength          = newUserValidationError(UserFieldUID, errorx.InvalidArgumentReasonGreaterThanMaximumLength, "user's uid must not be longer than 128 characters")
	ErrUserEmailEmpty        = newUserValidationError(UserFieldEmail, errorx.InvalidArgumentReasonIsEmpty, "user's email can not be empty")
	ErrUserPhoneNumberEmpty  = newUserValidationError(UserFieldPhoneNumber, errorx.InvalidArgumentReasonIsEmpty, "user's phone number can not be empty")
	ErrUserDisplayNameEmpty  = newUserValidationError(UserFieldDisplayName, errorx.InvalidArgumentReasonIsEmpty, "user's display name can not be empty")
	ErrUserPhotoURLEmpty     = newUserValidationError(UserFieldPhotoURL, errorx.InvalidArgumentReasonIsEmpty, "user's photo url can not be empty")
	ErrUserPasswordMinLength = newUserValidationError(UserFieldRawPassword, errorx.InvalidArgumentReasonSmallerThanMinimumLength, "user's password length must be larger than 6")
	ErrUserPasswordMaxLength = newUserValidationError(UserFieldRawPassword, errorx.InvalidArgumentReasonGreaterThanMaximumLength, "user's password length can not be larger than 1024")
)
