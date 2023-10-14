package errcode

import (
	"errors"
	"fmt"
	"strings"
)

type Code int

var (
	Empty    Code
	Internal Code = 1
	Invalid  Code = 2
)
var (
	ErrUserEmailIsEmpty                  = errors.New("email cannot be empty")
	ErrUserEmailExists                   = errors.New("email already exists")
	ErrUserFirstNameIsEmpty              = errors.New("first name cannot be empty")
	ErrUserLastNameIsEmpty               = errors.New("last name cannot be empty")
	ErrUserFullNameIsEmpty               = errors.New("full name cannot be empty")
	ErrUserCountryIsEmpty                = errors.New("country cannot be empty")
	ErrUserLocationIsEmpty               = errors.New("location cannot be empty")
	ErrUserLocationsAreInvalid           = errors.New("locations are invalid")
	ErrUserPhoneNumberIsWrongType        = errors.New("phone number is wrong type")
	ErrUserPhoneNumberIsDuplicate        = errors.New("primary and secondary phone number is duplicate")
	ErrUserPrimaryPhoneNumberIsRedundant = errors.New("primary phone number is redundant")
	ErrUserUserGroupDoesNotExist         = errors.New("user group does not exist")
	ErrUserExternalUserIDExists          = errors.New("external user id already exists")
)

type UserError struct {
	Code      Code
	FieldName string
	Message   string
	Err       error
}

func (err UserError) Error() string {
	return fmt.Sprintf(`Code: %v, field: %s, message: %s, error: %v`, err.Code, err.FieldName, err.Message, err.Err)
}

type UserNotFoundErr struct {
	UserIDs []string
}

func (err UserNotFoundErr) Error() string {
	return fmt.Sprintf(`cannot found users with ids: "%s"`, strings.Join(err.UserIDs, ", "))
}

func NewUserNotFoundErr(userIDs ...string) UserNotFoundErr {
	return UserNotFoundErr{
		UserIDs: userIDs,
	}
}
