package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_services "github.com/manabie-com/backend/mock/invoicemgmt/services"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_InvoiceScheduleChecker(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockOrganizationRepo := new(mock_repositories.MockOrganizationRepo)
	mockInvoiceScheduleRepo := new(mock_repositories.MockInvoiceScheduleRepo)
	mockInvoiceScheduleHistoryRepo := new(mock_repositories.MockInvoiceScheduleHistoryRepo)
	mockInvoiceScheduleStudentRepo := new(mock_repositories.MockInvoiceScheduleStudentRepo)
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	// Generate Invoice Mocks
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceBillItemRepo := new(mock_repositories.MockInvoiceBillItemRepo)
	mockOrderServiceClient := new(mock_services.OrderService)

	mockOrganizations := []*entities.Organization{
		{
			OrganizationID: pgtype.Text{String: "1"},
			Name:           pgtype.Text{String: "Organization 1"},
		},
		{
			OrganizationID: pgtype.Text{String: "2"},
			Name:           pgtype.Text{String: "Organization 2"},
		},
		{
			OrganizationID: pgtype.Text{String: "3"},
			Name:           pgtype.Text{String: "Organization 3"},
		},
	}

	now := time.Now().UTC()
	testErr := errors.New("test error")

	price := &big.Int{}
	price.SetInt64(100)

	mockScheduledInvoiceToday := []*entities.InvoiceSchedule{}
	mockBilledBillItems := [][]*entities.BillItem{}
	billItemsWithReviewRequiredTag := [][]*entities.BillItem{}
	billItemsWithCreatedDateAfterCutoffDate := [][]*entities.BillItem{}
	mockHistoryIDs := []string{}
	contexts := []interface{}{}
	for i := range mockOrganizations {
		mockScheduledInvoiceToday = append(mockScheduledInvoiceToday, &entities.InvoiceSchedule{
			InvoiceScheduleID: pgtype.Text{String: fmt.Sprintf("test-invoice-schedule-id-%d", i)},
			InvoiceDate:       pgtype.Timestamptz{Time: now},
			Status:            pgtype.Text{String: "SCHEDULED"},
			UserID:            database.Text("school admin"),
		})

		mockBilledBillItems = append(mockBilledBillItems, []*entities.BillItem{
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 1},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 2},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 3},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 4},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 5},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 6},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
				AdjustmentPrice:        database.Numeric(10),
				IsReviewed:             pgtype.Bool{Bool: true},
			},
		})

		billItemsWithReviewRequiredTag = append(billItemsWithReviewRequiredTag, []*entities.BillItem{
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 1},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 2},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 3},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 4},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 5},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 6},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
				AdjustmentPrice:        database.Numeric(10),
				IsReviewed:             pgtype.Bool{Bool: false},
			},
		})

		billItemsWithCreatedDateAfterCutoffDate = append(billItemsWithCreatedDateAfterCutoffDate, []*entities.BillItem{
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 1},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 2},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 3},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 4},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
				CreatedAt:              database.Timestamptz(time.Now().Add(72 * time.Hour)),
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 5},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
				CreatedAt:              database.Timestamptz(time.Now().Add(72 * time.Hour)),
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 6},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
				AdjustmentPrice:        database.Numeric(10),
				IsReviewed:             pgtype.Bool{Bool: true},
				CreatedAt:              database.Timestamptz(time.Now().Add(72 * time.Hour)),
			},
		})

		mockHistoryIDs = append(mockHistoryIDs, fmt.Sprintf("test-history-id-%d", i))
		contexts = append(contexts, mock.Anything)
	}

	successfulResp := &invoice_pb.InvoiceScheduleCheckerResponse{
		Successful: true,
	}

	// Init service
	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		InternalOrderService: mockOrderServiceClient,

		OrganizationRepo:           mockOrganizationRepo,
		InvoiceScheduleRepo:        mockInvoiceScheduleRepo,
		InvoiceScheduleHistoryRepo: mockInvoiceScheduleHistoryRepo,
		InvoiceScheduleStudentRepo: mockInvoiceScheduleStudentRepo,

		UnleashClient: mockUnleashClient,
	}

	testcases := []TestCase{
		{
			name: "happy test case - scheduled invoice generated successfully",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "happy test case - bill item with review required tag",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(billItemsWithReviewRequiredTag[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(1).Return(billItemsWithReviewRequiredTag[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(1).Return(billItemsWithReviewRequiredTag[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(1).Return(billItemsWithReviewRequiredTag[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "happy test case - bill item created date is after the cutoff date",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(billItemsWithCreatedDateAfterCutoffDate[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(1).Return(billItemsWithCreatedDateAfterCutoffDate[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(1).Return(billItemsWithCreatedDateAfterCutoffDate[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(1).Return(billItemsWithCreatedDateAfterCutoffDate[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "happy test case - no scheduled invoice today",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				}
			},
		},
		{
			name: "negative test case - validation error",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid InvoiceDate value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test case - error on fetching organizations",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("s.OrganizationRepo.GetOrganizations error: %v", testErr)),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(nil, testErr)
			},
		},
		{
			name: "negative test case - error on fetching invoice schedule",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleRepo.GetByStatusAndInvoiceDate error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(nil, testErr)
				}
			},
		},
		{
			name: "negative test case - error on fetching bill item by statuses",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.BillItemRepo.FindByStatuses error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)
					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(nil, testErr)
				}
			},
		},
		{
			name: "negative test case - error occurred while generating invoice",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(pgtype.Text{}, testErr)
					mockTx.On("Rollback", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - error on updating invoice schedule status",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleRepo.Update error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					// retry
					for j := 0; j < 10; j++ {
						mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
						mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(testErr)
						mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
						mockTx.On("Rollback", mock.Anything).Once().Return(nil)
					}

				}
			},
		},
		{
			name: "negative test case - there is an error with invoice and save student history errors",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.saveStudentHistoryError error: s.InvoiceScheduleStudentRepo.CreateMultiple error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(pgtype.Text{}, testErr)
					mockTx.On("Rollback", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					// retry
					for j := 0; j < 10; j++ {
						mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
						mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
						mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(testErr)
						mockTx.On("Rollback", mock.Anything).Once().Return(nil)
					}

				}
			},
		},
		{
			name: "negative test case - billing with ADJUSTMENT_BILLING type has no present adjustment_price and error on gen invoice",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItems := []*entities.BillItem{
						{
							BillItemSequenceNumber: pgtype.Int4{Int: 1},
							StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
							ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
							BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
							FinalPrice:             pgtype.Numeric{Int: price},
							BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
							IsReviewed:             pgtype.Bool{Bool: true},
						},
					}
					mockBillItems = append(mockBillItems, mockBilledBillItems[i]...)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBillItems, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(pgtype.Text{}, testErr)
					mockTx.On("Rollback", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - billing with ADJUSTMENT_BILLING type has no present adjustment_price",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)
					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return([]*entities.BillItem{
						{
							BillItemSequenceNumber: pgtype.Int4{Int: 1},
							StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
							ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
							BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
							FinalPrice:             pgtype.Numeric{Int: price},
							BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
							IsReviewed:             pgtype.Bool{Bool: true},
						},
					}, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - billing with ADJUSTMENT_BILLING type has no present adjustment_price",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)
					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return([]*entities.BillItem{
						{
							BillItemSequenceNumber: pgtype.Int4{Int: 1},
							StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
							ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
							BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
							FinalPrice:             pgtype.Numeric{Int: price},
							BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()),
							AdjustmentPrice:        database.Numeric(10),
							IsReviewed:             pgtype.Bool{Bool: true},
						},
					}, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - error on creating invoice schedule history",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleHistoryRepo.Create error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return("", testErr)
				}
			},
		},
		{
			name: "negative test case - error on duplicate invoice schedule history",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr([]*entities.Organization{mockOrganizations[0]}, fmt.Errorf("invoice schedule %s history already exists or another process is currently running", mockScheduledInvoiceToday[0].InvoiceScheduleID.String))),
			setup: func(ctx context.Context) {

				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return([]*entities.Organization{mockOrganizations[0]}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[0], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[0], nil)
				mockInvoiceScheduleHistoryRepo.On("Create", contexts[0], mockDB, mock.Anything).Once().Return("", errors.New("\"invoice_schedule_history_invoice_schedule_id_key\" (SQLSTATE 23505)"))
			},
		},
		{
			name: "negative test case - error on updating scheduled invoice history",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleHistoryRepo.UpdateWithFields error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndInvoiceDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					// retry
					for j := 0; j < 10; j++ {
						mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
						mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(testErr)
						mockTx.On("Rollback", mock.Anything).Once().Return(nil)
					}
				}
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.InvoiceScheduleChecker(testCase.ctx, testCase.req.(*invoice_pb.InvoiceScheduleCheckerRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockInvoiceRepo,
				mockInvoiceBillItemRepo,
				mockBillItemRepo,
				mockOrganizationRepo,
				mockInvoiceScheduleRepo,
				mockInvoiceScheduleHistoryRepo,
				mockInvoiceScheduleStudentRepo,
				mockTx,
				mockUnleashClient,
			)
		})
	}
}

func TestInvoiceModifierService_InvoiceScheduleChecker_CheckScheduledDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockOrganizationRepo := new(mock_repositories.MockOrganizationRepo)
	mockInvoiceScheduleRepo := new(mock_repositories.MockInvoiceScheduleRepo)
	mockInvoiceScheduleHistoryRepo := new(mock_repositories.MockInvoiceScheduleHistoryRepo)
	mockInvoiceScheduleStudentRepo := new(mock_repositories.MockInvoiceScheduleStudentRepo)
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	// Generate Invoice Mocks
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceBillItemRepo := new(mock_repositories.MockInvoiceBillItemRepo)
	mockOrderServiceClient := new(mock_services.OrderService)

	mockOrganizations := []*entities.Organization{
		{
			OrganizationID: pgtype.Text{String: "1"},
			Name:           pgtype.Text{String: "Organization 1"},
		},
		{
			OrganizationID: pgtype.Text{String: "2"},
			Name:           pgtype.Text{String: "Organization 2"},
		},
		{
			OrganizationID: pgtype.Text{String: "3"},
			Name:           pgtype.Text{String: "Organization 3"},
		},
	}

	now := time.Now().UTC()
	testErr := errors.New("test error")

	price := &big.Int{}
	price.SetInt64(100)

	mockScheduledInvoiceToday := []*entities.InvoiceSchedule{}
	mockBilledBillItems := [][]*entities.BillItem{}
	mockHistoryIDs := []string{}
	contexts := []interface{}{}
	for i := range mockOrganizations {
		mockScheduledInvoiceToday = append(mockScheduledInvoiceToday, &entities.InvoiceSchedule{
			InvoiceScheduleID: pgtype.Text{String: fmt.Sprintf("test-invoice-schedule-id-%d", i)},
			InvoiceDate:       pgtype.Timestamptz{Time: now},
			ScheduledDate:     pgtype.Timestamptz{Time: now.Add(24 * time.Hour)},
			Status:            pgtype.Text{String: "SCHEDULED"},
			UserID:            database.Text("school admin"),
		})

		mockBilledBillItems = append(mockBilledBillItems, []*entities.BillItem{
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 1},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 2},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 3},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 4},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 5},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 6},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
				AdjustmentPrice:        database.Numeric(10),
				IsReviewed:             pgtype.Bool{Bool: true},
			},
		})

		mockHistoryIDs = append(mockHistoryIDs, fmt.Sprintf("test-history-id-%d", i))
		contexts = append(contexts, mock.Anything)
	}

	successfulResp := &invoice_pb.InvoiceScheduleCheckerResponse{
		Successful: true,
	}

	// Init service
	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		InternalOrderService: mockOrderServiceClient,

		OrganizationRepo:           mockOrganizationRepo,
		InvoiceScheduleRepo:        mockInvoiceScheduleRepo,
		InvoiceScheduleHistoryRepo: mockInvoiceScheduleHistoryRepo,
		InvoiceScheduleStudentRepo: mockInvoiceScheduleStudentRepo,

		UnleashClient: mockUnleashClient,
	}

	testcases := []TestCase{
		{
			name: "happy test case - scheduled invoice generated successfully",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "happy test case - no scheduled invoice today",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				}
			},
		},
		{
			name: "negative test case - error on fetching invoice schedule",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleRepo.GetByStatusAndScheduledDate error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(nil, testErr)
				}
			},
		},
		{
			name: "negative test case - error on fetching bill item by statuses",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.BillItemRepo.FindByStatuses error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)
					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(nil, testErr)
				}
			},
		},
		{
			name: "negative test case - error occurred while generating invoice",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(pgtype.Text{}, testErr)
					mockTx.On("Rollback", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - error on updating invoice schedule status",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleRepo.Update error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					// retry
					for j := 0; j < 10; j++ {
						mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
						mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(testErr)
						mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
						mockTx.On("Rollback", mock.Anything).Once().Return(nil)
					}

				}
			},
		},
		{
			name: "negative test case - there is an error with invoice and save student history errors",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.saveStudentHistoryError error: s.InvoiceScheduleStudentRepo.CreateMultiple error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return(pgtype.Text{}, testErr)
					mockTx.On("Rollback", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					// retry
					for j := 0; j < 10; j++ {
						mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
						mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
						mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(testErr)
						mockTx.On("Rollback", mock.Anything).Once().Return(nil)
					}
				}
			},
		},
		{
			name: "negative test case - billing with ADJUSTMENT_BILLING type has no present adjustment_price",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)
					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return([]*entities.BillItem{
						{
							BillItemSequenceNumber: pgtype.Int4{Int: 1},
							StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
							ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
							BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
							FinalPrice:             pgtype.Numeric{Int: price},
							BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
							IsReviewed:             pgtype.Bool{Bool: true},
						},
					}, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - billing with ADJUSTMENT_BILLING type has no present adjustment_price",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)
					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return([]*entities.BillItem{
						{
							BillItemSequenceNumber: pgtype.Int4{Int: 1},
							StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
							ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
							BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
							FinalPrice:             pgtype.Numeric{Int: price},
							BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()),
							AdjustmentPrice:        database.Numeric(10),
							IsReviewed:             pgtype.Bool{Bool: true},
						},
					}, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleStudentRepo.On("CreateMultiple", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative test case - error on creating invoice schedule history",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleHistoryRepo.Create error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return("", testErr)
				}
			},
		},
		{
			name: "negative test case - error on duplicate invoice schedule history",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr([]*entities.Organization{mockOrganizations[0]}, fmt.Errorf("invoice schedule %s history already exists or another process is currently running", mockScheduledInvoiceToday[0].InvoiceScheduleID.String))),
			setup: func(ctx context.Context) {

				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return([]*entities.Organization{mockOrganizations[0]}, nil)

				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[0], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[0], nil)
				mockInvoiceScheduleHistoryRepo.On("Create", contexts[0], mockDB, mock.Anything).Once().Return("", errors.New("\"invoice_schedule_history_invoice_schedule_id_key\" (SQLSTATE 23505)"))
			},
		},
		{
			name: "negative test case - error on updating scheduled invoice history",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.InvoiceScheduleHistoryRepo.UpdateWithFields error: %v", testErr))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					// retry
					for j := 0; j < 10; j++ {
						mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
						mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(testErr)
						mockTx.On("Rollback", mock.Anything).Once().Return(nil)
					}
				}
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.InvoiceScheduleChecker(testCase.ctx, testCase.req.(*invoice_pb.InvoiceScheduleCheckerRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockInvoiceRepo,
				mockInvoiceBillItemRepo,
				mockBillItemRepo,
				mockOrganizationRepo,
				mockInvoiceScheduleRepo,
				mockInvoiceScheduleHistoryRepo,
				mockInvoiceScheduleStudentRepo,
				mockTx,
				mockUnleashClient,
			)
		})
	}
}

