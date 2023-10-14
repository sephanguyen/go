package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

func (s *DataMigrationModifierService) InsertInvoiceBillItemDataMigration(ctx context.Context) error {
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Fetch invoices that have reference1 and migrated_at
		invoices, err := s.InvoiceRepo.RetrievedMigratedInvoices(ctx, tx)
		if err != nil {
			return fmt.Errorf("InvoiceRepo.RetrievedMigratedInvoices err: %v", err)
		}

		if len(invoices) == 0 {
			s.logger.Info("there are no migrated invoices")
			return nil
		}

		for _, invoice := range invoices {
			err := s.migrateInvoiceBillItems(ctx, tx, invoice)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *DataMigrationModifierService) getMappedBillSeqNumbers(ctx context.Context, db database.QueryExecer, invoice *entities.Invoice) (map[int32]struct{}, error) {
	invoiceBillItems, err := s.InvoiceBillItemRepo.FindAllByInvoiceID(ctx, db, invoice.InvoiceID.String)
	if err != nil {
		return nil, fmt.Errorf("error InvoiceBillItemRepo.FindAllByInvoiceID err: %v reference_id: %v", err, invoice.InvoiceReferenceID2.String)
	}

	seqNumbers := make(map[int32]struct{})
	for _, e := range *invoiceBillItems {
		seqNumbers[e.BillItemSequenceNumber.Int] = struct{}{}
	}

	return seqNumbers, nil
}

func (s *DataMigrationModifierService) migrateInvoiceBillItems(ctx context.Context, db database.QueryExecer, invoice *entities.Invoice) error {
	// Fetch bill_item list to get bill item details.
	billItemList, err := s.BillItemRepo.RetrieveBillItemsByInvoiceReferenceNum(ctx, db, invoice.InvoiceReferenceID2.String)
	if err != nil {
		return fmt.Errorf("error BillItemRepo.RetrieveBillItemsByInvoiceReferenceNum err: %v reference_id: %v", err, invoice.InvoiceReferenceID2.String)
	}

	// Get the bill item sequence number that are already mapped with invoice
	seqNumbers, err := s.getMappedBillSeqNumbers(ctx, db, invoice)
	if err != nil {
		return err
	}

	for _, billItem := range billItemList {
		// Filter out the bill items that are already mapped to an invoice to prevent duplicate data
		// This can prevent generating dirty data when this script run multiple times
		if _, ok := seqNumbers[billItem.BillItemSequenceNumber.Int]; ok {
			s.logger.Infof("skipping bill item %v since it is already mapped to invoice %v", billItem.BillItemSequenceNumber.Int, invoice.InvoiceID.String)
			continue
		}

		invoiceBillItem, err := createInvoiceBillItemEntity(invoice, billItem)
		if err != nil {
			return err
		}

		err = s.InvoiceBillItemRepo.Create(ctx, db, invoiceBillItem)
		if err != nil {
			return fmt.Errorf("error InvoiceBillItemRepo.Create err: %v reference_id: %v", err, invoice.InvoiceReferenceID2.String)
		}
	}

	return nil
}

func createInvoiceBillItemEntity(invoice *entities.Invoice, billItem *entities.BillItem) (*entities.InvoiceBillItem, error) {
	e := &entities.InvoiceBillItem{}
	database.AllNullEntity(e)

	if err := multierr.Combine(
		e.InvoiceID.Set(invoice.InvoiceID.String),
		e.BillItemSequenceNumber.Set(billItem.BillItemSequenceNumber.Int),
		e.PastBillingStatus.Set(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
		e.CreatedAt.Set(invoice.CreatedAt.Time),
		e.MigratedAt.Set(time.Now().UTC()),
	); err != nil {
		return nil, err
	}

	return e, nil
}
