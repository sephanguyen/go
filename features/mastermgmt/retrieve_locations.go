package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) retrieveLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rand.Seed(time.Now().UnixNano())
	req := &mpb.RetrieveLocationsRequest{
		IsArchived: rand.Intn(2) == 1,
	}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataReaderServiceClient(s.MasterMgmtConn).RetrieveLocations(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) mustReturnCorrectLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*mpb.RetrieveLocationsResponse)
	gotLocations := rsp.GetLocations()
	var mapLocation = map[string]string{}
	for _, l := range gotLocations {
		mapLocation[l.LocationId] = l.ParentLocationId
	}
	// check locations are expected
	expectedLocations := stepState.CenterIDs
	for _, locationID := range expectedLocations {
		if _, found := mapLocation[locationID]; !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected location id (%s) is not found", locationID)
		}
	}
	// check locations have parent that not found in list response
	for lID, parentID := range mapLocation {
		if len(parentID) == 0 {
			continue
		}
		_, foundParent := mapLocation[parentID]
		if !foundParent {
			return StepStateToContext(ctx, stepState), fmt.Errorf("location id (%s) is returned not correct", lID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
