package student_submission

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userRetrieveSubmissions(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).
		ListSubmissions(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListSubmissionsRequest{
			StudyPlanItemIdentities: stepState.StudyPlanIdentities,
		})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) retrieveStudentSubmissionsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.ListSubmissionsResponse)
	for _, s := range resp.Submissions {
		studyPlanItem := s.StudyPlanItemIdentity
		if studyPlanItem.LearningMaterialId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong LearningMaterialId of %v", s.SubmissionId)
		}
		if studyPlanItem.StudyPlanId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudyPlanId of %v", s.SubmissionId)
		}
		if studyPlanItem.StudentId == nil || studyPlanItem.StudentId.Value == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudentId of %v", s.SubmissionId)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
