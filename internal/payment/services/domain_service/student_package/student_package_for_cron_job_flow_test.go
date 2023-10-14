package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	mockServices "github.com/manabie-com/backend/mock/payment/services/domain_service"

	"github.com/jackc/pgtype"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentPackage_GetStudentPackagesForCronJob(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                     *mockDb.Ext
		mockStudentPackageRepo *mockRepositories.MockStudentPackageRepo
		mockStudentCourseRepo  *mockRepositories.MockStudentCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when get student packages for cronjob",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, fmt.Sprintf("error when getting student packages for cron job: %s", constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				mockStudentPackageRepo.On("GetStudentPackagesForCronjobByDay", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentPackages{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				mockStudentPackageRepo.On("GetStudentPackagesForCronjobByDay", mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentPackages{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockStudentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			mockStudentCourseRepo = new(mockRepositories.MockStudentCourseRepo)

			testCase.Setup(testCase.Ctx)

			studentPackageService := &StudentPackageService{
				StudentPackageRepo: mockStudentPackageRepo,
				StudentCourseRepo:  mockStudentCourseRepo,
			}

			_, err := studentPackageService.GetStudentPackagesForCronJob(testCase.Ctx, db)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockStudentPackageRepo, mockStudentCourseRepo)
		})
	}
}

