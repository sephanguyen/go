package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) generateUser(ctx context.Context, id, schoolID, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	organizationID := switchSchoolIDStringToSchoolID(schoolID)

	switch group {
	case cpb.UserGroup_USER_GROUP_STUDENT.String():
		if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		return s.aValidStudentWithSchoolID(ctx, id, organizationID)

	case cpb.UserGroup_USER_GROUP_TEACHER.String():
		return s.aValidTeacherProfileWithID(ctx, id, int32(organizationID))

	case cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String():
		return s.aValidSchoolAdminProfileWithId(ctx, id, group, organizationID)

	default:
		return s.aSignedInAdminWithProfileId(ctx, id)
	}
}

func (s *suite) userUpdateProfileSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	reqProfile := stepState.Request.(*pb.UpdateUserProfileRequest).Profile
	currentUserID := reqProfile.Id

	user := new(entity.LegacyUser)
	fieldName, values := user.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id = $1", strings.Join(fieldName, ","))
	err := s.BobDBTrace.QueryRow(ctx, query, &currentUserID).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := compareUserInDBAndRequest(reqProfile, user); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user profile was not updated correctly: %s", err.Error())
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) userCannotUpdateUserProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return ctx, errors.New("expected response has err but actual is nil")
	}
	return ctx, nil
}

func (s *suite) theSignedInUserUpdateProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.userUpdateProfileSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userUpdateProfileSubscription: %w", err)
	}

	stepState.Request = generateUpdateUserProfileRequest(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateUserProfile(ctx, stepState.Request.(*pb.UpdateUserProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserUpdateAnotherUserProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.userUpdateProfileSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userUpdateProfileSubscription: %w", err)
	}

	req := generateUpdateUserProfileRequest(ctx)
	req.Profile.Id = newID()
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateUserProfile(ctx, stepState.Request.(*pb.UpdateUserProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserUpdateUserProfileWithoutMandatoryField(ctx context.Context, missingField string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.userUpdateProfileSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userUpdateProfileSubscription: %w", err)
	}

	req := generateUpdateUserProfileRequest(ctx)
	if missingField == "name" {
		req.Profile.Name = ""
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateUserProfile(ctx, stepState.Request.(*pb.UpdateUserProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateProfileSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{nats.StartTime(time.Now()), nats.ManualAck(), nats.AckWait(2 * time.Second)},
	}
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	handleUserDeviceTokenUpdatedEvent := func(ctx context.Context, data []byte) (bool, error) {
		msg := &pb.EvtUserInfo{}
		err := proto.Unmarshal(data, msg)
		if err != nil {
			return true, err
		}
		switch req := stepState.Request.(type) {
		case *pb.UpdateUserProfileRequest:
			if req.Profile.Name == msg.Name {
				select {
				case stepState.FoundChanForJetStream <- stepState.Request:
					return false, nil
				case <-ctx.Done():
					return true, ctx.Err()
				}
			}
		case *pb.UpdateUserDeviceTokenRequest:
			if req.UserId == msg.UserId && req.DeviceToken == msg.DeviceToken {
				select {
				case stepState.FoundChanForJetStream <- stepState.Request:
					return false, nil
				case <-ctx.Done():
					return true, ctx.Err()
				}
			}
		}
		return false, nil
	}
	sub, err := s.JSM.Subscribe(constants.SubjectUserDeviceTokenUpdated, opts, handleUserDeviceTokenUpdatedEvent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userUpdateProfileSubscription: S.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserDeviceTokenCreatedSubscription: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func compareUserInDBAndRequest(reqProfile *pb.UpdateUserProfileRequest_UserProfile, user *entity.LegacyUser) error {
	firstNameReq, lastNameReq := helper.SplitNameToFirstNameAndLastName(reqProfile.Name)
	switch {
	case firstNameReq != user.FirstName.String || lastNameReq != user.LastName.String:
		return errors.New("name doesn't match")
	case reqProfile.Group != user.Group.String:
		return errors.New("user_group doesn't match")
	case reqProfile.Avatar != user.Avatar.String:
		return errors.New("avatar doesn't match")
	}

	return nil
}

func generateUpdateUserProfileRequest(ctx context.Context) *pb.UpdateUserProfileRequest {
	userGroup := ""
	userID := ""
	claims := interceptors.JWTClaimsFromContext(ctx)
	if claims != nil {
		userGroup = claims.Manabie.UserGroup
		userID = claims.Manabie.UserID
	}
	random := newID()
	profile := &pb.UpdateUserProfileRequest_UserProfile{
		Id:          userID,
		Name:        random,
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: random,
		Email:       fmt.Sprintf("%s@example.com", random),
		Avatar:      fmt.Sprintf("http://avatar-%s", random),
		DeviceToken: fmt.Sprintf("random device %s", random),
		Group:       userGroup,
	}

	return &pb.UpdateUserProfileRequest{
		Profile: profile,
	}
}
