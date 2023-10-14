package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type InvoiceScheduleStudentRepo struct {
}

func (r *InvoiceScheduleStudentRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, invoiceScheduleStudents []*entities.InvoiceScheduleStudent) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleStudentRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.InvoiceScheduleStudent) {
		stmt := `INSERT INTO invoice_schedule_student (
			invoice_schedule_student_id, 
			invoice_schedule_history_id, 
			student_id,
			error_details,
			actual_error_details,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		`

		b.Queue(stmt, &u.InvoiceSchedulesStudentID, &u.InvoiceScheduleHistoryID, &u.StudentID, &u.ErrorDetails, &u.ActualErrorDetails, &u.CreatedAt)
	}

	b := &pgx.Batch{}
	now := time.Now()

	for _, u := range invoiceScheduleStudents {
		_ = u.CreatedAt.Set(now)
		_ = u.InvoiceSchedulesStudentID.Set(idutil.ULIDNow())
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(invoiceScheduleStudents); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("invoiceScheduleStudents not inserted")
		}
	}

	return nil
}
