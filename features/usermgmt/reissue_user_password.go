package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

func (s *suite) reissuesUsersPasswordWithNonexistingUser(ctx context.Context, caller string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, caller)

	req := &pb.ReissueUserPasswordRequest{
		UserId:      newID(),
		NewPassword: newID(),
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = s.reissuePassword(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theOwnerReissuesUsersPassword(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existedUser := stepState.SrcUser
	token, err := s.generateExchangeToken(existedUser.GetUID(), stepState.CurrentUserGroup)
	if err != nil {
		return nil, err
	}
	newPassword := newID()
	stepState.SrcUser = user.NewUser(user.WithUID(existedUser.GetUID()), user.WithEmail(existedUser.GetEmail()), user.WithRawPassword(newPassword))

	ctx = common.ValidContext(ctx, OrgIDFromCtx(ctx), existedUser.GetUID(), token)

	req := &pb.ReissueUserPasswordRequest{
		UserId:      stepState.SrcUser.GetUID(),
		NewPassword: stepState.SrcUser.GetRawPassword(),
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = s.reissuePassword(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) reissuesUsersPasswordWhenMissingField(ctx context.Context, caller, fieldMissing string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, caller)

	req := &pb.ReissueUserPasswordRequest{
		UserId:      stepState.SrcUser.GetUID(),
		NewPassword: stepState.SrcUser.GetRawPassword(),
	}
	switch fieldMissing {
	case "user id":
		req.UserId = ""
	case "new password":
		req.NewPassword = ""
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = s.reissuePassword(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) reissuesUsersPassword(ctx context.Context, caller string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, caller)

	existedUser := stepState.SrcUser
	newPassword := newID()

	stepState.SrcUser = user.NewUser(user.WithUID(existedUser.GetUID()), user.WithEmail(existedUser.GetEmail()), user.WithRawPassword(newPassword))
	req := &pb.ReissueUserPasswordRequest{
		UserId:      stepState.SrcUser.GetUID(),
		NewPassword: stepState.SrcUser.GetRawPassword(),
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = s.reissuePassword(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCanSignInWithTheNewPassword(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	email := stepState.SrcUser.GetEmail()
	password := stepState.SrcUser.GetRawPassword()
	if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], email, password); err != nil {
		return ctx, errors.Wrap(err, "loginIdentityPlatform")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserCreateUser(ctx context.Context, accountType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, fmt.Sprintf("Get feature toggle error(%s)", pkg_unleash.FeatureToggleUserNameStudentParent))
	}
	switch accountType {
	case StaffRoleTeacher:
		return s.createRandomStaff(ctx, "user group was granted teacher role")
	case student:
		student, err := CreateStudent(ctx, s.UserMgmtConn, nil, getChildrenLocation(OrgIDFromCtx(ctx)))
		if err != nil {
			return nil, fmt.Errorf("create student: %s", err.Error())
		}
		stepState.SrcUser = &entity.LegacyUser{
			ID:         database.Text(student.StudentProfile.Student.UserProfile.UserId),
			Email:      database.Text(student.StudentProfile.Student.UserProfile.Email),
			LoginEmail: database.Text(student.StudentProfile.Student.UserProfile.Email),
		}
		stepState.CurrentUserGroup = constant.UserGroupStudent

	case parent:
		student, err := CreateStudent(ctx, s.UserMgmtConn, nil, getChildrenLocation(OrgIDFromCtx(ctx)))
		if err != nil {
			return nil, fmt.Errorf("create student: %s", err.Error())
		}
		parent, err := CreateParent(ctx, s.UserMgmtConn, nil, student.StudentProfile.Student.UserProfile.UserId)
		if err != nil {
			return nil, fmt.Errorf("create parent: %s", err.Error())
		}
		loginEmail := parent.ParentProfiles[0].Parent.UserProfile.Email
		if isEnableUsername {
			loginEmail = parent.ParentProfiles[0].Parent.UserProfile.UserId + constant.LoginEmailPostfix
		}
		stepState.SrcUser = &entity.LegacyUser{
			ID:         database.Text(parent.ParentProfiles[0].Parent.UserProfile.UserId),
			Email:      database.Text(parent.ParentProfiles[0].Parent.UserProfile.Email),
			LoginEmail: database.Text(loginEmail),
		}
		stepState.CurrentUserGroup = constant.UserGroupParent
	}

	return ctx, nil
}

func (s *suite) reissuePassword(ctx context.Context, req *pb.ReissueUserPasswordRequest) (*pb.ReissueUserPasswordResponse, error) {
	return pb.NewUserModifierServiceClient(s.UserMgmtConn).ReissueUserPassword(ctx, req)
}
