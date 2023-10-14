package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type InvoiceBillItemRepo struct {
}

func (r *InvoiceBillItemRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceBillItem) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceBillItemRepo.Create")
	defer span.End()
	now := time.Now()

	if e.CreatedAt.Time.IsZero() {
		err := e.CreatedAt.Set(now)
		if err != nil {
			return fmt.Errorf("multierr.Combine CreatedAt.Set: %w", err)
		}
	}

	err := e.UpdatedAt.Set(now)
	if err != nil {
		return fmt.Errorf("err UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"invoice_bill_item_id", "resource_path"}, db.Exec)

	if err != nil {
		return fmt.Errorf("err insert InvoiceBillItemRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert InvoiceBillItemRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceBillItemRepo) FindAllByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.InvoiceBillItems, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceBillItemRepo.FindAllByInvoiceID")
	defer span.End()

	e := &entities.InvoiceBillItem{}
	fields, _ := e.FieldMap()

	var invoiceBillItems entities.InvoiceBillItems

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, &invoiceID).ScanAll(&invoiceBillItems)

	if err != nil {
		return nil, err
	}

	return &invoiceBillItems, nil
}
