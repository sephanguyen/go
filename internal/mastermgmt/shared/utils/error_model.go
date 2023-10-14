package utils

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gotest.tools/assert"
)

// Get an error with detailed information about field violation/ file content
func GetValidationError(errs []*errdetails.BadRequest_FieldViolation) error {
	st := status.New(codes.InvalidArgument, "data is not valid, please check")
	br := &errdetails.BadRequest{}

	// st.WithDetails currently not support non-UTF8 characters.
	errs = sliceutils.Map(errs, removeNonUTF8FromError)

	br.FieldViolations = errs
	st, err := st.WithDetails(br)
	if err != nil {
		// If this errored, it will always error
		// here, so better panic so we can figure
		// out why than have this silently passing.
		panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
	}
	return st.Err()
}

func removeNonUTF8FromError(e *errdetails.BadRequest_FieldViolation) *errdetails.BadRequest_FieldViolation {
	if e == nil {
		return e
	}
	return &errdetails.BadRequest_FieldViolation{
		Description: removeNonUTF8(e.Description),
		Field:       e.Field,
	}
}

func removeNonUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.Map(func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		return r
	}, s)
}

func AssertBadRequestErrorModel(t *testing.T, expectedErr *errdetails.BadRequest, err error) {
	st := status.Convert(err)
	for _, detail := range st.Details() {
		switch errorType := detail.(type) {
		case *errdetails.BadRequest:
			expectedViolations := expectedErr.GetFieldViolations()
			violations := errorType.GetFieldViolations()
			assert.Equal(t, len(expectedViolations), len(violations))
			for i, v := range expectedViolations {
				assert.Equal(t, v.Field, violations[i].Field)
				assert.Equal(t, v.Description, violations[i].Description)
			}

		default:
			t.Error("Error model should be bad request")
		}
	}
}
