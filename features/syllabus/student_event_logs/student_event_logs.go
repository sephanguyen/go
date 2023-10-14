package student_event_logs

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// nolint: gosec
func (s *Suite) studentCreateEventLog(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx((ctx), stepState.Token)
	types := []string{"study_guide_finished", "video_finished", "learning_objective", "quiz_answer_selected"}
	randomNo := rand.Intn(len(types))
	mockLog := []*epb.StudentEventLog{
		{
			EventId:   idutil.ULIDNow(),
			EventType: types[randomNo],
			Payload: &epb.StudentEventLogPayload{
				StudyPlanItemId: "study-plan-item-id-1",
			},
			CreatedAt: timestamppb.Now(),
		},
		{
			EventId:   idutil.ULIDNow(),
			EventType: "learning_objective",
			Payload: &epb.StudentEventLogPayload{
				TopicId: "TopicId",
			},
			CreatedAt: timestamppb.Now(),
		}}
	err := utils.GenerateStudentEventLogs(ctx, mockLog, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error generate student event log: %w", err)
	}

	stepState.StudentEventLogs = mockLog

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentEventLogMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	query := `SELECT count(*) FROM student_event_logs WHERE student_id = $1`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.UserID).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if count != 2 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of student_event_logs %d, got %d\nid: %s", 2, count, stepState.UserID)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentEventLogWithStudyPlanItemIDColumnMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var count int

	query := `SELECT count(*) FROM student_event_logs WHERE student_id = $1 AND study_plan_item_id = $2`
	err := s.EurekaDB.QueryRow(ctx, query, stepState.UserID, &stepState.StudentEventLogs[0].Payload.StudyPlanItemId).Scan(&count)

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
	}

	if count != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of student_event_logs %d, got %d", 1, count)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
