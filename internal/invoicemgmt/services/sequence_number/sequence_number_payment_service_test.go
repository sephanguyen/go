package seqnumberservice

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name             string
	ctx              context.Context
	payment          *entities.Payment
	expectedPayment  *entities.Payment
	payments         []*entities.Payment
	expectedPayments []*entities.Payment
	expectedErr      error
	setup            func(ctx context.Context)
}

var testError = errors.New("test error")

func TestPaymentSequenceNumberService_SetPaymentSeqNumber(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const (
		ctxUserID = "user-id"
	)

	mockDB := new(mock_database.Ext)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)

	s := &PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	testCases := []TestCase{
		{
			name:            "happy case",
			ctx:             ctx,
			payment:         &entities.Payment{PaymentSequenceNumber: database.Int4(0)},
			expectedPayment: &entities.Payment{PaymentSequenceNumber: database.Int4(2)},
			setup: func(ctx context.Context) {
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockDB).Once().Return(int32(1), nil)
			},
		},
		{
			name:            "error case",
			ctx:             ctx,
			payment:         &entities.Payment{PaymentSequenceNumber: database.Int4(0)},
			expectedPayment: &entities.Payment{PaymentSequenceNumber: database.Int4(0)},
			setup: func(ctx context.Context) {
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockDB).Once().Return(int32(0), testError)
			},
			expectedErr: fmt.Errorf("paymentRepo.GetLatestPaymentSequenceNumber err: %v", testError),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := func() error {
				err := s.InitLatestSeqNumber(testCase.ctx, mockDB)
				if err != nil {
					return err
				}

				err = s.AssignSeqNumberToPayment(testCase.payment)
				if err != nil {
					return err
				}

				return nil
			}()

			assert.Equal(t, testCase.expectedErr, err)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedPayment.PaymentSequenceNumber.Int, testCase.payment.PaymentSequenceNumber.Int)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockPaymentRepo,
			)
		})
	}
}

func TestPaymentSequenceNumberService_SetPaymentSeqNumbers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const (
		ctxUserID = "user-id"
	)

	mockDB := new(mock_database.Ext)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)

	s := &PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	testCases := []TestCase{
		{
			name:             "happy case",
			ctx:              ctx,
			payments:         []*entities.Payment{{PaymentSequenceNumber: database.Int4(0)}, {PaymentSequenceNumber: database.Int4(0)}, {PaymentSequenceNumber: database.Int4(0)}},
			expectedPayments: []*entities.Payment{{PaymentSequenceNumber: database.Int4(2)}, {PaymentSequenceNumber: database.Int4(3)}, {PaymentSequenceNumber: database.Int4(4)}},
			setup: func(ctx context.Context) {
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockDB).Once().Return(int32(1), nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := func() error {
				err := s.InitLatestSeqNumber(testCase.ctx, mockDB)
				if err != nil {
					return err
				}

				err = s.AssignSeqNumberToPayments(testCase.payments)
				if err != nil {
					return err
				}

				return nil
			}()

			assert.Equal(t, testCase.expectedErr, err)

			if testCase.expectedErr != nil {
				for i := 0; i < len(testCase.expectedPayments); i++ {
					assert.Equal(t, testCase.expectedPayments[i].PaymentSequenceNumber.Int, testCase.payments[i].PaymentSequenceNumber.Int)
				}

			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockPaymentRepo,
			)
		})
	}
}
