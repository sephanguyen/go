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
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	pmpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentPackage_voidStudentPackageDataForCancelOrder(t *testing.T) {
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
			Name: "Fail case: Error when revert by student id and course id",
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
					OrderID:   pgtype.Text{},
					PackageID: pgtype.Text{},
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
			ExpectedErr: status.Errorf(codes.Internal, "error when revert student package access path with student_id = %s and course_id = %s and error = %v", constant.StudentID, constant.CourseID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student package access path by student id and course id",
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
					OrderID:   pgtype.Text{},
					PackageID: pgtype.Text{},
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
			ExpectedErr: status.Errorf(codes.Internal, "error when get student package access path with student_id = %s and course_id = %s and error = %v", constant.StudentID, constant.CourseID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get package",
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
			ExpectedErr: status.Errorf(codes.Internal, "error when get package with package_id = %s and error = %v", constant.PackageID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.QuantityType_QUANTITY_TYPE_SLOT.String(),
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get package",
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
			ExpectedErr: status.Errorf(codes.Internal, "error when get quantity type with package_type = %s and error = %v", pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get package",
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
			ExpectedErr: status.Errorf(codes.Internal, "error when get student package order with student_package_id = %s and order_id = %s and error = %v", constant.StudentPackageID, constant.OrderID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
				}, nil)
				packageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pmpb.PackageType_PACKAGE_TYPE_FREQUENCY.String(),
						Status: pgtype.Present,
					},
				}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pmpb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when delete student package order by id",
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
					},
					FromStudentPackageOrderID: pgtype.Text{
						String: constant.FromStudentPackageOrderID,
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase studentPackageOrderPosition = CurrentStudentPackage, error when set current student package order by time and student package id",
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
				}, nil)
				studentPackageOrderService.On("RevertStudentPackageOrderByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Incase studentPackageOrderPosition = CurrentStudentPackage, error when upsert student package",
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
				studentPackageRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Incase studentPackageOrderPosition = CurrentStudentPackage, error when upsert student course",
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
				studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Incase studentPackageOrderPosition = CurrentStudentPackage, error when insert student package log",
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
			orderItemCourseReq := testCase.Req.([]interface{})[1].(entities.OrderItemCourse)
			studentPackageAccessPathReq := testCase.Req.([]interface{})[2].(entities.StudentPackageAccessPath)

			resp, err := s.voidStudentPackageDataForCancelOrder(testCase.Ctx, db, voidStudentPackageArgs, orderItemCourseReq, studentPackageAccessPathReq)

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

func TestStudentPackage_voidStudentPackageForCancelOrder(t *testing.T) {
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
			ExpectedErr: status.Errorf(codes.Internal, "error when get map order item course for void update order with order_id = %s and package_id = %s and error = %v", constant.OrderID, constant.ProductID, constant.ErrDefault),
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
							String: "",
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
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{
					constant.CourseID: entities.OrderItemCourse{
						CourseID: pgtype.Text{
							String: constant.CourseID,
							Status: pgtype.Present,
						},
					},
				}, nil)
				studentPackageAccessPathRepo.On("RevertByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageAccessPathRepo.On("GetByStudentIDAndCourseID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackageAccessPath{
					StudentPackageID: pgtype.Text{
						String: constant.StudentPackageID,
						Status: pgtype.Present,
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

			resp, err := s.voidStudentPackageForCancelOrder(testCase.Ctx, db, voidStudentPackageArgs, mapStudentCourseKeyWithStudentPackageAccessPath)

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

func TestStudentPackage_MutationStudentPackageForCancelOrder(t *testing.T) {
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
	// packageProperties := entities.PackageProperties{
	// 	AllCourseInfo: []entities.CourseInfo{
	// 		{
	// 			CourseID:      constant.CourseID,
	// 			Name:          constant.CourseName,
	// 			NumberOfSlots: 1,
	// 			Weight:        6,
	// 		},
	// 	},
	// 	CanWatchVideo:     []string{constant.CourseID},
	// 	CanViewStudyGuide: []string{constant.CourseID},
	// 	CanDoQuiz:         []string{constant.CourseID},
	// 	LimitOnlineLesson: 0,
	// 	AskTutor: &entities.AskTutorCfg{
	// 		TotalQuestionLimit: 0,
	// 		LimitDuration:      "",
	// 	},
	// }
	// packagePropertiesJson, _ := json.Marshal(packageProperties)
	// studentPackageObject := entities.StudentPackages{
	// 	ID: pgtype.Text{
	// 		String: constant.StudentPackageID,
	// 		Status: pgtype.Present,
	// 	},
	// 	StudentID: pgtype.Text{
	// 		String: constant.StudentID,
	// 		Status: pgtype.Present,
	// 	},
	// 	PackageID: pgtype.Text{
	// 		String: constant.PackageID,
	// 		Status: pgtype.Present,
	// 	},
	// 	StartAt: pgtype.Timestamptz{
	// 		Time:   now,
	// 		Status: pgtype.Present,
	// 	},
	// 	EndAt: pgtype.Timestamptz{
	// 		Time:   now.AddDate(0, 4, 0),
	// 		Status: pgtype.Present,
	// 	},
	// 	Properties: pgtype.JSONB{
	// 		Bytes:  packagePropertiesJson,
	// 		Status: pgtype.Present,
	// 	},
	// 	IsActive: pgtype.Bool{
	// 		Bool:   false,
	// 		Status: pgtype.Present,
	// 	},
	// 	LocationIDs: pgtype.TextArray{
	// 		Elements: []pgtype.Text{
	// 			{
	// 				String: constant.LocationID,
	// 				Status: pgtype.Present,
	// 			},
	// 		},
	// 		Status: pgtype.Present,
	// 	},
	// 	CreatedAt: pgtype.Timestamptz{
	// 		Time:             now,
	// 		Status:           pgtype.Present,
	// 		InfinityModifier: 0,
	// 	},
	// 	UpdatedAt: pgtype.Timestamptz{
	// 		Time:   now,
	// 		Status: pgtype.Present,
	// 	},
	// 	DeletedAt: pgtype.Timestamptz{
	// 		Status: pgtype.Null,
	// 	},
	// }
	// studentPackageObjectJSON, _ := json.Marshal(studentPackageObject)
	var mapStudentPackageAccessPath = make(map[string]entities.StudentPackageAccessPath)
	key := fmt.Sprintf("%v_%v", constant.StudentID, constant.CourseID)
	mapStudentPackageAccessPath[key] = entities.StudentPackageAccessPath{
		StudentPackageID: pgtype.Text{
			String: constant.StudentPackageID,
			Status: pgtype.Present,
		},
	}

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get map student course key with student package access path by student_ids",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student product for update by student product_id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student package for update by student product id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
						CourseItems: []*pb.CourseItem{
							{
								CourseId: constant.CourseID,
							},
						},
						StartDate: timestamppb.New(time.Now().Add(-3 * time.Hour)),
					},
					IsOneTimeProduct: true,
					Timezone:         7,
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when DeleteStudentPackageOrderByID",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
						CourseItems: []*pb.CourseItem{
							{
								CourseId: constant.CourseID,
							},
						},
						StartDate: timestamppb.New(time.Now().Add(-3 * time.Hour)),
					},
					IsOneTimeProduct: true,
					Timezone:         7,
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: "student_package_order_id",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student package by id",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
						CourseItems: []*pb.CourseItem{
							{
								CourseId: constant.CourseID,
							},
						},
						StartDate: timestamppb.New(time.Now().Add(-3 * time.Hour)),
					},
					IsOneTimeProduct: true,
					Timezone:         7,
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: "student_package_order_id",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when cancel student package in past time",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
						CourseItems: []*pb.CourseItem{
							{
								CourseId: constant.CourseID,
							},
						},
						StartDate: timestamppb.New(time.Now().Add(-3 * time.Hour)),
					},
					IsOneTimeProduct: true,
					Timezone:         7,
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: status.Errorf(codes.Internal, "error when cancel student package in past time with student_package_id = student_package_id and student_package_order_id = student_package_order_id"),
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: "student_package_order_id",
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   false,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: "student_package_id",
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -5, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						String: "student_package_id",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: "Fail case: Error when SetCurrentStudentPackageOrderByTimeAndStudentPackageID",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
						CourseItems: []*pb.CourseItem{
							{
								CourseId: constant.CourseID,
							},
						},
						StartDate: timestamppb.New(time.Now().Add(-3 * time.Hour)),
					},
					IsOneTimeProduct: true,
					Timezone:         7,
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: "student_package_order_id",
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: "student_package_id",
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -5, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						String: "student_package_id",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when InsertStudentPackageOrder",
			Ctx:  ctx,
			Req: []interface{}{
				utils.OrderItemData{
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
					OrderItem: &pmpb.OrderItem{
						StudentProductId: wrapperspb.String("student_product_id_1"),
						CourseItems: []*pb.CourseItem{
							{
								CourseId: constant.CourseID,
							},
						},
						StartDate: timestamppb.New(time.Now().Add(-3 * time.Hour)),
					},
					IsOneTimeProduct: true,
					Timezone:         7,
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				studentPackageOrderService.On("GetStudentPackageOrderByStudentPackageIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					ID: pgtype.Text{
						String: "student_package_order_id",
						Status: pgtype.Present,
					},
					IsCurrentStudentPackage: pgtype.Bool{
						Bool:   true,
						Status: pgtype.Present,
					},
					StudentPackageID: pgtype.Text{
						String: "student_package_id",
						Status: pgtype.Present,
					},
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -5, 0),
						Status: pgtype.Present,
					},
					EndAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, -2, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("DeleteStudentPackageOrderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{
					ID: pgtype.Text{
						String: "student_package_id",
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("GetCurrentStudentPackageOrderByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{
					StartAt: pgtype.Timestamptz{
						Time:   now.AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
				}, nil)
				studentPackageOrderService.On("SetCurrentStudentPackageOrderByTimeAndStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
				studentPackageOrderService.On("InsertStudentPackageOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
			mapStudentCourseKeyWithStudentPackageAccessPath := testCase.Req.([]interface{})[0].(utils.OrderItemData)

			resp, err := s.MutationStudentPackageForCancelOrder(testCase.Ctx, db, mapStudentCourseKeyWithStudentPackageAccessPath)

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
