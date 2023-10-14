package bob

import (
	"context"
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) aListOfLocationTypesInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	listLocationTypes := []struct {
		locationTypeId       string
		name                 string
		parentLocationTypeId string
		archived             bool
		expected             bool
	}{
		// satisfied
		{locationTypeId: "locationtype-id-1", name: "org test", expected: true},
		{locationTypeId: "locationtype-id-2", name: "brand test", parentLocationTypeId: "locationtype-id-1", expected: true},
		{locationTypeId: "locationtype-id-3", name: "area test", parentLocationTypeId: "locationtype-id-1", expected: true},
		{locationTypeId: "locationtype-id-4", name: "center test", parentLocationTypeId: "locationtype-id-2", expected: true},
		{locationTypeId: "locationtype-id-10", name: "center-10", parentLocationTypeId: "locationtype-id-2", expected: true},

		// unsatisfied
		{locationTypeId: "locationtype-id-5", name: "test-5", archived: true},
		{locationTypeId: "locationtype-id-6", name: "test-6", parentLocationTypeId: "locationtype-id-5"},
		{locationTypeId: "locationtype-id-7", name: "test-7", parentLocationTypeId: "locationtype-id-6"},
		{locationTypeId: "locationtype-id-8", name: "test-8", parentLocationTypeId: "locationtype-id-10", archived: true},
		{locationTypeId: "locationtype-id-9", name: "test-9", parentLocationTypeId: "locationtype-id-8"},
	}

	addedRandom := "-" + stepState.Random

	for _, lt := range listLocationTypes {
		lt.locationTypeId += addedRandom
		lt.name += addedRandom
		if lt.parentLocationTypeId != "" {
			lt.parentLocationTypeId += addedRandom
		}

		stmt := `INSERT INTO location_types (location_type_id,name,parent_location_type_id, is_archived,updated_at,created_at) VALUES($1,$2,$3,$4,now(),now()) 
				ON CONFLICT ON CONSTRAINT location_types_pkey DO NOTHING`
		_, err := db.Exec(ctx, stmt, lt.locationTypeId,
			lt.name,
			NewNullString(lt.parentLocationTypeId),
			lt.archived)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location types with `id:%s`, %v", lt.locationTypeId, err)
		}
		if lt.expected {
			stepState.LocationTypesID = append(stepState.LocationTypesID, lt.locationTypeId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveLocationTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveLocationTypesRequest{}
	stepState.Response, stepState.ResponseErr = bpb.NewMasterDataReaderServiceClient(s.Conn).RetrieveLocationTypes(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustReturnCorrectLocationTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLocationTypesResponse)
	gotLocationTypes := rsp.GetLocationTypes()
	var mapLocationTypes = map[string]string{}
	for _, lt := range gotLocationTypes {
		mapLocationTypes[lt.LocationTypeId] = lt.ParentLocationTypeId
	}
	expectedLocationTypeIDs := stepState.LocationTypesID
	// check expected location types
	for _, ltID := range expectedLocationTypeIDs {
		if _, found := mapLocationTypes[ltID]; !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected location type id:`%s` is not found", ltID)
		}
	}
	// check location types have parent that not found in list response
	for ltID, parentID := range mapLocationTypes {
		if len(parentID) == 0 {
			continue
		} else {
			_, foundParent := mapLocationTypes[parentID]
			if !foundParent {
				return StepStateToContext(ctx, stepState), fmt.Errorf("location type id:`%s` haven't parent", ltID)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
