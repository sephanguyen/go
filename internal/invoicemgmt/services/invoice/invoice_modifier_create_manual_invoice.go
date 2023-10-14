package invoicesvc

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type genInvoiceResult struct {
	InvoiceDetail *invoice_pb.GenerateInvoiceDetail
	InvoiceData   *invoice_pb.GenerateInvoicesResponse_InvoicesData
	Err           error
}

const (
	workerNumberReference = 1000

	// Set maxRetry to 50 since there are some instance that it exceeded 10 retries
	createManualInvoiceRetry = 10
)

type generateInvoiceParam struct {
	invoice                   *invoice_pb.GenerateInvoiceDetail
	enableReviewOrderChecking bool
}

func (s *InvoiceModifierService) GenerateInvoices(ctx context.Context, req *invoice_pb.GenerateInvoicesRequest) (*invoice_pb.GenerateInvoicesResponse, error) {
	// validate the request. This will return an InvalidArgument error when the request is invalid.
	if err := s.validateGenerateInvoiceInput(req); err != nil {
		return &invoice_pb.GenerateInvoicesResponse{
			Successful:   false,
			InvoicesData: []*invoice_pb.GenerateInvoicesResponse_InvoicesData{},
		}, err
	}

	enableReviewOrderChecking, err := s.UnleashClient.IsFeatureEnabled(constant.EnableReviewOrderChecking, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.UnleashClient.IsFeatureEnabled err: %v", err))
	}

	invoiceLength := len(req.Invoices)
	errors := make([]*invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError, 0, invoiceLength)

	workerNumber := workerNumberReference
	if invoiceLength < workerNumberReference {
		workerNumber = invoiceLength
	}

	var wg sync.WaitGroup
	wp := utils.NewWorkerPool(workerNumber)
	wp.Run() // Run the workers

	resChan := make(chan genInvoiceResult, len(req.Invoices)) // output channel
	// the output of the process will be pushed to the output channel
	for _, invoice := range req.Invoices {
		// validate each invoice
		if err := s.validateEachInvoiceDetail(invoice); err != nil {
			errors = append(errors, &invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
				InvoiceDetail: invoice,
				Error:         err.Error(),
			})
			continue
		}

		// this is important. Reassign to prevent data race on input
		input := invoice

		wg.Add(1)
		wp.AddTask(func() {
			defer wg.Done()

			res := s.generateInvoice(ctx, &generateInvoiceParam{invoice: input, enableReviewOrderChecking: enableReviewOrderChecking})

			resChan <- res
		})
	}

	go func() {
		// Close the worker pool (closes the function channel)
		log.Println("GenerateInvoice: Closing func channel")
		wp.Close()

		// Wait for the all tasks to finish
		log.Println("GenerateInvoice: Waiting for workers")
		wg.Wait()

		// Close the output error channel
		log.Println("GenerateInvoice: Closing result channel")
		close(resChan)

		log.Println("GenerateInvoice: Result channel closed")
	}()

	invoicesData := make([]*invoice_pb.GenerateInvoicesResponse_InvoicesData, 0, len(resChan))

	// fetch data from output channel
	for res := range resChan {
		if res.Err != nil {
			errors = append(errors, &invoice_pb.GenerateInvoicesResponse_GenerateInvoiceResponseError{
				InvoiceDetail: res.InvoiceDetail,
				Error:         res.Err.Error(),
			})
			continue
		}

		invoicesData = append(invoicesData, res.InvoiceData)
	}

	// Return success if no error
	return &invoice_pb.GenerateInvoicesResponse{
		Successful:   len(errors) == 0,
		InvoicesData: invoicesData,
		Errors:       errors,
	}, nil
}

func (s *InvoiceModifierService) validateGenerateInvoiceInput(req *invoice_pb.GenerateInvoicesRequest) error {
	if len(req.Invoices) == 0 {
		return status.Error(codes.InvalidArgument, "Invoices cannot be empty")
	}

	return nil
}

