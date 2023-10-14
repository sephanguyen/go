package flashcard

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userCreateFlashcardStudy(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp, err := sspb.NewFlashcardClient(s.EurekaConn).
		CreateFlashCardStudy(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.CreateFlashCardStudyRequest{
			StudyPlanId: stepState.StudyPlanID,
			LmId:        stepState.FlashcardID,
			StudentId:   stepState.StudentID,
			Paging: &cpb.Paging{
				Limit:  0,
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 1},
			},
		})
	stepState.Response, stepState.ResponseErr = resp, err
	if err == nil {
		stepState.StudySetID = resp.StudySetId
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemCreatesFlashcardProgressionCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.NewFlashcardClient: %w", stepState.ResponseErr)
	}

	resp := stepState.Response.(*sspb.CreateFlashCardStudyResponse)

	// retrieve FlashcardProgression from DB
	fcp := &entities.FlashcardProgression{}
	fields, values := fcp.FieldMap()
	if err := s.EurekaDB.QueryRow(
		ctx,
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE student_id = $1 AND learning_material_id = $2 AND study_plan_id = $3 AND deleted_at IS NULL",
			strings.Join(fields, ", "),
			fcp.TableName(),
		),
		database.Text(stepState.StudentID),
		database.Text(stepState.FlashcardID),
		database.Text(stepState.StudyPlanID),
	).Scan(values...); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error query FlashcardProgression: %w", err)
	}
	// retrieve flashcard quizzes from DB
	fcQuizzes := entities.Quizzes{}
	q := entities.Quiz{}
	fields, _ = q.FieldMap()
	if err := database.Select(ctx, s.EurekaDB,
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE lo_ids[1] = $1 AND deleted_at IS NULL",
			strings.Join(fields, ", "),
			q.TableName(),
		),
		database.Text(stepState.FlashcardID),
	).ScanAll(&fcQuizzes); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Query flashcard quizzes: %w", err)
	}
	// validate persistency
	if resp.StudySetId != fcp.StudySetID.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Unexpected StudySetId, expected: %s, got: %s", resp.StudySetId, fcp.StudySetID.String)
	}
	if resp.StudyingIndex != fcp.StudyingIndex.Int {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Unexpected StudyingIndex, expected: %d, got: %d", resp.StudyingIndex, fcp.StudyingIndex.Int)
	}
	if len(resp.Items) != len(fcQuizzes) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Unexpected Quizzes, expected: %v, got: %v", len(resp.Items), len(fcQuizzes))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
