package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/internal_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateStudentProductStatus(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID               = "user-id"
		orderId                 = "order-id-123"
		invalidOrdersParameters = "invalid orders parameters"
	)

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		mockDB                *mockDb.Ext
		mockTx                *mockDb.Tx
		mockOrderService      *mockServices.IOrderServiceForInternalService
		studentProductService *mockServices.IStudentProductServiceForInternalService
	)

	testcases := []utils.TestCase{
		{
			Name:         "fail getting student products for cancellation",
			Ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			ExpectedErr:  constant.ErrDefault,
			ExpectedResp: &pb.UpdateStudentProductStatusResponse{},
			Req: &pb.UpdateStudentProductStatusRequest{
				StudentProductLabel: []string{pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String()},
			},
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductsByStudentProductLabel", mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
				mockDB.On("Begin", mock.Anything).Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name:        "happy case withdrawal orders with err in update",
			Ctx:         interceptors.ContextWithUserID(ctx, ctxUserID),
			ExpectedErr: nil,
			ExpectedResp: &pb.UpdateStudentProductStatusResponse{
				Errors: []*pb.UpdateStudentProductStatusResponse_UpdateStudentProductStatusError{
					{
						StudentProductId: "1",
						Error:            constant.ErrDefault.Error(),
					}, {
						StudentProductId: "2",
						Error:            constant.ErrDefault.Error(),
					},
				},
			},
			Req: &pb.UpdateStudentProductStatusRequest{
				StudentProductLabel: []string{pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String()},
			},
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductsByStudentProductLabel", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, nil)
				studentProductService.On("CancelStudentProduct", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockDB.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
		{
			Name:         "happy case graduation orders",
			Ctx:          interceptors.ContextWithUserID(ctx, ctxUserID),
			ExpectedErr:  nil,
			ExpectedResp: &pb.UpdateStudentProductStatusResponse{},
			Req: &pb.UpdateStudentProductStatusRequest{
				StudentProductLabel: []string{pb.StudentProductLabel_GRADUATION_SCHEDULED.String()},
			},
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductsByStudentProductLabel", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, nil)
				studentProductService.On("CancelStudentProduct", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockDB.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			mockDB = new(mockDb.Ext)
			mockTx = new(mockDb.Tx)
			mockOrderService = new(mockServices.IOrderServiceForInternalService)
			studentProductService = new(mockServices.IStudentProductServiceForInternalService)
			s := &InternalService{
				DB:                    mockDB,
				orderService:          mockOrderService,
				studentProductService: studentProductService,
			}

			testCase.Setup(testCase.Ctx)

			resp, err := s.UpdateStudentProductStatus(testCase.Ctx, testCase.Req.(*pb.UpdateStudentProductStatusRequest))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.NotNil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}
			mock.AssertExpectationsForObjects(t, mockDB, mockTx, mockOrderService)
		})
	}
}
