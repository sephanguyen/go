package eureka

import (
	"context"
	"fmt"
	"strconv"

	ec "github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/golibs/constants"
)

func (s *suite) logins(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	var err error
	switch role {
	case teacherRawText:
		ctx, stepState.TeacherID, stepState.TeacherToken, err = s.signedInAs(ctx, ec.RoleTeacher)
		stepState.AuthToken = stepState.TeacherToken
		return StepStateToContext(ctx, stepState), err
	case schoolAdminRawText, adminRawText:
		ctx, stepState.SchoolAdminID, stepState.SchoolAdminToken, err = s.signedInAs(ctx, ec.RoleSchoolAdmin)
		stepState.AuthToken = stepState.SchoolAdminToken
		return StepStateToContext(ctx, stepState), err
	case hqStaffRawText:
		ctx, stepState.HqStaffID, stepState.HqStaffToken, err = s.signedInAs(ctx, ec.RoleHQStaff)
		stepState.AuthToken = stepState.HqStaffToken
		return StepStateToContext(ctx, stepState), err
	case centerLeadRawText, centerManagerRawText, centerStaffRawText:
		// At this point we haven't implemented the new roles yet, so we will use parent permissions
		// on center lead, center manager, and center staff because they are most similar.
		// Please change this to use the new roles when they are implemented.
		ctx, err := s.aSignedIn(ctx, "parent")
		return StepStateToContext(ctx, stepState), err
	case studentRawText:
		ctx, stepState.StudentID, stepState.StudentToken, err = s.signedInAs(ctx, ec.RoleStudent)
		stepState.AuthToken = stepState.StudentToken
		return StepStateToContext(ctx, stepState), err
	case parentRawText:

		ctx, _, stepState.AuthToken, err = s.signedInAs(ctx, ec.RoleParent)
		return StepStateToContext(ctx, stepState), err
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("not support this role %s", role)
	}
}
