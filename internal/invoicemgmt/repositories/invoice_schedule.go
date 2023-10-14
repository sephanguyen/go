package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type InvoiceScheduleRepo struct {
}

func (r *InvoiceScheduleRepo) GetByStatusAndInvoiceDate(ctx context.Context, db database.QueryExecer, status string, invoiceDate time.Time) (*entities.InvoiceSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.GetByStatusAndInvoiceDate")
	defer span.End()

	invoiceSchedule := &entities.InvoiceSchedule{}
	fields, _ := invoiceSchedule.FieldMap()

	// Format the invoice_date to YYYY-MM-DD to easily compare using a string
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = $1 AND to_char(invoice_date, 'YYYY-MM-DD') = $2 AND (is_archived = false or is_archived IS null) ORDER BY created_at DESC", strings.Join(fields, ","), invoiceSchedule.TableName())

	// Format the time.Time to date with same format YYYY-MM-DD
	timeStr := invoiceDate.Format("2006-01-02")

	err := database.Select(ctx, db, query, status, timeStr).ScanOne(invoiceSchedule)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Errorf("db.QueryRowEx").Error())
	}
	return invoiceSchedule, nil
}

func (r *InvoiceScheduleRepo) GetByStatusAndScheduledDate(ctx context.Context, db database.QueryExecer, status string, scheduledDate time.Time) (*entities.InvoiceSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.GetByStatusAndScheduledDate")
	defer span.End()

	invoiceSchedule := &entities.InvoiceSchedule{}
	fields, _ := invoiceSchedule.FieldMap()

	// Format the scheduled_date to YYYY-MM-DD to easily compare using a string
	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = $1 AND to_char(scheduled_date, 'YYYY-MM-DD') = $2 AND (is_archived = false or is_archived IS null) ORDER BY created_at DESC", strings.Join(fields, ","), invoiceSchedule.TableName())

	// Format the time.Time to date with same format YYYY-MM-DD
	timeStr := scheduledDate.Format("2006-01-02")

	err := database.Select(ctx, db, query, status, timeStr).ScanOne(invoiceSchedule)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Errorf("db.QueryRowEx").Error())
	}
	return invoiceSchedule, nil
}

func (r *InvoiceScheduleRepo) GetCurrentEarliestInvoiceSchedule(ctx context.Context, db database.QueryExecer, status string) (*entities.InvoiceSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.GetCurrentEarliestInvoiceSchedule")
	defer span.End()

	invoiceSchedule := &entities.InvoiceSchedule{}
	fields, _ := invoiceSchedule.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE status = $1 AND invoice_date > now() AND (is_archived = false or is_archived IS null) ORDER BY invoice_date ASC LIMIT 1", strings.Join(fields, ","), invoiceSchedule.TableName())

	err := database.Select(ctx, db, query, status).ScanOne(invoiceSchedule)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Errorf("db.QueryRowEx").Error())
	}
	return invoiceSchedule, nil
}

func (r *InvoiceScheduleRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceSchedule) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.Create")
	defer span.End()
	now := time.Now()

	userID := interceptors.UserIDFromContext(ctx)

	if err := multierr.Combine(
		e.UserID.Set(userID),
		e.InvoiceScheduleID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)

	if err != nil {
		return fmt.Errorf("err insert InvoiceSchedule: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert InvoiceSchedule: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Cancels schedule if it exists
func (r *InvoiceScheduleRepo) CancelScheduleIfExists(ctx context.Context, db database.QueryExecer, invoiceScheduleDate time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.CancelScheduleIfExists")
	defer span.End()

	e := &entities.InvoiceSchedule{}

	// Before comparing the dates, convert them first in UTC so that they will be compared in the same timezone
	// This will not affect the value or timezone of the saved invoice date
	query := fmt.Sprintf("UPDATE %s SET updated_at = NOW(), status = $1 WHERE (invoice_date AT TIME ZONE 'UTC')::Date = ($2 AT TIME ZONE 'UTC')::Date", e.TableName())

	_, err := db.Exec(ctx, query, invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_CANCELLED.String(), invoiceScheduleDate)

	if err != nil {
		return fmt.Errorf("err CancelScheduleIfExists InvoiceSchedule: %w", err)
	}

	return nil
}

func (r *InvoiceScheduleRepo) RetrieveInvoiceScheduleByID(ctx context.Context, db database.QueryExecer, invoiceScheduleID string) (*entities.InvoiceSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.RetrieveInvoiceByInvoiceID")
	defer span.End()

	e := &entities.InvoiceSchedule{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_schedule_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, &invoiceScheduleID).ScanOne(e)

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *InvoiceScheduleRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.InvoiceSchedule) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "invoice_schedule_id", []string{"status", "remarks", "is_archived", "updated_at"})
	if err != nil {
		return fmt.Errorf("err update InvoiceSchedule: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update InvoiceSchedule: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceScheduleRepo) FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.InvoiceSchedule, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleRepo.FindAll")
	defer span.End()

	e := &entities.InvoiceSchedule{}
	fields, _ := e.FieldMap()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE resource_path = $1", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt, resourcePath)
	if err != nil {
		return nil, err
	}

	invoiceSchedules := []*entities.InvoiceSchedule{}
	defer rows.Close()
	for rows.Next() {
		invoiceSchedule := new(entities.InvoiceSchedule)
		database.AllNullEntity(invoiceSchedule)

		_, fieldValues := invoiceSchedule.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		invoiceSchedules = append(invoiceSchedules, invoiceSchedule)
	}

	return invoiceSchedules, nil
}
