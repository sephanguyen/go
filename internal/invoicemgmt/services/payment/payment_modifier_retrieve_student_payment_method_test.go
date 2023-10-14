package paymentsvc

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

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_RetrieveStudentPaymentMethod(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)

	s := &PaymentModifierService{
		DB:                       mockDB,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
	}

	student := &entities.Student{
		StudentID: database.Text("123"),
	}

	studentPaymentDetail := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              student.StudentID,
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	studentPaymentDetailEmptyPaymentMethod := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              student.StudentID,
		PaymentMethod:          database.Text(""),
	}

	testError := errors.New("test error")

	testcases := []TestCase{
		{
			name: "Happy case retrieve student payment detail payment method direct debit",
			ctx:  ctx,
			req: &invoice_pb.RetrieveStudentPaymentMethodRequest{
				StudentId: "123",
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveStudentPaymentMethodResponse{
				Successful:    true,
				StudentId:     "123",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentPaymentDetail, nil)
			},
		},
		{
			name: "Happy case retrieve student payment detail payment method convenience store",
			ctx:  ctx,
			req: &invoice_pb.RetrieveStudentPaymentMethodRequest{
				StudentId: "123",
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveStudentPaymentMethodResponse{
				Successful:    true,
				StudentId:     "123",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
			},
			setup: func(ctx context.Context) {
				studentPaymentDetail.PaymentMethod = database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String())
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentPaymentDetail, nil)
			},
		},
		{
			name: "Student with no default payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveStudentPaymentMethodRequest{
				StudentId: "555",
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveStudentPaymentMethodResponse{
				Successful:    true,
				StudentId:     "555",
				PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "Error on retrieving student payment detail record",
			ctx:  ctx,
			req: &invoice_pb.RetrieveStudentPaymentMethodRequest{
				StudentId: "123",
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByStudentID: %v on student id: %v", testError, "123")),
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "Error has empty student id on request",
			ctx:  ctx,
			req: &invoice_pb.RetrieveStudentPaymentMethodRequest{
				StudentId: "",
			},
			expectedErr: status.Error(codes.FailedPrecondition, "student id cannot be empty"),
			setup: func(ctx context.Context) {
				// mock nothing
			},
		},
		{
			name: "Student with empty default payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveStudentPaymentMethodRequest{
				StudentId: "555",
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveStudentPaymentMethodResponse{
				Successful:    true,
				StudentId:     "555",
				PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentPaymentDetailEmptyPaymentMethod, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveStudentPaymentMethod(testCase.ctx, testCase.req.(*invoice_pb.RetrieveStudentPaymentMethodRequest))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t,
				mockDB,
				mockStudentPaymentDetailRepo,
			)
		})
	}

}
