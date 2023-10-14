package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type TimeSlotRepo struct{}

func (tsr *TimeSlotRepo) Upsert(ctx context.Context, db database.QueryExecer, timeSlotList []*domain.TimeSlot, locationIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "TimeSlotRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for _, locationID := range locationIDs {
		for _, ts := range timeSlotList {
			ts.LocationID = locationID
			timeSlot, err := NewTimeSlotFromEntity(ts)
			if timeSlot.TimeSlotID.String == "" {
				err = multierr.Append(err, timeSlot.TimeSlotID.Set(idutil.ULIDNow()))
			}
			if err != nil {
				return err
			}
			fields, args := timeSlot.FieldMap()
			placeHolders := database.GeneratePlaceholders(len(fields))
			query := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT time_slot_internal_id_location_id_unique 
			DO UPDATE SET updated_at = now(), start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time`,
				timeSlot.TableName(),
				strings.Join(fields, ","),
				placeHolders)
			b.Queue(query, args...)
		}
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("time slots is not inserted")
		}
	}
	return nil
}
