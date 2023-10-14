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
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentPackage_voidStudentPackageDataForUpdateOrder(t *testing.T) {
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
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student package order by student package id and order id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when soft delete student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when revert student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when set current student package order by time and student package id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when upsert student package",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when upsert student course",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when update student package order",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when create student package log",
			Ctx:  ctx,
			Req: []interface{}{
				utils.VoidStudentPackageArgs{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
							Status: pgtype.Present,
						},
					},
					StudentProduct: entities.StudentProduct{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
					Product:  entities.Product{},
					IsCancel: false,
				},
				entities.OrderItemCourse{
					OrderID: pgtype.Text{},
					PackageID: pgtype.Text{
						String: constant.PackageID,
						Status: pgtype.Present,
					},
					CourseID: pgtype.Text{
						String: constant.CourseID,
						Status: pgtype.Present,
					},
					CourseName:        pgtype.Text{},
					CourseSlot:        pgtype.Int4{},
					OrderItemCourseID: pgtype.Text{},
				},
				entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: "",
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
				StudentPackageRepo:           studentPackageRepo,
				StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
				StudentPackageClassRepo:      studentPackageClassRepo,
				StudentCourseRepo:            studentCourseRepo,
				OrderItemCourseRepo:          orderItemCourseRepo,
				OrderItemRepo:                orderItemRepo,

				StudentProductRepo:             studentProductRepo,
				StudentPackageLogRepo:          studentPackageLogRepo,
				StudentPackageOrderRepo:        studentPackageOrderRepo,
				PackageRepo:                    packageRepo,
				PackageQuantityTypeMappingRepo: packageQuantityTypeMappingRepo,
				ProductRepo:                    productRepo,
				StudentPackageOrderService:     studentPackageOrderService,
			}
			voidStudentPackageArgs := testCase.Req.([]interface{})[0].(utils.VoidStudentPackageArgs)
			orderItemCourseReq := testCase.Req.([]interface{})[1].(entities.OrderItemCourse)
			studentPackageAccessPathReq := testCase.Req.([]interface{})[2].(entities.StudentPackageAccessPath)

			resp, err := s.voidStudentPackageDataForUpdateOrder(testCase.Ctx, db, voidStudentPackageArgs, orderItemCourseReq, studentPackageAccessPathReq)

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

func TestStudentPackage_voidStudentPackageForUpdateOrder(t *testing.T) {
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
							String: "",
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get order item by student product id",
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
							String: "",
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, constant.ErrDefault)
			},
		},
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
							String: "",
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
				}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: Incase Void student package for create order",
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
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{
					constant.CourseID: {
						OrderID: pgtype.Text{
							String: "update_order_id",
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
					},
				}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
				}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{}, nil)
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
		{
			Name: "Happy case: Incase Void student package for Cancel order",
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
							StartDate:        timestamppb.New(now),
							EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
				}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{
					constant.CourseID: {
						OrderID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
					},
				}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
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
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Happy case: Incase Void student package for Update order",
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
							StartDate:        timestamppb.New(now),
							EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{
					constant.CourseID: {
						OrderID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						PackageID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						CourseSlot: pgtype.Int4{
							Int:    3,
							Status: pgtype.Present,
						},
					},
				}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
				}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string]entities.OrderItemCourse{
					constant.CourseID: {
						OrderID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						PackageID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
						CourseSlot: pgtype.Int4{
							Int:    6,
							Status: pgtype.Present,
						},
					},
				}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: constant.StudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
				productRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
			voidStudentPackageArgs := testCase.Req.([]interface{})[0].(utils.VoidStudentPackageArgs)
			mapStudentCourseKeyWithStudentPackageAccessPath := testCase.Req.([]interface{})[1].(map[string]entities.StudentPackageAccessPath)

			resp, err := s.voidStudentPackageForUpdateOrder(testCase.Ctx, db, voidStudentPackageArgs, mapStudentCourseKeyWithStudentPackageAccessPath)

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

