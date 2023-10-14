package learning_objective

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userCreatesACourseAndAddStudentsIntoTheCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateCourse: %w", err)
	}
	stepState.CourseID = courseID

	studentIDs, err := utils.InsertMultiUserIntoBob(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.BobDB, 1)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("InsertMultiUserIntoBob: %w", err)
	}
	stepState.StudentIDs = studentIDs

	courseStudents, err := utils.AValidCourseWithIDs(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}
	stepState.CourseStudents = courseStudents

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userAddsAMasterStudyPlanWithTheCreatedBook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateStudyPlan: %w", err)
	}
	stepState.StudyPlanID = studyPlanID

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereIsExamLOExistedInTopic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp, err := sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("exam-lo-name+%s", stepState.TopicIDs[0]),
			},
			MaximumAttempt: wrapperspb.Int32(10),
			ApproveGrading: false,
			GradeCapping:   false,
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("InsertExamLO: %w", err)
	}

	stepState.LearningMaterialID = resp.LearningMaterialId

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) listLOProgression(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewLearningObjectiveClient(s.EurekaConn).RetrieveLOProgression(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.RetrieveLOProgressionRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LearningMaterialID,
			StudentId:          wrapperspb.String(stepState.StudentIDs[0]),
		},
		Paging: &cpb.Paging{
			Limit: 100,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreAnswers(ctx context.Context, expect int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	answerNum := 0
	res := stepState.Response.(*sspb.RetrieveLOProgressionResponse)
	if res.SessionId != stepState.SessionID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect value session_id expected: %s but got: %s", stepState.SessionID, res.SessionId)
	}

	for _, item := range res.Items {
		if item.QuizAnswer != nil {
			answerNum++
		}
	}

	if expect != answerNum {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number answer, expect: %d but got: %d", expect, answerNum)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createQuizzes(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	quizLO := utils.GenerateQuizLOProtobufMessage(3, stepState.LearningMaterialID)
	stepState.QuizLOList = append(stepState.QuizLOList, quizLO...)

	if _, err := utils.UpsertQuizzes(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, quizLO); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsertQuizzes: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateQuizTestV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.SessionID = idutil.ULIDNow()

	res, err := sspb.NewQuizClient(s.EurekaConn).CreateQuizTestV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.CreateQuizTestV2Request{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LearningMaterialID,
			StudentId:          wrapperspb.String(stepState.StudentIDs[0]),
		},
		SessionId: stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		KeepOrder: true,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("create quiz test v2: %w", err)
	}

	stepState.ShuffledQuizSetID = res.GetShuffleQuizSetId()

	return utils.StepStateToContext(ctx, stepState), nil
}
