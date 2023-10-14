package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	types "github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aSignedInWithSchool(ctx context.Context, role string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccountV2(ctx, role)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ASignedInWithSchool(ctx context.Context, role string, schoolID int) (context.Context, error) {
	return s.aSignedInWithSchool(ctx, role, schoolID)
}
func (s *suite) aValidStudentWithSchoolID(ctx context.Context, id string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	sql := "UPDATE students SET school_id = $1 WHERE student_id = $2"
	_, err := s.DB.Exec(ctx, sql, &schoolID, &id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aSignedInSchoolAdminWithSchoolID(ctx context.Context, group string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	ctx, err := s.aValidSchoolAdminProfileWithId(ctx, id, group, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, group)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aSignedInTeacherWithSchoolID(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	ctx, err := s.aValidTeacherProfileWithId(ctx, id, int32(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentTeacherID = id
	stepState.CurrentUserID = id

	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidSchoolAdminProfileWithId(ctx context.Context, id, userGroup string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entities.SchoolAdmin{}
	database.AllNullEntity(&c)

	c.SchoolAdminID.Set(id)
	c.SchoolID.Set(schoolID)
	now := time.Now()
	if err := c.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(now); err != nil {
		return nil, err
	}

	num := rand.Int()
	u := entities.User{}
	database.AllNullEntity(&u)

	u.ID = c.SchoolAdminID
	u.LastName.Set(fmt.Sprintf("valid-school-admin-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-school-admin-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-school-admin-%d", num))
	u.Country.Set(pb.COUNTRY_VN.String())
	if userGroup == "" {
		userGroup = entities.UserGroupSchoolAdmin
	}
	u.Group.Set(userGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = c.CreatedAt
	u.UpdatedAt = c.UpdatedAt
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)

	userRepo := repositories.UserRepo{}

	err := userRepo.Create(ctx, s.DB, &u)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	schoolAdminRepo := repositories.SchoolAdminRepo{}
	err = schoolAdminRepo.CreateMultiple(ctx, s.DB, []*entities.SchoolAdmin{&c})

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	ug := entities.UserGroup{}
	database.AllNullEntity(&ug)

	ug.UserID.Set(id)
	ug.GroupID.Set(userGroup)
	ug.UpdatedAt.Set(now)
	ug.CreatedAt.Set(now)
	ug.IsOrigin.Set(true)
	ug.Status.Set(entities.UserGroupStatusActive)

	userGroupRepo := repositories.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.DB, &ug)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	err = s.createUserAccessPath(ctx, id)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidSchoolAdminProfileWithId createUserAccessPath %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userUpdatedHisOwnProfile(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	currentID := t.Subject()
	profile := generateUserProfile(group)
	profile.Id = currentID
	stepState.Request = &pb.UpdateUserProfileRequest{
		Profile: profile,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userAskBobToDoUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.createUserDeviceTokenCreatedSubscription(ctx)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserDeviceTokenCreatedSubscription: %v", err)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).UpdateUserProfile(s.signedCtx(ctx), stepState.Request.(*pb.UpdateUserProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserAskBobToDoUpdate(ctx context.Context) (context.Context, error) {
	return s.userAskBobToDoUpdate(ctx)
}
func (s *suite) userUpdatedOtherProfile(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	profile := generateUserProfile(group)
	if group == entities.UserGroupTeacher {
		id := s.newID()

		ctx, err := s.aValidTeacherProfileWithId(ctx, id, 1)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		profile.Id = id
	} else if group == entities.UserGroupSchoolAdmin {
		id := s.newID()

		ctx, err := s.aValidSchoolAdminProfileWithId(ctx, id, group, 1)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		profile.Id = id
	} else {
		profile.Id = stepState.CurrentStudentID
	}
	stepState.Request = &pb.UpdateUserProfileRequest{
		Profile: profile,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustUpdateProfile(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	reqProfile := stepState.Request.(*pb.UpdateUserProfileRequest).Profile
	currentUserID := reqProfile.Id
	if arg1 == "his own" {
		//his own profile
		t, err := jwt.ParseString(stepState.AuthToken)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		currentUserID = t.Subject()
	}

	user := new(entities.User)
	fieldName, values := user.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id = $1", strings.Join(fieldName, ","))
	err := s.DBPostgres.QueryRow(ctx, query, &currentUserID).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	updatedAt, err := types.TimestampProto(user.UpdatedAt.Time)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	updatedProfile := &pb.UserProfile{
		Id:          user.ID.String,
		Name:        user.GetName(),
		Country:     pb.Country(pb.Country_value[user.Country.String]),
		PhoneNumber: user.PhoneNumber.String,
		Avatar:      user.Avatar.String,
		UserGroup:   user.Group.String,
		Email:       "",
		DeviceToken: "",
		CreatedAt:   nil,
		UpdatedAt:   updatedAt,
	}
	//api does not update device token and created at
	reqProfile.DeviceToken = ""
	reqProfile.CreatedAt = nil
	reqProfile.Email = ""

	if !reflect.DeepEqual(reqProfile, updatedProfile) {
		return StepStateToContext(ctx, stepState), errors.New("bob did not update his profile")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) BobMustUpdateProfile(ctx context.Context, arg1 string) (context.Context, error) {
	return s.bobMustUpdateProfile(ctx, arg1)
}
func generateUserProfile(group string) *pb.UserProfile {
	return &pb.UserProfile{
		Name:        fmt.Sprintf("user %d", rand.Int()),
		Country:     pb.COUNTRY_VN,
		PhoneNumber: fmt.Sprintf("+849%d", rand.Int()),
		Email:       fmt.Sprintf("valid-%d@email.com", rand.Int()),
		Avatar:      fmt.Sprintf("http://avatar-%d", rand.Int()),
		DeviceToken: fmt.Sprintf("random device %d", rand.Int()),
		UserGroup:   group,
		CreatedAt:   nil,
		UpdatedAt:   &types.Timestamp{Seconds: time.Now().Unix()},
	}
}
func (s *suite) userUpdatedProfileWithUserGroupNamePhoneEmailSchool(ctx context.Context, typeProfile, userGroup, name, phone, email string, schoolID int) (context.Context, error) {
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
		Country:     pb.COUNTRY_VN,
		PhoneNumber: phone,
		Email:       email,
		Avatar:      fmt.Sprintf("http://avatar-%d", num),
		DeviceToken: fmt.Sprintf("random device %d", num),
		UserGroup:   userGroup,
		CreatedAt:   nil,
		UpdatedAt:   &types.Timestamp{Seconds: time.Now().Unix()},
	}

	stepState := StepStateFromContext(ctx)
	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	id := t.Subject()
	if typeProfile == "other" {
		id = s.newID()
		generateUser := func(group string) (context.Context, error) {
			switch group {
			case pb.USER_GROUP_STUDENT.String():
				{
					if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
						return StepStateToContext(ctx, stepState), err
					}
					return s.aValidStudentWithSchoolID(ctx, id, schoolID)
				}
			case pb.USER_GROUP_TEACHER.String():
				{
					return s.aValidTeacherProfileWithId(ctx, id, int32(schoolID))
				}
			case pb.USER_GROUP_SCHOOL_ADMIN.String():
				{
					return s.aValidSchoolAdminProfileWithId(ctx, id, group, schoolID)
				}
			default:
				{ //admin
					return s.aSignedInAdminWithProfileId(ctx, id)
				}
			}
		}

		if ctx, err := generateUser(profile.UserGroup); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	profile.Id = id
	stepState.Request = &pb.UpdateUserProfileRequest{
		Profile: profile,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserUpdatedProfileWithUserGroupNamePhoneEmailSchool(ctx context.Context, typeProfile, userGroup, name, phone, email string, schoolID int) (context.Context, error) {
	return s.userUpdatedProfileWithUserGroupNamePhoneEmailSchool(ctx, typeProfile, userGroup, name, phone, email, schoolID)
}
