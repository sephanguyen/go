package task_assignment

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
)

func (s *Suite) userUpdateValidTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	updateTaskAssignmentReq := &sspb.UpdateTaskAssignmentRequest{
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: stepState.LearningMaterialIDs[0],
				Name:               "task-assignment-name updated",
			},
			Attachments:               []string{"attachment-1-updated", "attachment-2-updated"},
			Instruction:               "instruction-updated",
			RequireDuration:           true,
			RequireCompleteDate:       true,
			RequireUnderstandingLevel: true,
			RequireCorrectness:        true,
			RequireAttachment:         false,
			RequireAssignmentNote:     false,
		},
	}
	stepState.Response, stepState.ResponseErr = sspb.NewTaskAssignmentClient(s.EurekaConn).UpdateTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), updateTaskAssignmentReq)
	stepState.Request = updateTaskAssignmentReq
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTheTaskAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := stepState.Request.(*sspb.UpdateTaskAssignmentRequest)

	query := `SELECT name, instruction FROM task_assignment WHERE learning_material_id = $1`
	var name, instruction pgtype.Text
	if err := s.EurekaDB.QueryRow(ctx, query, &req.TaskAssignment.Base.LearningMaterialId).Scan(&name, &instruction); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if name.String != req.TaskAssignment.Base.Name || instruction.String != req.TaskAssignment.Instruction {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("the task assignment was not updated")
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