func TestStudentPackage_UpsertStudentPackageDataForCronjob(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                                 *mockDb.Ext
		mockStudentPackageRepo             *mockRepositories.MockStudentPackageRepo
		mockStudentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		mockStudentPackageOrderService     *mockServices.StudentPackageOrderService
		mockStudentCourseRepo              *mockRepositories.MockStudentCourseRepo
		mockPackageRepo                    *mockRepositories.MockPackageRepo
		mockPackageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		mockStudentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		now                                = time.Now().UTC()
	)
	packageProperties := entities.PackageProperties{
		AllCourseInfo: []entities.CourseInfo{
			{
				CourseID:      constant.CourseID,
				Name:          constant.CourseName,
				NumberOfSlots: 1,
				Weight:        6,
			},
		},
		CanWatchVideo:     []string{constant.CourseID},
		CanViewStudyGuide: []string{constant.CourseID},
		CanDoQuiz:         []string{constant.CourseID},
		LimitOnlineLesson: 0,
		AskTutor: &entities.AskTutorCfg{
			TotalQuestionLimit: 0,
			LimitDuration:      "",
		},
	}
	packagePropertiesJson, _ := json.Marshal(packageProperties)
	studentPackageObject := entities.StudentPackages{
		ID: pgtype.Text{
			String: constant.StudentPackageID,
			Status: pgtype.Present,
		},
		StudentID: pgtype.Text{
			String: constant.StudentID,
			Status: pgtype.Present,
		},
		PackageID: pgtype.Text{
			String: constant.PackageID,
			Status: pgtype.Present,
		},
		StartAt: pgtype.Timestamptz{
			Time:   now.AddDate(0, 1, 0),
			Status: pgtype.Present,
		},
		EndAt: pgtype.Timestamptz{
			Time:   now.AddDate(0, 4, 0),
			Status: pgtype.Present,
		},
		Properties: pgtype.JSONB{
			Bytes:  packagePropertiesJson,
			Status: pgtype.Present,
		},
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
		CreatedAt: pgtype.Timestamptz{
			Time:             now,
			Status:           pgtype.Present,
			InfinityModifier: 0,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:   now,
			Status: pgtype.Present,
		},
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
	}
	studentPackageObjectJSON, _ := json.Marshal(studentPackageObject)

	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when set current student package order",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, fmt.Sprintf("error when set current student package order with student_package_id=%s: %s", constant.StudentPackageID, constant.ErrDefault)),
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: there is no current student package order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			Name:        "Failed case: Error when get package by id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when get package quantity type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				mockPackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when upsert student package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				mockPackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				mockStudentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				mockStudentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageOrderService.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			},
		},
		{
			Name:        "Failed case: Error when upsert student course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				mockPackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				mockStudentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				mockStudentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageOrderService.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			},
		},
		{
			Name:        "Failed case: Error when create student package log",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{Status: pgtype.Present},
					PackageID: pgtype.Text{Status: pgtype.Present},
					StartAt:   pgtype.Timestamptz{Status: pgtype.Present},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive:    pgtype.Bool{Status: pgtype.Present},
					LocationIDs: pgtype.TextArray{Status: pgtype.Present},
					CreatedAt:   pgtype.Timestamptz{Status: pgtype.Present},
					UpdatedAt:   pgtype.Timestamptz{Status: pgtype.Present},
					DeletedAt:   pgtype.Timestamptz{Status: pgtype.Present},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				mockPackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				mockStudentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				mockStudentPackageOrderService.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Failed case: Error when update student package order execute status",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{Status: pgtype.Present},
					PackageID: pgtype.Text{Status: pgtype.Present},
					StartAt:   pgtype.Timestamptz{Status: pgtype.Present},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive:    pgtype.Bool{Status: pgtype.Present},
					LocationIDs: pgtype.TextArray{Status: pgtype.Present},
					CreatedAt:   pgtype.Timestamptz{Status: pgtype.Present},
					UpdatedAt:   pgtype.Timestamptz{Status: pgtype.Present},
					DeletedAt:   pgtype.Timestamptz{Status: pgtype.Present},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				mockPackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				mockStudentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageOrderService.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentPackages{
					ID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{Status: pgtype.Present},
					PackageID: pgtype.Text{Status: pgtype.Present},
					StartAt:   pgtype.Timestamptz{Status: pgtype.Present},
					EndAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive:    pgtype.Bool{Status: pgtype.Present},
					LocationIDs: pgtype.TextArray{Status: pgtype.Present},
					CreatedAt:   pgtype.Timestamptz{Status: pgtype.Present},
					UpdatedAt:   pgtype.Timestamptz{Status: pgtype.Present},
					DeletedAt:   pgtype.Timestamptz{Status: pgtype.Present},
				},
			},
			Setup: func(ctx context.Context) {
				mockStudentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID:                   pgtype.Text{String: constant.StudentPackageOrderID},
					StartAt:              pgtype.Timestamptz{Time: now.AddDate(0, 1, 0), Status: pgtype.Present},
					StudentPackageObject: pgtype.JSONB{Bytes: studentPackageObjectJSON, Status: pgtype.Present},
				}, nil)
				mockPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				mockPackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				mockStudentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockStudentPackageOrderService.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockStudentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			mockStudentPackageOrderService = new(mockServices.StudentPackageOrderService)
			mockStudentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			mockPackageRepo = new(mockRepositories.MockPackageRepo)
			mockPackageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			mockStudentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)

			testCase.Setup(testCase.Ctx)

			studentPackageService := &StudentPackageService{
				StudentPackageRepo:             mockStudentPackageRepo,
				StudentCourseRepo:              mockStudentCourseRepo,
				StudentPackageOrderRepo:        mockStudentPackageOrderRepo,
				StudentPackageOrderService:     mockStudentPackageOrderService,
				PackageRepo:                    mockPackageRepo,
				PackageQuantityTypeMappingRepo: mockPackageQuantityTypeMappingRepo,
				StudentPackageLogRepo:          mockStudentPackageLogRepo,
			}
			studentPackageReq := testCase.Req.([]interface{})[0].(entities.StudentPackages)
			_, _, err := studentPackageService.UpsertStudentPackageDataForCronjob(testCase.Ctx, db, studentPackageReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockStudentPackageRepo, mockStudentCourseRepo, mockPackageRepo, mockPackageQuantityTypeMappingRepo, mockStudentPackageOrderService)
		})
	}
}
