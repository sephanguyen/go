package calendar

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"

	"github.com/jackc/pgx/v4"
)

func (s *suite) checkSyncLocationData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)
	query := "SELECT location_id from locations"
	rows, err := s.CalendarDBTrace.Query(ctx, query)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	var locationTypeIDs []string
	for rows.Next() {
		var locationTypeID string
		if err = rows.Scan(&locationTypeID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		locationTypeIDs = append(locationTypeIDs, locationTypeID)
	}
	expectedLocations := stepState.LocationIDs
	for _, l := range expectedLocations {
		if !golibs.InArrayString(l, locationTypeIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found location=%s ", l)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createAListOfLocationsInBobDB(ctx context.Context) (context.Context, error) {
	return s.someExistingLocations(ctx)
}

func (s *suite) checkSyncLocationTypeData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)
	query := "SELECT location_type_id from location_types"
	rows, err := s.CalendarDBTrace.Query(ctx, query)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	var locationTypeIds []string
	for rows.Next() {
		var locationTypeID string
		if err = rows.Scan(&locationTypeID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		locationTypeIds = append(locationTypeIds, locationTypeID)
	}
	expectedLocationTypes := stepState.LocationTypeIDs
	for _, l := range expectedLocationTypes {
		if !golibs.InArrayString(l, locationTypeIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found location type=%s ", l)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createAListOfLocationTypesInBobDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationTypes := []struct {
		locationTypeID       string
		name                 string
		parentLocationTypeID string
	}{
		{locationTypeID: "lt1", name: "name-1"},
		{locationTypeID: "lt2", name: "name-2", parentLocationTypeID: "lt1"},
		{locationTypeID: "lt3", name: "name-3", parentLocationTypeID: "lt2"},
		{locationTypeID: "lt7", name: "name-7"},
		{locationTypeID: "lt4", name: "name-4"},
		{locationTypeID: "lt5", name: "name-5", parentLocationTypeID: "lt4"},
		{locationTypeID: "lt6", name: "name-6", parentLocationTypeID: "lt5"},
		{locationTypeID: "lt8", name: "name-8", parentLocationTypeID: "lt7"},
	}
	b := &pgx.Batch{}
	for _, l := range locationTypes {
		query := `INSERT INTO location_types (location_type_id,name,parent_location_type_id,updated_at,created_at) VALUES($1,$2,$3,now(),now()) 
				ON CONFLICT ON CONSTRAINT location_types_pkey DO UPDATE SET updated_at = now()`
		b.Queue(query, l.locationTypeID, l.name, NullString(l.parentLocationTypeID))
	}
	batchResults := s.BobDBTrace.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("batchResults.Exec():%w", err)
		}
	}
	stepState.LocationTypeIDs = append(stepState.LocationTypeIDs, func() (res []string) {
		for _, l := range locationTypes {
			res = append(res, l.locationTypeID)
		}
		return res
	}()...)
	return StepStateToContext(ctx, stepState), nil
}

func NullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
