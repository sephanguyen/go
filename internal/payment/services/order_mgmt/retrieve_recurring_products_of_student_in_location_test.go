package ordermgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetrieveRecurringProductsOfStudentInLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductServiceForRetrieveRecurringProduct
		orderService          *mockServices.IOrderServiceForRetrieveRecurringProduct
	)

	studentID := uuid.New().String()
	locationID := uuid.New().String()

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Invalid orderType",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_NEW,
			},
			ExpectedErr: fmt.Errorf("invalid orderType: %s", pb.OrderType_ORDER_TYPE_NEW),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Failed case: error when get student products from location",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_WITHDRAWAL,
			},
			ExpectedErr: fmt.Errorf("error when getting student products of student %s in location %s: %v", studentID, locationID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductService.On("GetActiveRecurringProductsOfStudentInLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: error when get student products from location for LOA",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_LOA,
			},
			ExpectedErr: fmt.Errorf("error when getting student products of student %s in location %s for LOA: %v", studentID, locationID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductService.On("GetRecurringProductsOfStudentInLocationForLOA", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: error when get student product ids from location for Resume",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_RESUME,
			},
			ExpectedErr: fmt.Errorf("error when getting student product ID of student %s in location %s for resume: %v", studentID, locationID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderService.On("GetStudentProductIDsForResume", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: error when get student product ids from location for Resume",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_RESUME,
			},
			ExpectedErr: fmt.Errorf("error when getting student products of student %s in location %s for resume: %v", studentID, locationID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderService.On("GetStudentProductIDsForResume", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{"student_product_id"}, nil)
				studentProductService.On("GetStudentProductsByStudentProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_GRADUATE,
			},
			Setup: func(ctx context.Context) {
				studentProductService.On("GetActiveRecurringProductsOfStudentInLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name: "Happy case For resume",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveRecurringProductsOfStudentInLocationRequest{
				StudentId:  studentID,
				LocationId: locationID,
				OrderType:  pb.OrderType_ORDER_TYPE_RESUME,
			},
			Setup: func(ctx context.Context) {
				orderService.On("GetStudentProductIDsForResume", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentProductService = new(mockServices.IStudentProductServiceForRetrieveRecurringProduct)
			orderService = new(mockServices.IOrderServiceForRetrieveRecurringProduct)

			testCase.Setup(testCase.Ctx)
			s := &RetrieveRecurringProductsOfStudentInLocation{
				DB:                    db,
				StudentProductService: studentProductService,
				OrderService:          orderService,
			}
			req := testCase.Req.(*pb.RetrieveRecurringProductsOfStudentInLocationRequest)
			_, err := s.RetrieveRecurringProductsOfStudentInLocation(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
