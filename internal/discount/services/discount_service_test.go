package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockServices "github.com/manabie-com/backend/mock/discount/services/domain_service"
	mockNats "github.com/manabie-com/backend/mock/golibs/nats"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDiscountSubscription_SubscribeToOrderWithProductInfoLog(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		jsm *mockNats.JetStreamManagement
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&nats.Subscription{}, nil)
			},
		},
		{
			Name:        "Fail case: error parsing data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&nats.Subscription{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		jsm = new(mockNats.JetStreamManagement)

		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				JSM: jsm,
			}

			err := s.SubscribeToOrderWithProductInfoLog()
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, jsm)
		})
	}
}

func TestDiscountService_TrackValidSiblingDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		discountTrackerService *mockServices.MockDiscountTrackerService
	)

	studentProductDiscountGroups := []entities.ProductDiscountGroup{
		{
			StudentProduct: entities.StudentProduct{
				StudentID:           pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				LocationID:          pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				StudentProductID:    pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				ProductID:           pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				ProductStatus:       pgtype.Text{String: paymentPb.StudentProductStatus_ORDERED.String(), Status: pgtype.Present},
				StudentProductLabel: pgtype.Text{String: paymentPb.StudentProductLabel_CREATED.String(), Status: pgtype.Present},
				StartDate:           pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
				EndDate:             pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
			},
			ProductGroups: []entities.ProductGroup{
				{
					ProductGroupID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					DiscountType:   pgtype.Text{String: paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String(), Status: pgtype.Present},
				},
			},
			DiscountType: paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String(),
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        "Happy case: student product successfully tracked",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("TrackDiscount", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: error on tracking student product",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("TrackDiscount", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			discountTrackerService = new(mockServices.MockDiscountTrackerService)

			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				Logger:                 zap.NewNop(),
				DiscountTrackerService: discountTrackerService,
			}

			_, err := s.TrackValidSiblingDiscount(testCase.Ctx, mock.Anything, studentProductDiscountGroups)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, discountTrackerService)
		})
	}
}

func TestDiscountService_TagValidSiblingDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		discountTrackerService *mockServices.MockDiscountTrackerService
		discountTagService     *mockServices.MockDiscountTagService
	)

	studentID1 := "1"
	studentID2 := "2"

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: error on retrieve discount tracking history by student IDs",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("RetrieveSiblingDiscountTrackingHistoriesByStudentIDs", ctx, mock.Anything, mock.Anything).Return(
					map[string][]entities.StudentDiscountTracker{},
					map[string][]entities.StudentDiscountTracker{},
					constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error on retrieve discount tag ID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("RetrieveSiblingDiscountTrackingHistoriesByStudentIDs", ctx, mock.Anything, mock.Anything).Return(
					map[string][]entities.StudentDiscountTracker{
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 2), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
							},
						},
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID2, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
							},
						},
					},
					map[string][]entities.StudentDiscountTracker{
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 2), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
							},
						},
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID2, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
							},
						},
					},
					nil)
				discountTagService.On("RetrieveDiscountTagIDsByDiscountType", ctx, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error on update discount tags",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("RetrieveSiblingDiscountTrackingHistoriesByStudentIDs", ctx, mock.Anything, mock.Anything).Return(
					map[string][]entities.StudentDiscountTracker{
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 2), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
							},
						},
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID2, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
							},
						},
					},
					map[string][]entities.StudentDiscountTracker{
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 2), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
							},
						},
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID2, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
							},
						},
					},
					nil)
				discountTagService.On("RetrieveDiscountTagIDsByDiscountType", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				discountTagService.On("UpdateDiscountTagOfStudentIDWithTimeSegment", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error on soft delete old tags with 0 tracking data",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("RetrieveSiblingDiscountTrackingHistoriesByStudentIDs", ctx, mock.Anything, mock.Anything).Return(
					map[string][]entities.StudentDiscountTracker{
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
						},
						studentID2: {},
					},
					map[string][]entities.StudentDiscountTracker{
						studentID1: {},
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
						},
					},
					nil)
				discountTagService.On("RetrieveDiscountTagIDsByDiscountType", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				discountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("RetrieveSiblingDiscountTrackingHistoriesByStudentIDs", ctx, mock.Anything, mock.Anything).Return(
					map[string][]entities.StudentDiscountTracker{
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 2), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
							},
						},
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID2, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
							},
						},
					},
					map[string][]entities.StudentDiscountTracker{
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, 2), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
							},
						},
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID2, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 10), Status: pgtype.Present},
							},
						},
					},
					nil)
				discountTagService.On("RetrieveDiscountTagIDsByDiscountType", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				discountTagService.On("UpdateDiscountTagOfStudentIDWithTimeSegment", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Happy case: soft delete old tags with 0 tracking data",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountTrackerService.On("RetrieveSiblingDiscountTrackingHistoriesByStudentIDs", ctx, mock.Anything, mock.Anything).Return(
					map[string][]entities.StudentDiscountTracker{
						studentID1: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
						},
						studentID2: {},
					},
					map[string][]entities.StudentDiscountTracker{
						studentID1: {},
						studentID2: {
							{
								StudentID:               pgtype.Text{String: studentID1, Status: pgtype.Present},
								StudentProductStartDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
								StudentProductEndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
							},
						},
					},
					nil)
				discountTagService.On("RetrieveDiscountTagIDsByDiscountType", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				discountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			discountTrackerService = new(mockServices.MockDiscountTrackerService)
			discountTagService = new(mockServices.MockDiscountTagService)

			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				Logger:                 zap.NewNop(),
				DiscountTagService:     discountTagService,
				DiscountTrackerService: discountTrackerService,
			}

			err := s.TagValidSiblingDiscount(testCase.Ctx, []string{studentID1, studentID2})
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, discountTagService, discountTrackerService)
		})
	}
}

