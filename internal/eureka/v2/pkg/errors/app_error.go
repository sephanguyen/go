package errors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppError struct {
	RootErr error    `json:"-"`
	Message string   `json:"message"`
	Key     ErrorKey `json:"errorKey"`
}

type ErrorKey string

const (
	ErrGeneral ErrorKey = "ErrGeneral"
	// Usecase layer

	ErrEntityNotFound ErrorKey = "ErrEntityNotFound"

	// General error for repo layer

	ErrLearnosityRequestFailed ErrorKey = "ErrLearnosityRequestFailed"
	ErrDB                      ErrorKey = "ErrDB"
	ErrNoRowsExisted           ErrorKey = "ErrNoRowsExisted"
	ErrNoRowsAffected          ErrorKey = "ErrNoRowsAffected"
	ErrAPIRespondNotFound      ErrorKey = "ErrAPIRespondNotFound"

	// ErrConversion for some convert, casting, asserting type errors
	ErrConversion ErrorKey = "ErrConversion"

	ErrInputValidation ErrorKey = "ErrInputValidation"
)

func (a *AppError) Error() string {
	return fmt.Errorf("[%s] %s: %v", a.Key, a.Message, a.RootError()).Error()
}

// RootError
// Get root error recursively.
// If the deepest level is nil, so return the 2nd deepest level.
func (a *AppError) RootError() error {
	if err, ok := a.RootErr.(*AppError); ok {
		if err.RootErr == nil {
			return err
		}
		if innerErr, ok := err.RootErr.(*AppError); ok {
			if innerErr.RootErr == nil {
				return innerErr
			}
			return innerErr.RootError()
		}
		return err.RootErr
	}
	return a.RootErr
}

func NewAppError(key ErrorKey, msg string, root error) *AppError {
	return &AppError{
		Key:     key,
		Message: msg,
		RootErr: root,
	}
}

func NewConversionError(msg string, root error) *AppError {
	return NewAppError(ErrConversion, msg, root)
}

func NewDBError(msg string, root error) *AppError {
	return NewAppError(ErrDB, msg, root)
}

func NewNoRowsUpdatedError(msg string, root error) *AppError {
	return NewAppError(ErrNoRowsAffected, msg, root)
}

func NewEntityNotFoundError(msg string, root error) *AppError {
	return NewAppError(ErrEntityNotFound, msg, root)
}

func NewNoRowsExistedError(msg string, root error) *AppError {
	return NewAppError(ErrNoRowsExisted, msg, root)
}

func NewLearnosityError(msg string, root error) *AppError {
	return NewAppError(ErrLearnosityRequestFailed, msg, root)
}

func NewValidationError(msg string, root error) *AppError {
	return NewAppError(ErrInputValidation, msg, root)
}

// New creates a new error.
// root error can be nil
func New(msg string, root error) *AppError {
	return NewAppError(ErrGeneral, msg, root)
}

func CheckErrType(errType ErrorKey, err error) bool {
	if appError, ok := err.(*AppError); ok {
		return appError.Key == errType
	}
	return false
}

func GetGrpcStatusCode(err error, errMap map[ErrorKey]codes.Code) codes.Code {
	code := codes.Internal
	if appError, ok := err.(*AppError); ok {
		if statusCode, ok := errMap[appError.Key]; ok {
			return statusCode
		}
	}
	return code
}

func NewGrpcError(err error, errMap map[ErrorKey]codes.Code) error {
	return status.Error(GetGrpcStatusCode(err, errMap), fmt.Errorf("%w", err).Error())
}

// IsPgxNoRows Check if an error is a wrapper of or a pgx.ErrNoRows
// Not applied for `AppError`
func IsPgxNoRows(err error) bool {
	if err == pgx.ErrNoRows {
		return true
	}

	// Recursive check for underlying error
	if err = errors.Unwrap(err); err != nil {
		return IsPgxNoRows(err)
	}

	return false
}
