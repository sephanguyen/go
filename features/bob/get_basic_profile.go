package bob

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/lestrrat-go/jwx/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) aUserIdGetBasicProfileRequest(ctx context.Context, userIDData string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch userIDData {
	case "valid":
		t, _ := jwt.ParseString(stepState.AuthToken)
		stepState.Request = &pb.GetBasicProfileRequest{
			UserIds: []string{t.Subject()},
		}
	case "invalid":
		stepState.Request = &pb.GetBasicProfileRequest{
			UserIds: []string{"invalid"},
		}
	case "missing id":
		stepState.Request = &pb.GetBasicProfileRequest{
			UserIds: nil,
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRetrievesBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).GetBasicProfile(contextWithToken(s, ctx), stepState.Request.(*pb.GetBasicProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRetrievesBasicProfileWithMissingMetadata(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).GetBasicProfile(context.Background(), stepState.Request.(*pb.GetBasicProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustReturnsBasicProfile(ctx context.Context, total int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetBasicProfileResponse)
	if len(resp.Profiles) != total {
		return StepStateToContext(ctx, stepState), fmt.Errorf("total profile returns does not match")
	}

	if len(resp.Profiles) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	req := stepState.Request.(*pb.GetBasicProfileRequest)

	for _, userId := range req.UserIds {
		found := false
		for _, p := range resp.Profiles {
			if p.UserId == userId {
				found = true
				break
			}
			if p.UserGroup == "" {
				return StepStateToContext(ctx, stepState), fmt.Errorf("bob did not return usergroup")
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found userID: " + userId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCannotRetrievesBasicProfileWhenMissing(ctx context.Context, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return nil, fmt.Errorf("expecting err but got nil")
	}

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	switch data {
	case "userId":
		if stt.Code() != codes.InvalidArgument {
			return nil, fmt.Errorf("expecting %s but got", stt.Code())
		}
		if stt.Message() != "missing userIds" {
			return nil, fmt.Errorf("expecting %s but got", stt.Message())
		}
	case "metadata":
		if stt.Code() != codes.InvalidArgument {
			return nil, fmt.Errorf("expecting %s but got", stt.Code())
		}
		if stt.Message() != "missing package name" {
			return nil, fmt.Errorf("expecting missing package name but got %s", stt.Message())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
