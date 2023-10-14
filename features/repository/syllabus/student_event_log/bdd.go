package student_event_log

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

type Suite utils.Suite[StepState]

func (s *Suite) insertStudyPlanItem(ctx context.Context) error {
	s.StepState.StudyPlanItemID = idutil.ULIDNow()
	s.StepState.StudyPlanID = idutil.ULIDNow()
	s.StepState.MasterStudyPlanID = s.StepState.StudyPlanID
	now := time.Now()

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO study_plans (study_plan_id, master_study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
		database.Text(s.StepState.StudyPlanID),
		database.Text(s.StepState.MasterStudyPlanID),
		database.Timestamptz(now),
	); err != nil {
		return fmt.Errorf("insert study_plans: %w", err)
	}
	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO study_plan_items (study_plan_item_id, study_plan_id, created_at, updated_at) VALUES($1, $2, $3, $3);`,
		database.Text(s.StepState.StudyPlanItemID),
		database.Text(s.StepState.StudyPlanID),
		database.Timestamptz(now),
	); err != nil {
		return fmt.Errorf("insert study_plan_items: %w", err)
	}

	return nil
}

func (s *Suite) upsertStudentEventLogByEventType(ctx context.Context, eventType, createdAt string) error {
	at, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return err
	}

	types := []string{"study_guide_finished", "video_finished", "learning_objective", "quiz_answer_selected"}
	randomNo := rand.Intn(len(types))
	payload := &epb.StudentEventLogPayload{
		LoId:            s.StepState.LearningMaterialID,
		Event:           eventType,
		SessionId:       idutil.ULIDNow(),
		StudyPlanItemId: s.StepState.StudyPlanItemID,
	}

	if _, err := s.DB.Exec(
		ctx,
		`INSERT INTO student_event_logs (event_id, student_id, event_type, payload, study_plan_item_id, created_at) VALUES($1, $2, $3, $4, $5, $6);`,
		database.Varchar(idutil.ULIDNow()),
		database.Text(s.StepState.StudentID),
		database.Varchar(types[randomNo]),
		database.JSONB(payload),
		database.Text(s.StepState.StudyPlanItemID),
		database.Timestamptz(at),
	); err != nil {
		return err
	}

	return nil
}
