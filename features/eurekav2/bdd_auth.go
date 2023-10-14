package eurekav2

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

const (
	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"
)

func (s *suite) takeAnSignedInUser(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.SignedAsAccountV2(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("SignedAsAccountV2: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnRootContext(ctx context.Context) context.Context {
	return common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
}

func (s *suite) SignedAsAccountV2(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	roleWithLocation := usermgmt.RoleWithLocation{}
	adminCtx := s.returnRootContext(ctx)
	switch account {
	case unauthenticatedType:
		stepState.AuthToken = "random-token"
		stepState.UserID = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case "staff granted role school admin":
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case "staff granted role hq staff":
		roleWithLocation.RoleName = constant.RoleHQStaff
	case "staff granted role centre lead":
		roleWithLocation.RoleName = constant.RoleCentreLead
	case "staff granted role centre manager":
		roleWithLocation.RoleName = constant.RoleCentreManager
	case "staff granted role centre staff":
		roleWithLocation.RoleName = constant.RoleCentreStaff
	case "staff granted role teacher":
		roleWithLocation.RoleName = constant.RoleTeacher
	case "staff granted role teacher lead":
		roleWithLocation.RoleName = constant.RoleTeacherLead
	case studentType:
		roleWithLocation.RoleName = constant.RoleStudent
	case schoolAdminType:
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case teacherType:
		roleWithLocation.RoleName = constant.RoleTeacher
	case parentType:
		roleWithLocation.RoleName = constant.RoleParent
	}

	roleWithLocation.LocationIDs = []string{constants.ManabieOrgLocation}

	authInfo, err := usermgmt.SignIn(adminCtx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.CommonSuite.StepState.FirebaseAddress, s.Connections.UserMgmtConn, roleWithLocation, []string{constants.ManabieOrgLocation})

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("SignIn: %w", err)
	}

	stepState.UserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = constants.ManabieOrgLocation

	if account == studentType {
		stepState.StudentID = authInfo.UserID
	} else if account == teacherType {
		stepState.TeacherID = authInfo.UserID
	}

	ctx = common.ValidContext(ctx, constants.ManabieSchool, authInfo.UserID, authInfo.Token)

	return StepStateToContext(ctx, stepState), nil
}