func (s *InvoiceModifierService) validateEachInvoiceDetail(invoiceDetail *invoice_pb.GenerateInvoiceDetail) error {
	if strings.TrimSpace(invoiceDetail.StudentId) == "" {
		return status.Error(codes.InvalidArgument, "Student ID cannot be empty")
	}

	if len((invoiceDetail.BillItemIds)) == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Bill Items of student %s cannot be empty", invoiceDetail.StudentId))
	}

	return nil
}

func (s *InvoiceModifierService) generateInvoice(ctx context.Context, param *generateInvoiceParam) genInvoiceResult {
	var data *invoice_pb.GenerateInvoicesResponse_InvoicesData
	err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
			billItemList, err := s.validateAndGetBillItem(ctx, tx, param)
			if err != nil {
				return err
			}

			invoiceID, err := s.createInvoice(ctx, tx, param.invoice, billItemList)
			if err != nil {
				return err
			}

			data = &invoice_pb.GenerateInvoicesResponse_InvoicesData{
				InvoiceId: invoiceID,
			}

			return nil
		})

		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return false, fmt.Errorf("repo.Create: %w", err)
		}

		log.Printf("Retrying creating invoice. Attempt: %d \n", attempt)
		return attempt < createManualInvoiceRetry, fmt.Errorf("cannot generate invoice data, err %v", err)
	}, createManualInvoiceRetry)

	if err != nil {
		return genInvoiceResult{
			InvoiceDetail: param.invoice,
			InvoiceData:   nil,
			Err:           err,
		}
	}

	return genInvoiceResult{
		InvoiceDetail: param.invoice,
		InvoiceData:   data,
		Err:           nil,
	}
}

func (s *InvoiceModifierService) createInvoice(ctx context.Context, tx pgx.Tx, invoice *invoice_pb.GenerateInvoiceDetail, billItems []*entities.BillItem) (string, error) {
	// Create single invoice
	invoiceID, err := s.InvoiceRepo.Create(ctx, tx, genInvoiceData(invoice))
	if err != nil {
		return "", fmt.Errorf("error Invoice Create: %v", err)
	}

	// Create invoice_bill_item and generate bill item update request
	updateBillItems, err := s.processBillItems(ctx, tx, invoiceID, billItems)
	if err != nil {
		return "", fmt.Errorf("error Process Bill Items: %v", err)
	}

	// Update the bill item status using payment service
	// If the context contains useInternalOrderService key, use the internal service
	userInfo := golibs.UserInfoFromCtx(ctx)
	changeBillItemStatusesRequest := &payment_pb.UpdateBillItemStatusRequest{
		UpdateBillItems: updateBillItems,
		OrganizationId:  userInfo.ResourcePath,
		CurrentUserId:   userInfo.UserID,
	}

	resp, err := s.InternalOrderService.UpdateBillItemStatus(ctx, changeBillItemStatusesRequest)
	if err != nil {
		return "", fmt.Errorf("error Update When Bill Items Statuses Changed: %v", err)
	}

	// Check for any validation errors from the payment service
	if resp != nil && len(resp.Errors) > 0 {
		var errorList []string

		for _, err := range resp.Errors {
			errorStr := fmt.Sprintf("BillItemSequenceNumber %v with error %v", err.BillItemSequenceNumber, err.Error)
			errorList = append(errorList, errorStr)
		}

		return "", fmt.Errorf("error UpdateBillItemStatus: %v", strings.Join(errorList, ","))
	}

	return invoiceID.String, nil
}

