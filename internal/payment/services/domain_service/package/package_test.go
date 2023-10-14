package service

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	HappyCaseQuantityTypeSlot           = constant.HappyCase + "(quantity_type = pb.QuantityType_QUANTITY_TYPE_SLOT)"
	HappyCaseQuantityTypeSlotPerWeek    = constant.HappyCase + "(quantity_type = pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK)"
	FailCaseCourseHasSlotGreaterThanMax = "Fail case: Error when course has slot greater than max slot of the course"
	CourseID2                           = "CourseID_2"
	CourseID3                           = "CourseID_3"
)

func TestPackageService_convertMapCourseAndStudentCourse(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when missing slot (quantity_type == QuantityType_QUANTITY_TYPE_SLOT)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.CourseItemMissingSlotField),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						CourseItems: []*pb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     nil,
								Slot:       nil,
							},
						},
					},
				},
				pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when missing slot (quantity_type == QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.CourseItemMissingSlotField),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						CourseItems: []*pb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     nil,
								Slot:       nil,
							},
						},
					},
				},
				pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String(),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when missing slot (quantity_type == QuantityType_QUANTITY_TYPE_SLOT)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "course item for weight is missing weight field"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						CourseItems: []*pb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     nil,
								Slot:       nil,
							},
						},
					},
				},
				pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String(),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case(quantity_type == QuantityType_QUANTITY_TYPE_SLOT)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						CourseItems: []*pb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     nil,
								Slot:       &wrapperspb.Int32Value{Value: 5},
							},
						},
					},
				},
				pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when missing slot (quantity_type == QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						CourseItems: []*pb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight:     nil,
								Slot:       &wrapperspb.Int32Value{Value: 5},
							},
						},
					},
				},
				pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String(),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case (quantity_type == QuantityType_QUANTITY_TYPE_COURSE_WEIGHT)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						CourseItems: []*pb.CourseItem{
							{
								CourseId:   constant.CourseID,
								CourseName: constant.CourseName,
								Weight: &wrapperspb.Int32Value{
									Value: 5,
								},
								Slot: nil,
							},
						},
					},
				},
				pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String(),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			quantityType := testCase.Req.([]interface{})[1].(string)
			_, _, _, err := convertMapCourseAndStudentCourse(orderItemData, quantityType)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
			mock.AssertExpectationsForObjects(t)
		})
	}
}

