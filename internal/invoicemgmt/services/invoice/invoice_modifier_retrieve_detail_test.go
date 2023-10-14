package invoicesvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_RetrieveInvoiceInfo(t *testing.T) {

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mocked objects
	mockDb := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceBillItemRepo := new(mock_repositories.MockInvoiceBillItemRepo)
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)

	s := &InvoiceModifierService{
		DB:                  mockDb,
		InvoiceRepo:         mockInvoiceRepo,
		PaymentRepo:         mockPaymentRepo,
		InvoiceBillItemRepo: mockInvoiceBillItemRepo,
		BillItemRepo:        mockBillItemRepo,
	}

	// Convert numeric to float64 without error
	pgNumericToFloat64NoErr := func(number pgtype.Numeric) float64 {
		result, _ := utils.GetFloat64ExactValueAndDecimalPlaces(number, "2")

		return result
	}

	// Entity objects
	invoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}
	invoice.SubTotal.Set(80)
	invoice.Total.Set(81042605.55)

	invoiceDraft := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceDraft.SubTotal.Set(80)
	invoiceDraft.Total.Set(81042605.55)

	payment := &entities.Payment{
		PaymentDueDate: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
	}

	invoiceBillItemsEmpty := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{})

	invoiceBillItemsSingle := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{
		{
			BillItemSequenceNumber: database.Int4(1),
		},
	})

	invoiceBillItemsMultiple := (entities.InvoiceBillItems)([]*entities.InvoiceBillItem{
		{
			BillItemSequenceNumber: database.Int4(1),
		},
		{
			BillItemSequenceNumber: database.Int4(2),
		},
	})

	billItem := &entities.BillItem{
		DiscountAmountType: database.Text("PERCENTAGE"),
		ProductDescription: database.Text("Test"),
		TaxPercentage:      database.Int4(10),
	}

	billItem.DiscountAmountValue.Set(0.00)
	billItem.DiscountAmount.Set(20.00)
	billItem.TaxAmount.Set(0.00)
	billItem.FinalPrice.Set(80.00)

	testcases := []TestCase{
		{
			name: "happy test case - invoice with single invoice bill item",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "123",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceInfoResponse{
				Successful:  true,
				DueDate:     timestamppb.New(payment.PaymentDueDate.Time),
				CreatedDate: timestamppb.New(invoice.CreatedAt.Time),
				Status:      invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoice.Status.String]),
				SubTotal:    pgNumericToFloat64NoErr(invoice.SubTotal),
				Total:       pgNumericToFloat64NoErr(invoice.Total),
				BillItems: []*invoice_pb.RetrieveInvoiceInfoBillItem{{
					BillItemId:          billItem.BillItemSequenceNumber.Int,
					Description:         billItem.ProductDescription.String,
					DiscountAmountType:  billItem.DiscountAmountType.String,
					DiscountAmountValue: pgNumericToFloat64NoErr(billItem.DiscountAmountValue),
					DiscountAmount:      pgNumericToFloat64NoErr(billItem.DiscountAmount),
					TaxPercentage:       billItem.TaxPercentage.Int,
					TaxAmount:           pgNumericToFloat64NoErr(billItem.TaxAmount),
					Amount:              pgNumericToFloat64NoErr(billItem.TaxAmount) + pgNumericToFloat64NoErr(billItem.FinalPrice),
				}},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(payment, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(&invoiceBillItemsSingle, nil)
				mockBillItemRepo.On("FindByID", ctx, mockDb, mock.Anything).Once().Return(billItem, nil)
			},
		},
		{
			name: "happy test case - invoice with multiple invoice bill items",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "123",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceInfoResponse{
				Successful:  true,
				DueDate:     timestamppb.New(payment.PaymentDueDate.Time),
				CreatedDate: timestamppb.New(invoice.CreatedAt.Time),
				Status:      invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoice.Status.String]),
				SubTotal:    pgNumericToFloat64NoErr(invoice.SubTotal),
				Total:       pgNumericToFloat64NoErr(invoice.Total),
				BillItems: []*invoice_pb.RetrieveInvoiceInfoBillItem{
					{
						BillItemId:          billItem.BillItemSequenceNumber.Int,
						Description:         billItem.ProductDescription.String,
						DiscountAmountType:  billItem.DiscountAmountType.String,
						DiscountAmountValue: pgNumericToFloat64NoErr(billItem.DiscountAmountValue),
						DiscountAmount:      pgNumericToFloat64NoErr(billItem.DiscountAmount),
						TaxPercentage:       billItem.TaxPercentage.Int,
						TaxAmount:           pgNumericToFloat64NoErr(billItem.TaxAmount),
						Amount:              pgNumericToFloat64NoErr(billItem.TaxAmount) + pgNumericToFloat64NoErr(billItem.FinalPrice),
					},
					{
						BillItemId:          billItem.BillItemSequenceNumber.Int,
						Description:         billItem.ProductDescription.String,
						DiscountAmountType:  billItem.DiscountAmountType.String,
						DiscountAmountValue: pgNumericToFloat64NoErr(billItem.DiscountAmountValue),
						DiscountAmount:      pgNumericToFloat64NoErr(billItem.DiscountAmount),
						TaxPercentage:       billItem.TaxPercentage.Int,
						TaxAmount:           pgNumericToFloat64NoErr(billItem.TaxAmount),
						Amount:              pgNumericToFloat64NoErr(billItem.TaxAmount) + pgNumericToFloat64NoErr(billItem.FinalPrice),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(payment, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(&invoiceBillItemsMultiple, nil)
				mockBillItemRepo.On("FindByID", ctx, mockDb, mock.Anything).Return(billItem, nil)
				mockBillItemRepo.On("FindByID", ctx, mockDb, mock.Anything).Return(billItem, nil)
			},
		},
		{
			name: "happy case - invoice with no invoice bill items",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "123",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceInfoResponse{
				Successful:  true,
				DueDate:     timestamppb.New(payment.PaymentDueDate.Time),
				CreatedDate: timestamppb.New(invoice.CreatedAt.Time),
				Status:      invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoice.Status.String]),
				SubTotal:    pgNumericToFloat64NoErr(invoice.SubTotal),
				Total:       pgNumericToFloat64NoErr(invoice.Total),
				BillItems:   []*invoice_pb.RetrieveInvoiceInfoBillItem{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(payment, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(&invoiceBillItemsEmpty, nil)
				mockBillItemRepo.On("FindByID", ctx, mockDb, mock.Anything).Return(billItem, nil)
			},
		},
		{
			name: "happy case - invoice with no invoice bill items and no payment record",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "123",
			},
			expectedResp: &invoice_pb.RetrieveInvoiceInfoResponse{
				Successful:  true,
				DueDate:     timestamppb.New(payment.PaymentDueDate.Time),
				CreatedDate: timestamppb.New(invoice.CreatedAt.Time),
				Status:      invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoice.Status.String]),
				SubTotal:    pgNumericToFloat64NoErr(invoice.SubTotal),
				Total:       pgNumericToFloat64NoErr(invoice.Total),
				BillItems:   []*invoice_pb.RetrieveInvoiceInfoBillItem{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(&invoiceBillItemsEmpty, nil)
				mockBillItemRepo.On("FindByID", ctx, mockDb, mock.Anything).Return(billItem, nil)
			},
		},
		{
			name: "negative test case - invoice not found",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "1",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test case - error getting invoice record",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "1",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test case - invoice in draft status, not accepted",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "1",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invoice should not be in DRAFT status"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoiceDraft, nil)
			},
		},
		{
			name: "negative test case - error getting payment record",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "1",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test case - error getting invoice bill item record",
			ctx:  ctx,
			req: &invoice_pb.RetrieveInvoiceInfoRequest{
				InvoiceIdString: "1",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(invoice, nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(payment, nil)
				mockInvoiceBillItemRepo.On("FindAllByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.RetrieveInvoiceInfo(testCase.ctx, testCase.req.(*invoice_pb.RetrieveInvoiceInfoRequest))

			if err != nil {
				fmt.Printf("Test case name: %v 	Response err: %v\n", testCase.name, err)
			}

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			if testCase.expectedResp != nil {
				assert.Equal(t, len(testCase.expectedResp.(*invoice_pb.RetrieveInvoiceInfoResponse).BillItems), len(response.BillItems))
				assert.True(t, checkEquality(testCase.expectedResp.(*invoice_pb.RetrieveInvoiceInfoResponse), response))
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockInvoiceRepo, mockPaymentRepo, mockBillItemRepo, mockInvoiceBillItemRepo)
		})
	}

}

