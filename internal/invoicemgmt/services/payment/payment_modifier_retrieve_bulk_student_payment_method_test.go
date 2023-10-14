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

func TestPaymentModifierService_RetrieveBulkStudentPaymentMethod(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockStudentPaymentDetailRepo := new(mock_repositories.MockStudentPaymentDetailRepo)

	s := &PaymentModifierService{
		DB:                       mockDB,
		StudentPaymentDetailRepo: mockStudentPaymentDetailRepo,
	}

	studentOne := &entities.Student{
		StudentID: database.Text("test-student-single"),
	}

	studentTwo := &entities.Student{
		StudentID: database.Text("test-student-single2"),
	}

	studentThree := &entities.Student{
		StudentID: database.Text("test-student-single3"),
	}

	singleStudentIDRequest := []string{studentOne.StudentID.String}

	testError := errors.New("test error")

	studentWithStudentPaymentMethodDD := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              studentOne.StudentID,
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}

	studentWithStudentPaymentMethodCC := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              studentOne.StudentID,
		PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
	}

	studentPaymentDetailEmptyPaymentMethod := &entities.StudentPaymentDetail{
		StudentPaymentDetailID: database.Text("123"),
		StudentID:              studentOne.StudentID,
		PaymentMethod:          database.Text(""),
	}

	multiStudentIDRequest := []string{studentOne.StudentID.String, studentTwo.StudentID.String, studentThree.StudentID.String}

	testcases := []TestCase{
		{
			name: "Happy case retrieve single student payment detail payment method direct debit",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: singleStudentIDRequest,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
				Successful: true,
				StudentPaymentMethods: []*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
					{
						StudentId:     studentOne.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodDD, nil)
			},
		},
		{
			name: "Happy case retrieve single student payment detail payment method convenience store",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: singleStudentIDRequest,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
				Successful: true,
				StudentPaymentMethods: []*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
					{
						StudentId:     studentOne.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodCC, nil)
			},
		},
		{
			name: "Happy case single student with no default payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: singleStudentIDRequest,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
				Successful: true,
				StudentPaymentMethods: []*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
					{
						StudentId:     studentOne.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "Happy case single student with empty default payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: singleStudentIDRequest,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
				Successful: true,
				StudentPaymentMethods: []*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
					{
						StudentId:     studentOne.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentPaymentDetailEmptyPaymentMethod, nil)
			},
		},
		{
			name: "Happy Case multiple student with empty payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: multiStudentIDRequest,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
				Successful: true,
				StudentPaymentMethods: []*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
					{
						StudentId:     studentOne.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
					},
					{
						StudentId:     studentTwo.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
					},
					{
						StudentId:     studentThree.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
					},
				},
			},
			setup: func(ctx context.Context) {
				studentWithStudentPaymentMethodCC.StudentID = studentTwo.StudentID
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentPaymentDetailEmptyPaymentMethod, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodCC, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "Happy Case multiple student with payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: multiStudentIDRequest,
			},
			expectedErr: nil,
			expectedResp: &invoice_pb.RetrieveBulkStudentPaymentMethodResponse{
				Successful: true,
				StudentPaymentMethods: []*invoice_pb.RetrieveBulkStudentPaymentMethodResponse_StudentPaymentMethod{
					{
						StudentId:     studentOne.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
					},
					{
						StudentId:     studentTwo.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
					},
					{
						StudentId:     studentThree.StudentID.String,
						PaymentMethod: invoice_pb.PaymentMethod_NO_DEFAULT_PAYMENT,
					},
				},
			},
			setup: func(ctx context.Context) {
				studentWithStudentPaymentMethodCC.StudentID = studentTwo.StudentID
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodDD, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodCC, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "Error has empty student ids on request",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: []string{},
			},
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("error request student ids cannot be empty")),
			setup: func(ctx context.Context) {
				// do nothing
			},
		},
		{
			name: "Error multiple student with payment method",
			ctx:  ctx,
			req: &invoice_pb.RetrieveBulkStudentPaymentMethodRequest{
				StudentIds: multiStudentIDRequest,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByStudentID: test error on student id: test-student-single3")),
			setup: func(ctx context.Context) {
				studentWithStudentPaymentMethodCC.StudentID = studentTwo.StudentID
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodDD, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(studentWithStudentPaymentMethodCC, nil)
				mockStudentPaymentDetailRepo.On("FindByStudentID", ctx, mockDB, mock.Anything).Once().Return(nil, testError)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveBulkStudentPaymentMethod(testCase.ctx, testCase.req.(*invoice_pb.RetrieveBulkStudentPaymentMethodRequest))
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
