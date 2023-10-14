package mastermgmt

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) getPageInfoBySlug(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &mpb.GetPageInfoBySlugRequest{
		Slug:          "4-editing-table-data",
		ApplicationId: "636475935bf0c92b9021e512",
	}
	stepState.Response, stepState.ResponseErr = mpb.NewAppsmithServiceClient(s.Connections.MasterMgmtConn).
		GetPageInfoBySlug(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnCorrectAppsmithPage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil && stepState.ResponseErr.Error() != status.Error(codes.Internal, mongo.ErrNoDocuments.Error()).Error() {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}
