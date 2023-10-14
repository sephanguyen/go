package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/internal_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInternalService_UpdateStudentPackage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                                *mockDb.Ext
		tx                                *mockDb.Tx
		mockUpcomingStudentPackageService *mockServices.IUpcomingStudentPackageForInternalService
		mockStudentPackageService         *mockServices.IStudentPackageForInternalService
		mockUpcomingStudentCourseService  *mockServices.IUpcomingStudentCourseForInternalService
		mockPackageService                *mockServices.IPackageForInternalService
		mockSubscriptionService           *mockServices.ISubscriptionServiceForInternalService
		mockStudentPackageOrderService    *mockServices.IStudentPackageOrderForInternalService
		//now                               = time.Now()
	)

	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when get student package for cronjob",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         &pb.UpdateStudentPackageForCronjobRequest{},
			Setup: func(ctx context.Context) {
				mockStudentPackageService.On("GetStudentPackagesForCronJob", mock.Anything, mock.Anything).Return([]entities.StudentPackages{
					{
						ID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						PackageID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						StartAt: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
						EndAt: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
						Properties: pgtype.JSONB{},
						IsActive: pgtype.Bool{
							Bool:   false,
							Status: pgtype.Present,
						},
						LocationIDs: pgtype.TextArray{
							Elements: []pgtype.Text{
								{
									String: constant.LocationID,
									Status: pgtype.Present,
								},
							},
							Status: pgtype.Present,
						},
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when upsert upsert student package data for cron job",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectErrorMessages: []*pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError{
				{
					UpcomingStudentPackageId: constant.UpcomingStudentPackageID,
					Error:                    fmt.Sprintf("Error when upserting student package data by student package order: %v", constant.ErrDefault),
					StudentPackageId:         constant.StudentPackageID,
					StudentPackageOrderId:    constant.StudentPackageOrderID,
				},
			},
			Req: &pb.UpdateStudentPackageForCronjobRequest{},
			Setup: func(ctx context.Context) {
				mockStudentPackageService.On("GetStudentPackagesForCronJob", mock.Anything, mock.Anything).Return([]entities.StudentPackages{
					{
						ID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						PackageID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						StartAt: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
						EndAt: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
						Properties: pgtype.JSONB{},
						IsActive: pgtype.Bool{
							Bool:   false,
							Status: pgtype.Present,
						},
						LocationIDs: pgtype.TextArray{
							Elements: []pgtype.Text{
								{
									String: constant.LocationID,
									Status: pgtype.Present,
								},
							},
							Status: pgtype.Present,
						},
					},
				}, nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				mockStudentPackageService.On("UpsertStudentPackageDataForCronjob", mock.Anything, mock.Anything, mock.Anything).Return(&npb.EventStudentPackage{}, &entities.StudentPackageOrder{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				mockStudentPackageOrderService.On("UpdateExecuteError", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			},
		},
		{
			Name: "Failed case: Error when publish student package for create order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectErrorMessages: []*pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError{
				{
					UpcomingStudentPackageId: constant.UpcomingStudentPackageID,
					Error:                    fmt.Sprintf("Error when upserting student package data by student package order: %v", constant.ErrDefault),
					StudentPackageId:         constant.StudentPackageID,
					StudentPackageOrderId:    constant.StudentPackageOrderID,
				},
			},
			Req: &pb.UpdateStudentPackageForCronjobRequest{},
			Setup: func(ctx context.Context) {
				mockStudentPackageService.On("GetStudentPackagesForCronJob", mock.Anything, mock.Anything).Return([]entities.StudentPackages{
					{
						ID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						PackageID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						StartAt: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
						EndAt: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
						Properties: pgtype.JSONB{},
						IsActive: pgtype.Bool{
							Bool:   false,
							Status: pgtype.Present,
						},
						LocationIDs: pgtype.TextArray{
							Elements: []pgtype.Text{
								{
									String: constant.LocationID,
									Status: pgtype.Present,
								},
							},
							Status: pgtype.Present,
						},
					},
				}, nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				mockStudentPackageService.On("UpsertStudentPackageDataForCronjob", mock.Anything, mock.Anything, mock.Anything).Return(&npb.EventStudentPackage{}, &entities.StudentPackageOrder{}, nil)
				mockSubscriptionService.On("PublishStudentPackageForCreateOrder", mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				mockStudentPackageOrderService.On("UpdateExecuteError", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			mockUpcomingStudentPackageService = new(mockServices.IUpcomingStudentPackageForInternalService)
			mockUpcomingStudentCourseService = new(mockServices.IUpcomingStudentCourseForInternalService)
			mockStudentPackageService = new(mockServices.IStudentPackageForInternalService)
			mockPackageService = new(mockServices.IPackageForInternalService)
			mockSubscriptionService = new(mockServices.ISubscriptionServiceForInternalService)
			mockStudentPackageOrderService = new(mockServices.IStudentPackageOrderForInternalService)

			testCase.Setup(testCase.Ctx)

			s := &InternalService{
				DB:                         db,
				studentPackageService:      mockStudentPackageService,
				packageService:             mockPackageService,
				subscriptionService:        mockSubscriptionService,
				studentPackageOrderService: mockStudentPackageOrderService,
			}

			resp, err := s.UpdateStudentPackage(testCase.Ctx, testCase.Req.(*pb.UpdateStudentPackageForCronjobRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Nil(t, resp)
			} else {
				if len(resp.Errors) > 0 {
					assert.Equal(t, len(resp.Errors), len(testCase.ExpectErrorMessages.([]*pb.UpdateStudentPackageForCronjobResponse_UpdateStudentPackageForCronjobError)))
				}
				assert.NotNil(t, resp)
				assert.True(t, resp.Successful)
			}

			mock.AssertExpectationsForObjects(t, mockUpcomingStudentPackageService, mockUpcomingStudentCourseService, mockStudentPackageService, mockSubscriptionService)
		})
	}

}
