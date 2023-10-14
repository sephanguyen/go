package bob

import (
	"context"

	types "github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) userRetrievesHisOwnProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).GetCurrentUserProfile(s.signedCtx(ctx), &pb.GetCurrentUserProfileRequest{})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnsUserOwnProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	if stepState.Response == nil {
		return StepStateToContext(ctx, stepState), errors.New("bob does not return user profile")
	}
	pProfile := stepState.Response.(*pb.GetCurrentUserProfileResponse).Profile
	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	currentUserID := t.Subject()

	userId := new(pgtype.Text)
	userId.Set(currentUserID)

	query := `SELECT user_id, country, name, avatar, phone_number, email, device_token, allow_notification, ` +
		`user_group, updated_at, created_at, is_tester FROM users WHERE user_id = $1`
	row := s.DB.QueryRow(ctx, query, userId)
	eProfile := new(entities_bob.User)
	row.Scan(
		&eProfile.ID,
		&eProfile.Country,
		&eProfile.LastName,
		&eProfile.Avatar,
		&eProfile.PhoneNumber,
		&eProfile.Email,
		&eProfile.DeviceToken,
		&eProfile.AllowNotification,
		&eProfile.Group,
		&eProfile.UpdatedAt,
		&eProfile.CreatedAt,
		&eProfile.IsTester,
	)

	if !isEqualUserEnAndPb(pProfile, eProfile) {
		return StepStateToContext(ctx, stepState), errors.New("return profile is not correct")
	}
	return StepStateToContext(ctx, stepState), nil
}
func isEqualUserEnAndPb(p *pb.UserProfile, e *entities_bob.User) bool {
	if p.CreatedAt == nil {
		p.CreatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	if p.UpdatedAt == nil {
		p.UpdatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	updatedAt := &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	createdAt := &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}
	return (e.GetName() == p.Name) && (pb.Country(pb.Country_value[e.Country.String]) == p.Country) &&
		(e.PhoneNumber.String == p.PhoneNumber) && (e.Email.String == p.Email) && (e.Avatar.String == p.Avatar) &&
		(e.DeviceToken.String == p.DeviceToken) && (e.Group.String == p.UserGroup) && updatedAt.Equal(p.UpdatedAt) &&
		createdAt.Equal(p.CreatedAt)
}
