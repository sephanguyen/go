package bob

import (
	"context"
	"database/sql"
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func buildAccessPath(rootLocation, rand string, locationPrefixes []string) string {
	rs := rootLocation
	for _, str := range locationPrefixes {
		rs += "/" + str + rand
	}
	return rs
}

func (s *suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	addedRandom := "-" + stepState.Random
	listLocation := []struct {
		locationID       string
		name             string
		parentLocationID string
		archived         bool
		expected         bool
		accessPath       string
	}{ // satisfied
		{locationID: "1" + addedRandom, parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1"})},
		{locationID: "2" + addedRandom, parentLocationID: "1" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2"})},
		{locationID: "3" + addedRandom, parentLocationID: "2" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2", "3"})},
		{locationID: "7" + addedRandom, parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7"})},
		// unsatisfied
		{locationID: "4" + addedRandom, parentLocationID: stepState.LocationID, archived: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4"})},
		{locationID: "5" + addedRandom, parentLocationID: "4" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "6" + addedRandom, parentLocationID: "5" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "8" + addedRandom, parentLocationID: "7" + addedRandom, archived: true, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7", "8"})},
	}

	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,name,parent_location_id, is_archived, access_path) VALUES($1,$2,$3,$4,$5) 
				ON CONFLICT DO NOTHING`
		_, err := db.Exec(ctx, stmt, l.locationID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived, l.accessPath)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
			stepState.CenterIDs = append(stepState.CenterIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveLocationsRequest{}
	stepState.Response, stepState.ResponseErr = bpb.NewMasterDataReaderServiceClient(s.Conn).RetrieveLocations(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) mustReturnCorrectLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLocationsResponse)
	gotLocations := rsp.GetLocations()
	var mapLocation = map[string]string{}
	for _, l := range gotLocations {
		mapLocation[l.LocationId] = l.ParentLocationId
	}
	// check locations are expected
	expectedLocations := stepState.LocationIDs
	for _, locationID := range expectedLocations {
		if _, found := mapLocation[locationID]; !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected location id(%s) is not found", locationID)
		}
	}
	// check locations have parent that not found in list response
	for lID, parentID := range mapLocation {
		if len(parentID) == 0 {
			continue
		}
		_, foundParent := mapLocation[parentID]
		if !foundParent {
			return StepStateToContext(ctx, stepState), fmt.Errorf("location id(%s) is returned not correct", lID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