func checkEquality(actualResp *invoice_pb.RetrieveInvoiceInfoResponse, expectedResp *invoice_pb.RetrieveInvoiceInfoResponse) bool {

	if actualResp.DueDate.String() != expectedResp.DueDate.String() {
		// if string representation fails, try the Timestamp value to check nullity
		if actualResp.DueDate != nil && expectedResp.DueDate != nil {
			fmt.Printf("DueDate: expected %v but got %v\n", expectedResp.DueDate, actualResp.DueDate)
			return false
		}
	}

	if actualResp.CreatedDate.String() != expectedResp.CreatedDate.String() {
		// if string representation fails, try the Timestamp value to check nullity
		if actualResp.CreatedDate != nil && expectedResp.CreatedDate != nil {
			fmt.Printf("CreatedDate: expected %v but got %v\n", expectedResp.CreatedDate, actualResp.CreatedDate)
			return false
		}
	}

	if actualResp.Status != expectedResp.Status {
		fmt.Printf("Status: expected %v but got %v\n", expectedResp.Status, actualResp.Status)
		return false
	}

	if actualResp.SubTotal != expectedResp.SubTotal {
		fmt.Printf("SubTotal: expected %v but got %v\n", expectedResp.SubTotal, actualResp.SubTotal)
		return false
	}

	if len(actualResp.BillItems) != len(expectedResp.BillItems) {
		fmt.Printf("BillItems: expected %v but got %v\n", len(expectedResp.BillItems), len(actualResp.BillItems))
		return false
	}

	for i := 0; i < len(actualResp.BillItems); i++ {
		actualBillItem := actualResp.BillItems[i]
		expectedBillItem := expectedResp.BillItems[i]

		if actualBillItem.Description != expectedBillItem.Description {
			fmt.Printf("Description: expected %v but got %v\n", actualBillItem.Description, expectedBillItem.Description)
			return false
		}

		if actualBillItem.DiscountAmountType != expectedBillItem.DiscountAmountType {
			fmt.Printf("DiscountAmountType: expected %v but got %v\n", actualBillItem.DiscountAmountType, expectedBillItem.DiscountAmountType)
			return false
		}

		if actualBillItem.BillItemId != expectedBillItem.BillItemId {
			fmt.Printf("BillItemId: expected %v but got %v\n", actualBillItem.BillItemId, expectedBillItem.BillItemId)
			return false
		}

		if actualBillItem.DiscountAmountValue != expectedBillItem.DiscountAmountValue {
			fmt.Printf("DiscountAmountValue: expected %v but got %v\n", actualBillItem.DiscountAmountValue, expectedBillItem.DiscountAmountValue)
			return false
		}

		if actualBillItem.DiscountAmount != expectedBillItem.DiscountAmount {
			fmt.Printf("DiscountAmount: expected %v but got %v\n", actualBillItem.DiscountAmount, expectedBillItem.DiscountAmount)
			return false
		}

		if actualBillItem.TaxPercentage != expectedBillItem.TaxPercentage {
			fmt.Printf("TaxPercentage: expected %v but got %v\n", actualBillItem.TaxPercentage, expectedBillItem.TaxPercentage)
			return false
		}

		if actualBillItem.Amount != expectedBillItem.Amount {
			fmt.Printf("Amount: expected %v but got %v\n", actualBillItem.Amount, expectedBillItem.Amount)
			return false
		}
	}

	return true
}
