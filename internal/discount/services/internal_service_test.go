package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepo "github.com/manabie-com/backend/mock/discount/repositories"
	mockServices "github.com/manabie-com/backend/mock/discount/services/domain_service"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestInternalService_ValidateProductAndPublishUpdateOrderEvent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.MockStudentProductService
		discountEventService  *mockServices.MockDiscountEventService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error on RetrieveStudentProductByID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Product has update scheduled tag",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					StudentProductLabel: pgtype.Text{String: paymentPb.StudentProductLabel_UPDATE_SCHEDULED.String(), Status: pgtype.Present},
				}, nil)
				discountEventService.On("PublishNotificationForStudentProductWithScheduleTag", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error on PublishEventForUpdateStudentProduct",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					EndDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
				}, nil)
				discountEventService.On("PublishEventForUpdateStudentProduct", ctx, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy Case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					EndDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
				}, nil)
				discountEventService.On("PublishEventForUpdateStudentProduct", ctx, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Happy Case: Product start date is after current date",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 10), Status: pgtype.Present},
					EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
				}, nil)
				discountEventService.On("PublishEventForUpdateStudentProduct", ctx, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.MockStudentProductService)
			discountEventService = new(mockServices.MockDiscountEventService)

			testCase.Setup(testCase.Ctx)
			s := &InternalService{
				Logger:                zap.NewNop(),
				DB:                    db,
				StudentProductService: studentProductService,
				DiscountEventService:  discountEventService,
			}
			err := s.ValidateProductAndPublishUpdateOrderEvent(testCase.Ctx, mock.Anything, entities.Discount{})
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductService, discountEventService)
		})
	}
}

func TestInternalService_RetrieveCurrentDiscountOfStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.MockStudentProductService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error on RetrieveDiscountOfStudentProduct",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveDiscountOfStudentProduct", ctx, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy Case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveDiscountOfStudentProduct", ctx, mock.Anything).Return(entities.Discount{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.MockStudentProductService)

			testCase.Setup(testCase.Ctx)
			s := &InternalService{
				DB:                    db,
				StudentProductService: studentProductService,
			}
			_, err := s.RetrieveCurrentDiscountOfStudentProduct(testCase.Ctx, mock.Anything)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductService)
		})
	}
}

func TestInternalService_RetrieveActiveStudentProductsOfStudentInLocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.MockStudentProductService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error on RetrieveActiveStudentProductsOfStudentInLocation",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveActiveStudentProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy Case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveActiveStudentProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.MockStudentProductService)

			testCase.Setup(testCase.Ctx)
			s := &InternalService{
				DB:                    db,
				StudentProductService: studentProductService,
			}
			_, err := s.RetrieveActiveStudentProductsOfStudentInLocation(testCase.Ctx, mock.Anything, mock.Anything)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductService)
		})
	}
}

func TestInternalService_RetrieveStudentsCandidateForDiscountUpdateOnDate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		discountTagService *mockServices.MockDiscountTagService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error on RetrieveDiscountTagsWithActivityOnDate",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy Case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveUserIDsWithActivityOnDate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountTagService = new(mockServices.MockDiscountTagService)

			testCase.Setup(testCase.Ctx)
			s := &InternalService{
				DB:                 db,
				DiscountTagService: discountTagService,
			}
			_, err := s.RetrieveStudentsCandidateForDiscountUpdateOnDate(testCase.Ctx, time.Now())
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountTagService)
		})
	}
}

func TestInternalService_RetrieveHighestDiscountOfStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		discountTagService *mockServices.MockDiscountTagService
		discountRepo       *mockRepo.MockDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error on RetrieveDiscountEligibilityOfStudentProduct",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveDiscountEligibilityOfStudentProduct", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: No discount available",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveDiscountEligibilityOfStudentProduct", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, nil)
				discountRepo.On("GetMaxDiscountByTypeAndDiscountTagIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
				discountRepo.On("GetMaxProductDiscountByProductID", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy Case: Org discount available",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveDiscountEligibilityOfStudentProduct", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, nil)
				discountRepo.On("GetMaxDiscountByTypeAndDiscountTagIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				}, nil)
			},
		},
		{
			Name:        "Happy Case: Product discount available",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveDiscountEligibilityOfStudentProduct", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, nil)
				discountRepo.On("GetMaxDiscountByTypeAndDiscountTagIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				discountRepo.On("GetMaxProductDiscountByProductID", ctx, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountTagService = new(mockServices.MockDiscountTagService)
			discountRepo = new(mockRepo.MockDiscountRepo)

			testCase.Setup(testCase.Ctx)
			s := &InternalService{
				DB:                 db,
				DiscountTagService: discountTagService,
				DiscountRepo:       discountRepo,
			}
			_, err := s.RetrieveHighestDiscountOfStudentProduct(testCase.Ctx, mock.Anything, mock.Anything, mock.Anything)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountTagService, discountRepo)
		})
	}
}
