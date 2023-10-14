package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
)

func (s *suite) ARandomNumberInRange(ctx context.Context, rangeInt int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Random = strconv.Itoa(rand.Intn(rangeInt) + 1)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfLocationsVariantTypesInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	null := pgtype.Timestamptz{
		Status: pgtype.Null,
	}
	listLocation := []struct {
		locationID       string
		name             string
		locationType     string
		parentLocationID string
		isArchived       bool
		deletedAt        pgtype.Timestamptz
		expected         bool
	}{
		{locationID: "location-id-1", locationType: "locationtype-id-1", deletedAt: null},
		{locationID: "location-id-2", parentLocationID: "location-id-1", locationType: "locationtype-id-2", deletedAt: null},
		{locationID: "location-id-3", parentLocationID: "location-id-1", locationType: "locationtype-id-2", deletedAt: null},
		{locationID: "location-id-4", parentLocationID: "location-id-1", locationType: "locationtype-id-2", deletedAt: null},

		{locationID: "location-id-0", locationType: "locationtype-id-1", deletedAt: null},
		{locationID: "location-id-5", parentLocationID: "location-id-0", locationType: "locationtype-id-3", deletedAt: null},
		{locationID: "location-id-6", parentLocationID: "location-id-0", locationType: "locationtype-id-3", deletedAt: null},
		{locationID: "location-id-7", name: "Hue area test", parentLocationID: "location-id-0", locationType: "locationtype-id-3", expected: true, deletedAt: null},

		{locationID: "location-id-8", name: "Hue center test 1", parentLocationID: "location-id-4", locationType: "locationtype-id-4", deletedAt: null, expected: true},
		{locationID: "location-id-9", name: "Hue center test 2", parentLocationID: "location-id-4", locationType: "locationtype-id-4", deletedAt: null, expected: true},
		{locationID: "location-id-10", name: "HN center test", parentLocationID: "location-id-4", locationType: "locationtype-id-4", deletedAt: null},
		{locationID: "location-id-11", name: "HCM center test", parentLocationID: "location-id-4", locationType: "locationtype-id-4", deletedAt: null},
		{locationID: "location-id-12", name: "Hue center test 3", parentLocationID: "location-id-4", locationType: "locationtype-id-4", deletedAt: database.Timestamptz(time.Now())},
		{locationID: "location-id-13", name: "Hue center test 4", parentLocationID: "location-id-4", locationType: "locationtype-id-4", isArchived: true, deletedAt: null},
	}

	addedRandom := "-" + stepState.Random
	for _, l := range listLocation {
		l.locationID += addedRandom
		l.locationType += addedRandom

		if l.parentLocationID != "" {
			l.parentLocationID += addedRandom
		}
		if l.name != "" {
			l.name = stepState.Random + "-" + l.name
		}

		stmt := `INSERT INTO locations (location_id,name,parent_location_id, location_type,deleted_at,is_archived) VALUES($1,$2,$3,$4,$5,$6) 
				ON CONFLICT ON CONSTRAINT locations_pkey DO NOTHING`
		_, err := db.Exec(ctx, stmt, l.locationID,
			l.name,
			NewNullString(l.parentLocationID),
			l.locationType,
			l.deletedAt,
			l.isArchived,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveLowestLevelLocations(ctx context.Context, locationIDsStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationIDs := []string{}
	if len(locationIDsStr) > 0 {
		locationIDs = strings.Split(locationIDsStr, ",")
		for i := 0; i < len(locationIDs); i++ {
			locationIDs[i] = locationIDs[i] + "-" + stepState.Random
		}
	}

	req := &bpb.RetrieveLowestLevelLocationsRequest{
		Name:        stepState.Random + "-" + "Hue",
		Limit:       5,
		LocationIds: locationIDs,
	}
	stepState.Response, stepState.ResponseErr = bpb.NewMasterDataReaderServiceClient(s.Conn).RetrieveLowestLevelLocations(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustReturnLowestLevelLocations(ctx context.Context, locationIDsStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLowestLevelLocationsResponse)
	gotLocations := rsp.GetLocations()
	total := len(gotLocations)
	if len(locationIDsStr) > 0 {
		filteredLocationIDs := strings.Split(locationIDsStr, ",")
		for i := 0; i < len(filteredLocationIDs); i++ {
			filteredLocationIDs[i] = filteredLocationIDs[i] + "-" + stepState.Random
		}
		for _, l := range gotLocations {
			if !slices.Contains(filteredLocationIDs, l.LocationId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected location (id:%s) not found ", l.LocationId)
			}
			if total != len(filteredLocationIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("total number of locations is not equal, expected : %d ,got : %d", len(filteredLocationIDs), total)
			}
		}
	} else {
		locationIDs := make([]string, total)
		for _, l := range gotLocations {
			locationIDs = append(locationIDs, l.LocationId)
		}
		expectedLocations := stepState.LocationIDs
		if len(expectedLocations) != total {
			return StepStateToContext(ctx, stepState), fmt.Errorf("total number of locations is not equal, expected : %d ,got : %d", len(expectedLocations), total)
		}
		for _, l := range expectedLocations {
			if !golibs.InArrayString(l, locationIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected location (id:%s) not found ", l)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