func TestPackageService_verifyPackageCourse(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                             *mockDb.Ext
		PackageRepo                    *mockRepositories.MockPackageRepo
		PackageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		PackageCourseRepo              *mockRepositories.MockPackageCourseRepo
		StudentCourseRepo              *mockRepositories.MockStudentCourseRepo
		StudentPackageByOrderRepo      *mockRepositories.MockStudentPackageByOrderRepo
		OrderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get package courses by package id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting package course have err %v", constant.ErrDefault),
			Req:         utils.PackageInfo{},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when missing mandatory course in bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Missing mandatory course with id %s in bill item", constant.CourseID),
			Req: utils.PackageInfo{
				Package: entities.Package{PackageID: pgtype.Text{String: constant.PackageID}},
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int: 10,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Fail case: Error when course-weight between database and orderItem doesn't equal",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"This course with id %v doesn't equal course-weight between database and orderItem %v %v %v",
				constant.CourseID,
				pgtype.Present,
				10,
				8),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
				Package:      entities.Package{PackageID: pgtype.Text{String: constant.PackageID}},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Weight: &wrapperspb.Int32Value{
							Value: 8,
						},
						Slot: &wrapperspb.Int32Value{
							Value: 8,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int: 10,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name: FailCaseCourseHasSlotGreaterThanMax,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				constant.CourseHasSlotGreaterThanMaxSlot,
				constant.CourseID),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				Package:      entities.Package{PackageID: pgtype.Text{String: constant.PackageID}},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 18,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name: FailCaseCourseHasSlotGreaterThanMax,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				constant.CourseHasSlotGreaterThanMaxSlot,
				constant.CourseID),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				Package:      entities.Package{PackageID: pgtype.Text{String: constant.PackageID}},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 18,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseCourseHasSlotGreaterThanMax,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Some courses is order item not available in package courses"),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				Package:      entities.Package{PackageID: pgtype.Text{String: constant.PackageID}},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 8,
						},
					},
					"constant.CourseID_212": {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 8,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: "constant.CourseID_5",
						},
						MandatoryFlag: pgtype.Bool{
							Bool: false,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: CourseID2,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: false,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseCourseHasSlotGreaterThanMax,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "package is missing max slot"),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				Package:      entities.Package{PackageID: pgtype.Text{String: constant.PackageID}},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 8,
						},
						Weight: &wrapperspb.Int32Value{Value: 10},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when course has slot greater than max slot of the course (quantity_type != QuantityType_QUANTITY_TYPE_COURSE_WEIGHT)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Package with id %s has slot greater than max slot allowed for the package", constant.PackageID),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				Package: entities.Package{
					PackageID: pgtype.Text{String: constant.PackageID},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 8,
						},
						Weight: &wrapperspb.Int32Value{Value: 10},
					},
				},
				Quantity: 20,
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.PackageInfo{
				QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				Package: entities.Package{
					PackageID: pgtype.Text{String: constant.PackageID},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				},
				MapCourseInfo: map[string]*pb.CourseItem{
					constant.CourseID: {
						CourseId:   constant.CourseID,
						CourseName: constant.CourseName,
						Slot: &wrapperspb.Int32Value{
							Value: 8,
						},
						Weight: &wrapperspb.Int32Value{Value: 10},
					},
				},
				Quantity: 10,
			},
			Setup: func(ctx context.Context) {
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			PackageRepo = &mockRepositories.MockPackageRepo{}
			PackageQuantityTypeMappingRepo = &mockRepositories.MockPackageQuantityTypeMappingRepo{}
			PackageCourseRepo = &mockRepositories.MockPackageCourseRepo{}
			StudentCourseRepo = &mockRepositories.MockStudentCourseRepo{}
			StudentPackageByOrderRepo = &mockRepositories.MockStudentPackageByOrderRepo{}
			OrderItemCourseRepo = &mockRepositories.MockOrderItemCourseRepo{}
			s := &PackageService{
				PackageRepo:                    PackageRepo,
				PackageQuantityTypeMappingRepo: PackageQuantityTypeMappingRepo,
				PackageCourseRepo:              PackageCourseRepo,
				OrderItemCourseRepo:            OrderItemCourseRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.PackageInfo)

			err := s.verifyPackageCourse(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, PackageRepo, PackageQuantityTypeMappingRepo, PackageCourseRepo, StudentCourseRepo, StudentPackageByOrderRepo, OrderItemCourseRepo)
		})
	}
}

