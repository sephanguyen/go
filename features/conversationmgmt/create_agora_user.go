package conversationmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/cucumber/godog"
)

type CreateAgoraUserSuite struct {
	*common.ConversationMgmtSuite
}

func (c *SuiteConstructor) InitCreateAgoraUser(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &CreateAgoraUserSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates a students with first name is "([^"]*)" and last name is "([^"]*)"$`:                                 s.CreatesAStudentWithFirstNameAndLastName,
		`^a new staff with role teacher is created$`:                                                                                s.StaffWithRoleTeacher,
		`^agora teacher is created successfully$`:                                                                                   s.agoraUserIsCreatedSuccessfully,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CreateAgoraUserSuite) agoraUserIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	staffID := commonState.CurrentStaff.ID

	query := `
		SELECT count(*)
		FROM agora_user au
		WHERE au.user_id = $1
		AND au.deleted_at IS NULL;
	`

	err := try.Do(func(attempt int) (bool, error) {
		var userCount int
		err := s.TomDBConn.QueryRow(ctx, query, database.Text(staffID)).Scan(&userCount)
		if err != nil {
			return false, fmt.Errorf("failed QueryRow: %+v", err)
		}

		if userCount == 0 && attempt < 10 {
			time.Sleep(2 * time.Second)
			return attempt < 10, nil
		}

		if userCount != 1 {
			return false, fmt.Errorf("expected find 1 agora user with id %s", staffID)
		}
		return false, nil
	})

	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
