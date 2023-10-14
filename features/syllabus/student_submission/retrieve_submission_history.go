package student_submission

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/constants"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) studentDoTestOf(ctx context.Context, lmType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentID := stepState.StudentIDs[0]
	studentToken, err := s.AuthHelper.GenerateExchangeToken(studentID, constants.RoleStudent)
	ctx = s.AuthHelper.SignedCtx(ctx, studentToken)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.AuthHelper.GenerateExchangeToken: %w", err)
	}

	resp, err := sspb.NewQuizClient(s.EurekaConn).CreateQuizTestV2(ctx, &sspb.CreateQuizTestV2Request{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LearningMaterialID,
			StudentId:          wrapperspb.String(studentID),
		},
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		KeepOrder: true,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.NewQuizClient(s.EurekaConn).CreateQuizTestV2: %w", err)
	}

	var lmTypePb sspb.LearningMaterialType
	switch lmType {
	case "learning objective":
		lmTypePb = sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE
	case "flashcard":
		lmTypePb = sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD
	}

	stepState.ShuffleQuizSetID = resp.ShuffleQuizSetId
	for _, quiz := range resp.Quizzes {
		req := &sspb.CheckQuizCorrectnessRequest{
			ShuffledQuizSetId: stepState.ShuffleQuizSetID,
			QuizId:            quiz.Core.ExternalId,
			LmType:            lmTypePb,
		}
		switch lmType {
		case "learning objective":
			req.Answer = []*sspb.Answer{{Format: &sspb.Answer_SelectedIndex{1}}}
		case "flashcard":
			req.Answer = []*sspb.Answer{{Format: &sspb.Answer_FilledText{"abc"}}}
		}
		if _, err := sspb.NewQuizClient(s.EurekaConn).CheckQuizCorrectness(ctx, req); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("sspb.NewQuizClient(s.EurekaConn).CheckQuizCorrectness: %w", err)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userRetrieveStudentSubmissionHistory(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewStudentSubmissionServiceClient(s.EurekaConn).
		RetrieveSubmissionHistory(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.RetrieveSubmissionHistoryRequest{
			SetId: stepState.ShuffleQuizSetID,
		})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnsCorrectStudentSubmissionHistory(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("userRetrieveStudentSubmissionHistory: %w", stepState.ResponseErr)
	}
	resp := stepState.Response.(*sspb.RetrieveSubmissionHistoryResponse)
	if len(resp.Logs) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("student's submission is missing")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