func TestPackageService_VerifyPackageDataAndUpsertRelateData(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                             *mockDb.Ext
		PackageRepo                    *mockRepositories.MockPackageRepo
		PackageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		PackageCourseRepo              *mockRepositories.MockPackageCourseRepo
		StudentCourseRepo              *mockRepositories.MockStudentCourseRepo
		StudentPackageByOrderRepo      *mockRepositories.MockStudentPackageByOrderRepo
		OrderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get package by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting package have err %v", constant.ErrDefault),
			Req:         utils.OrderItemData{},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get package quantity type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting quantity type have err %v", constant.ErrDefault),
			Req:         utils.OrderItemData{},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
					},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				}, nil)
				PackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when convert map course and student course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.CourseItemMissingSlotField),
			Req: utils.OrderItemData{
				PackageInfo: utils.PackageInfo{
					Package: entities.Package{
						Product: entities.Product{
							AvailableFrom: pgtype.Timestamptz{
								Time:   time.Now(),
								Status: pgtype.Present,
							},
							AvailableUntil: pgtype.Timestamptz{
								Time:   time.Now().Add(1 * time.Minute),
								Status: pgtype.Present,
							},
						},
					},
				},
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
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
						Status: pgtype.Present,
					},
				},
				IsOneTimeProduct: true,
				OrderItem: &pb.OrderItem{
					CourseItems: []*pb.CourseItem{
						{
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Weight:     nil,
							Slot:       nil,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
					},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				}, nil)
				PackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
			},
		},
		{
			Name:        "Fail case: Error when verify package course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting package course have err %v", constant.ErrDefault),
			Req: utils.OrderItemData{
				PackageInfo: utils.PackageInfo{
					Package: entities.Package{
						Product: entities.Product{
							AvailableFrom: pgtype.Timestamptz{
								Time:   time.Now(),
								Status: pgtype.Present,
							},
							AvailableUntil: pgtype.Timestamptz{
								Time:   time.Now().Add(1 * time.Minute),
								Status: pgtype.Present,
							},
						},
					},
				},
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
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
						Status: pgtype.Present,
					},
				},
				IsOneTimeProduct: true,
				OrderItem: &pb.OrderItem{
					CourseItems: []*pb.CourseItem{
						{
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Weight:     nil,
							Slot:       &wrapperspb.Int32Value{Value: 5},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
					},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				}, nil)
				PackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when verify package course",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				constant.CourseHasSlotGreaterThanMaxSlot,
				constant.CourseID),
			Req: utils.OrderItemData{
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
					Package: entities.Package{
						PackageID: pgtype.Text{String: constant.PackageID},
						MaxSlot: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
					MapCourseInfo: map[string]*pb.CourseItem{
						constant.CourseID: {
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Slot: &wrapperspb.Int32Value{
								Value: 18,
							},
							Weight: &wrapperspb.Int32Value{Value: 10},
						},
					},
					Quantity: 20,
				},
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
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
						Status: pgtype.Present,
					},
				},
				IsOneTimeProduct: true,
				OrderItem: &pb.OrderItem{
					CourseItems: []*pb.CourseItem{
						{
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Slot:       &wrapperspb.Int32Value{Value: 18},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
					},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				}, nil)
				PackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when multi create order item course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order item course with error %v", constant.ErrDefault),
			Req: utils.OrderItemData{
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
					Package: entities.Package{
						PackageID: pgtype.Text{String: constant.PackageID},
						MaxSlot: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
					MapCourseInfo: map[string]*pb.CourseItem{
						constant.CourseID: {
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Slot: &wrapperspb.Int32Value{
								Value: 8,
							},
							Weight: &wrapperspb.Int32Value{Value: 10},
						},
					},
					Quantity: 20,
				},
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
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
						Status: pgtype.Present,
					},
				},
				IsOneTimeProduct: true,
				OrderItem: &pb.OrderItem{
					CourseItems: []*pb.CourseItem{
						{
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Slot:       &wrapperspb.Int32Value{Value: 5},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
					},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				}, nil)
				PackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
				OrderItemCourseRepo.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
					Package: entities.Package{
						PackageID: pgtype.Text{String: constant.PackageID},
						MaxSlot: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
					MapCourseInfo: map[string]*pb.CourseItem{
						constant.CourseID: {
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Slot: &wrapperspb.Int32Value{
								Value: 8,
							},
							Weight: &wrapperspb.Int32Value{Value: 10},
						},
					},
					Quantity: 20,
				},
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
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
						Status: pgtype.Present,
					},
				},
				IsOneTimeProduct: true,
				OrderItem: &pb.OrderItem{
					CourseItems: []*pb.CourseItem{
						{
							CourseId:   constant.CourseID,
							CourseName: constant.CourseName,
							Slot:       &wrapperspb.Int32Value{Value: 5},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				PackageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{
						String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(),
					},
					MaxSlot: pgtype.Int4{
						Int:    10,
						Status: pgtype.Present,
					},
				}, nil)
				PackageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT, nil)
				PackageCourseRepo.On("GetByPackageIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return([]entities.PackageCourse{
					{
						PackageID: pgtype.Text{
							String: constant.PackageID,
						},
						CourseID: pgtype.Text{
							String: constant.CourseID,
						},
						MandatoryFlag: pgtype.Bool{
							Bool: true,
						},
						MaxSlotsPerCourse: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
						CourseWeight: pgtype.Int4{
							Int:    10,
							Status: pgtype.Present,
						},
					},
				}, nil)
				OrderItemCourseRepo.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			PackageRepo = &mockRepositories.MockPackageRepo{}
			PackageQuantityTypeMappingRepo = &mockRepositories.MockPackageQuantityTypeMappingRepo{}
			PackageCourseRepo = &mockRepositories.MockPackageCourseRepo{}
			StudentCourseRepo = &mockRepositories.MockStudentCourseRepo{}
			StudentPackageByOrderRepo = &mockRepositories.MockStudentPackageByOrderRepo{}
			OrderItemCourseRepo = &mockRepositories.MockOrderItemCourseRepo{}
			s := &PackageService{
				PackageRepo:                    PackageRepo,
				PackageQuantityTypeMappingRepo: PackageQuantityTypeMappingRepo,
				PackageCourseRepo:              PackageCourseRepo,
				OrderItemCourseRepo:            OrderItemCourseRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			_, err := s.VerifyPackageDataAndUpsertRelateData(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, PackageRepo, PackageQuantityTypeMappingRepo, PackageCourseRepo, StudentCourseRepo, StudentPackageByOrderRepo, OrderItemCourseRepo)
		})
	}
}

