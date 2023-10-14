package lessonmgmt

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) RetrieveLowestLevelLocations(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveLowestLevelLocationsRequest{
		Name:   "",
		Limit:  100,
		Offset: 0,
	}
	res, err := bpb.NewMasterDataReaderServiceClient(s.CommonSuite.BobConn).
		RetrieveLowestLevelLocations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.ResponseErr = err
	stepState.Response = res
	if err == nil {
		for _, loc := range res.Locations {
			stepState.LowestLevelLocationIDs = append(stepState.LowestLevelLocationIDs, loc.LocationId)
		}
	}
	stepState.Request = req
	return nil
}

func (s *Suite) RetrieveLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveLocationsRequest{}
	stepState.Response, stepState.ResponseErr = bpb.NewMasterDataReaderServiceClient(s.BobConn).
		RetrieveLocations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}
