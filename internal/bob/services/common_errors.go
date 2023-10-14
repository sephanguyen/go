package services

import (
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emailAlreadyExist = "emailAlreadyExist"
	phoneAlreadyExist = "phoneAlreadyExist"
)

var (
	emailAlreadyExistMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        emailAlreadyExist,
				Subject:     "registration",
				Description: "email already exist",
			},
		},
	}

	phoneAlreadyExistMsg = &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{
			{
				Type:        phoneAlreadyExist,
				Subject:     "registration",
				Description: "phone number already exist",
			},
		},
	}
)

// ToStatusError used by all services
func ToStatusError(err error) error {
	return toStatusError(err)
}

func toStatusError(err error) error {
	switch e := errors.Cause(err).(type) {
	case *pgconn.PgError:
		switch e.Code {
		case pgerrcode.ForeignKeyViolation: // foreign_key_violation
			return status.Error(codes.InvalidArgument, e.Message)
		case pgerrcode.UniqueViolation: // unique_violation
			stt := status.New(codes.AlreadyExists, e.Message)
			if e.ConstraintName == "users_phone_un" {
				stt, _ = stt.WithDetails(phoneAlreadyExistMsg)
			} else if e.ConstraintName == "users_email_un" {
				stt, _ = stt.WithDetails(emailAlreadyExistMsg)
			}

			return stt.Err()
		}
	}

	return status.Convert(err).Err()
}
