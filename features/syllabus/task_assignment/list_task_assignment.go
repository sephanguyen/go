package task_assignment

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userListTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ListTaskAssignmentReq := &sspb.ListTaskAssignmentRequest{
		LearningMaterialIds: stepState.LearningMaterialIDs,
	}
	stepState.Response, stepState.ResponseErr = sspb.NewTaskAssignmentClient(s.EurekaConn).ListTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), ListTaskAssignmentReq)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnTaskAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.ListTaskAssignmentResponse)
	if len(response.TaskAssignments) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of task_assignment, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.TaskAssignments))
	}

	for _, assignment := range response.TaskAssignments {
		if !golibs.InArrayString(assignment.Base.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected task_assignment id: %q in list %v of task_assignments ids: %q", assignment.Base.LearningMaterialId, stepState.LearningMaterialIDs, response.TaskAssignments)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
