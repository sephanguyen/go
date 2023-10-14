package study_plan

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) ListStudentStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	req := &sspb.ListStudentStudyPlansRequest{Paging: nil, CourseId: stepState.CourseID, StudentIds: []string{stepState.StudentIDs[0]}, BookIds: []string{stepState.BookID}, Grades: []int32{1}}
	res, err := sspb.NewStudyPlanClient(s.EurekaConn).ListStudentStudyPlan(ctx, req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.Response = res

	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) OurSysReturnCorrectStudentStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.ListStudentStudyPlansResponse)
	respSP := resp.GetStudyPlans()
	if len(respSP) != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Expect 1 sp, atual: %v, studentID: %s", len(respSP), stepState.StudentIDs[0])
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
