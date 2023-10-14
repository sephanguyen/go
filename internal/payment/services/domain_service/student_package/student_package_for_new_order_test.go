package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	mockServices "github.com/manabie-com/backend/mock/payment/services/domain_service"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pmpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentPackage_getPositionOfStudentPackageOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                             *mockDb.Ext
		studentPackageRepo             *mockRepositories.MockStudentPackageRepo
		studentPackageAccessPathRepo   *mockRepositories.MockStudentPackageAccessPathRepo
		studentPackageClassRepo        *mockRepositories.MockStudentPackageClassRepo
		studentCourseRepo              *mockRepositories.MockStudentCourseRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		studentProductRepo             *mockRepositories.MockStudentProductRepo
		studentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		studentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		productRepo                    *mockRepositories.MockProductRepo
		studentPackageOrderService     *mockServices.StudentPackageOrderService
		now                            = time.Now().UTC()
	)
	testcases := []utils.TestCase{
		{
			Name: "Happy case: CurrentStudentPackage when is_current_student_package = true",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				},
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: error when missing student package id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
				},
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  status.Errorf(codes.Internal, "error when missing student package id while getting position of student package order"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: error when get current student package order by student package id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				},
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  status.Errorf(codes.Internal, "error when get current student package order by student package id with student_package_id = %v: %v", constant.StudentPackageID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, status.Errorf(codes.Internal, "error when get current student package order by student package id with student_package_id = %v: %v", constant.StudentPackageID, constant.ErrDefault))
			},
		},
		{
			Name: "Happy case: PastStudentPackage when student package order before current student package order",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				},
			},
			ExpectedResp: entities.PastStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 2, 0),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: FutureStudentPackage when student package order after current student package order",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 2, 0),
						Status: pgtype.Present,
					},
				},
			},
			ExpectedResp: entities.FutureStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			studentPackageAccessPathRepo = new(mockRepositories.MockStudentPackageAccessPathRepo)
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			studentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			packageRepo = new(mockRepositories.MockPackageRepo)
			packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			studentPackageOrderService = new(mockServices.StudentPackageOrderService)

			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageRepo:             studentPackageRepo,
				StudentPackageAccessPathRepo:   studentPackageAccessPathRepo,
				StudentPackageClassRepo:        studentPackageClassRepo,
				StudentCourseRepo:              studentCourseRepo,
				OrderItemCourseRepo:            orderItemCourseRepo,
				OrderItemRepo:                  orderItemRepo,
				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			studentPackageOrderReq := testCase.Req.([]interface{})[0].(entities.StudentPackageOrder)
			resp, err := s.getPositionOfStudentPackageOrder(testCase.Ctx, db, studentPackageOrderReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_voidStudentPackageDataForCreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                             *mockDb.Ext
		studentPackageRepo             *mockRepositories.MockStudentPackageRepo
		studentPackageAccessPathRepo   *mockRepositories.MockStudentPackageAccessPathRepo
		studentPackageClassRepo        *mockRepositories.MockStudentPackageClassRepo
		studentCourseRepo              *mockRepositories.MockStudentCourseRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		studentProductRepo             *mockRepositories.MockStudentProductRepo
		studentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		studentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		productRepo                    *mockRepositories.MockProductRepo
		studentPackageOrderService     *mockServices.StudentPackageOrderService
		now                            = time.Now().UTC()
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
			Time:   now,
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
			Name: "Fail case: Error when get student package order by student package id and order id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when void student package in PAST",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  status.Errorf(codes.Internal, "error when void student package in past time with student_package_id = %v and student_package_order_id = %v", constant.StudentPackageID, constant.StudentPackageOrderID),
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 2, 0),
						Status: pgtype.Present,
					},
				}, status.Errorf(codes.Internal, "error when void student package in past time with student_package_id = %v and student_package_order_id = %v", constant.StudentPackageID, constant.StudentPackageOrderID))
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when deleting student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when get student package by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when set current student package order by time and student package id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase position = CurrentStudentPackage, error when upsert student package",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.StudentPackages{}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					UserID: pgtype.Text{
						String: constant.UserID,
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:             now.AddDate(0, 4, 0),
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Null,
					},
				}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.QuantityType_QUANTITY_TYPE_SLOT.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: Incase position = CurrentStudentPackage, error when upsert student course",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					UserID: pgtype.Text{
						String: constant.UserID,
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:             now.AddDate(0, 4, 0),
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Null,
					},
				}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.QuantityType_QUANTITY_TYPE_SLOT.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: Incase position = CurrentStudentPackage, error when cancel student package by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				studentPackageRepo.On("CancelByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentCourseRepo.On("CancelByStudentPackageIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("SoftDeleteByStudentPackageIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: Incase position = CurrentStudentPackage, error when cancel student course bv student package id and course id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				studentPackageRepo.On("CancelByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("CancelByStudentPackageIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageAccessPathRepo.On("SoftDeleteByStudentPackageIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: Incase position = CurrentStudentPackage, error when delete student package access path by student package ids",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				studentPackageRepo.On("CancelByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("CancelByStudentPackageIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("SoftDeleteByStudentPackageIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: Incase position = CurrentStudentPackage, error when create student package log",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{},
				constant.CourseID,
				entities.StudentPackageAccessPath{},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				studentPackageRepo.On("CancelByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("CancelByStudentPackageIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("SoftDeleteByStudentPackageIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{},
					Product:        entities.Product{},
					IsCancel:       false,
				},
				constant.CourseID,
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: constant.StudentID,
					Package: &npb.EventStudentPackage_Package{
						CourseIds: []string{constant.CourseID},
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
					},
					IsActive: false,
				},
				LocationIds: database.FromTextArray(pgtype.TextArray{
					Elements: []pgtype.Text{{
						String: constant.LocationID,
						Status: pgtype.Present,
					}},
					Status: pgtype.Present,
				}),
			},
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				studentPackageRepo.On("CancelByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("CancelByStudentPackageIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("SoftDeleteByStudentPackageIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			studentPackageAccessPathRepo = new(mockRepositories.MockStudentPackageAccessPathRepo)
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			studentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			packageRepo = new(mockRepositories.MockPackageRepo)
			packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			studentPackageOrderService = new(mockServices.StudentPackageOrderService)

			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageRepo:             studentPackageRepo,
				StudentPackageAccessPathRepo:   studentPackageAccessPathRepo,
				StudentPackageClassRepo:        studentPackageClassRepo,
				StudentCourseRepo:              studentCourseRepo,
				OrderItemCourseRepo:            orderItemCourseRepo,
				OrderItemRepo:                  orderItemRepo,
				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			voidStudentPackageArgs := testCase.Req.([]interface{})[0].(utils.VoidStudentPackageArgs)
			courseID := testCase.Req.([]interface{})[1].(string)
			studentPackageAccessPath := testCase.Req.([]interface{})[2].(entities.StudentPackageAccessPath)
			resp, err := s.voidStudentPackageDataForCreateOrder(testCase.Ctx, db, voidStudentPackageArgs, courseID, studentPackageAccessPath)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_convertStudentPackageDataByStudentPackageOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                             *mockDb.Ext
		studentPackageRepo             *mockRepositories.MockStudentPackageRepo
		studentPackageAccessPathRepo   *mockRepositories.MockStudentPackageAccessPathRepo
		studentPackageClassRepo        *mockRepositories.MockStudentPackageClassRepo
		studentCourseRepo              *mockRepositories.MockStudentCourseRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		studentProductRepo             *mockRepositories.MockStudentProductRepo
		studentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		studentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		productRepo                    *mockRepositories.MockProductRepo
		studentPackageOrderService     *mockServices.StudentPackageOrderService
		now                            = time.Now().UTC()
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
			Time:   now,
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
			Name: "Fail case: error when get package by id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					UserID: pgtype.Text{
						String: constant.UserID,
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:             now.AddDate(0, 4, 0),
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Null,
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  status.Errorf(codes.Internal, fmt.Sprintf("Error when get package by package id: %v", constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error when get package quantity type mapping by package type",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					UserID: pgtype.Text{
						String: constant.UserID,
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:             now.AddDate(0, 4, 0),
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Null,
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  status.Errorf(codes.Internal, fmt.Sprintf("Error when get package quantity type mapping: %v", constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					UserID: pgtype.Text{
						String: constant.UserID,
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:             now.AddDate(0, 4, 0),
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Null,
					},
				},
			},
			ExpectedResp: []interface{}{
				entities.StudentPackages{
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
						Time:   now,
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
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
				},
				entities.StudentCourse{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: constant.LocationID,
						Status: pgtype.Present,
					},
					StudentStartDate: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					StudentEndDate: pgtype.Timestamptz{
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
					},
					CourseSlot: pgtype.Int4{
						Int:    1,
						Status: pgtype.Present,
					},
					CourseSlotPerWeek: pgtype.Int4{
						Int:    0,
						Status: pgtype.Null,
					},
					Weight: pgtype.Int4{
						Int:    0,
						Status: pgtype.Null,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   now,
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Null,
					},
					PackageType: pgtype.Text{
						String: pmpb.QuantityType_QUANTITY_TYPE_SLOT.String(),
						Status: pgtype.Present,
					},
					ResourcePath: pgtype.Text{
						Status: pgtype.Null,
					},
				},
				&npb.EventStudentPackage{
					StudentPackage: &npb.EventStudentPackage_StudentPackage{
						StudentId: constant.StudentID,
						Package: &npb.EventStudentPackage_Package{
							CourseIds: []string{constant.CourseID},
							StartDate: timestamppb.New(now),
							EndDate:   timestamppb.New(now.AddDate(0, 4, 0)),
							LocationIds: database.FromTextArray(pgtype.TextArray{
								Elements: []pgtype.Text{{
									String: constant.LocationID,
									Status: pgtype.Present,
								}},
								Status: pgtype.Present,
							}),
							StudentPackageId: constant.StudentPackageID,
						},
						IsActive: true,
					},
					LocationIds: database.FromTextArray(pgtype.TextArray{
						Elements: []pgtype.Text{{
							String: constant.LocationID,
							Status: pgtype.Present,
						}},
						Status: pgtype.Present,
					}),
				},
			},
			Setup: func(ctx context.Context) {
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.QuantityType_QUANTITY_TYPE_SLOT.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			studentPackageAccessPathRepo = new(mockRepositories.MockStudentPackageAccessPathRepo)
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			studentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			packageRepo = new(mockRepositories.MockPackageRepo)
			packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			studentPackageOrderService = new(mockServices.StudentPackageOrderService)

			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageRepo:             studentPackageRepo,
				StudentPackageAccessPathRepo:   studentPackageAccessPathRepo,
				StudentPackageClassRepo:        studentPackageClassRepo,
				StudentCourseRepo:              studentCourseRepo,
				OrderItemCourseRepo:            orderItemCourseRepo,
				OrderItemRepo:                  orderItemRepo,
				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			studentPackageOrderReq := testCase.Req.([]interface{})[0].(entities.StudentPackageOrder)

			studentPackageResp, studentCourseResp, eventMessageResp, err := s.convertStudentPackageDataByStudentPackageOrder(testCase.Ctx, db, studentPackageOrderReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedStudentPackageResp := testCase.ExpectedResp.([]interface{})[0].(entities.StudentPackages)
				expectedStudentCourseResp := testCase.ExpectedResp.([]interface{})[1].(entities.StudentCourse)
				expectedEventMessageResp := testCase.ExpectedResp.([]interface{})[2].(*npb.EventStudentPackage)

				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, expectedStudentPackageResp, studentPackageResp)
				assert.Equal(t, expectedStudentCourseResp, studentCourseResp)
				assert.Equal(t, expectedEventMessageResp, eventMessageResp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_voidStudentPackageForCreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                             *mockDb.Ext
		studentPackageRepo             *mockRepositories.MockStudentPackageRepo
		studentPackageAccessPathRepo   *mockRepositories.MockStudentPackageAccessPathRepo
		studentPackageClassRepo        *mockRepositories.MockStudentPackageClassRepo
		studentCourseRepo              *mockRepositories.MockStudentCourseRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		studentProductRepo             *mockRepositories.MockStudentProductRepo
		studentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		studentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		productRepo                    *mockRepositories.MockProductRepo
		studentPackageOrderService     *mockServices.StudentPackageOrderService
		now                            = time.Now().UTC()
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
			Time:   now,
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
	//studentPackageObjectJSON, _ := json.Marshal(studentPackageObject)
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get map order item course by order id and package id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					Product: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						}},
				},
				map[string]entities.StudentPackageAccessPath{
					constant.CourseID: entities.StudentPackageAccessPath{},
				},
			},
			ExpectedResp: []*npb.EventStudentPackage{
				{
					StudentPackage: &npb.EventStudentPackage_StudentPackage{
						StudentId: constant.StudentID,
						Package: &npb.EventStudentPackage_Package{
							CourseIds: []string{constant.CourseID},
							LocationIds: database.FromTextArray(pgtype.TextArray{
								Elements: []pgtype.Text{{
									String: constant.LocationID,
									Status: pgtype.Present,
								}},
								Status: pgtype.Present,
							}),
							StudentPackageId: constant.StudentPackageID,
						},
						IsActive: false,
					},
					LocationIds: database.FromTextArray(pgtype.TextArray{
						Elements: []pgtype.Text{{
							String: constant.LocationID,
							Status: pgtype.Present,
						}},
						Status: pgtype.Present,
					}),
				},
			},
			ExpectedErr: status.Errorf(codes.Internal, "error when get map order item course for void create order with order_id = %s and package_id = %s and error = %v", constant.OrderID, constant.ProductID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{
					constant.CourseID: entities.OrderItemCourse{
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						}},
				},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: []*npb.EventStudentPackage{
				{
					StudentPackage: &npb.EventStudentPackage_StudentPackage{
						StudentId: constant.StudentID,
						Package: &npb.EventStudentPackage_Package{
							CourseIds: []string{constant.CourseID},
							LocationIds: database.FromTextArray(pgtype.TextArray{
								Elements: []pgtype.Text{{
									String: constant.LocationID,
									Status: pgtype.Present,
								}},
								Status: pgtype.Present,
							}),
							StudentPackageId: constant.StudentPackageID,
						},
						IsActive: false,
					},
					LocationIds: database.FromTextArray(pgtype.TextArray{
						Elements: []pgtype.Text{{
							String: constant.LocationID,
							Status: pgtype.Present,
						}},
						Status: pgtype.Present,
					}),
				},
			},
			Setup: func(ctx context.Context) {
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{
					constant.CourseID: entities.OrderItemCourse{
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
					},
				}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{String: constant.StudentPackageOrderID},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(studentPackageObject, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				studentPackageRepo.On("CancelByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("CancelByStudentPackageIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("SoftDeleteByStudentPackageIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			studentPackageAccessPathRepo = new(mockRepositories.MockStudentPackageAccessPathRepo)
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			studentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			packageRepo = new(mockRepositories.MockPackageRepo)
			packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			studentPackageOrderService = new(mockServices.StudentPackageOrderService)

			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageRepo:             studentPackageRepo,
				StudentPackageAccessPathRepo:   studentPackageAccessPathRepo,
				StudentPackageClassRepo:        studentPackageClassRepo,
				StudentCourseRepo:              studentCourseRepo,
				OrderItemCourseRepo:            orderItemCourseRepo,
				OrderItemRepo:                  orderItemRepo,
				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			voidStudentPackageArgs := testCase.Req.([]interface{})[0].(utils.VoidStudentPackageArgs)
			mapStudentCourseKeyWithStudentPackageAccessPath := testCase.Req.([]interface{})[1].(map[string]entities.StudentPackageAccessPath)

			resp, err := s.voidStudentPackageForCreateOrder(testCase.Ctx, db, voidStudentPackageArgs, mapStudentCourseKeyWithStudentPackageAccessPath)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_MutationStudentPackageForCreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                             *mockDb.Ext
		studentPackageRepo             *mockRepositories.MockStudentPackageRepo
		studentPackageAccessPathRepo   *mockRepositories.MockStudentPackageAccessPathRepo
		studentPackageClassRepo        *mockRepositories.MockStudentPackageClassRepo
		studentCourseRepo              *mockRepositories.MockStudentCourseRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		studentProductRepo             *mockRepositories.MockStudentProductRepo
		studentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		studentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		productRepo                    *mockRepositories.MockProductRepo
		studentPackageOrderService     *mockServices.StudentPackageOrderService
		now                            = time.Now().UTC()
	)

	var testcases = []utils.TestCase{
		{
			Name: "Fail case: error when get package by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{Package: entities.Package{
						PackageStartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						PackageEndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					}},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  status.Errorf(codes.Internal, "error when get position for student package by time with student_package_id = %s, error = %v", constant.StudentPackageID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when upsert student package",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present}},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when upsert student course",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present}},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when insert student package order",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present}},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Incase CurrentStudentPackage, error when create student package log",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present}},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase FutureStudentPackage, error when insert student package order",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.FutureStudentPackage, nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase FutureStudentPackage, error when insert student package access path",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s1", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.FutureStudentPackage, nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
				&pmpb.CourseItem{CourseId: constant.CourseID},
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%s_%s", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: constant.StudentID,
					Package: &npb.EventStudentPackage_Package{
						CourseIds: []string{constant.CourseID},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now.AddDate(0, 4, 0)),
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
					},
					IsActive: true,
				},
				LocationIds: database.FromTextArray(pgtype.TextArray{
					Elements: []pgtype.Text{{
						String: constant.LocationID,
						Status: pgtype.Present,
					}},
					Status: pgtype.Present,
				}),
			},
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			studentPackageAccessPathRepo = new(mockRepositories.MockStudentPackageAccessPathRepo)
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			studentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			packageRepo = new(mockRepositories.MockPackageRepo)
			packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			studentPackageOrderService = new(mockServices.StudentPackageOrderService)

			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageRepo:             studentPackageRepo,
				StudentPackageAccessPathRepo:   studentPackageAccessPathRepo,
				StudentPackageClassRepo:        studentPackageClassRepo,
				StudentCourseRepo:              studentCourseRepo,
				OrderItemCourseRepo:            orderItemCourseRepo,
				OrderItemRepo:                  orderItemRepo,
				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			orderItemDataReq := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			courseReq := testCase.Req.([]interface{})[1].(*pmpb.CourseItem)
			mapStudentCourseWithStudentPackageAccessPathReq := testCase.Req.([]interface{})[2].(map[string]entities.StudentPackageAccessPath)

			eventMessageResp, err := s.upsertStudentPackageDataForNewOrder(testCase.Ctx, db, orderItemDataReq, courseReq, mapStudentCourseWithStudentPackageAccessPathReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, eventMessageResp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_upsertStudentPackageDataForNewOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                             *mockDb.Ext
		studentPackageRepo             *mockRepositories.MockStudentPackageRepo
		studentPackageAccessPathRepo   *mockRepositories.MockStudentPackageAccessPathRepo
		studentPackageClassRepo        *mockRepositories.MockStudentPackageClassRepo
		studentCourseRepo              *mockRepositories.MockStudentCourseRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		studentProductRepo             *mockRepositories.MockStudentProductRepo
		studentPackageLogRepo          *mockRepositories.MockStudentPackageLogRepo
		studentPackageOrderRepo        *mockRepositories.MockStudentPackageOrderRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		productRepo                    *mockRepositories.MockProductRepo
		studentPackageOrderService     *mockServices.StudentPackageOrderService
		now                            = time.Now().UTC()
	)

	var testcases = []utils.TestCase{
		{
			Name: "Fail case: error when get map student course key with student package access path by student ids",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedResp: []*npb.EventStudentPackage{},
			ExpectedErr:  status.Errorf(codes.Internal, "error when get map student course key with student package access path by student ids with student_id = %s, err = %v", constant.StudentID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: false,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						Package: entities.Package{
							PackageStartDate: pgtype.Timestamptz{
								Time:   now,
								Status: pgtype.Present,
							},
							PackageEndDate: pgtype.Timestamptz{
								Time:   now.AddDate(0, 4, 0),
								Status: pgtype.Present,
							},
						}},
					StudentProduct: entities.StudentProduct{
						StartDate: pgtype.Timestamptz{
							Time:   now,
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   now.AddDate(0, 4, 0),
							Status: pgtype.Present,
						},
					},
					OrderItem: &pmpb.OrderItem{
						CourseItems: []*pmpb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
					},
				},
			},
			ExpectedResp: []*npb.EventStudentPackage{
				{
					StudentPackage: &npb.EventStudentPackage_StudentPackage{
						StudentId: constant.StudentID,
						Package: &npb.EventStudentPackage_Package{
							CourseIds: []string{constant.CourseID},
							StartDate: timestamppb.New(now),
							EndDate:   timestamppb.New(now.AddDate(0, 4, 0)),
							LocationIds: database.FromTextArray(pgtype.TextArray{
								Elements: []pgtype.Text{{
									String: constant.LocationID,
									Status: pgtype.Present,
								}},
								Status: pgtype.Present,
							}),
							StudentPackageId: constant.StudentPackageID,
						},
						IsActive: true,
					},
					LocationIds: database.FromTextArray(pgtype.TextArray{
						Elements: []pgtype.Text{{
							String: constant.LocationID,
							Status: pgtype.Present,
						}},
						Status: pgtype.Present,
					}),
				},
			},
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: constant.StudentPackageID,
							Status: pgtype.Present,
						},
					},
				}, nil)
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentPackageRepo = new(mockRepositories.MockStudentPackageRepo)
			studentPackageAccessPathRepo = new(mockRepositories.MockStudentPackageAccessPathRepo)
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			studentCourseRepo = new(mockRepositories.MockStudentCourseRepo)
			orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentPackageLogRepo = new(mockRepositories.MockStudentPackageLogRepo)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			packageRepo = new(mockRepositories.MockPackageRepo)
			packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			studentPackageOrderService = new(mockServices.StudentPackageOrderService)

			testCase.Setup(testCase.Ctx)

			s := &StudentPackageService{
				StudentPackageRepo:             studentPackageRepo,
				StudentPackageAccessPathRepo:   studentPackageAccessPathRepo,
				StudentPackageClassRepo:        studentPackageClassRepo,
				StudentCourseRepo:              studentCourseRepo,
				OrderItemCourseRepo:            orderItemCourseRepo,
				OrderItemRepo:                  orderItemRepo,
				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			orderItemDataReq := testCase.Req.([]interface{})[0].(utils.OrderItemData)

			eventMessageResp, err := s.MutationStudentPackageForCreateOrder(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, eventMessageResp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}
