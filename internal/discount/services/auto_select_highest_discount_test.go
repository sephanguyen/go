package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockServices "github.com/manabie-com/backend/mock/discount/services/domain_service"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInternalService_AutoSelectHighestDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		discountTagService    *mockServices.MockDiscountTagService
		studentProductService *mockServices.MockStudentProductService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail Case: Error on discountTagService.RetrieveUserIDsWithActivityOnDate",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Req:         &pb.AutoSelectHighestDiscountRequest{},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail Case: Error on studentProductService.RetrieveActiveStudentProductsOfStudentInLocation",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Req:         &pb.AutoSelectHighestDiscountRequest{},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{
					mock.Anything,
				}, nil)
				studentProductService.On("RetrieveActiveStudentProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy Case: No tag update on date",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Req:         &pb.AutoSelectHighestDiscountRequest{},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
			},
		},
		{
			Name:        "Happy Case: With candidate students",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Req:         &pb.AutoSelectHighestDiscountRequest{},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{
					mock.Anything,
				}, nil)
				studentProductService.On("RetrieveActiveStudentProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountTagService = new(mockServices.MockDiscountTagService)
			studentProductService = new(mockServices.MockStudentProductService)

			testCase.Setup(testCase.Ctx)
			s := &InternalService{
				DB:                    db,
				DiscountTagService:    discountTagService,
				StudentProductService: studentProductService,
			}

			req := testCase.Req.(*pb.AutoSelectHighestDiscountRequest)
			_, err := s.AutoSelectHighestDiscount(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountTagService, studentProductService)
		})
	}
}
