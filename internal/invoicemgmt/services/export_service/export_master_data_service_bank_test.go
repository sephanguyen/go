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

func TestInvoiceModifierService_ExportBank(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockBankRepo := new(mock_repositories.MockBankRepo)

	s := &ExportMasterDataService{
		DB:       mockDB,
		BankRepo: mockBankRepo,
	}

	mockBank := []*entities.Bank{}
	for i := 0; i < 3; i++ {
		mockBank = append(mockBank, &entities.Bank{
			BankID:           database.Text(fmt.Sprintf("test-bank-id-%d", i)),
			BankCode:         database.Text(fmt.Sprintf("test-bank-code-%d", i)),
			BankName:         database.Text(fmt.Sprintf("test-bank-name-%d", i)),
			BankNamePhonetic: database.Text(fmt.Sprintf("test-bank-name-phonetic-%d", i)),
			IsArchived:       database.Bool(false),
		})
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name:        "Happy Case with bank records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankRepo.On("FindAll", ctx, mockDB).Once().Return(mockBank, nil)
			},
		},
		{
			name:        "Happy Case with no bank records",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBankRepo.On("FindAll", ctx, mockDB).Once().Return([]*entities.Bank{}, nil)
			},
		},
		{
			name:        "Error on BankRepo.FindAll",
			ctx:         ctx,
			req:         &invoice_pb.ExportBankRequest{},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("s.BankRepo.FindAll err: %v", testError)),
			setup: func(ctx context.Context) {
				mockBankRepo.On("FindAll", ctx, mockDB).Once().Return(nil, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ExportBank(testCase.ctx, testCase.req.(*invoice_pb.ExportBankRequest))

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
		})

		mock.AssertExpectationsForObjects(t, mockDB, mockBankRepo)
	}

}
