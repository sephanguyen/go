package service

import (
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToStatusError(t *testing.T) {
	t.Parallel()
	assert.Equal(t, status.Error(codes.InvalidArgument, "InvalidArgument"), ToStatusError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation, Message: "InvalidArgument"}))
	assert.Equal(t, status.Error(codes.AlreadyExists, "AlreadyExists"), ToStatusError(&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "AlreadyExists"}))
	assert.Equal(t, status.Error(codes.Unknown, pgx.ErrNoRows.Error()), ToStatusError(pgx.ErrNoRows))
}

func TestToStatusErrorPhoneAlreadyExist(t *testing.T) {
	t.Parallel()
	actual := toStatusError(&pgconn.PgError{
		Code:           pgerrcode.UniqueViolation,
		Message:        "users_phone_un AlreadyExists",
		ConstraintName: "users_phone_un",
	})

	stt := status.Convert(actual)
	for _, d := range stt.Details() {
		switch info := d.(type) {
		case *errdetails.PreconditionFailure:
			assert.Equal(t, phoneAlreadyExistMsg.Violations[0].Type, info.Violations[0].Type)
			assert.Equal(t, phoneAlreadyExistMsg.Violations[0].Subject, info.Violations[0].Subject)
			assert.Equal(t, phoneAlreadyExistMsg.Violations[0].Description, info.Violations[0].Description)
		default:
			t.Error("unexpected error detail returned")
		}
	}
}

func TestToStatusErrorEmailAlreadyExist(t *testing.T) {
	t.Parallel()
	actual := toStatusError(&pgconn.PgError{
		Code:           pgerrcode.UniqueViolation,
		Message:        "users_email_un AlreadyExists",
		ConstraintName: "users_email_un",
	})

	stt := status.Convert(actual)
	for _, d := range stt.Details() {
		switch info := d.(type) {
		case *errdetails.PreconditionFailure:
			assert.Equal(t, emailAlreadyExistMsg.Violations[0].Type, info.Violations[0].Type)
			assert.Equal(t, emailAlreadyExistMsg.Violations[0].Subject, info.Violations[0].Subject)
			assert.Equal(t, emailAlreadyExistMsg.Violations[0].Description, info.Violations[0].Description)
		default:
			t.Error("unexpected error detail returned")
		}
	}
}
