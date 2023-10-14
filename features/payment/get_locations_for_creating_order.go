package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) createUserAndAssignToLocations(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	locationIDs := stepState.LocationIDs
	ctx, err := s.signedAsAccountWithLocations(ctx, userGroup, locationIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getLocationsForCreatingOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.GetLocationsForCreatingOrderRequest{}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	resp, err := client.GetLocationsForCreatingOrder(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response = resp

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkLocationsForCreatingOrder(ctx context.Context, typeOfLocation string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetLocationsForCreatingOrderResponse)
	locationsResp := resp.LocationInfos
	locationIDsResp := sliceutils.Map(resp.LocationInfos, func(t *pb.LocationInfo) string {
		return t.LocationId
	})

	switch typeOfLocation {
	case "empty":
		if len(locationsResp) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("length of locations should be empty")
		}
	case "list":
		if len(locationsResp) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("length of locations should not be empty")
		}
		expectedLocationIDs := stepState.LowestLevelLocationIDs
		for _, expectedLocationID := range expectedLocationIDs {
			if !sliceutils.Contains(locationIDsResp, expectedLocationID) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("missing expected location id")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
