package assignment

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) userInsertAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	insertAssignmentReq := &sspb.InsertAssignmentRequest{
		Assignment: &sspb.AssignmentBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    "assignment-name",
			},
			Attachments:            []string{"attachment-1", "attachment-2"},
			Instruction:            "instruction",
			MaxGrade:               10,
			IsRequiredGrade:        true,
			AllowResubmission:      false,
			RequireAttachment:      true,
			AllowLateSubmission:    false,
			RequireAssignmentNote:  true,
			RequireVideoSubmission: false,
		},
	}
	insertAssignmentResp, err := sspb.NewAssignmentClient(s.EurekaConn).InsertAssignment(s.AuthHelper.SignedCtx((ctx), stepState.Token), insertAssignmentReq)

	stepState.Request = insertAssignmentReq
	stepState.Response = insertAssignmentResp
	stepState.ResponseErr = err

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) assignmentMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.InsertAssignmentResponse)

	query := `SELECT count(*) FROM assignment WHERE learning_material_id = $1`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, &resp.LearningMaterialId).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of assignment %d, got %d", 1, count)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
