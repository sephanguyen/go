package invoicemgmt

import (
	"context"
	"fmt"
	"os"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/yasuo/constant"
)

// signedAsAccount used for signing in an account. This method will create the user first based on its user group and will generate an exchange token.
func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		userGroup string
		role      string
		err       error
	)

	switch group {
	case invoiceConst.UserGroupUnauthenticated:
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case invoiceConst.UserGroupStudent:
		ctx, err = s.createStudent(ctx)
		userGroup = constant.UserGroupStudent // change to phone if have an error
		role = userConstant.RoleStudent
	case invoiceConst.UserGroupSchoolAdmin:
		ctx, err = s.createSchoolAdmin(ctx)
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleSchoolAdmin
	case invoiceConst.UserGroupTeacher:
		ctx, err = s.createTeacher(ctx)
		userGroup = constant.UserGroupTeacher
		role = userConstant.RoleTeacher
	case invoiceConst.UserGroupParent:
		ctx, err = s.createParent(ctx)
		userGroup = constant.UserGroupParent
		role = userConstant.RoleParent
	case invoiceConst.UserGroupHQStaff:
		ctx, err = s.createSchoolAdmin(ctx)
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleHQStaff
	case invoiceConst.UserGroupCentreManager:
		ctx, err = s.createSchoolAdmin(ctx)
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleCentreManager
	case invoiceConst.UserGroupCentreStaff:
		ctx, err = s.createSchoolAdmin(ctx)
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleCentreStaff
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Save to user group
	err = InsertEntities(
		stepState,
		s.EntitiesCreator.CreateUserGroupV2(ctx, s.BobDBTrace, role),
		s.EntitiesCreator.CreateUserGroupMember(ctx, s.BobDBTrace),
		s.EntitiesCreator.CreateGrantedRole(ctx, s.BobDBTrace, role),
		s.EntitiesCreator.CreateGrantedRoleAccessPath(ctx, s.BobDBTrace, role),
		s.EntitiesCreator.CreateUserAccessPath(ctx, s.BobDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Wait for kafka sync of bob entities
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	// Generate Auth Token
	stepState.AuthToken, err = s.generateExchangeToken(stepState.CurrentUserID, userGroup, int64(stepState.CurrentSchoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createParent creates a parent user
func (s *suite) createParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateParent(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createStudent creates a student user with access path
func (s *suite) createStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID := idutil.ULIDNow()

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateStudent(ctx, s.BobDBTrace, studentID),
		s.EntitiesCreator.CreateUserAccessPathForStudent(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// sync time for user basic info
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	return StepStateToContext(ctx, stepState), nil
}

// createSchoolAdmin creates a school admin user
func (s *suite) createSchoolAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateSchoolAdmin(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	claim := interceptors.JWTClaimsFromContext(ctx)
	claim.Manabie.UserID = stepState.CurrentUserID
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)

	return StepStateToContext(ctx, stepState), nil
}

// createTeacher creates a teacher user
func (s *suite) createTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateTeacher(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

// createLocation creates a location
func (s *suite) createLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateLocation(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func createTempFile(tempDir, filename string) (*os.File, error) {
	fileTemp, err := os.Create(fmt.Sprintf("%s-%s", tempDir, filename))
	if err != nil {
		return nil, err
	}

	return fileTemp, nil
}

func cleanup(tempDir string) {
	if err := os.RemoveAll(tempDir); err != nil {
		logger.Warnf("os.RemoveAll: %v", err)
	}
}
