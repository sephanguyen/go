package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) generateAmountStudentParentWithoutUserGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	sampleAmoumtUser := 10
	// init IDs holder for assigning to StepState
	studentIDs := make([]string, 0)
	parentIDs := make([]string, 0)

	for index := 0; index < sampleAmoumtUser; index++ {
		// init parents assigned student payload and create parent & student
		req, err := s.addMultipleParentDataToCreateParentReq(ctx, 2)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Request = req

		if ctx, err := s.createMultipleNewParents(ctx, schoolAdminType, "2"); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when create parent: %w", err)
		}

		req = stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
		res := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse)
		// save student and parent ids
		studentIDs = append(studentIDs, req.StudentId)
		for _, parentProfile := range res.ParentProfiles {
			parentIDs = append(parentIDs, parentProfile.Parent.UserProfile.UserId)
		}
	}

	stepState.StudentIds = studentIDs
	stepState.ParentIDs = parentIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunMigrationJobAddDefaultUserGroup(ctx context.Context) (context.Context, error) {
	usermgmt.RunMigrationAddDefaultUserGroupForStudentParent(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	})
	return ctx, nil
}

func (s *suite) assertPreviousStudentAndParentHasUserGroup(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userIDsMustBeMigrated := append(stepState.StudentIds, stepState.ParentIDs...)

	fieldsName, _ := new(entity.UserGroupMember).FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM user_group_member WHERE user_id = ANY($1)`, strings.Join(fieldsName, ", "))
	rows, err := s.BobDBTrace.Query(ctx, query, database.TextArray(userIDsMustBeMigrated))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("query error when finding user group member: %w", err)
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows error when finding user group member: %w", err)
	}

	userGroupMembers := []*entity.UserGroupMember{}
	for rows.Next() {
		userGroupMember := new(entity.UserGroupMember)
		_, fields := userGroupMember.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("scan error when finding user group member: %w", err)
		}
		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	migratedUser := len(userGroupMembers)
	mustBeMigratedUser := len(userIDsMustBeMigrated)
	if migratedUser != mustBeMigratedUser {
		return StepStateToContext(ctx, stepState), fmt.Errorf("got %d users had not migrated", mustBeMigratedUser-migratedUser)
	}

	req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
	for _, userGroupMember := range userGroupMembers {
		if userGroupMember.ResourcePath.String != fmt.Sprint(req.GetSchoolId()) {
			return StepStateToContext(ctx, stepState),
				fmt.Errorf(
					"user group member of %s expect %s, but got %s",
					userGroupMember.UserID.String,
					fmt.Sprint(req.GetSchoolId()),
					userGroupMember.ResourcePath.String,
				)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
