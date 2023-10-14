package assignment

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
)

func (s *Suite) userUpdateAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.InsertAssignmentResponse)

	updateAssignmentReq := &sspb.UpdateAssignmentRequest{
		Assignment: &sspb.AssignmentBase{
			Base: &sspb.LearningMaterialBase{
				LearningMaterialId: resp.LearningMaterialId,
				Name:               "assignment-name updated",
			},
			Attachments:            []string{"attachment-1-updated", "attachment-2-updated"},
			Instruction:            "instruction-updated",
			MaxGrade:               12,
			IsRequiredGrade:        false,
			AllowResubmission:      true,
			RequireAttachment:      false,
			AllowLateSubmission:    true,
			RequireAssignmentNote:  false,
			RequireVideoSubmission: true,
		},
	}
	updateAssignmentResp, err := sspb.NewAssignmentClient(s.EurekaConn).UpdateAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), updateAssignmentReq)

	stepState.Request = updateAssignmentReq
	stepState.Response = updateAssignmentResp
	stepState.ResponseErr = err

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) assignmentMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := stepState.Request.(*sspb.UpdateAssignmentRequest)

	query := `SELECT name, instruction FROM assignment WHERE learning_material_id = $1`
	var name, instruction pgtype.Text
	if err := s.EurekaDB.QueryRow(ctx, query, &req.Assignment.Base.LearningMaterialId).Scan(&name, &instruction); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if name.String != req.Assignment.Base.Name || instruction.String != req.Assignment.Instruction {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("the general assignment was not updated")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
