package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	oldpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"google.golang.org/protobuf/proto"
)

const (
	parent  = "parent"
	student = "student"
)

func (s *suite) aProfileOfUserWithUsergroupNamePhoneEmailSchool(ctx context.Context, profileType, userGroup, name, phone, email string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := rand.Int()
	if name != "" {
		name = fmt.Sprintf(name, num)
	}
	if phone != "" {
		phone = fmt.Sprintf(phone, num)
	}
	if email != "" {
		email = fmt.Sprintf(email, num)
	}
	profile := &pb.UserProfile{
		Name:        name,
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: phone,
		Email:       email,
		Avatar:      fmt.Sprintf("http://avatar-%d", num),
		DeviceToken: fmt.Sprintf("random device %d", num),
		UserGroup:   userGroup,
	}

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	id := t.Subject()

	if profileType == "other" {
		id = s.newID()
		currentToken := stepState.AuthToken
		generateUser := func(group string) (context.Context, error) {
			switch group {
			case cpb.UserGroup_USER_GROUP_STUDENT.String():
				{
					if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
						return StepStateToContext(ctx, stepState), err
					}

					return s.aValidStudentWithSchoolID(ctx, id, convertSchoolID(schoolID))
				}
			case cpb.UserGroup_USER_GROUP_TEACHER.String():
				{
					return s.aValidTeacherProfileWithId(ctx, id, int32(convertSchoolID(schoolID)))
				}
			case cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String():
				{
					return s.aValidSchoolAdminProfileWithId(ctx, id, group, convertSchoolID(schoolID))
				}
			default:
				{ // admin
					return s.aSignedInAdminWithProfileId(ctx, id)
				}
			}
		}

		if ctx, err := generateUser(profile.UserGroup); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.AuthToken = currentToken
	}
	profile.Id = id
	if profileType == "own" && phone == "" {
		stmt := "UPDATE users SET phone_number = null WHERE user_id = $1"
		_, err := s.DB.Exec(ctx, stmt, &id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	stepState.Request = &pb.UpdateUserProfileRequest{
		Profile: profile,
	}
	return StepStateToContext(ctx, stepState), nil
}

func convertSchoolID(schoolID int) int {
	if schoolID == 3 {
		return constants.ManabieSchool
	}
	return schoolID
}

func (s *suite) aSignedInUserWithSchool(ctx context.Context, role string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch role {
	case "school admin":
		return s.signedAsAccountV2(ctx, "school admin")
	case parent:
		return s.signedAsAccountV2(ctx, "parent")
	case student:
		{
			return s.signedAsAccountV2(ctx, "student")
		}
	case "teacher":
		return s.signedAsAccountV2(ctx, "staff granted role teacher")
	case "admin":
		return s.signedAsAccountV2(ctx, "school admin")
	case "unauthenticated":
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eventMustBePublishedToChannel(ctx context.Context, eventType, channelSubject string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(3 * time.Second)

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	select {
	case <-stepState.FoundChanForJetStream:
		switch eventType {
		case "EvtUserInfo":
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), errors.New("eventType is not valid")
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}

func (s *suite) profileOfUserMustBeUpdated(ctx context.Context, profileType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	reqProfile := stepState.Request.(*pb.UpdateUserProfileRequest).Profile
	currentUserID := reqProfile.Id
	if profileType == "own" {
		// his own profile
		t, err := jwt.ParseString(stepState.AuthToken)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		currentUserID = t.Subject()
	}

	user := new(entities.User)
	fieldName, values := user.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id = $1", strings.Join(fieldName, ","))
	err := s.DB.QueryRow(ctx, query, &currentUserID).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow %w", err)
	}

	updatedProfile := &pb.UserProfile{
		Id:          user.ID.String,
		Name:        user.GetName(),
		Country:     cpb.Country(cpb.Country_value[user.Country.String]),
		PhoneNumber: user.PhoneNumber.String,
		Avatar:      user.Avatar.String,
		UserGroup:   user.Group.String,
		Email:       "",
		DeviceToken: "",
	}
	// api does not update device token and created at
	reqProfile.DeviceToken = ""
	reqProfile.CreatedAt = nil
	reqProfile.Email = ""

	if !proto.Equal(reqProfile, updatedProfile) {
		return StepStateToContext(ctx, stepState), errors.New("user profile was not updated correctly")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

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
		case *pb.UpdateProfileRequest:
			if req.Name == msg.Name {
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
		return StepStateToContext(ctx, stepState), fmt.Errorf("S.JSM.Subscribe: %v", err)
	}

	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserDeviceTokenCreatedSubscription: %v", err)
	}

	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.Conn).UpdateUserProfile(s.signedCtx(ctx), stepState.Request.(*pb.UpdateUserProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInAdminWithProfileId(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	num := rand.Int()

	u := entities_bob.User{}
	database.AllNullEntity(&u)
	u.ID.Set(id)
	u.LastName.Set(fmt.Sprintf("valid-admin-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-admin-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-admin-%d", num))
	u.Country.Set(oldpb.COUNTRY_VN.String())
	u.Group.Set(entities_bob.UserGroupAdmin)
	u.CreatedAt.Set(now)
	u.UpdatedAt.Set(now)

	userRepo := repositories.UserRepo{}
	err := userRepo.Create(ctx, s.DB, &u)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	return StepStateToContext(ctx, stepState), nil
}