func TestDiscountService_UpdateProductDiscountTracking(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		studentProductService  *mockServices.MockStudentProductService
		discountTrackerService *mockServices.MockDiscountTrackerService
	)

	studentProductDiscountGroups := []entities.ProductDiscountGroup{
		{
			StudentProduct: entities.StudentProduct{
				StartDate:                   pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
				EndDate:                     pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
				StudentProductID:            pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				UpdatedFromStudentProductID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: error retrieving student product by ID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error updating student product tracking",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				discountTrackerService.On("UpdateTrackingDurationByStudentProduct", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				discountTrackerService.On("UpdateTrackingDurationByStudentProduct", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			discountTrackerService = new(mockServices.MockDiscountTrackerService)
			studentProductService = new(mockServices.MockStudentProductService)

			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				Logger:                 zap.NewNop(),
				DiscountTrackerService: discountTrackerService,
				StudentProductService:  studentProductService,
			}

			_, err := s.UpdateProductDiscountTrackingData(testCase.Ctx, studentProductDiscountGroups)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, discountTrackerService, studentProductService)
		})
	}
}

func TestDiscountService_UpdateTrackingDataForVoidOrders(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		studentProductService  *mockServices.MockStudentProductService
		discountTrackerService *mockServices.MockDiscountTrackerService
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: error retrieving student product by order ID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductsByOrderID", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error updating tracking duration from student product",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductsByOrderID", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						StudentProductID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					},
				}, nil)
				discountTrackerService.On("UpdateTrackingDurationByStudentProduct", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("RetrieveStudentProductsByOrderID", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						StudentProductID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					},
				}, nil)
				discountTrackerService.On("UpdateTrackingDurationByStudentProduct", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			discountTrackerService = new(mockServices.MockDiscountTrackerService)
			studentProductService = new(mockServices.MockStudentProductService)

			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				Logger:                 zap.NewNop(),
				DiscountTrackerService: discountTrackerService,
				StudentProductService:  studentProductService,
			}

			_, err := s.UpdateTrackingDataForVoidOrders(testCase.Ctx, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, discountTrackerService, studentProductService)
		})
	}
}
