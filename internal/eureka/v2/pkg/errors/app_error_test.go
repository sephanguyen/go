package errors

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewAppError(t *testing.T) {
	t.Run("1 level root app", func(t *testing.T) {
		t.Parallel()
		// Arrange
		rootErr := New("root error", nil)
		key := ErrGeneral
		msg := "test error message"

		// Act
		appErr := NewAppError(key, msg, rootErr)

		// Assert
		assert.Equal(t, rootErr, appErr.RootErr)
		assert.Equal(t, key, appErr.Key)
		assert.Equal(t, msg, appErr.Message)
	})
}

func TestAppError_RootError(t *testing.T) {
	t.Run("Recursively get the deepest level of error when root is not nil", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err1 := fmt.Errorf("%s", "root level 1 error")
		err2 := New("wrapper level 2 error", err1)
		err3 := New("wrapper level 3 error", err2)
		rootErr := New("wrapper level 4 error", err3)
		key := ErrGeneral
		msg := "some message"

		// Act
		appErr := NewAppError(key, msg, rootErr)
		actual := appErr.RootError()

		// Assert
		assert.Equal(t, err1, actual)
	})

	t.Run("Recursively get the 2nd deepest level of error when the deepest is nil", func(t *testing.T) {
		// Arrange
		err1 := New("root level 1 error", nil)
		err2 := New("wrapper level 2 error", err1)
		rootErr := New("wrapper level 3 error", err2)
		key := ErrGeneral
		msg := "some message"

		// Act
		appErr := NewAppError(key, msg, rootErr)
		actual := appErr.RootError()

		// Assert
		assert.Equal(t, err1, actual)
	})
}

func TestAppError_Error(t *testing.T) {
	t.Run("Return text of the highest and deepest level only", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err1 := fmt.Errorf("%s", "root level 0 error")
		rootErr := New("root level 1 error", err1)
		key := ErrGeneral
		msg := "test error message"

		// Act
		appErr := NewAppError(key, msg, rootErr)
		actual := appErr.Error()

		// Assert
		assert.Equal(t, "[ErrGeneral] test error message: root level 0 error", actual)
	})

	t.Run("Return text of the highest and 2nd deepest when root is nil", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err1 := New("root level 0 error", nil)
		err2 := New("root level 1 error", err1)
		rootErr := New("root level 2 error", err2)
		key := ErrGeneral
		msg := "test error message"

		// Act
		appErr := NewAppError(key, msg, rootErr)
		actual := appErr.Error()

		// Assert
		assert.Equal(t, "[ErrGeneral] test error message: [ErrGeneral] root level 0 error: <nil>", actual)
	})
}

func TestCheckErrType(t *testing.T) {
	t.Run("Error is AppError with matching ErrorKey", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err := NewAppError(ErrGeneral, "test error message", nil)

		// Act
		result := CheckErrType(ErrGeneral, err)

		// Assert
		assert.True(t, result)
	})

	t.Run("Error is AppError with non-matching ErrorKey", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err := NewAppError(ErrConversion, "test error conversion", nil)

		// Act
		result := CheckErrType(ErrGeneral, err)

		// Assert
		assert.False(t, result)
	})

	t.Run("Error is not AppError", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err := fmt.Errorf("%s", "some random err")

		// Act
		result := CheckErrType(ErrGeneral, err)

		// Assert
		assert.False(t, result)
	})

	t.Run("Error is nil", func(t *testing.T) {
		t.Parallel()
		// Arrange
		var err error

		// Act
		result := CheckErrType(ErrGeneral, err)

		// Assert
		assert.False(t, result)
	})
}

