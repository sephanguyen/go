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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInvoiceModifierService_ExportBankMapping(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockBankMappingRepo := new(mock_repositories.MockBankMappingRepo)

	s := &ExportMasterDataService{
		DB:              mockDB,
		BankMappingRepo: mockBankMappingRepo,
	}

	mockBankMapping := []*entities.BankMapping{}
	for i := 0; i < 3; i++ {
		mockBankMapping = append(mockBankMapping, &entities.BankMapping{
			BankMappingID: database.Text(fmt.Sprintf("test-bank-mapping-id-%d", i)),
			BankID:        database.Text(fmt.Sprintf("test-bank-id-%d", i)),
			PartnerBankID: database.Text(fmt.Sprintf("test-partner-bank-id-%d", i)),
			Remarks:       database.Text(fmt.Sprintf("test-remarks-%d", i)),
			IsArchived:    database.Bool(false),
		})
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name:        "Happy Case with bank mapping records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankMappingRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankMappingRepo.On("FindAll", ctx, mockDB).Once().Return(mockBankMapping, nil)
			},
		},
		{
			name:        "Happy Case with no bank mapping records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankMappingRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankMappingRepo.On("FindAll", ctx, mockDB).Once().Return([]*entities.BankMapping{}, nil)
			},
		},
		{
			name:        "Error on BankMappingRepo.FindAll",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankMappingRequest{},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BankMappingRepo.FindAll err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBankMappingRepo.On("FindAll", ctx, mockDB).Once().Return(nil, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ExportBankMapping(testCase.ctx, testCase.req.(*invoice_pb.ExportBankMappingRequest))

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

			mock.AssertExpectationsForObjects(t, mockDB, mockBankMappingRepo)

		})
	}

}
