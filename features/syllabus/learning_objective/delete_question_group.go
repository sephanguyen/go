package learning_objective

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) deleteExistingQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.DeleteQuestionGroupRequest{
		QuestionGroupId: stepState.QuestionGroupID,
	}
	if len(req.QuestionGroupId) == 0 {
		return nil, fmt.Errorf("question group id not found")
	}

	quizRepo := repositories.QuizRepo{}
	quizzes, err := quizRepo.GetByQuestionGroupID(ctx, s.EurekaDBTrace, database.Text(req.QuestionGroupId))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	quizzesExternalIDs := []string{}

	for _, quiz := range quizzes {
		quizzesExternalIDs = append(quizzesExternalIDs, quiz.ExternalID.String)
	}

	stepState.ExternalIDs = quizzesExternalIDs

	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	res, err := sspb.NewQuestionServiceClient(s.EurekaConn).
		DeleteQuestionGroup(ctx, req)

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = res, err
	if err == nil {
		stepState.ExistingQuestionHierarchy.ExcludeQuestionGroupIDs([]string{req.QuestionGroupId})
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) questionGroupDeleted(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), nil
	}

	req := stepState.Request.(*sspb.DeleteQuestionGroupRequest)
	questionGroupID := req.QuestionGroupId

	quizSetRepo := repositories.QuizSetRepo{}
	quizSet, err := quizSetRepo.GetQuizSetByLoID(ctx, s.EurekaDBTrace, database.Text(stepState.LearningObjectiveID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	quizRepo := repositories.QuizRepo{}
	quizzes, err := quizRepo.GetByQuestionGroupID(ctx, s.EurekaDBTrace, database.Text(questionGroupID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if len(quizzes) != 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expect quizzes length is 0 but got %d", len(quizzes))
	}

	questionHierarchy := entities.QuestionHierarchy{}
	externalIDs := []string{}

	quizSet.QuizExternalIDs.AssignTo(&externalIDs)
	quizSet.QuestionHierarchy.AssignTo(&questionHierarchy)

	for _, id := range stepState.ExternalIDs {
		if sliceutils.Contains(externalIDs, id) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("quiz %s still exist in external ids", id)
		}
	}

	for _, obj := range questionHierarchy {
		if obj.ID == questionGroupID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("question group %s still exist in question hierarchy", questionGroupID)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
