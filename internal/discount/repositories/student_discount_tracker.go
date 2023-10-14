package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"

	"go.uber.org/multierr"
)

type StudentDiscountTrackerRepo struct {
}

func (r *StudentDiscountTrackerRepo) GetByID(ctx context.Context, db database.QueryExecer, id string) (entities.StudentDiscountTracker, error) {
	studentDiscountTracker := &entities.StudentDiscountTracker{}
	studentDiscountTrackerFieldNames, studentDiscountTrackerFieldValues := studentDiscountTracker.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			discount_tracker_id = $1
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentDiscountTrackerFieldNames, ","),
		studentDiscountTracker.TableName(),
	)
	row := db.QueryRow(ctx, stmt, id)
	err := row.Scan(studentDiscountTrackerFieldValues...)
	if err != nil {
		return entities.StudentDiscountTracker{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *studentDiscountTracker, nil
}

func (r *StudentDiscountTrackerRepo) GetActiveTrackingByStudentIDs(ctx context.Context, db database.QueryExecer, studentID []string) ([]entities.StudentDiscountTracker, error) {
	studentDiscountTrackerEntity := &entities.StudentDiscountTracker{}
	studentDiscountTrackerFieldNames, _ := studentDiscountTrackerEntity.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			student_id = ANY($1)
		AND
			student_product_end_date > NOW()
		AND
			deleted_at IS NULL
		AND
			student_product_status != 'CANCELLED'
		ORDER BY
			student_product_start_date ASC,
			student_product_end_date ASC
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentDiscountTrackerFieldNames, ","),
		studentDiscountTrackerEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, studentID)
	if err != nil {
		return []entities.StudentDiscountTracker{}, fmt.Errorf("row.Scan: %w", err)
	}
	defer rows.Close()

	discountTrackerRecords := []entities.StudentDiscountTracker{}
	for rows.Next() {
		discountTracker := new(entities.StudentDiscountTracker)
		_, fieldValues := discountTracker.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		discountTrackerRecords = append(discountTrackerRecords, *discountTracker)
	}
	return discountTrackerRecords, nil
}

func (r *StudentDiscountTrackerRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentDiscountTracker) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentDiscountTrackerRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.DiscountTrackerID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
		e.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentDiscountTracker: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert StudentDiscountTracker: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentDiscountTrackerRepo) UpdateTrackingDurationByStudentProduct(ctx context.Context, db database.QueryExecer, studentProduct entities.StudentProduct) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentDiscountTrackerRepo.UpdateTrackingDurationByStudentProduct")
	defer span.End()

	e := &entities.StudentDiscountTracker{}
	stmt := fmt.Sprintf(`UPDATE public.%s
		SET
			student_product_start_date = $1,
			student_product_end_date = $2,
			student_product_status = $3
		WHERE student_product_id = $4;`,
		e.TableName())

	cmdTag, err := db.Exec(ctx, stmt, studentProduct.StartDate, studentProduct.EndDate, studentProduct.ProductStatus, studentProduct.StudentProductID)
	if err != nil {
		return fmt.Errorf("err update student discount tracker: %w", err)
	}

	if cmdTag.RowsAffected() < 1 {
		return fmt.Errorf("updating student discount tracker for student product id %v have %d RowsAffected", studentProduct.StudentProductID, cmdTag.RowsAffected())
	}

	return
}
