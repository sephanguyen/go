package bob

import (
	"context"
	"errors"

	"github.com/manabie-com/backend/internal/bob/entities"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/lestrrat-go/jwx/jwt"
)

func (s *suite) aUserIdRetrieveBasicProfileRequest(ctx context.Context, userId string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch userId {
	case "valid":
		t, _ := jwt.ParseString(stepState.AuthToken)
		stepState.Request = &bpb.RetrieveBasicProfileRequest{
			UserIds: []string{t.Subject()},
		}
	case "invalid":
		stepState.Request = &bpb.RetrieveBasicProfileRequest{
			UserIds: []string{"invalid"},
		}
	case "missing":
		stepState.Request = &bpb.RetrieveBasicProfileRequest{}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserRetrieveBasicProfileRequest(ctx context.Context, userRole string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &bpb.RetrieveBasicProfileRequest{}

	if userRole == "student" {
		t, _ := jwt.ParseString(stepState.AuthToken)
		stepState.Request = &bpb.RetrieveBasicProfileRequest{
			UserIds: []string{t.Subject()},
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserRetrievesBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = bpb.NewUserReaderServiceClient(s.Conn).RetrieveBasicProfile(contextWithToken(s, ctx), stepState.Request.(*bpb.RetrieveBasicProfileRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserRetrievesBasicProfileWithoutMetadata(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = bpb.NewUserReaderServiceClient(s.Conn).RetrieveBasicProfile(context.Background(), stepState.Request.(*bpb.RetrieveBasicProfileRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aBobMustReturnsBasicProfile(ctx context.Context, total int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*bpb.RetrieveBasicProfileResponse)

	if len(resp.Profiles) != total {
		return StepStateToContext(ctx, stepState), errors.New("total profile returns does not match")
	}
	if len(resp.Profiles) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	req := stepState.Request.(*bpb.RetrieveBasicProfileRequest)
	for _, userId := range req.UserIds {
		found := false
		for _, profile := range resp.Profiles {
			if userId == profile.UserId {
				if err := s.validateUserProfile(ctx, userId, profile); err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				found = true
				break
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), errors.New("not found userID: " + userId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateUserProfile(ctx context.Context, userId string, profile *cpb.BasicProfile) error {
	stmt := `SELECT name, given_name, user_group, country FROM users WHERE user_id=$1 LIMIT 1`
	rows, err := s.DB.Query(ctx, stmt, userId)
	if err != nil {
		return err
	}
	defer rows.Close()
	user := entities.User{}
	for rows.Next() {
		err := rows.Scan(
			&user.LastName,
			&user.GivenName,
			&user.Group,
			&user.Country,
		)
		if err != nil {
			return err
		}
	}

	if profile.Name != user.GetName() {
		return errors.New("username in response is incorrect")
	}
	if profile.Country != cpb.Country(cpb.Country_value[user.Country.String]) {
		return errors.New("country in response is incorrect")
	}
	if profile.Group != cpb.UserGroup(cpb.UserGroup_value[user.Group.String]) {
		return errors.New("user_group in response is incorrect")
	}

	return nil
}
