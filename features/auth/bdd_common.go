package auth

import (
	"context"
	"strconv"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

func StepStateFromContext(ctx context.Context) *common.StepState {
	state := ctx.Value(common.StepStateKey{})
	if state == nil {
		return &common.StepState{}
	}
	return state.(*common.StepState)
}

func StepStateToContext(ctx context.Context, state *common.StepState) context.Context {
	return context.WithValue(ctx, common.StepStateKey{}, state)
}

// func (s *suite) userSignedInAs(ctx context.Context, role string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)

// 	ctx = s.signedIn(ctx, constants.ManabieSchool, role)
// 	stepState.CurrentUserID = s.MapOrgStaff[constants.ManabieSchool][usermgmt.GetRoleFromConstant(role)].UserID
// 	stepState.AuthToken = s.MapOrgStaff[constants.ManabieSchool][usermgmt.GetRoleFromConstant(role)].Token
// 	stepState.OrganizationID = strconv.Itoa(constants.ManabieSchool)
// 	stepState.CurrentSchoolID = constants.ManabieSchool
// 	ctx = interceptors.ContextWithUserID(ctx, stepState.CurrentUserID)
// 	ctx = interceptors.ContextWithUserGroup(ctx, usermgmt.GetLegacyUserGroupFromConstant(role))

// 	return StepStateToContext(ctx, stepState), nil
// }

func (s *suite) userSignedInAsInOrganization(ctx context.Context, role string, org string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orgID := 0

	switch org {
	case "manabie":
		orgID = constants.ManabieSchool
	case "jprep":
		orgID = constants.JPREPSchool
	case "kec-demo":
		orgID = constants.KECDemo
	}

	ctx = s.signedIn(ctx, orgID, role)
	stepState.CurrentUserID = s.MapOrgStaff[orgID][usermgmt.GetRoleFromConstant(role)].UserID
	stepState.AuthToken = s.MapOrgStaff[orgID][usermgmt.GetRoleFromConstant(role)].Token
	stepState.OrganizationID = strconv.Itoa(orgID)
	stepState.CurrentSchoolID = int32(orgID)
	ctx = interceptors.ContextWithUserID(ctx, stepState.CurrentUserID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signedIn(ctx context.Context, orgID int, role string) context.Context {
	authInfo := s.getAuthInfo(orgID, role)
	ctx = contextWithToken(ctx, authInfo.Token)
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(orgID),
			UserID:       authInfo.UserID,
			UserGroup:    usermgmt.GetLegacyUserGroupFromConstant(role),
		},
	})

	return ctx
}

func (s *suite) getAuthInfo(orgID int, account string) common.AuthInfo {
	switch role := usermgmt.GetRoleFromConstant(account); role {
	case "unauthenticatedType":
		return common.AuthInfo{
			UserID: newID(),
			Token:  "invalidToken",
		}

	default:
		return s.MapOrgStaff[orgID][role]
	}
}
