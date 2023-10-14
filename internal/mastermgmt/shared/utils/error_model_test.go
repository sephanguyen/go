package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_GetValidationError(t *testing.T) {
	t.Parallel()

	t.Run("should remove utf-8 values", func(t *testing.T) {
		// arrange
		errs := []*errdetails.BadRequest_FieldViolation{
			{
				Field:       "field1",
				Description: string([]byte{0xff, 0xfe, 0xfd}) + " sample utf8",
			},
		}
		st := status.New(codes.InvalidArgument, "data is not valid, please check")
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "field1",
					Description: " sample utf8",
				},
			},
		}

		expectedErr, _ := st.WithDetails(br)

		// act
		res := GetValidationError(errs)

		// assert
		assert.Equal(t, res, expectedErr.Err())
	})
}