func TestStudentPackage_updateStudentPackageDataForNonCompleteUpdateOrder(t *testing.T) {
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
			Name: "Fail case: Error when get student package order by id and time",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student package by id",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when upsert student package by id",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						Status: pgtype.Present,
					},
					PackageID: pgtype.Text{
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					EndAt: pgtype.Timestamptz{
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					LocationIDs: pgtype.TextArray{
						Elements:   nil,
						Dimensions: nil,
						Status:     pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
				}, nil)
				studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when upsert student course",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						Status: pgtype.Present,
					},
					PackageID: pgtype.Text{
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					EndAt: pgtype.Timestamptz{
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					LocationIDs: pgtype.TextArray{
						Elements:   nil,
						Dimensions: nil,
						Status:     pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
				}, nil)
				studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when update student package order",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						Status: pgtype.Present,
					},
					PackageID: pgtype.Text{
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					EndAt: pgtype.Timestamptz{
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					LocationIDs: pgtype.TextArray{
						Elements:   nil,
						Dimensions: nil,
						Status:     pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
				}, nil)
				studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when insert new student package order",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageObject: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						Status: pgtype.Present,
					},
					PackageID: pgtype.Text{
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					EndAt: pgtype.Timestamptz{
						Status: pgtype.Present,
					},
					Properties: pgtype.JSONB{
						Bytes:  studentPackageObjectJSON,
						Status: pgtype.Present,
					},
					IsActive: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					LocationIDs: pgtype.TextArray{
						Elements:   nil,
						Dimensions: nil,
						Status:     pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:             time.Time{},
						Status:           pgtype.Present,
						InfinityModifier: 0,
					},
				}, nil)
				studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
							Status: pgtype.Present,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
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
					IsOneTimeProduct: true,
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: map[string]*pmpb.CourseItem{
							constant.CourseID: {
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     wrapperspb.Int32(6),
								Slot:       wrapperspb.Int32(1),
							},
						},
						QuantityType: pmpb.QuantityType_QUANTITY_TYPE_SLOT,
						Package: entities.Package{
							PackageType: pgtype.Text{
								String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
								Status: pgtype.Present,
							},
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				constant.UpcomingStudentPackageID,
				constant.CourseID,
				map[string]entities.StudentPackageAccessPath{
					fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID): {
						StudentPackageID: pgtype.Text{
							String: "",
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
						LocationIds: database.FromTextArray(pgtype.TextArray{
							Elements: []pgtype.Text{{
								String: constant.LocationID,
								Status: pgtype.Present,
							}},
							Status: pgtype.Present,
						}),
						StudentPackageId: constant.StudentPackageID,
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(studentPackageObject, nil)
				studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("UpdateStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
			courseID := testCase.Req.([]interface{})[2].(string)
			mapStudentCourseWithStudentPackageAccessPath := testCase.Req.([]interface{})[3].(map[string]entities.StudentPackageAccessPath)

			resp, err := s.updateStudentPackageDataForNonCompleteUpdateOrder(testCase.Ctx, db, orderItemDataReq, courseID, mapStudentCourseWithStudentPackageAccessPath)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_updateStudentPackageDataForCompleteUpdateOrder(t *testing.T) {
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
			Name: "Fail case: Error when get student package order by id and time",
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				&pmpb.CourseItem{
					CourseId:   constant.CourseID,
					CourseName: constant.CourseName,
					Weight:     wrapperspb.Int32(6),
					Slot:       wrapperspb.Int32(1),
				},
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when upsert student package",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				&pmpb.CourseItem{
					CourseId:   constant.CourseID,
					CourseName: constant.CourseName,
					Weight:     wrapperspb.Int32(6),
					Slot:       wrapperspb.Int32(1),
				},
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when upsert student course",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				&pmpb.CourseItem{
					CourseId:   constant.CourseID,
					CourseName: constant.CourseName,
					Weight:     wrapperspb.Int32(6),
					Slot:       wrapperspb.Int32(1),
				},
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when delete student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				&pmpb.CourseItem{
					CourseId:   constant.CourseID,
					CourseName: constant.CourseName,
					Weight:     wrapperspb.Int32(6),
					Slot:       wrapperspb.Int32(1),
				},
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when insert student package order ",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{String: constant.StudentID},
					},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{
							String: constant.ProductID,
							Status: pgtype.Present,
						},
					},
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				&pmpb.CourseItem{
					CourseId:   constant.CourseID,
					CourseName: constant.CourseName,
					Weight:     wrapperspb.Int32(6),
					Slot:       wrapperspb.Int32(1),
				},
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
							Status: pgtype.Present,
						},
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
					IsOneTimeProduct: true,
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
						EffectiveDate: timestamppb.New(now),
					},
				},
				&pmpb.CourseItem{
					CourseId:   constant.CourseID,
					CourseName: constant.CourseName,
					Weight:     wrapperspb.Int32(6),
					Slot:       wrapperspb.Int32(1),
				},
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
						StartDate:        timestamppb.New(now),
						EndDate:          timestamppb.New(now.AddDate(0, 4, 0)),
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
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
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
						Time:   now.AddDate(0, 4, 0),
						Status: pgtype.Present,
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
						Bool:   true,
						Status: pgtype.Present,
					},
					CreatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					UpdatedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					DeletedAt: pgtype.Timestamptz{
						Time:   time.Time{},
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
					IsExecutedByCronJob: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					ExecutedError: pgtype.Text{
						String: "",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
			courseItem := testCase.Req.([]interface{})[1].(*pmpb.CourseItem)
			studentPackageAccessPath := testCase.Req.([]interface{})[2].(entities.StudentPackageAccessPath)

			resp, err := s.updateStudentPackageDataForCompleteUpdateOrder(testCase.Ctx, db, orderItemDataReq, courseItem, studentPackageAccessPath)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}

func TestStudentPackage_MutationStudentPackageForUpdateOrder(t *testing.T) {
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
			Name: "Fail case: error when get order item by student product id",
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
						StudentProductId: wrapperspb.String(constant.StudentPackageID),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error when get map order item course by order id and package id",
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
						StudentProductId: wrapperspb.String(constant.StudentPackageID),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error when get student product for update by student product id",
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
						StudentProductId: wrapperspb.String(constant.StudentPackageID),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String(constant.StudentPackageID),
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
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: Upsert student package data when add new data in update order",
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
					IsOneTimeProduct: true,
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
						StudentProductId: wrapperspb.String(constant.StudentPackageID),
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
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentPackageOrderService.On("GetPositionForStudentPackageByTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.CurrentStudentPackage, nil)
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
						StudentProductId: wrapperspb.String(constant.StudentPackageID),
						CourseItems:      []*pmpb.CourseItem{},
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
				orderItemRepo.On("GetOrderItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
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

			_, err := s.MutationStudentPackageForUpdateOrder(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageRepo, studentPackageAccessPathRepo,
				studentPackageClassRepo, studentCourseRepo, orderItemCourseRepo, orderItemRepo, studentProductRepo, studentPackageLogRepo,
				studentPackageOrderRepo, packageRepo, packageQuantityTypeMappingRepo, productRepo)
		})
	}
}
