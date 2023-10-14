package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/pkg/errors"
)

const (
	amountUserToTest = 10
)

func (s *suite) createUserGroupWithRoleName(ctx context.Context, roleName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// sign in admin to create user group
	ctx, err := s.signedAsAccount(ctx, schoolAdminType)
	if err != nil {
		return nil, fmt.Errorf("s.signedAsAccount:%s : %w", schoolAdminType, err)
	}
	resp, err := s.createUserGroupWithRoleNames(ctx, []string{roleName})
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "s.createUserGroupWithRoleNames")
	}

	stepState.CurrentUserGroup = resp.UserGroupId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existedStaffInDBOfASchool(ctx context.Context, userGroup string, schoolName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resourcePath := fmt.Sprint(SchoolNameWithResourcePath[schoolName])
	userIDs, err := s.createAmountUserOfSchool(auth.InjectFakeJwtToken(ctx, resourcePath), userGroup, resourcePath, amountUserToTest)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "s.createAmountUserOfSchool")
	}

	// store to StepState for the next usages
	stepState.UserIDs = userIDs
	stepState.SchoolID = resourcePath
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunMigrationAssignUsergroupToSpecifyStaff(ctx context.Context, amountType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// pick amount user ids stored of the previous step in StepState
	var userIDsSequence string
	switch amountType {
	case "half":
		amountUserIDs := len(stepState.UserIDs) / 2
		userIDsSequence = strings.Join(stepState.UserIDs[:amountUserIDs], usermgmt.Separator)
		stepState.NumberOfIds = amountUserIDs
	case "none":
		userIDsSequence = ""
		stepState.NumberOfIds = 0
	}

	usermgmt.RunMigrationAssignUsergroupToSpecificStaff(
		auth.InjectFakeJwtToken(ctx, fmt.Sprint(stepState.SchoolID)),
		&configurations.Config{
			Common:     s.Cfg.Common,
			PostgresV2: s.Cfg.PostgresV2,
		},
		stepState.SchoolID,
		stepState.CurrentUserGroup,
		userIDsSequence,
	)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userMustHaveUserGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userGroupID := stepState.CurrentUserGroup
	schoolID := stepState.SchoolID
	totalUserIDs := database.TextArray(stepState.UserIDs)

	var expectedMigratedUserIDs []string
	// if no pick ids -> pick all ids
	if stepState.NumberOfIds == 0 {
		expectedMigratedUserIDs = stepState.UserIDs
	} else {
		expectedMigratedUserIDs = stepState.UserIDs[:stepState.NumberOfIds]
	}

	// assert by counting existed user group member in DB vs pass id to migration
	query := `
	  SELECT COUNT(ugm.*)
	  FROM user_group_member ugm

	  WHERE
	    ugm.resource_path = $1 AND
	    ugm.user_group_id = $2 AND
	    ugm.user_id = any($3)
	`
	counted := 0
	if err := s.BobDB.QueryRow(auth.InjectFakeJwtToken(ctx, stepState.SchoolID), query, schoolID, userGroupID, totalUserIDs).Scan(&counted); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrapf(err, "count user group member of user group %s was assigned to users", userGroupID)
	}

	if len(expectedMigratedUserIDs) != counted {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d users have %s user groups but %d", len(expectedMigratedUserIDs), userGroupID, counted)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createAmountUserOfSchool(ctx context.Context, userGroup string, schoolID string, amountUsers int) ([]string, error) {
	var userIDs []string
	for index := 0; index < amountUsers; index++ {
		userID := newID()
		if _, err := s.generateUser(ctx, userID, schoolID, userGroup); err != nil {
			return nil, errors.Wrap(err, "s.generateUser")
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}
