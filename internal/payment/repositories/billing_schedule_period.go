package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type BillingSchedulePeriodRepo struct{}

func (r *BillingSchedulePeriodRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, billingPeriodID string) (entities.BillingSchedulePeriod, error) {
	billingSchedulePeriod := &entities.BillingSchedulePeriod{}
	fieldNames, fieldValues := billingSchedulePeriod.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_period_id = $1 AND is_archived = false
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingSchedulePeriod.TableName(),
	)
	row := db.QueryRow(ctx, stmt, billingPeriodID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.BillingSchedulePeriod{}, fmt.Errorf(constant.RowScanError, err)
	}
	return *billingSchedulePeriod, nil
}

func (r *BillingSchedulePeriodRepo) GetPeriodIDsByScheduleIDAndStartTimeForUpdate(ctx context.Context, db database.QueryExecer, billingScheduleID string, startTime time.Time) ([]pgtype.Text, error) {
	var billingSchedulePeriods []pgtype.Text
	billingSchedulePeriod := &entities.BillingSchedulePeriod{}
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_id = $1 AND is_archived = false AND start_date > $2
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		"billing_schedule_period_id",
		billingSchedulePeriod.TableName(),
	)
	rows, err := db.Query(ctx, stmt, billingScheduleID, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var periodID pgtype.Text
		err = rows.Scan(&periodID)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billingSchedulePeriods = append(billingSchedulePeriods, periodID)
	}
	return billingSchedulePeriods, nil
}

func (r *BillingSchedulePeriodRepo) GetPeriodIDsInRangeTimeByScheduleID(
	ctx context.Context,
	db database.QueryExecer,
	billingScheduleID string,
	startTime time.Time,
	endTime time.Time,
) ([]pgtype.Text, error) {
	var billingSchedulePeriods []pgtype.Text
	billingSchedulePeriod := &entities.BillingSchedulePeriod{}
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_id = $1
			AND is_archived = false
			AND start_date >= $2
			AND end_date <= $3
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		"billing_schedule_period_id",
		billingSchedulePeriod.TableName(),
	)
	rows, err := db.Query(ctx, stmt, billingScheduleID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var periodID pgtype.Text
		err = rows.Scan(&periodID)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billingSchedulePeriods = append(billingSchedulePeriods, periodID)
	}
	return billingSchedulePeriods, nil
}

func (r *BillingSchedulePeriodRepo) GetLatestPeriodByScheduleIDForUpdate(ctx context.Context, db database.QueryExecer, billingScheduleID string) (billingSchedulePeriod entities.BillingSchedulePeriod, err error) {
	fieldNames, fieldValues := (&billingSchedulePeriod).FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE billing_schedule_id = $1
				ORDER BY start_date DESC
				LIMIT 1
			 	FOR NO KEY UPDATE
		 `,
		strings.Join(fieldNames, ","),
		billingSchedulePeriod.TableName(),
	)

	row := db.QueryRow(ctx, stmt, billingScheduleID)
	err = row.Scan(fieldValues...)
	return
}

func (r *BillingSchedulePeriodRepo) GetAllBillingPeriodsByBillingScheduleID(ctx context.Context, db database.QueryExecer, billingScheduleID string) ([]entities.BillingSchedulePeriod, error) {
	var billingSchedulePeriods []entities.BillingSchedulePeriod
	billingSchedulePeriod := &entities.BillingSchedulePeriod{}
	fieldNames, fieldValues := billingSchedulePeriod.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_id = $1 AND is_archived = false
		ORDER BY
			start_date ASC
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingSchedulePeriod.TableName(),
	)
	rows, err := db.Query(ctx, stmt, billingScheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		billingSchedulePeriods = append(billingSchedulePeriods, *billingSchedulePeriod)
	}
	return billingSchedulePeriods, nil
}

func (r *BillingSchedulePeriodRepo) GetPeriodByScheduleIDAndEndTime(ctx context.Context, db database.QueryExecer, billingScheduleID string, endDateOfStudentProduct time.Time) (entities.BillingSchedulePeriod, error) {
	billingSchedulePeriod := &entities.BillingSchedulePeriod{}
	fieldNames, fieldValues := billingSchedulePeriod.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_id = $1
			AND start_date <= $2
			AND end_date >= $2
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingSchedulePeriod.TableName(),
	)
	row := db.QueryRow(ctx, stmt, billingScheduleID, endDateOfStudentProduct)
	err := row.Scan(fieldValues...)
	if err != nil {
		return *billingSchedulePeriod, err
	}
	return *billingSchedulePeriod, nil
}

// Create creates BillingSchedulePeriodRepo entity
func (r *BillingSchedulePeriodRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedulePeriod) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingSchedulePeriodRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.BillingSchedulePeriodID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert BillingSchedulePeriod: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert BillingSchedulePeriod: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates BillingSchedulePeriodRepo entity
func (r *BillingSchedulePeriodRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedulePeriod) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingSchedulePeriodRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "billing_schedule_period_id", []string{
		"name",
		"billing_schedule_id",
		"start_date",
		"end_date",
		"billing_date",
		"remarks",
		"is_archived",
		"updated_at",
	})
	if err != nil {
		return fmt.Errorf("err update BillingSchedulePeriod: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update BillingSchedulePeriod: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *BillingSchedulePeriodRepo) GetNextBillingSchedulePeriod(ctx context.Context, db database.QueryExecer, billingScheduleID string, endTime time.Time) (entities.BillingSchedulePeriod, error) {
	billingSchedulePeriod := entities.BillingSchedulePeriod{}
	fieldNames, fieldValues := billingSchedulePeriod.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_id = $1 AND start_date > $2 AND is_archived = false 
		ORDER BY start_date ASC
		LIMIT 1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingSchedulePeriod.TableName(),
	)
	row := db.QueryRow(ctx, stmt, billingScheduleID, endTime)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.BillingSchedulePeriod{}, err
	}
	return billingSchedulePeriod, nil
}

func (r *BillingSchedulePeriodRepo) GetLatestBillingSchedulePeriod(
	ctx context.Context,
	db database.QueryExecer,
	billingScheduleID string,
) (entities.BillingSchedulePeriod, error) {
	billingSchedulePeriod := entities.BillingSchedulePeriod{}
	fieldNames, fieldValues := billingSchedulePeriod.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_id = $1 AND is_archived = false 
		ORDER BY end_date DESC
		LIMIT 1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingSchedulePeriod.TableName(),
	)
	row := db.QueryRow(ctx, stmt, billingScheduleID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.BillingSchedulePeriod{}, err
	}
	return billingSchedulePeriod, nil
}
