package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

func (s *Suite) createDateInfoToCalendarDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	dateInfo := []struct {
		date       string
		locationID string
	}{
		{
			date:       "2022-09-01",
			locationID: "lo-1",
		},
		{
			date:       "2022-09-02",
			locationID: "lo-1",
		},
		{
			date:       "2022-09-03",
			locationID: "lo-1",
		},
	}
	b := &pgx.Batch{}
	for _, l := range dateInfo {
		query := `INSERT INTO day_info (date,location_id) VALUES($1,$2) 
		ON CONFLICT DO NOTHING`
		b.Queue(query, l.date, l.locationID)
	}
	batchResults := s.CalendarDBTrace.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("batchResults.Exec():%w", err)
		}
	}
	for _, d := range dateInfo {
		stepState.DateInfoIDs = append(stepState.DateInfoIDs, d.date)
	}
	stepState.LocationID = dateInfo[0].locationID
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) dateInfoSynced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for sync process done
	time.Sleep(5 * time.Second)
	query := "SELECT count(*) from day_info where date = ANY($1) and location_id = $2"
	var count int
	dateInfoIds := stepState.DateInfoIDs
	err := s.BobDBTrace.QueryRow(ctx, query, stepState.DateInfoIDs, stepState.LocationID).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(dateInfoIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not equal date info expected %d,got %d", len(dateInfoIds), count)
	}
	return StepStateToContext(ctx, stepState), nil
}
