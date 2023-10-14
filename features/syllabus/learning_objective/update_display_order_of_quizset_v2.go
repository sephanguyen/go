package learning_objective

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) updateDisplayOrder(ctx context.Context, first int, second int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	quizSetRepo := repositories.QuizSetRepo{}

	quizSet, err := quizSetRepo.GetQuizSetByLoID(ctx, s.EurekaDBTrace, database.Text(stepState.LearningObjectiveID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	questionHierarchy := entities.QuestionHierarchy{}
	if err := questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if first > len(questionHierarchy)-1 || second > len(questionHierarchy)-1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("index should not be greater than %d", len(questionHierarchy)-1)
	}

	questionHierarchy[first], questionHierarchy[second] = questionHierarchy[second], questionHierarchy[first]

	questionHierarchyReq := make([]*sspb.QuestionHierarchy, 0)
	for _, questionHierarchyObj := range questionHierarchy {
		questionHierarchyReq = append(questionHierarchyReq, &sspb.QuestionHierarchy{
			Id:          questionHierarchyObj.ID,
			Type:        sspb.QuestionHierarchyType(sspb.QuestionHierarchyType_value[string(questionHierarchyObj.Type)]),
			ChildrenIds: questionHierarchyObj.ChildrenIDs,
		})
	}

	loID := stepState.LearningObjectiveID

	req := &sspb.UpdateDisplayOrderOfQuizSetV2Request{
		LearningMaterialId: loID,
		QuestionHierarchy:  questionHierarchyReq,
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = sspb.NewQuestionServiceClient(s.EurekaConn).UpdateDisplayOrderOfQuizSetV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateDisplayOrderInQuestionGroup(ctx context.Context, first int, second int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	quizSetRepo := repositories.QuizSetRepo{}
	quizSet, err := quizSetRepo.GetQuizSetByLoID(ctx, s.EurekaDBTrace, database.Text(stepState.LearningObjectiveID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	questionHierarchy := entities.QuestionHierarchy{}
	if err := questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	for _, questionHierarchyObj := range questionHierarchy {
		if questionHierarchyObj.ID == stepState.QuestionGroupID {
			childrenIDs := questionHierarchyObj.ChildrenIDs

			if first > len(childrenIDs)-1 || second > len(childrenIDs)-1 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("index should not be greater than %d", len(childrenIDs)-1)
			}

			childrenIDs[first], childrenIDs[second] = childrenIDs[second], childrenIDs[first]
		}
	}

	questionHierarchyReq := make([]*sspb.QuestionHierarchy, 0)
	for _, questionHierarchyObj := range questionHierarchy {
		questionHierarchyReq = append(questionHierarchyReq, &sspb.QuestionHierarchy{
			Id:          questionHierarchyObj.ID,
			Type:        sspb.QuestionHierarchyType(sspb.QuestionHierarchyType_value[string(questionHierarchyObj.Type)]),
			ChildrenIds: questionHierarchyObj.ChildrenIDs,
		})
	}
	loID := stepState.LearningObjectiveID

	req := &sspb.UpdateDisplayOrderOfQuizSetV2Request{
		LearningMaterialId: loID,
		QuestionHierarchy:  questionHierarchyReq,
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = sspb.NewQuestionServiceClient(s.EurekaConn).UpdateDisplayOrderOfQuizSetV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newDisplayOrderIsUpdated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), nil
	}

	req := stepState.Request.(*sspb.UpdateDisplayOrderOfQuizSetV2Request)
	reqQuestionHierarchy := req.QuestionHierarchy

	quizSetRepo := repositories.QuizSetRepo{}
	quizSet, err := quizSetRepo.GetQuizSetByLoID(ctx, s.EurekaDBTrace, database.Text(stepState.LearningObjectiveID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	dbQuestionHierarchy := entities.QuestionHierarchy{}
	if err := dbQuestionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	dbQuizExtIDs := database.FromTextArray(quizSet.QuizExternalIDs)

	// check len
	if len(reqQuestionHierarchy) != len(dbQuestionHierarchy) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("length mismatch, expected %d but got %d", len(reqQuestionHierarchy), len(dbQuestionHierarchy))
	}

	// check question hierarchy content
	reqQuizExtIDs := []string{}
	for idx := range reqQuestionHierarchy {
		reqQuestionHierarchyObj := reqQuestionHierarchy[idx]
		dbQuestionHierarchyObj := dbQuestionHierarchy[idx]

		if reqQuestionHierarchyObj.Id != dbQuestionHierarchyObj.ID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("id mismatch, expected %s but got %s", reqQuestionHierarchyObj.Id, dbQuestionHierarchyObj.ID)
		}

		if reqQuestionHierarchyObj.Type.String() != string(dbQuestionHierarchyObj.Type) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("type mismatch, expected %s but got %s", reqQuestionHierarchyObj.Type.String(), string(dbQuestionHierarchyObj.Type))
		}

		if reqQuestionHierarchyObj.Type == sspb.QuestionHierarchyType_QUESTION {
			reqQuizExtIDs = append(reqQuizExtIDs, reqQuestionHierarchyObj.Id)
		}

		if !reflect.DeepEqual(reqQuestionHierarchyObj.ChildrenIds, dbQuestionHierarchyObj.ChildrenIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("children ids is not equal, expected %s but got %s", strings.Join(reqQuestionHierarchyObj.ChildrenIds, ","), strings.Join(dbQuestionHierarchyObj.ChildrenIDs, ","))
		}
		reqQuizExtIDs = append(reqQuizExtIDs, reqQuestionHierarchyObj.ChildrenIds...)
	}

	// check ext ids
	if !reflect.DeepEqual(reqQuizExtIDs, dbQuizExtIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("quiz ext ids is not equal, expected %s but got %s", strings.Join(reqQuizExtIDs, ","), strings.Join(dbQuizExtIDs, ","))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
