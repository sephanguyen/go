package mastermgmt

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) adminExecuteCustomScript(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &mpb.ExecuteCustomEntityRequest{
		Sql: "CREATE TABLE IF NOT EXISTS table_test(UserID int)",
	}
	stepState.Response, stepState.ResponseErr = mpb.NewCustomEntityServiceClient(s.Connections.MasterMgmtConn).
		ExecuteCustomEntity(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}