func TestInvoiceModifierService_InvoiceScheduleChecker_ReviewOrderDisabled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockOrganizationRepo := new(mock_repositories.MockOrganizationRepo)
	mockInvoiceScheduleRepo := new(mock_repositories.MockInvoiceScheduleRepo)
	mockInvoiceScheduleHistoryRepo := new(mock_repositories.MockInvoiceScheduleHistoryRepo)
	mockInvoiceScheduleStudentRepo := new(mock_repositories.MockInvoiceScheduleStudentRepo)
	mockBillItemRepo := new(mock_repositories.MockBillItemRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	// Generate Invoice Mocks
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceBillItemRepo := new(mock_repositories.MockInvoiceBillItemRepo)
	mockOrderServiceClient := new(mock_services.OrderService)

	mockOrganizations := []*entities.Organization{
		{
			OrganizationID: pgtype.Text{String: "1"},
			Name:           pgtype.Text{String: "Organization 1"},
		},
		{
			OrganizationID: pgtype.Text{String: "2"},
			Name:           pgtype.Text{String: "Organization 2"},
		},
		{
			OrganizationID: pgtype.Text{String: "3"},
			Name:           pgtype.Text{String: "Organization 3"},
		},
	}

	now := time.Now().UTC()

	price := &big.Int{}
	price.SetInt64(100)

	mockScheduledInvoiceToday := []*entities.InvoiceSchedule{}
	mockBilledBillItems := [][]*entities.BillItem{}
	mockHistoryIDs := []string{}
	contexts := []interface{}{}
	mockBilledBillItemsReviewRequired := [][]*entities.BillItem{}

	for i := range mockOrganizations {
		mockScheduledInvoiceToday = append(mockScheduledInvoiceToday, &entities.InvoiceSchedule{
			InvoiceScheduleID: pgtype.Text{String: fmt.Sprintf("test-invoice-schedule-id-%d", i)},
			InvoiceDate:       pgtype.Timestamptz{Time: now},
			ScheduledDate:     pgtype.Timestamptz{Time: now.Add(24 * time.Hour)},
			Status:            pgtype.Text{String: "SCHEDULED"},
			UserID:            database.Text("school admin"),
		})

		mockBilledBillItems = append(mockBilledBillItems, []*entities.BillItem{
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 1},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 2},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 3},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 4},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 5},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: true},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 6},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
				AdjustmentPrice:        database.Numeric(10),
				IsReviewed:             pgtype.Bool{Bool: true},
			},
		})

		mockBilledBillItemsReviewRequired = append(mockBilledBillItemsReviewRequired, []*entities.BillItem{
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 1},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 2},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 3},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 4},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-1-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 5},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-2-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				IsReviewed:             pgtype.Bool{Bool: false},
			},
			{
				BillItemSequenceNumber: pgtype.Int4{Int: 6},
				StudentID:              pgtype.Text{String: fmt.Sprintf("test-student-3-from-%d", i)},
				ResourcePath:           pgtype.Text{String: fmt.Sprintf("resource-path-%d", i)},
				BillStatus:             pgtype.Text{String: payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()},
				FinalPrice:             pgtype.Numeric{Int: price},
				BillType:               database.Text(payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String()),
				AdjustmentPrice:        database.Numeric(10),
				IsReviewed:             pgtype.Bool{Bool: false},
			},
		})

		mockHistoryIDs = append(mockHistoryIDs, fmt.Sprintf("test-history-id-%d", i))
		contexts = append(contexts, mock.Anything)
	}

	successfulResp := &invoice_pb.InvoiceScheduleCheckerResponse{
		Successful: true,
	}

	// Init service
	s := &InvoiceModifierService{
		DB:                   mockDB,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceBillItemRepo:  mockInvoiceBillItemRepo,
		BillItemRepo:         mockBillItemRepo,
		InternalOrderService: mockOrderServiceClient,

		OrganizationRepo:           mockOrganizationRepo,
		InvoiceScheduleRepo:        mockInvoiceScheduleRepo,
		InvoiceScheduleHistoryRepo: mockInvoiceScheduleHistoryRepo,
		InvoiceScheduleStudentRepo: mockInvoiceScheduleStudentRepo,

		UnleashClient: mockUnleashClient,
	}

	testError := errors.New("test-error")

	testcases := []TestCase{
		{
			name: "happy test case - scheduled invoice generated successfully for bill item with review required tag",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItemsReviewRequired[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItemsReviewRequired[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItemsReviewRequired[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItemsReviewRequired[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "happy test case - scheduled invoice generated successfully",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)

					// Generate Invoice related
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(true, nil)
					mockDB.On("Begin", mock.Anything).Times(3).Return(mockTx, nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][0], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][1], nil)
					mockBillItemRepo.On("FindByID", contexts[i], mockTx, mock.Anything).Times(2).Return(mockBilledBillItems[i][2], nil)
					mockInvoiceRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(3).Return((&entities.Invoice{}).InvoiceID, nil)
					mockInvoiceBillItemRepo.On("Create", contexts[i], mockTx, mock.Anything).Times(6).Return(nil)
					mockOrderServiceClient.On("UpdateBillItemStatus", mock.Anything, mock.Anything).Times(3).Return(nil, nil)
					mockTx.On("Commit", mock.Anything).Times(3).Return(nil)

					mockUnleashClient.On("IsFeatureEnabled", constant.EnableRetryFailedInvoiceSchedule, mock.Anything).Once().Return(false, nil)

					mockDB.On("Begin", mock.Anything).Once().Return(mockTx, nil)
					mockInvoiceScheduleRepo.On("Update", contexts[i], mockTx, mock.Anything).Once().Return(nil)
					mockInvoiceScheduleHistoryRepo.On("UpdateWithFields", contexts[i], mockTx, mock.Anything, mock.Anything).Once().Return(nil)
					mockTx.On("Commit", mock.Anything).Once().Return(nil)
				}
			},
		},
		{
			name: "negative case - error on IsFeatureEnabled",
			ctx:  ctx,
			req: &invoice_pb.InvoiceScheduleCheckerRequest{
				InvoiceDate: timestamppb.New(now),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, genTenantErr(mockOrganizations, fmt.Errorf("s.UnleashClient.IsFeatureEnabled err: %v", testError))),
			setup: func(ctx context.Context) {
				mockOrganizationRepo.On("GetOrganizations", ctx, mockDB).Once().Return(mockOrganizations, nil)

				for i := range mockOrganizations {
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
					mockInvoiceScheduleRepo.On("GetByStatusAndScheduledDate", contexts[i], mockDB, mock.Anything, mock.Anything).Once().Return(mockScheduledInvoiceToday[i], nil)
					mockInvoiceScheduleHistoryRepo.On("Create", contexts[i], mockDB, mock.Anything).Once().Return(mockHistoryIDs[i], nil)

					mockBillItemRepo.On("FindByStatuses", contexts[i], mockDB, mock.Anything).Once().Return(mockBilledBillItems[i], nil)
					mockUnleashClient.On("IsFeatureEnabled", constant.EnableReviewOrderChecking, mock.Anything).Once().Return(false, testError)
				}
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.InvoiceScheduleChecker(testCase.ctx, testCase.req.(*invoice_pb.InvoiceScheduleCheckerRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockInvoiceRepo,
				mockInvoiceBillItemRepo,
				mockBillItemRepo,
				mockOrganizationRepo,
				mockInvoiceScheduleRepo,
				mockInvoiceScheduleHistoryRepo,
				mockInvoiceScheduleStudentRepo,
				mockTx,
				mockUnleashClient,
			)
		})
	}
}

func genTenantErr(orgs []*entities.Organization, err error) string {
	m := make(map[string]error)

	for _, org := range orgs {
		m[org.Name.String] = err
	}

	return genTenantErrorStr(orgs, m)
}