func TestGetGrpcStatusCode(t *testing.T) {
	errMap := map[ErrorKey]codes.Code{
		ErrConversion:    codes.InvalidArgument,
		ErrNoRowsExisted: codes.NotFound,
	}

	t.Run("AppError with matching ErrorKey", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err := NewAppError(ErrConversion, "test error message", nil)

		// Act
		result := GetGrpcStatusCode(err, errMap)

		// Assert
		assert.Equal(t, codes.InvalidArgument, result)
	})

	t.Run("AppError with non-matching ErrorKey: return codes.Internal", func(t *testing.T) {
		t.Parallel()
		// Arrange
		var errKeyRd ErrorKey = "RandomKey"
		err := NewAppError(errKeyRd, "test error message", nil)

		// Act
		result := GetGrpcStatusCode(err, errMap)

		// Assert
		assert.Equal(t, codes.Internal, result)
	})

	t.Run("Error is not AppError: return codes.Internal", func(t *testing.T) {
		t.Parallel()
		// Arrange
		err := fmt.Errorf("some error %s", "")

		// Act
		result := GetGrpcStatusCode(err, errMap)

		// Assert
		assert.Equal(t, codes.Internal, result)
	})

	t.Run("Error is nil: return codes.Internal", func(t *testing.T) {
		t.Parallel()
		// Arrange
		var err error

		// Act
		result := GetGrpcStatusCode(err, errMap)

		// Assert
		assert.Equal(t, codes.Internal, result)
	})
}

func TestNewGrpcError(t *testing.T) {
	errMap := map[ErrorKey]codes.Code{
		ErrConversion:    codes.InvalidArgument,
		ErrNoRowsExisted: codes.NotFound,
	}

	t.Run("AppError with matching ErrorKey", func(t *testing.T) {
		// Arrange
		err := NewAppError(ErrConversion, "test error message", nil)
		expectedErr := status.Error(GetGrpcStatusCode(err, errMap), fmt.Errorf("%w", err).Error())

		// Act
		result := NewGrpcError(err, errMap)

		// Assert
		assert.Equal(t, codes.InvalidArgument, status.Code(result))
		assert.Equal(t, expectedErr, result)
	})

	t.Run("AppError with non-matching ErrorKey", func(t *testing.T) {
		// Arrange
		var errKeyRd ErrorKey = "RandomKey"
		err := NewAppError(errKeyRd, "test error message", nil)
		expectedErr := status.Error(GetGrpcStatusCode(err, errMap), fmt.Errorf("%w", err).Error())

		// Act
		result := NewGrpcError(err, errMap)

		// Assert
		assert.Equal(t, codes.Internal, status.Code(result))
		assert.Equal(t, expectedErr, result)
	})

	t.Run("Error is not AppError", func(t *testing.T) {
		// Arrange
		err := fmt.Errorf("some error %s", "")
		expectedErr := status.Error(GetGrpcStatusCode(err, errMap), fmt.Errorf("%w", err).Error())

		// Act
		result := NewGrpcError(err, errMap)

		// Assert
		assert.Equal(t, codes.Internal, status.Code(result))
		assert.Equal(t, expectedErr, result)
	})

	t.Run("Error is nil", func(t *testing.T) {
		// Arrange
		var err error
		expectedErr := status.Error(GetGrpcStatusCode(err, errMap), fmt.Errorf("%w", err).Error())

		// Act
		result := NewGrpcError(err, errMap)

		// Assert
		assert.Equal(t, codes.Internal, status.Code(result))
		assert.Equal(t, expectedErr, result)
	})
}

func TestIsPgxNoRows(t *testing.T) {
	t.Run("Error is pgx.ErrNoRows", func(t *testing.T) {
		// Arrange
		err := pgx.ErrNoRows

		// Act
		result := IsPgxNoRows(err)

		// Assert
		assert.True(t, result)
	})

	t.Run("Error is wrapped with pgx.ErrNoRows", func(t *testing.T) {
		// Arrange
		wrapper := fmt.Errorf("wrapped: %w", pgx.ErrNoRows)
		wrapper2 := fmt.Errorf("wrapped 2: %w", wrapper)

		// Act
		result := IsPgxNoRows(wrapper)
		result2 := IsPgxNoRows(wrapper2)

		// Assert
		assert.True(t, result)
		assert.True(t, result2)
	})

	t.Run("Error is not pgx.ErrNoRows", func(t *testing.T) {
		// Arrange
		err := fmt.Errorf("wrapped: %w", pgx.ErrTxClosed)

		// Act
		result := IsPgxNoRows(err)

		// Assert
		assert.False(t, result)
	})

	t.Run("Error is nil", func(t *testing.T) {
		// Arrange
		var err error

		// Act
		result := IsPgxNoRows(err)

		// Assert
		assert.False(t, result)
	})
}