func TestPackageService_GetAllPackagesForExport(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		packageRepo *mockRepositories.MockPackageRepo
		productRepo *mockRepositories.MockProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get data for export package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				packageRepo.On("GetPackagesForExport", ctx, db).Return([]*entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get product for export package data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				packageRepo.On("GetPackagesForExport", ctx, db).Return([]*entities.Package{}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get product for export package data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				packageRepo.On("GetPackagesForExport", ctx, db).Return([]*entities.Package{}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get product for export package data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				packageRepo.On("GetPackagesForExport", ctx, db).Return([]*entities.Package{}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when missing product info",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Missing product info with id"),
			Setup: func(ctx context.Context) {
				packageRepo.On("GetPackagesForExport", ctx, db).Return([]*entities.Package{
					{
						PackageID: pgtype.Text{
							String: "package_product_1",
							Status: pgtype.Present,
						},
					},
					{
						PackageID: pgtype.Text{
							String: "package_product_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "package_product_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				packageRepo.On("GetPackagesForExport", ctx, db).Return([]*entities.Package{
					{
						PackageID: pgtype.Text{
							String: "package_product_1",
							Status: pgtype.Present,
						},
					},
					{
						PackageID: pgtype.Text{
							String: "package_product_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
				productRepo.On("GetByIDsForExport", ctx, db, mock.Anything).Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "package_product_2",
							Status: pgtype.Present,
						},
					},
					{
						ProductID: pgtype.Text{
							String: "package_product_1",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			packageRepo = &mockRepositories.MockPackageRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			s := &PackageService{
				PackageRepo: packageRepo,
				ProductRepo: productRepo,
			}
			testCase.Setup(testCase.Ctx)

			_, err := s.GetAllPackagesForExport(testCase.Ctx, db)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, packageRepo, productRepo)
		})
	}
}

func TestPackageService_GetTotalAssociatedPackageWithCourseIDAndPackageID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		packageCourseFeeRepo      *mockRepositories.MockPackageCourseFeeRepo
		packageCourseMaterialRepo *mockRepositories.MockPackageCourseMaterialRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get package course fee",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				packageCourseFeeRepo.On("GetToTalAssociatedByCourseIDAndPackageID", ctx, db, mock.Anything, mock.Anything).Return(int32(0), constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get package course material",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				packageCourseFeeRepo.On("GetToTalAssociatedByCourseIDAndPackageID", ctx, db, mock.Anything, mock.Anything).Return(int32(0), nil)
				packageCourseMaterialRepo.On("GetToTalAssociatedByCourseIDAndPackageID", ctx, db, mock.Anything, mock.Anything).Return(int32(0), constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				packageCourseFeeRepo.On("GetToTalAssociatedByCourseIDAndPackageID", ctx, db, mock.Anything, mock.Anything).Return(int32(0), nil)
				packageCourseMaterialRepo.On("GetToTalAssociatedByCourseIDAndPackageID", ctx, db, mock.Anything, mock.Anything).Return(int32(0), nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			packageCourseFeeRepo = &mockRepositories.MockPackageCourseFeeRepo{}
			packageCourseMaterialRepo = &mockRepositories.MockPackageCourseMaterialRepo{}
			s := &PackageService{
				PackageCourseFeeRepo:      packageCourseFeeRepo,
				PackageCourseMaterialRepo: packageCourseMaterialRepo,
			}
			testCase.Setup(testCase.Ctx)

			_, err := s.GetTotalAssociatedPackageWithCourseIDAndPackageID(testCase.Ctx, db, "1", []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, packageCourseFeeRepo, packageCourseMaterialRepo)
		})
	}
}
