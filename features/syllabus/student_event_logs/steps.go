package student_event_logs

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

type StepState struct {
	Token               string
	UserID              string
	StudentIDs          []string
	StudentEventLogs    []*epb.StudentEventLog
	EStudentEventLogs_1 []*entities.StudentEventLog
	StudyPlanID         string
	StudyPlanItemID     string
	LmIDs               []string
	StudyPlanItemIDs    []string
	Response            interface{}
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		// BEGIN====common=====BEGIN
		`^<student_event_logs>a signed in "([^"]*)"$`: s.aSignedIn,
		`^<learning_time>a signed in "([^"]*)"$`:      s.aSignedIn,

		// insert student event logs
		`^student insert event log$`:                                         s.studentCreateEventLog,
		`^student insert event log for learning time$`:                       s.studentCreateEventLogv2,
		`^student event log must be created$`:                                s.studentEventLogMustBeCreated,
		`^learning time is calculated$`:                                      s.calculateLearningTime,
		`^student event log must be created with study_plan_item_id column$`: s.studentEventLogWithStudyPlanItemIDColumnMustBeCreated,
		`^max score must be stored$`:                                         s.maxScoreMustBeStored,
		`^student\'s event log is stored$`:                                   s.studentsEventLogIsStored,
	}

	return steps
}

func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	userID, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.UserID = userID
	stepState.Token = authToken

	return utils.StepStateToContext(ctx, stepState), nil
}
