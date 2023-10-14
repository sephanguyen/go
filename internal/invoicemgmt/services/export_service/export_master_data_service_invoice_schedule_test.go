package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestInvoiceModifierService_ExportInvoiceSchedule(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockInvoiceScheduleRepo := new(mock_repositories.MockInvoiceScheduleRepo)

	s := &ExportMasterDataService{
		DB:                  mockDB,
		InvoiceScheduleRepo: mockInvoiceScheduleRepo,
	}

	mockInvoiceSchedule := []*entities.InvoiceSchedule{
		{
			InvoiceScheduleID: database.Text("test-id-1"),
			InvoiceDate:       database.Timestamptz(time.Now().AddDate(0, 0, 1)),
			Status:            database.Text("test-status"),
			IsArchived:        database.Bool(false),
			Remarks:           database.Text("test-remarks-1"),
			UserID:            database.Text("test-user-id"),
			CreatedAt:         database.Timestamptz(time.Now()),
			UpdatedAt:         database.Timestamptz(time.Now()),
			ResourcePath:      database.Text("test-resource-path"),
		},
		{
			InvoiceScheduleID: database.Text("test-id-2"),
			InvoiceDate:       database.Timestamptz(time.Now().AddDate(0, 0, 2)),
			Status:            database.Text("test-status"),
			IsArchived:        database.Bool(false),
			Remarks:           database.Text("test-remarks-1"),
			UserID:            database.Text("test-user-id"),
			CreatedAt:         database.Timestamptz(time.Now()),
			UpdatedAt:         database.Timestamptz(time.Now()),
			ResourcePath:      database.Text("test-resource-path"),
		},
		{
			InvoiceScheduleID: database.Text("test-id-3"),
			InvoiceDate:       database.Timestamptz(time.Now().AddDate(0, 0, 3)),
			Status:            database.Text("test-status"),
			IsArchived:        database.Bool(true),
			Remarks:           database.Text("test-remarks-1"),
			UserID:            database.Text("test-user-id"),
			CreatedAt:         database.Timestamptz(time.Now()),
			UpdatedAt:         database.Timestamptz(time.Now()),
			ResourcePath:      database.Text("test-resource-path"),
		},
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name:        "Happy Case with invoice schedule records",
			ctx:         ctx,
			req:         &invoice_pb.ExportInvoiceScheduleRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceScheduleRepo.On("FindAll", ctx, mockDB).Once().Return(mockInvoiceSchedule, nil)
			},
		},
		{
			name:        "Happy Case with no invoice schedule records",
			ctx:         ctx,
			req:         &invoice_pb.ExportInvoiceScheduleRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockInvoiceScheduleRepo.On("FindAll", ctx, mockDB).Once().Return([]*entities.InvoiceSchedule{}, nil)
			},
		},
		{
			name:        "Error on InvoiceScheduleRepo.FindAll",
			ctx:         ctx,
			req:         &invoice_pb.ExportInvoiceScheduleRequest{},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.InvoiceScheduleRepo.FindAll err: %v", testError)),
			setup: func(ctx context.Context) {
				mockInvoiceScheduleRepo.On("FindAll", ctx, mockDB).Once().Return(nil, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ExportInvoiceSchedule(testCase.ctx, testCase.req.(*invoice_pb.ExportInvoiceScheduleRequest))

			if testCase.expectedErr == nil {

				if response == nil {
					t.Errorf("The response should not be nil")
				}

				if response.Data == nil {
					t.Errorf("The response data should not be nil")
				}

				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceScheduleRepo)

		})
	}

}
