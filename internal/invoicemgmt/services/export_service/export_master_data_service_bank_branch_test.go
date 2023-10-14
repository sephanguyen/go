package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	export_entities "github.com/manabie-com/backend/internal/invoicemgmt/export_entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_ExportBankBranch(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockBankBranchRepo := new(mock_repositories.MockBankBranchRepo)

	s := &ExportMasterDataService{
		DB:             mockDB,
		BankBranchRepo: mockBankBranchRepo,
	}

	mockBankBranchRecords := []*export_entities.BankBranchExport{
		{
			BankBranchID:           "test-id-1",
			BankBranchCode:         "123",
			BankBranchName:         "name1",
			BankBranchPhoneticName: "phonetic-name-1",
			BankCode:               "990",
			IsArchived:             false,
		},
		{
			BankBranchID:           "test-id-2",
			BankBranchCode:         "456",
			BankBranchName:         "name2",
			BankBranchPhoneticName: "phonetic-name-2",
			BankCode:               "991",
			IsArchived:             false,
		},
		{
			BankBranchID:           "test-id-3",
			BankBranchCode:         "789",
			BankBranchName:         "name3",
			BankBranchPhoneticName: "phonetic-name-3",
			BankCode:               "992",
			IsArchived:             false,
		},
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name:        "Happy Case export with bank branch records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankBranchRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindExportableBankBranches", ctx, mockDB).Once().Return(mockBankBranchRecords, nil)
			},
		},
		{
			name:        "Happy Case export with no bank branch records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankBranchRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindExportableBankBranches", ctx, mockDB).Once().Return([]*export_entities.BankBranchExport{}, nil)
			},
		},
		{
			name:        "Happy Case export with archive bank branch records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankBranchRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankBranchRecords[0].IsArchived = true
				mockBankBranchRepo.On("FindExportableBankBranches", ctx, mockDB).Once().Return(mockBankBranchRecords, nil)
			},
		},
		{
			name:        "Error on BankBranchRepo.FindExportableBankBranches",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankBranchRequest{},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BankBranchRepo.FindExportableBankBranches err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBankBranchRepo.On("FindExportableBankBranches", ctx, mockDB).Once().Return(nil, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ExportBankBranch(testCase.ctx, testCase.req.(*invoice_pb.ExportBankBranchRequest))

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

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockBankBranchRepo,
			)

		})
	}

}
