package bob

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) userCheckAnThatInDB(ctx context.Context, field, existent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	val := "for-sure-nonexistd-1231234123"
	if existent == "existed" {
		err := s.DB.QueryRow(ctx, "SELECT "+field+", user_id FROM users WHERE user_group = 'USER_GROUP_STUDENT' AND "+field+" IS NOT NULL ORDER BY random() LIMIT 1").
			Scan(&val, &stepState.expectingUserID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	req := &pb.CheckProfileRequest{}
	if field == "email" {
		req.Filter = &pb.CheckProfileRequest_Email{
			Email: val,
		}
	} else if field == "phone_number" {
		req.Filter = &pb.CheckProfileRequest_Phone{
			Phone: val,
		}
	}

	stepState.Request = req
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).CheckProfile(ctx, stepState.Request.(*pb.CheckProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) thatBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.CheckProfileResponse)
	if stepState.expectingUserID != "" && resp.Profile.UserId != stepState.expectingUserID {
		return ctx, fmt.Errorf("expecting user_id %s, got: %s", stepState.expectingUserID, resp.Profile.UserId)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCheckAnWithEmptyValue(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.CheckProfileRequest{}
	if field == "email" {
		req.Filter = &pb.CheckProfileRequest_Email{
			Email: "",
		}
	} else if field == "phone_number" {
		req.Filter = &pb.CheckProfileRequest_Phone{
			Phone: "",
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).CheckProfile(contextWithValidVersion(ctx), stepState.Request.(*pb.CheckProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}
