package mastermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

func (s *suite) compareFieldViolations(s1, s2 *errdetails.BadRequest_FieldViolation) bool {
	return s1.GetField() == s2.GetField() && s1.GetDescription() == s2.GetDescription()
}

// compare expected bad-request error with response error
func (s *suite) compareBadRequest(ctx context.Context, err error, br *errdetails.BadRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	st := status.Convert(err)
	for _, detail := range st.Details() {
		switch errorType := detail.(type) {
		case *errdetails.BadRequest:
			expectedViolations := br.GetFieldViolations()
			violations := errorType.GetFieldViolations()
			if len(expectedViolations) != len(violations) {
				_, _ = s.printBadRequest(ctx, stepState.ResponseErr)
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong field violation count, expected: %d, got: %d", len(expectedViolations), len(violations))
			}
			for i, v := range expectedViolations {
				if !s.compareFieldViolations(v, violations[i]) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("wrong field violation, expected: %v, got: %v", v, violations[i])
				}
			}

		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("%s", "error must be bad request")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) printBadRequest(ctx context.Context, err error) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	st := status.Convert(err)
	for _, detail := range st.Details() {
		switch errorType := detail.(type) {
		case *errdetails.BadRequest:
			violations := errorType.GetFieldViolations()
			for _, v := range violations {
				fmt.Printf("Field: %s, description: %s", v.Field, v.Description)
			}

		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("%s", "error must be bad request")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateULIDs(number int) []string {
	if number < 1 {
		number = 1
	}
	arr := make([]string, 0, number)
	for i := 0; i < number; i++ {
		arr = append(arr, idutil.ULID(time.Now().Add(time.Second*time.Duration(number))))
	}
	return arr
}
