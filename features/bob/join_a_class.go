package bob

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/lestrrat-go/jwx/jwt"
)

func (s *suite) aJoinClassRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.JoinClassRequest{}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AJoinClassRequest(ctx context.Context) (context.Context, error) {
	return s.aJoinClassRequest(ctx)
}
func (s *suite) userJoinAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).JoinClass(contextWithToken(s, ctx), stepState.Request.(*pb.JoinClassRequest))

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserJoinAClass(ctx context.Context) (context.Context, error) {
	return s.userJoinAClass(ctx)
}
func (s *suite) aClassCodeInJoinClassRequest(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg == "valid" {
		stepState.Request.(*pb.JoinClassRequest).ClassCode = stepState.CurrentClassCode
	}
	if arg == "wrong" {
		stepState.Request.(*pb.JoinClassRequest).ClassCode = "$1111111"
	}

	if stepState.CurrentStudentID == "" {
		t, _ := jwt.ParseString(stepState.AuthToken)
		stepState.CurrentStudentID = t.Subject()
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AClassCodeInJoinClassRequest(ctx context.Context, arg string) (context.Context, error) {
	return s.aClassCodeInJoinClassRequest(ctx, arg)
}
func (s *suite) studentSubscriptionMustHasIsWithPlanIdIs(ctx context.Context, key, value, plan_id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT COUNT(*) FROM student_subscriptions WHERE plan_id = $1 AND "
	if key == "end_duration" {
		n, err := strconv.Atoi(value)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		//2020-10-10: 10 character
		expiredAt := time.Now().AddDate(0, 0, int(n)).String()[0:10]
		hour := time.Now().UTC().Hour()
		query += fmt.Sprintf("end_time BETWEEN '%s %d:00:00' AND '%s %d:59:59'", expiredAt, hour, expiredAt, hour)
	} else {
		conds := fmt.Sprintf("%s = '%s'", key, value)
		if value == "NULL" {
			conds = key + " IS NULL"
		}
		query += conds
	}

	count := 0
	err := s.DB.QueryRow(ctx, query, plan_id).Scan(&count)

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found student_subscriptions has %s is %s", key, value)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) StudentSubscriptionMustHasIsWithPlanIdIs(ctx context.Context, key, value, plan_id string) (context.Context, error) {
	return s.studentSubscriptionMustHasIsWithPlanIdIs(ctx, key, value, plan_id)
}
func (s *suite) joinClassResponseMustReturn(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.JoinClassResponse)
	if arg1 == "ClassId" && resp.ClassId == 0 {
		return StepStateToContext(ctx, stepState), errors.New("bob does not return class id")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) JoinClassResponseMustReturn(ctx context.Context, arg1 string) (context.Context, error) {
	return s.joinClassResponseMustReturn(ctx, arg1)
}
