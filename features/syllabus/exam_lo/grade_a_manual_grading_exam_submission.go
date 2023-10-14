package exam_lo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userGradesASubmissionAnswersToStatus(ctx context.Context, statusChange string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	alloc := &entities.AllocateMarker{}
	database.AllNullEntity(alloc)
	uuid := idutil.ULIDNow()
	now := time.Now()
	if err := multierr.Combine(
		alloc.AllocateMarkerID.Set(uuid),
		alloc.TeacherID.Set(stepState.UserID),
		alloc.StudentID.Set(stepState.StudentID),
		alloc.StudyPlanID.Set(stepState.StudyPlanID),
		alloc.LearningMaterialID.Set(stepState.LearningMaterialID),
		alloc.CreatedAt.Set(now),
		alloc.CreatedBy.Set(fmt.Sprintf("school_admin_id_%v", uuid)),
		alloc.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup alloc: %w", err)
	}

	allocRepo := repositories.AllocateMarkerRepo{}
	if err := allocRepo.BulkUpsert(ctx, s.EurekaDB, []*entities.AllocateMarker{alloc}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat alloc: %w", err)
	}

	response := stepState.Response.(*sspb.ListExamLOSubmissionScoreResponse)
	teacherExamGrades := make([]*sspb.TeacherExamGrade, 0)
	stepState.TotalTeacherExamGrades = 0
	for _, scoreResp := range response.SubmissionScores {
		teacherExamGrade := &sspb.TeacherExamGrade{
			QuizId:            scoreResp.Core.ExternalId,
			TeacherPointGiven: wrapperspb.UInt32(scoreResp.GradedPoint.Value + uint32(rand.Intn(5)+3)),
			TeacherComment:    scoreResp.TeacherComment,
			Correctness:       scoreResp.Correctness,
			IsAccepted:        scoreResp.IsAccepted,
		}
		teacherExamGrades = append(teacherExamGrades, teacherExamGrade)
		stepState.TotalTeacherExamGrades++
	}

	_, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).GradeAManualGradingExamSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.GradeAManualGradingExamSubmissionRequest{
		SubmissionId:      stepState.SubmissionID,
		ShuffledQuizSetId: stepState.ShuffledQuizSetID,
		TeacherFeedback:   response.TeacherFeedback,
		SubmissionStatus:  sspb.SubmissionStatus(sspb.SubmissionStatus_value[statusChange]),
		TeacherExamGrades: teacherExamGrades,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsGradedScoreCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	query := `SELECT count(*) FROM exam_lo_submission_score WHERE submission_id = $1`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.SubmissionID).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if int32(count) != stepState.TotalTeacherExamGrades {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of TotalTeacherExamGrades %d, got %d", stepState.TotalTeacherExamGrades, count)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
