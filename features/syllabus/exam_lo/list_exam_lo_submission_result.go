package exam_lo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userListExamLOSubmissionResult(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.ListExamLOSubmissionResultRequest{
		StudyPlanItemIdentities: stepState.StudyPlanItemIdentities,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListExamLOSubmissionResult(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsListExamLoSubmissionResultCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.(*sspb.ListExamLOSubmissionResultResponse)
	if len(resp.Items) != len(stepState.StudyPlanItemIdentities) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("number of response items expected %v, got %v", len(stepState.StudyPlanItemIdentities), len(resp.Items))
	}

	numberOfExamLOSubmissionInfors := 0
	for _, item := range resp.Items {
		numberOfExamLOSubmissionInfors += len(item.ExamLoSubmissions.Items)

		for _, submission := range item.ExamLoSubmissions.Items {
			if submission.TotalGradedPoint == nil || submission.TotalGradedPoint.Value == 0 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("total_graded_point can't be zero")
			}
		}
	}
	if numberOfExamLOSubmissionInfors != len(stepState.ExamLOSubmissionEnts) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("number of exam lo submission infors expected %v, got %v", len(stepState.ExamLOSubmissionEnts), numberOfExamLOSubmissionInfors)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) upsertQuestionGroup(ctx context.Context, req *sspb.UpsertQuestionGroupRequest) (*sspb.UpsertQuestionGroupResponse, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(req.LearningMaterialId) == 0 {
		return nil, fmt.Errorf("lo ID dont have yet")
	}
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	return sspb.NewQuestionServiceClient(s.EurekaConn).
		UpsertQuestionGroup(ctx, req)
}

func (s *Suite) insertANewQuestionGroup(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.UpsertQuestionGroupRequest{
		LearningMaterialId: "lo-id-1",
		Name:               "name",
		Description:        "description",
		RichDescription: &cpb.RichText{
			Raw:      "raw rich text",
			Rendered: "rendered rich text",
		},
	}
	res, err := s.upsertQuestionGroup(ctx, req)
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = res, err
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.QuestionGroupID = res.QuestionGroupId
	return utils.StepStateToContext(ctx, stepState), nil
}
