package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type WorkingHoursRepo struct{}

func (wh *WorkingHoursRepo) Upsert(ctx context.Context, db database.QueryExecer, workingHoursList []*domain.WorkingHours, locationIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "WorkingHoursRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for _, locationID := range locationIDs {
		for _, w := range workingHoursList {
			w.LocationID = locationID
			workingHours, err := NewWorkingHoursFromEntity(w)
			if workingHours.WorkingHoursID.String == "" {
				err = multierr.Append(err, workingHours.WorkingHoursID.Set(idutil.ULIDNow()))
			}
			if err != nil {
				return err
			}
			fields, args := workingHours.FieldMap()
			placeHolders := database.GeneratePlaceholders(len(fields))
			query := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT unique__working_hour_location_id_day 
			DO UPDATE SET updated_at = now(), opening_time = EXCLUDED.opening_time, closing_time = EXCLUDED.closing_time`,
				workingHours.TableName(),
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
			return fmt.Errorf("working hours is not inserted")
		}
	}
	return nil
}

func (wh *WorkingHoursRepo) getWorkingHoursByID(ctx context.Context, db database.Ext, id string) (*WorkingHours, error) {
	workingHours := &WorkingHours{}
	fields, values := workingHours.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM working_hour
		WHERE working_hour_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	err := db.QueryRow(ctx, query, &id).Scan(values...)
	return workingHours, err
}

func (wh *WorkingHoursRepo) GetWorkingHoursByID(ctx context.Context, db database.Ext, id string) (*domain.WorkingHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "WorkingHoursRepo.GetWorkingHoursByID")
	defer span.End()

	result, err := wh.getWorkingHoursByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return result.ToWorkingHoursDomain(), nil
}
