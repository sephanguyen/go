package exam_lo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
)

func (s *Suite) userActionBulkSubmission(ctx context.Context, action string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var vAction sspb.ApproveGradingAction
	switch action {
	case "approve":
		vAction = sspb.ApproveGradingAction_APPROVE_ACTION_APPROVED
	case "reject":
		vAction = sspb.ApproveGradingAction_APPROVE_ACTION_REJECTED
	}

	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).BulkApproveRejectSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.BulkApproveRejectSubmissionRequest{
		ApproveGradingAction: vAction,
		SubmissionIds:        []string{stepState.SubmissionID},
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsStatusCorrectly(ctx context.Context, statusChange string, resultChange string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var status pgtype.Text
	var result pgtype.Text
	stmt := `SELECT status, result FROM exam_lo_submission WHERE submission_id = $1::TEXT`
	if err := database.Select(ctx, s.EurekaDB, stmt, stepState.SubmissionID).ScanFields(&status, &result); err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}

	if status.String != statusChange {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to status of submisson is %s, got %s", statusChange, status.String)
	}
	if result.String != resultChange {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to result of submisson is %s, got %s", resultChange, result.String)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