func (s *InvoiceModifierService) processBillItems(
	ctx context.Context,
	tx pgx.Tx,
	invoiceID pgtype.Text,
	billItems []*entities.BillItem,
) ([]*payment_pb.UpdateBillItemStatusRequest_UpdateBillItem, error) {
	updateBillItems := make([]*payment_pb.UpdateBillItemStatusRequest_UpdateBillItem, 0, len(billItems))

	// Loop billing items for that invoice
	for _, billItem := range billItems {
		// save 1 Invoice bill item if no issue
		invoiceBillItemData, err := genInvoiceBillItemData(invoiceID, billItem.BillItemSequenceNumber.Int, billItem.BillStatus.String)
		if err != nil {
			return nil, fmt.Errorf("error genInvoiceBillItemData %v: %v", billItem.BillItemSequenceNumber.Int, err)
		}

		err = s.InvoiceBillItemRepo.Create(ctx, tx, invoiceBillItemData)
		if err != nil {
			return nil, fmt.Errorf("error Invoice Billing Item Repo: %v", err)
		}

		billitem := &payment_pb.UpdateBillItemStatusRequest_UpdateBillItem{
			BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
			BillingStatusTo:        payment_pb.BillingStatus_BILLING_STATUS_INVOICED,
		}

		updateBillItems = append(updateBillItems, billitem)
	}

	return updateBillItems, nil
}

func genInvoiceData(invoice *invoice_pb.GenerateInvoiceDetail) *entities.Invoice {
	return &entities.Invoice{
		Status: pgtype.Text{
			String: invoice_pb.InvoiceStatus_DRAFT.String(),
			Status: pgtype.Present,
		},
		Type: pgtype.Text{
			String: invoice.InvoiceType.String(),
			Status: pgtype.Present,
		},
		StudentID: pgtype.Text{
			String: invoice.StudentId,
			Status: pgtype.Present,
		},
		SubTotal: pgtype.Numeric{
			Int:    big.NewInt(int64(invoice.SubTotal)),
			Status: pgtype.Present,
		},
		Total: pgtype.Numeric{
			Int:    big.NewInt(int64(invoice.Total)),
			Status: pgtype.Present,
		},
		IsExported: pgtype.Bool{
			Bool:   false,
			Status: pgtype.Present,
		},
		OutstandingBalance: pgtype.Numeric{
			Int:    big.NewInt(int64(invoice.Total)),
			Status: pgtype.Present,
		},
		AmountPaid: pgtype.Numeric{
			Int:    big.NewInt(0),
			Status: pgtype.Present,
		},
		AmountRefunded: pgtype.Numeric{
			Int:    big.NewInt(0),
			Status: pgtype.Present,
		},
		InvoiceReferenceID: pgtype.Text{
			Status: pgtype.Null,
		},
		MigratedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
		InvoiceReferenceID2: pgtype.Text{
			Status: pgtype.Null,
		},
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
	}
}

func genInvoiceBillItemData(invoiceID pgtype.Text, billingItemID int32, status string) (*entities.InvoiceBillItem, error) {
	e := new(entities.InvoiceBillItem)
	database.AllNullEntity(e)

	id := idutil.ULIDNow()
	if err := multierr.Combine(
		e.InvoiceBillItemID.Set(id),
		e.InvoiceID.Set(invoiceID.String),
		e.BillItemSequenceNumber.Set(billingItemID),
		e.PastBillingStatus.Set(status),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}

func (s *InvoiceModifierService) validateAndGetBillItem(ctx context.Context, db database.QueryExecer, param *generateInvoiceParam) ([]*entities.BillItem, error) {
	billItemList := []*entities.BillItem{}

	for _, billItemID := range param.invoice.BillItemIds {
		billItem, err := s.BillItemRepo.FindByID(ctx, db, billItemID)
		if err != nil {
			return nil, fmt.Errorf("s.BillItemRepo.FindByID err: %w", err)
		}

		// Check if the bill item has Review Required tag
		if !billItem.IsReviewed.Bool && param.enableReviewOrderChecking {
			return nil, status.Error(codes.InvalidArgument, "bill item should not contain Review Required tag")
		}

		// Check if the bill item is BILLED or PENDING
		if billItem.BillStatus.String == payment_pb.BillingStatus_BILLING_STATUS_BILLED.String() ||
			billItem.BillStatus.String == payment_pb.BillingStatus_BILLING_STATUS_PENDING.String() {
			billItemList = append(billItemList, billItem)
			continue
		}

		return nil, fmt.Errorf("bill item with ID %d has an invalid status %s", billItem.BillItemSequenceNumber, billItem.BillStatus.String)
	}

	return billItemList, nil
}
