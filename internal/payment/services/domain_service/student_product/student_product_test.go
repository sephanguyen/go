package service

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
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentProductService_CreateStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		packageRepo                  *mockRepositories.MockPackageRepo
		productRepo                  *mockRepositories.MockProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Err while create",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       true,
					IsDisableProRatingFlag: false,
					ProductType:            pb.ProductType_PRODUCT_TYPE_FEE,
					OrderItem: &pb.OrderItem{
						PackageAssociatedId: &wrapperspb.StringValue{
							Value: "package_associated_id",
						},
					},
					BillItems: nil,
				},
			},
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Err while get onetime package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					Order:       entities.Order{},
					StudentInfo: entities.Student{},
					ProductInfo: entities.Product{},
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: nil,
						Package: entities.Package{
							Product:          entities.Product{},
							PackageID:        pgtype.Text{},
							PackageType:      pgtype.Text{},
							MaxSlot:          pgtype.Int4{},
							PackageStartDate: pgtype.Timestamptz{},
							PackageEndDate:   pgtype.Timestamptz{},
							ResourcePath:     pgtype.Text{},
						},
						QuantityType:      0,
						StudentCourseSync: nil,
						Quantity:          0,
					},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       true,
					IsDisableProRatingFlag: false,
					ProductType:            pb.ProductType_PRODUCT_TYPE_PACKAGE,
					OrderItem:              &pb.OrderItem{},
					BillItems:              nil,
				},
			},
			Setup: func(ctx context.Context) {
				packageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
				studentProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Err while create is onetime product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					Order:       entities.Order{},
					StudentInfo: entities.Student{},
					ProductInfo: entities.Product{},
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: nil,
						Package: entities.Package{
							Product:          entities.Product{},
							PackageID:        pgtype.Text{},
							PackageType:      pgtype.Text{},
							MaxSlot:          pgtype.Int4{},
							PackageStartDate: pgtype.Timestamptz{},
							PackageEndDate:   pgtype.Timestamptz{},
							ResourcePath:     pgtype.Text{},
						},
						QuantityType:      0,
						StudentCourseSync: nil,
						Quantity:          0,
					},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       true,
					IsDisableProRatingFlag: false,
					ProductType:            pb.ProductType_PRODUCT_TYPE_PACKAGE,
					OrderItem:              &pb.OrderItem{},
					BillItems:              nil,
				},
			},
			Setup: func(ctx context.Context) {
				packageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Package{}, nil)
				studentProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Err while create not onetime product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					Order:       entities.Order{},
					StudentInfo: entities.Student{},
					ProductInfo: entities.Product{},
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: nil,
						Package: entities.Package{
							Product:          entities.Product{},
							PackageID:        pgtype.Text{},
							PackageType:      pgtype.Text{},
							MaxSlot:          pgtype.Int4{},
							PackageStartDate: pgtype.Timestamptz{},
							PackageEndDate:   pgtype.Timestamptz{},
							ResourcePath:     pgtype.Text{},
						},
						QuantityType:      0,
						StudentCourseSync: nil,
						Quantity:          0,
					},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            pb.ProductType_PRODUCT_TYPE_PACKAGE,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId:    nil,
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(constant.ErrDefault)
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate",
					ctx, db, mock.Anything,
				).Return(entities.BillingSchedulePeriod{}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					Order:       entities.Order{},
					StudentInfo: entities.Student{},
					ProductInfo: entities.Product{},
					PackageInfo: utils.PackageInfo{
						MapCourseInfo: nil,
						Package: entities.Package{
							Product:          entities.Product{},
							PackageID:        pgtype.Text{},
							PackageType:      pgtype.Text{},
							MaxSlot:          pgtype.Int4{},
							PackageStartDate: pgtype.Timestamptz{},
							PackageEndDate:   pgtype.Timestamptz{},
							ResourcePath:     pgtype.Text{},
						},
						QuantityType:      0,
						StudentCourseSync: nil,
						Quantity:          0,
					},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            pb.ProductType_PRODUCT_TYPE_PACKAGE,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId:    nil,
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: []interface{}{entities.StudentProduct{}, nil},
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(nil)
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate",
					ctx, db, mock.Anything,
				).Return(entities.BillingSchedulePeriod{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			packageRepo = &mockRepositories.MockPackageRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
				PackageRepo:                  packageRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, err := s.CreateStudentProduct(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_CreateAssociatedStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
	)

	var associatedProducts []*pb.ProductAssociation
	var mapKeyWithOrderItemData map[string]utils.OrderItemData

	var associatedProducts2 []*pb.ProductAssociation

	associatedProducts2 = append(associatedProducts2, &pb.ProductAssociation{
		PackageId:   "",
		CourseId:    "",
		ProductId:   "",
		ProductType: 0,
	})

	var associatedProducts3 []*pb.ProductAssociation

	associatedProducts3 = append(associatedProducts3, &pb.ProductAssociation{
		PackageId:   "1",
		CourseId:    "1",
		ProductId:   "1",
		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
	})

	testcases := []utils.TestCase{
		{
			Name: "associatedProducts is 0",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				associatedProducts,
				mapKeyWithOrderItemData,
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "getting order item data from associated product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				associatedProducts2,
				mapKeyWithOrderItemData,
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.FailedPrecondition, "getting order item data from associated product"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Failed while create",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				associatedProducts3,
				map[string]utils.OrderItemData{
					"1_1": {
						Order: entities.Order{
							OrderID:                 pgtype.Text{},
							StudentID:               pgtype.Text{},
							StudentFullName:         pgtype.Text{},
							LocationID:              pgtype.Text{},
							OrderSequenceNumber:     pgtype.Int4{},
							OrderComment:            pgtype.Text{},
							OrderStatus:             pgtype.Text{},
							OrderType:               pgtype.Text{},
							UpdatedAt:               pgtype.Timestamptz{},
							CreatedAt:               pgtype.Timestamptz{},
							ResourcePath:            pgtype.Text{},
							IsReviewed:              pgtype.Bool{},
							WithdrawalEffectiveDate: pgtype.Timestamptz{},
						},
						StudentInfo: entities.Student{
							StudentID:        pgtype.Text{},
							CurrentGrade:     pgtype.Int2{},
							EnrollmentStatus: pgtype.Text{},
							UpdatedAt:        pgtype.Timestamptz{},
							CreatedAt:        pgtype.Timestamptz{},
							DeletedAt:        pgtype.Timestamptz{},
							GradeID:          pgtype.Text{},
							ResourcePath:     pgtype.Text{},
						},
						ProductInfo: entities.Product{
							ProductID: pgtype.Text{
								String: "1",
								Status: 0,
							},
							Name: pgtype.Text{},
							ProductType: pgtype.Text{
								String: "PRODUCT_TYPE_FEE",
								Status: 0,
							},
							TaxID:                pgtype.Text{},
							AvailableFrom:        pgtype.Timestamptz{},
							AvailableUntil:       pgtype.Timestamptz{},
							CustomBillingPeriod:  pgtype.Timestamptz{},
							BillingScheduleID:    pgtype.Text{},
							DisableProRatingFlag: pgtype.Bool{},
							Remarks:              pgtype.Text{},
							IsArchived:           pgtype.Bool{},
							IsUnique:             pgtype.Bool{},
							UpdatedAt:            pgtype.Timestamptz{},
							CreatedAt:            pgtype.Timestamptz{},
							ResourcePath:         pgtype.Text{},
						},
						PackageInfo: utils.PackageInfo{
							MapCourseInfo:     nil,
							Package:           entities.Package{},
							QuantityType:      0,
							StudentCourseSync: nil,
							Quantity:          0,
						},
						StudentProduct: entities.StudentProduct{
							StudentProductID: pgtype.Text{},
							StudentID:        pgtype.Text{},
							ProductID: pgtype.Text{
								String: "1",
								Status: 0,
							},
							UpcomingBillingDate:         pgtype.Timestamptz{},
							StartDate:                   pgtype.Timestamptz{},
							EndDate:                     pgtype.Timestamptz{},
							ProductStatus:               pgtype.Text{},
							ApprovalStatus:              pgtype.Text{},
							UpdatedAt:                   pgtype.Timestamptz{},
							CreatedAt:                   pgtype.Timestamptz{},
							DeletedAt:                   pgtype.Timestamptz{},
							ResourcePath:                pgtype.Text{},
							LocationID:                  pgtype.Text{},
							UpdatedFromStudentProductID: pgtype.Text{},
							UpdatedToStudentProductID:   pgtype.Text{},
							StudentProductLabel:         pgtype.Text{},
							IsUnique:                    pgtype.Bool{},
							RootStudentProductID:        pgtype.Text{},
						},
						StudentName:            "",
						LocationName:           "",
						IsOneTimeProduct:       false,
						IsDisableProRatingFlag: false,
						ProductType:            pb.ProductType_PRODUCT_TYPE_FEE,
						OrderItem: &pb.OrderItem{
							ProductId:           "1",
							DiscountId:          nil,
							StartDate:           nil,
							CourseItems:         nil,
							ProductAssociations: nil,
							StudentProductId:    nil,
							EffectiveDate:       nil,
							CancellationDate:    nil,
							PackageAssociatedId: nil,
						},
						BillItems: nil,
					},
					"1": {
						Order:                  entities.Order{},
						StudentInfo:            entities.Student{},
						ProductInfo:            entities.Product{},
						PackageInfo:            utils.PackageInfo{},
						StudentProduct:         entities.StudentProduct{},
						StudentName:            "",
						LocationName:           "",
						IsOneTimeProduct:       false,
						IsDisableProRatingFlag: false,
						ProductType:            0,
						OrderItem:              nil,
						BillItems:              nil,
					},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentAssociatedProductRepo.On("GetMapAssociatedProducts", ctx, db, mock.Anything).Return(map[string]string{}, nil)
				studentAssociatedProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				associatedProducts3,
				map[string]utils.OrderItemData{
					"1_1": {
						Order: entities.Order{
							OrderID:                 pgtype.Text{},
							StudentID:               pgtype.Text{},
							StudentFullName:         pgtype.Text{},
							LocationID:              pgtype.Text{},
							OrderSequenceNumber:     pgtype.Int4{},
							OrderComment:            pgtype.Text{},
							OrderStatus:             pgtype.Text{},
							OrderType:               pgtype.Text{},
							UpdatedAt:               pgtype.Timestamptz{},
							CreatedAt:               pgtype.Timestamptz{},
							ResourcePath:            pgtype.Text{},
							IsReviewed:              pgtype.Bool{},
							WithdrawalEffectiveDate: pgtype.Timestamptz{},
						},
						StudentInfo: entities.Student{
							StudentID:        pgtype.Text{},
							CurrentGrade:     pgtype.Int2{},
							EnrollmentStatus: pgtype.Text{},
							UpdatedAt:        pgtype.Timestamptz{},
							CreatedAt:        pgtype.Timestamptz{},
							DeletedAt:        pgtype.Timestamptz{},
							GradeID:          pgtype.Text{},
							ResourcePath:     pgtype.Text{},
						},
						ProductInfo: entities.Product{
							ProductID: pgtype.Text{
								String: "1",
								Status: 0,
							},
							Name: pgtype.Text{},
							ProductType: pgtype.Text{
								String: "PRODUCT_TYPE_FEE",
								Status: 0,
							},
							TaxID:                pgtype.Text{},
							AvailableFrom:        pgtype.Timestamptz{},
							AvailableUntil:       pgtype.Timestamptz{},
							CustomBillingPeriod:  pgtype.Timestamptz{},
							BillingScheduleID:    pgtype.Text{},
							DisableProRatingFlag: pgtype.Bool{},
							Remarks:              pgtype.Text{},
							IsArchived:           pgtype.Bool{},
							IsUnique:             pgtype.Bool{},
							UpdatedAt:            pgtype.Timestamptz{},
							CreatedAt:            pgtype.Timestamptz{},
							ResourcePath:         pgtype.Text{},
						},
						PackageInfo: utils.PackageInfo{
							MapCourseInfo:     nil,
							Package:           entities.Package{},
							QuantityType:      0,
							StudentCourseSync: nil,
							Quantity:          0,
						},
						StudentProduct: entities.StudentProduct{
							StudentProductID: pgtype.Text{},
							StudentID:        pgtype.Text{},
							ProductID: pgtype.Text{
								String: "1",
								Status: 0,
							},
							UpcomingBillingDate:         pgtype.Timestamptz{},
							StartDate:                   pgtype.Timestamptz{},
							EndDate:                     pgtype.Timestamptz{},
							ProductStatus:               pgtype.Text{},
							ApprovalStatus:              pgtype.Text{},
							UpdatedAt:                   pgtype.Timestamptz{},
							CreatedAt:                   pgtype.Timestamptz{},
							DeletedAt:                   pgtype.Timestamptz{},
							ResourcePath:                pgtype.Text{},
							LocationID:                  pgtype.Text{},
							UpdatedFromStudentProductID: pgtype.Text{},
							UpdatedToStudentProductID:   pgtype.Text{},
							StudentProductLabel:         pgtype.Text{},
							IsUnique:                    pgtype.Bool{},
							RootStudentProductID:        pgtype.Text{},
						},
						StudentName:            "",
						LocationName:           "",
						IsOneTimeProduct:       false,
						IsDisableProRatingFlag: false,
						ProductType:            pb.ProductType_PRODUCT_TYPE_FEE,
						OrderItem: &pb.OrderItem{
							ProductId:           "1",
							DiscountId:          nil,
							StartDate:           nil,
							CourseItems:         nil,
							ProductAssociations: nil,
							StudentProductId:    nil,
							EffectiveDate:       nil,
							CancellationDate:    nil,
							PackageAssociatedId: nil,
						},
						BillItems: nil,
					},
					"1": {
						Order:                  entities.Order{},
						StudentInfo:            entities.Student{},
						ProductInfo:            entities.Product{},
						PackageInfo:            utils.PackageInfo{},
						StudentProduct:         entities.StudentProduct{},
						StudentName:            "",
						LocationName:           "",
						IsOneTimeProduct:       false,
						IsDisableProRatingFlag: false,
						ProductType:            0,
						OrderItem:              nil,
						BillItems:              nil,
					},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentAssociatedProductRepo.On("GetMapAssociatedProducts", ctx, db, mock.Anything).Return(map[string]string{}, nil)
				studentAssociatedProductRepo.On("Create",
					ctx, db, mock.Anything,
				).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			associatedProducts := testCase.Req.([]interface{})[0].([]*pb.ProductAssociation)
			mapKeyWithOrderItemData := testCase.Req.([]interface{})[1].(map[string]utils.OrderItemData)
			err := s.CreateAssociatedStudentProduct(testCase.Ctx, db, associatedProducts, mapKeyWithOrderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_MutationStudentProductForUpdateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "happy case one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       true,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						EffectiveDate:               nil,
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: entities.StudentProduct{
				StudentProductID:            pgtype.Text{},
				StudentID:                   pgtype.Text{},
				ProductID:                   pgtype.Text{},
				UpcomingBillingDate:         pgtype.Timestamptz{},
				StartDate:                   pgtype.Timestamptz{},
				EndDate:                     pgtype.Timestamptz{},
				ProductStatus:               pgtype.Text{},
				ApprovalStatus:              pgtype.Text{},
				UpdatedAt:                   pgtype.Timestamptz{},
				CreatedAt:                   pgtype.Timestamptz{},
				DeletedAt:                   pgtype.Timestamptz{},
				ResourcePath:                pgtype.Text{},
				LocationID:                  pgtype.Text{},
				UpdatedFromStudentProductID: pgtype.Text{},
				UpdatedToStudentProductID:   pgtype.Text{},
				StudentProductLabel: pgtype.Text{
					String: "UPDATED",
					Status: 2,
				},
				IsUnique: pgtype.Bool{},
				VersionNumber: pgtype.Int4{
					Int: 1,
				},
				RootStudentProductID: pgtype.Text{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx, db, mock.Anything,
				).Return(entities.StudentProduct{
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
				}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:       entities.Order{},
					StudentInfo: entities.Student{},
					ProductInfo: entities.Product{},
					PackageInfo: utils.PackageInfo{},
					StudentProduct: entities.StudentProduct{
						StudentProductID:            pgtype.Text{},
						StudentID:                   pgtype.Text{},
						ProductID:                   pgtype.Text{},
						UpcomingBillingDate:         pgtype.Timestamptz{},
						StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
						EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
						ProductStatus:               pgtype.Text{},
						ApprovalStatus:              pgtype.Text{},
						UpdatedAt:                   pgtype.Timestamptz{},
						CreatedAt:                   pgtype.Timestamptz{},
						DeletedAt:                   pgtype.Timestamptz{},
						ResourcePath:                pgtype.Text{},
						LocationID:                  pgtype.Text{},
						UpdatedFromStudentProductID: pgtype.Text{},
						UpdatedToStudentProductID:   pgtype.Text{},
						StudentProductLabel:         pgtype.Text{},
						IsUnique:                    pgtype.Bool{},
						RootStudentProductID:        pgtype.Text{},
					},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						EffectiveDate:               timestamppb.New(time.Now().AddDate(0, 0, 1)),
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: entities.StudentProduct{
				StudentProductID:            pgtype.Text{},
				StudentID:                   pgtype.Text{},
				ProductID:                   pgtype.Text{},
				UpcomingBillingDate:         pgtype.Timestamptz{},
				StartDate:                   pgtype.Timestamptz{},
				EndDate:                     pgtype.Timestamptz{},
				ProductStatus:               pgtype.Text{},
				ApprovalStatus:              pgtype.Text{},
				UpdatedAt:                   pgtype.Timestamptz{},
				CreatedAt:                   pgtype.Timestamptz{},
				DeletedAt:                   pgtype.Timestamptz{},
				ResourcePath:                pgtype.Text{},
				LocationID:                  pgtype.Text{},
				UpdatedFromStudentProductID: pgtype.Text{},
				UpdatedToStudentProductID:   pgtype.Text{},
				StudentProductLabel: pgtype.Text{
					String: "UPDATED",
					Status: 2,
				},
				IsUnique: pgtype.Bool{},
				VersionNumber: pgtype.Int4{
					Int: 1,
				},
				RootStudentProductID: pgtype.Text{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx, db, mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
				studentProductRepo.On("Create",
					ctx,
					db,
					mock.Anything,
				).Return(nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)

			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, _, err := s.MutationStudentProductForUpdateOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_MutationStudentProductForCancelOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Happy case one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       true,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						EffectiveDate:       timestamppb.New(time.Now()),
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					RootStudentProductID:        pgtype.Text{},
				}, nil)
				studentProductRepo.On("Create",
					ctx,
					db,
					mock.Anything,
				).Return(nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, _, err := s.MutationStudentProductForCancelOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_MutationStudentProductForGraduateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Fail Case: One time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: true,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.Internal, "updating student product label and status for graduate order is unimplemented"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail Case: Error on student product repo",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:           timestamppb.New(time.Now()),
						EndDate:             timestamppb.New(time.Now().AddDate(0, 1, 0)),
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail Case: Out of version",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:           nil,
						EndDate:             nil,
						EffectiveDate:       timestamppb.New(time.Now()),
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Error(codes.FailedPrecondition, "optimistic_locking_entity_version_mismatched"),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 2,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
			},
		},
		{
			Name: "Fail Case: Error on updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:                   nil,
						EndDate:                     nil,
						EffectiveDate:               timestamppb.New(time.Now()),
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						EffectiveDate:               timestamppb.New(time.Now()),
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)

				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, err := s.MutationStudentProductForGraduateOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_MutationStudentProductForWithdrawalOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Fail Case: One time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: true,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.Internal, "updating student product label and status for withdraw order is unimplemented"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail Case: Error on student product repo",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:           timestamppb.New(time.Now()),
						EndDate:             timestamppb.New(time.Now().AddDate(0, 1, 0)),
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					RootStudentProductID:        pgtype.Text{},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail Case: Error on updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:                   nil,
						EndDate:                     nil,
						EffectiveDate:               timestamppb.New(time.Now()),
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail Case: Out of version",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:           nil,
						EndDate:             nil,
						EffectiveDate:       timestamppb.New(time.Now()),
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Error(codes.FailedPrecondition, "optimistic_locking_entity_version_mismatched"),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 2,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						StartDate:           nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						EffectiveDate:               timestamppb.New(time.Now()),
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, err := s.MutationStudentProductForWithdrawalOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_MutationStudentProductForLOAOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Fail Case: One time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: true,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.Internal, "updating student product label and status for LOA order is unimplemented"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail Case: Error on student product repo",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:           timestamppb.New(time.Now()),
						EndDate:             timestamppb.New(time.Now().AddDate(0, 1, 0)),
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					RootStudentProductID:        pgtype.Text{},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail Case: Out of version",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:           timestamppb.New(time.Now()),
						EndDate:             timestamppb.New(time.Now().AddDate(0, 1, 0)),
						EffectiveDate:       nil,
						CancellationDate:    nil,
						PackageAssociatedId: nil,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Error(codes.FailedPrecondition, "optimistic_locking_entity_version_mismatched"),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 2,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
			},
		},
		{
			Name: "Fail Case: Error on updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:                   timestamppb.New(time.Now()),
						EndDate:                     timestamppb.New(time.Now().AddDate(0, 1, 0)),
						EffectiveDate:               nil,
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					Order:                  entities.Order{},
					StudentInfo:            entities.Student{},
					ProductInfo:            entities.Product{},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            "",
					LocationName:           "",
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            0,
					OrderItem: &pb.OrderItem{
						ProductId:           "",
						DiscountId:          nil,
						CourseItems:         nil,
						ProductAssociations: nil,
						StudentProductId: &wrapperspb.StringValue{
							Value: "1",
						},
						StartDate:                   timestamppb.New(time.Now()),
						EndDate:                     timestamppb.New(time.Now().AddDate(0, 1, 0)),
						EffectiveDate:               nil,
						CancellationDate:            nil,
						PackageAssociatedId:         nil,
						StudentProductVersionNumber: 1,
					},
					BillItems: nil,
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID",
					ctx,
					db,
					mock.Anything,
				).Return(entities.StudentProduct{
					StudentProductID:            pgtype.Text{},
					StudentID:                   pgtype.Text{},
					ProductID:                   pgtype.Text{},
					UpcomingBillingDate:         pgtype.Timestamptz{},
					StartDate:                   pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
					EndDate:                     pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					ProductStatus:               pgtype.Text{},
					ApprovalStatus:              pgtype.Text{},
					UpdatedAt:                   pgtype.Timestamptz{},
					CreatedAt:                   pgtype.Timestamptz{},
					DeletedAt:                   pgtype.Timestamptz{},
					ResourcePath:                pgtype.Text{},
					LocationID:                  pgtype.Text{},
					UpdatedFromStudentProductID: pgtype.Text{},
					UpdatedToStudentProductID:   pgtype.Text{},
					StudentProductLabel:         pgtype.Text{},
					IsUnique:                    pgtype.Bool{},
					VersionNumber: pgtype.Int4{
						Int: 1,
					},
					RootStudentProductID: pgtype.Text{},
				}, nil)
				studentProductRepo.On("UpdateWithVersionNumber",
					ctx,
					db,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, err := s.MutationStudentProductForLOAOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestGetUniqueProductsByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by student id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetUniqueProductsByStudentID", ctx, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetUniqueProductsByStudentID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{
					{
						ProductID: pgtype.Text{String: "1", Status: pgtype.Present},
					},
					{
						ProductID: pgtype.Text{String: "1", Status: pgtype.Present},
					},
					{
						ProductID:     pgtype.Text{String: "2", Status: pgtype.Present},
						EndDate:       pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
						ProductStatus: pgtype.Text{String: pb.StudentProductStatus_CANCELLED.String()},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo: studentProductRepo,
			}
			_, err := s.GetUniqueProductsByStudentID(testCase.Ctx, db, testCase.Req.(string))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestGetUniqueProductsByStudentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by student id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         []string{constant.StudentID},
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetUniqueProductsByStudentIDs", ctx, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         []string{constant.StudentID},
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetUniqueProductsByStudentIDs", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{
					{
						ProductID: pgtype.Text{String: "Product_1", Status: pgtype.Present},
						StudentID: pgtype.Text{String: "Student_1", Status: pgtype.Present},
					},
					{
						ProductID: pgtype.Text{String: "Product_1", Status: pgtype.Present},
						StudentID: pgtype.Text{String: "Student_1", Status: pgtype.Present},
					},
					{
						ProductID:     pgtype.Text{String: "Product_2", Status: pgtype.Present},
						EndDate:       pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
						ProductStatus: pgtype.Text{String: pb.StudentProductStatus_CANCELLED.String()},
						StudentID:     pgtype.Text{String: "Student_1", Status: pgtype.Present},
					},
					{
						ProductID:     pgtype.Text{String: "Product_2", Status: pgtype.Present},
						EndDate:       pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
						ProductStatus: pgtype.Text{String: pb.StudentProductStatus_CANCELLED.String()},
						StudentID:     pgtype.Text{String: "Student_2", Status: pgtype.Present},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo: studentProductRepo,
			}
			_, err := s.GetUniqueProductsByStudentIDs(testCase.Ctx, db, testCase.Req.([]string))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestEndDateOfUniqueRecurringProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		studentProductRepo     *mockRepositories.MockStudentProductRepo
		productRepo            *mockRepositories.MockProductRepo
		billSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: when disable pro rating is false",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					DisableProRatingFlag: pgtype.Bool{Bool: false, Status: pgtype.Present},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get billingSchedulePeriod",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					DisableProRatingFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				}, nil)
				billSchedulePeriodRepo.On("GetPeriodByScheduleIDAndEndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{}, constant.ErrDefault,
				)
			},
		},
		{
			Name:        "Happy case: when disable pro rating is true",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					DisableProRatingFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				}, nil)
				billSchedulePeriodRepo.On("GetPeriodByScheduleIDAndEndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						EndDate: pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
					}, nil,
				)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			billSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				BillingSchedulePeriodRepo: billSchedulePeriodRepo,
				ProductRepo:               productRepo,
			}
			_, err := s.EndDateOfUniqueRecurringProduct(testCase.Ctx, db, "1", time.Now())

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestGetStudentProductByStudentProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		studentProductRepo     *mockRepositories.MockStudentProductRepo
		productRepo            *mockRepositories.MockProductRepo
		billSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			billSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				BillingSchedulePeriodRepo: billSchedulePeriodRepo,
				ProductRepo:               productRepo,
			}
			_, err := s.GetStudentProductByStudentProductID(testCase.Ctx, db, "1")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestGetStudentProductByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		studentProductRepo     *mockRepositories.MockStudentProductRepo
		productRepo            *mockRepositories.MockProductRepo
		billSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when count student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("CountStudentProductIDsByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(10, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get by student product id and pagination",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("CountStudentProductIDsByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				studentProductRepo.On("GetByStudentIDAndLocationIDsWithPaging", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: when get by one time student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("CountStudentProductIDsByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				studentProductRepo.On("GetByStudentIDAndLocationIDsWithPaging", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{{}}, nil)
			},
		},
		{
			Name:        "Fail case: when get by recurring student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("CountStudentProductIDsByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				studentProductRepo.On("GetByStudentIDAndLocationIDsWithPaging", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{{
					StartDate:        pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
					EndDate:          pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
					StudentProductID: pgtype.Text{String: "1", Status: pgtype.Present},
				}}, nil)
				studentProductRepo.On("GetStudentProductIDsByRootStudentProductID", ctx, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: for recurring student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("CountStudentProductIDsByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(10, nil)
				studentProductRepo.On("GetByStudentIDAndLocationIDsWithPaging", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{{
					StartDate:        pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
					EndDate:          pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present},
					StudentProductID: pgtype.Text{String: "1", Status: pgtype.Present},
				}}, nil)
				studentProductRepo.On("GetStudentProductIDsByRootStudentProductID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{{
					StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
				},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			billSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				BillingSchedulePeriodRepo: billSchedulePeriodRepo,
				ProductRepo:               productRepo,
			}
			_, _, _, err := s.GetStudentProductByStudentIDAndLocationIDs(testCase.Ctx, db, "1", []string{"location_1", "location_2"}, int64(1), int64(10))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestGetStudentProductWithRootStudentProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		studentProductRepo     *mockRepositories.MockStudentProductRepo
		productRepo            *mockRepositories.MockProductRepo
		billSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when count student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductIDsByRootStudentProductID", ctx, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: when future product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductIDsByRootStudentProductID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{{
					StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
				},
				}, nil)
			},
		},
		{
			Name:        "Happy case: when present product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductIDsByRootStudentProductID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{
					{
						StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -2, 0), Status: pgtype.Present},
						EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
					},
					{
						StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, -1), Status: pgtype.Present},
						EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0), Status: pgtype.Present},
					},
				}, nil)
			},
		},
		{
			Name:        "Happy case: when past product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductIDsByRootStudentProductID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{
					{
						StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -3, 0), Status: pgtype.Present},
						EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, -2, 0), Status: pgtype.Present},
					},
					{
						StartDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, -1), Status: pgtype.Present},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			billSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				BillingSchedulePeriodRepo: billSchedulePeriodRepo,
				ProductRepo:               productRepo,
			}
			_, err := s.GetStudentProductWithRootStudentProductID(testCase.Ctx, db, "1")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestStudentProductService_GetStudentProductsByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
	)
	expectedResp := []entities.StudentProduct{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
			},
			StudentID: pgtype.Text{
				String: constant.StudentID,
			},
			ProductID: pgtype.Text{
				String: constant.ProductID,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "constant.StudentProductID",
			},
			StudentID: pgtype.Text{
				String: constant.StudentID,
			},
			ProductID: pgtype.Text{
				String: constant.ProductID,
			},
		},
	}
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when getting student product by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:         "Success case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: expectedResp,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return(expectedResp, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo: studentProductRepo,
			}
			resp, err := s.GetStudentProductsByStudentProductIDs(testCase.Ctx, db, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedStudentProducts := testCase.ExpectedResp.([]entities.StudentProduct)
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, len(expectedStudentProducts), len(resp))
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestGetStudentAssociatedProductByStudentProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                       *mockDb.Ext
		studentProductRepo       *mockRepositories.MockStudentProductRepo
		studentAssociatedProduct *mockRepositories.MockStudentAssociatedProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: when count associated product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentAssociatedProduct.On("CountAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything).Return(0, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: when get student associated product by student product ID with paging",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentAssociatedProduct.On("CountAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything).Return(2, nil)
				studentAssociatedProduct.On("GetAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: when get student product of product associated by student product ID",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentAssociatedProduct.On("CountAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything).Return(2, nil)
				studentAssociatedProduct.On("GetAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetStudentProductAssociatedByStudentProductID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentAssociatedProduct.On("CountAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything).Return(2, nil)
				studentAssociatedProduct.On("GetAssociatedProductIDsByStudentProductID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetStudentProductAssociatedByStudentProductID", ctx, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			studentAssociatedProduct = new(mockRepositories.MockStudentAssociatedProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				StudentAssociatedProductRepo: studentAssociatedProduct,
			}
			_, _, _, err := s.GetStudentAssociatedProductByStudentProductID(testCase.Ctx, db, "1", 0, 2)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, studentAssociatedProduct)
		})
	}
}

func TestStudentProductService_VoidStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                        *mockDb.Ext
		studentProductRepo        *mockRepositories.MockStudentProductRepo
		productRepo               *mockRepositories.MockProductRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Failed case: Error when getting student product for update by studenr product id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				pb.OrderType_ORDER_TYPE_NEW.String(),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when getting product by product id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				pb.OrderType_ORDER_TYPE_NEW.String(),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when student product label is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				pb.OrderType_ORDER_TYPE_NEW.String(),
			},
			ExpectedErr: status.Errorf(codes.Internal, "error when cannot void if any of the products have UPDATED/UPDATE_SCHEDULED/WITHDRAWAL_SCHEDULED/GRADUATION_SCHEDULED label"),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					StudentProductLabel: pgtype.Text{
						String: pb.StudentProductLabel_UPDATE_SCHEDULED.String(),
						Status: pgtype.Present,
					},
				}, nil)
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
			},
		},
		{
			Name: "Failed case: Error when student product label is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				pb.OrderType_ORDER_TYPE_NEW.String(),
			},
			ExpectedErr: status.Errorf(codes.Internal, "error when cannot void if any of the products have UPDATED/UPDATE_SCHEDULED/WITHDRAWAL_SCHEDULED/GRADUATION_SCHEDULED label"),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					StudentProductLabel: pgtype.Text{
						String: pb.StudentProductLabel_UPDATE_SCHEDULED.String(),
						Status: pgtype.Present,
					},
				}, nil)
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
			},
		},
		{
			Name: "Failed case: Error when voiding invalid order type",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				"invalid order type",
			},
			ExpectedErr: status.Errorf(codes.Internal, fmt.Sprintf("voiding %s order is unimplemented", "invalid order type")),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
			},
		},
		{
			Name: "Happy case (order_type = NEW)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				pb.OrderType_ORDER_TYPE_NEW.String(),
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					StudentProductLabel: pgtype.Text{
						String: pb.StudentProductLabel_CREATED.String(),
						Status: pgtype.Present,
					},
				}, nil)
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Happy case (order_type = UPDATE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				pb.OrderType_ORDER_TYPE_UPDATE.String(),
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetStudentProductForUpdateByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					StudentProductLabel: pgtype.Text{
						String: pb.StudentProductLabel_CREATED.String(),
						Status: pgtype.Present,
					},
				}, nil)
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					EndDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
				}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				ProductRepo:               productRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			studentProductIDReq := testCase.Req.([]interface{})[0].(string)
			orderTypeReq := testCase.Req.([]interface{})[1].(string)
			_, _, _, err := s.VoidStudentProduct(testCase.Ctx, db, studentProductIDReq, orderTypeReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, productRepo)
		})
	}
}

func TestStudentProductService_voidStudentProductForWithdrawal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                        *mockDb.Ext
		studentProductRepo        *mockRepositories.MockStudentProductRepo
		productRepo               *mockRepositories.MockProductRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Failed case: Error when getting latest period from product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: status.Errorf(codes.Internal, "can't get latest period from product id: %v", constant.ProductID),
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, status.Errorf(codes.Internal, "can't get latest period from product id: %v", constant.ProductID))
			},
		},
		{
			Name: "Failed case: Error when updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				ProductRepo:               productRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			studentProductReq := testCase.Req.([]interface{})[0].(entities.StudentProduct)
			productReq := testCase.Req.([]interface{})[1].(entities.Product)
			_, err := s.voidStudentProductForWithdrawal(testCase.Ctx, db, studentProductReq, productReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, productRepo)
		})
	}
}

func TestStudentProductService_voidStudentProductForCreate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                        *mockDb.Ext
		studentProductRepo        *mockRepositories.MockStudentProductRepo
		productRepo               *mockRepositories.MockProductRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Failed case: Error when updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				ProductRepo:               productRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			studentProductReq := testCase.Req.([]interface{})[0].(entities.StudentProduct)
			_, err := s.voidStudentProductForCreate(testCase.Ctx, db, studentProductReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, productRepo)
		})
	}
}

func TestStudentProductService_voidStudentProductForUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                        *mockDb.Ext
		studentProductRepo        *mockRepositories.MockStudentProductRepo
		productRepo               *mockRepositories.MockProductRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Failed case: Error when updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when getting student product by student product id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when updating previous student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{
					EndDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
					UpdatedFromStudentProductID: pgtype.Text{
						String: constant.StudentProductID,
						Status: 2,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					EndDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
				}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{
					EndDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
					UpdatedFromStudentProductID: pgtype.Text{
						String: constant.StudentProductID,
						Status: 2,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{
					EndDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
				}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				ProductRepo:               productRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			studentProductReq := testCase.Req.([]interface{})[0].(entities.StudentProduct)
			_, err := s.voidStudentProductForUpdate(testCase.Ctx, db, studentProductReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, productRepo)
		})
	}
}

func TestStudentProductService_voidStudentProductForCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                        *mockDb.Ext
		studentProductRepo        *mockRepositories.MockStudentProductRepo
		productRepo               *mockRepositories.MockProductRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Failed case: Error when getting latest period from product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: status.Errorf(codes.Internal, "can't get latest period from product id: %v", constant.ProductID),
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, status.Errorf(codes.Internal, "can't get latest period from product id: %v", constant.ProductID))
			},
		},
		{
			Name: "Failed case: Error when updating student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when get student product by student product id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{
					UpdatedFromStudentProductID: pgtype.Text{
						String: constant.StudentProductID,
						Status: pgtype.Present,
					},
				},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when updating previous student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{
					UpdatedFromStudentProductID: pgtype.Text{
						String: constant.StudentProductID,
						Status: pgtype.Present,
					},
				},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.StudentProduct{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.StudentProduct{
					UpdatedFromStudentProductID: pgtype.Text{
						String: constant.StudentProductID,
						Status: pgtype.Present,
					},
				},
				entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetLatestPeriodByScheduleIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentProductRepo.On("GetStudentProductByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.StudentProduct{}, nil)
				studentProductRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:        studentProductRepo,
				ProductRepo:               productRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			studentProductReq := testCase.Req.([]interface{})[0].(entities.StudentProduct)
			productReq := testCase.Req.([]interface{})[1].(entities.Product)
			_, err := s.voidStudentProductForCancel(testCase.Ctx, db, productReq, studentProductReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, productRepo)
		})
	}
}

func TestValidateProductSettingForCreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext
	)

	randomID, _ := uuid.NewUUID()
	testcases := []utils.TestCase{
		{
			Name: "Happy case: enrollment_required and status enrolled",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Student{StudentID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Order{LocationID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.ProductSetting{
					ProductID:            pgtype.Text{String: randomID.String(), Status: pgtype.Present},
					IsPausable:           pgtype.Bool{Bool: true, Status: pgtype.Present},
					IsEnrollmentRequired: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
				true,
			},
			ExpectedErr: nil,
		},
		{
			Name: "Happy case: not enrollment_required and status enrolled",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Student{StudentID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Order{LocationID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.ProductSetting{
					ProductID:            pgtype.Text{String: randomID.String(), Status: pgtype.Present},
					IsPausable:           pgtype.Bool{Bool: true, Status: pgtype.Present},
					IsEnrollmentRequired: pgtype.Bool{Bool: false, Status: pgtype.Present},
				},
				true,
			},
			ExpectedErr: nil,
		},
		{
			Name: "Happy case: not enrollment_required and status not enrolled",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Student{StudentID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Order{LocationID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.ProductSetting{
					ProductID:            pgtype.Text{String: randomID.String(), Status: pgtype.Present},
					IsPausable:           pgtype.Bool{Bool: true, Status: pgtype.Present},
					IsEnrollmentRequired: pgtype.Bool{Bool: false, Status: pgtype.Present},
				},
				false,
			},
			ExpectedErr: nil,
		},
		{
			Name: "Fail case: enrollment_required and status not enrolled",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Student{StudentID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.Order{LocationID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.ProductSetting{
					ProductID:            pgtype.Text{String: randomID.String(), Status: pgtype.Present},
					IsPausable:           pgtype.Bool{Bool: true, Status: pgtype.Present},
					IsEnrollmentRequired: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
				false,
			},
			ExpectedErr: status.Errorf(codes.Internal, "product %v has enrollment required tag but student %v not enrolled in location %v", randomID, randomID, randomID),
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			s := &StudentProductService{}

			orderItemData := utils.OrderItemData{
				ProductInfo:          testCase.Req.([]interface{})[0].(entities.Product),
				StudentInfo:          testCase.Req.([]interface{})[1].(entities.Student),
				Order:                testCase.Req.([]interface{})[2].(entities.Order),
				ProductSetting:       testCase.Req.([]interface{})[3].(entities.ProductSetting),
				IsEnrolledInLocation: testCase.Req.([]interface{})[4].(bool),
			}

			err := s.ValidateProductSettingForCreateOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db)
		})
	}
}

func TestValidateProductSettingForLOAOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext
	)

	randomID, _ := uuid.NewUUID()
	testcases := []utils.TestCase{
		{
			Name: "Happy case: product is_pausable true",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.ProductSetting{
					ProductID:            pgtype.Text{String: randomID.String(), Status: pgtype.Present},
					IsPausable:           pgtype.Bool{Bool: true, Status: pgtype.Present},
					IsEnrollmentRequired: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Fail case: product is_pausable false",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{String: randomID.String(), Status: pgtype.Present}},
				entities.ProductSetting{
					ProductID:            pgtype.Text{String: randomID.String(), Status: pgtype.Present},
					IsPausable:           pgtype.Bool{Bool: false, Status: pgtype.Present},
					IsEnrollmentRequired: pgtype.Bool{Bool: true, Status: pgtype.Present},
				},
			},
			ExpectedErr: status.Errorf(codes.Internal, "LOA order created for product %v but product is not pausable", randomID),
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			s := &StudentProductService{}

			orderItemData := utils.OrderItemData{
				ProductInfo:    testCase.Req.([]interface{})[0].(entities.Product),
				ProductSetting: testCase.Req.([]interface{})[1].(entities.ProductSetting),
			}

			err := s.ValidateProductSettingForLOAOrder(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db)
		})
	}
}

func TestStudentProductService_checkUniqueStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                           *mockDb.Ext
		studentProductRepo           *mockRepositories.MockStudentProductRepo
		billingSchedulePeriodRepo    *mockRepositories.MockBillingSchedulePeriodRepo
		studentAssociatedProductRepo *mockRepositories.MockStudentAssociatedProductRepo
		productRepo                  *mockRepositories.MockProductRepo
		now                          = time.Now()
	)
	testcases := []utils.TestCase{
		{
			Name: "Fail Case: error when get latest end date student product with product id and student id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: true,
					Order:            entities.Order{StudentID: pgtype.Text{String: constant.StudentID}},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{String: constant.ProductID},
						IsUnique:  pgtype.Bool{Bool: true},
					},
				},
				&entities.StudentProduct{},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.InvalidArgument, fmt.Sprintf("error when get latest end date student product with product id and student id: %v", constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetLatestEndDateStudentProductWithProductIDAndStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail Case: error when creating student product with conflict time range with previous student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: false,
					Order:            entities.Order{StudentID: pgtype.Text{String: constant.StudentID}},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{String: constant.ProductID},
						IsUnique:  pgtype.Bool{Bool: true},
					},
				},
				&entities.StudentProduct{
					StartDate: pgtype.Timestamptz{
						Time: now,
					},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.InvalidArgument, "creating return student product have error because it is unique product and it have conflict time range with previous student product"),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetLatestEndDateStudentProductWithProductIDAndStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{
					{
						EndDate: pgtype.Timestamptz{
							Time: now.AddDate(0, 0, 30),
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					IsOneTimeProduct: false,
					Order:            entities.Order{StudentID: pgtype.Text{String: constant.StudentID}},
					ProductInfo: entities.Product{
						ProductID: pgtype.Text{String: constant.ProductID},
						IsUnique:  pgtype.Bool{Bool: false},
					},
				},
				&entities.StudentProduct{},
			},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {

			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = &mockRepositories.MockStudentProductRepo{}
			billingSchedulePeriodRepo = &mockRepositories.MockBillingSchedulePeriodRepo{}
			studentAssociatedProductRepo = &mockRepositories.MockStudentAssociatedProductRepo{}
			productRepo = &mockRepositories.MockProductRepo{}
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:           studentProductRepo,
				BillingSchedulePeriodRepo:    billingSchedulePeriodRepo,
				StudentAssociatedProductRepo: studentAssociatedProductRepo,
				ProductRepo:                  productRepo,
			}
			orderItemDataReq := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			studentProductReq := testCase.Req.([]interface{})[1].(*entities.StudentProduct)
			err := s.checkUniqueStudentProduct(testCase.Ctx, db, orderItemDataReq, studentProductReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, billingSchedulePeriodRepo, studentAssociatedProductRepo, productRepo)
		})
	}
}

func TestStudentProductService_GetActiveRecurringProductsOfStudentInLocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentProductRepo                 *mockRepositories.MockStudentProductRepo
		productLocationRepo                *mockRepositories.MockProductLocationRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	expectedResp := []entities.StudentProduct{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
			},
			StudentID: pgtype.Text{
				String: constant.StudentID,
			},
			ProductID: pgtype.Text{
				String: constant.ProductID,
			},
		},
	}
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when get ignore student product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get ignore student product have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when getting active recurring products of student in location",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get active student product have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when get active student product of operation fee have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get active student product of operation fee have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
				studentProductRepo.On("GetActiveOperationFeeOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when get locationIDs of product of operation fee have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get locationIDs of product of operation fee have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResp, nil)
				studentProductRepo.On("GetActiveOperationFeeOfStudent", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						StudentProductID: pgtype.Text{
							String: "student_product_1",
						},
					},
				}, nil)
				productLocationRepo.On("GetLocationIDsWithProductID", ctx, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when get locationIDs of product of operation fee have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get locationIDs of product of operation fee have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResp, nil)
				studentProductRepo.On("GetActiveOperationFeeOfStudent", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						StudentProductID: pgtype.Text{
							String: "student_product_1",
						},
					},
				}, nil)
				productLocationRepo.On("GetLocationIDsWithProductID", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetLatestStatusEnrollmentByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{
					{
						StudentID: pgtype.Text{
							String: "student_1",
						},
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name:         "Success case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: expectedResp,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResp, nil)
				studentProductRepo.On("GetActiveOperationFeeOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name:         "Success case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: expectedResp,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResp, nil)
				studentProductRepo.On("GetActiveOperationFeeOfStudent", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						StudentProductID: pgtype.Text{
							String: "student_product_1",
						},
					},
				}, nil)
				productLocationRepo.On("GetLocationIDsWithProductID", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetLatestStatusEnrollmentByStudentIDAndLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{
					{
						StudentID: pgtype.Text{
							String: "student_1",
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			productLocationRepo = new(mockRepositories.MockProductLocationRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo:                 studentProductRepo,
				ProductLocationRepo:                productLocationRepo,
				StudentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			_, err := s.GetActiveRecurringProductsOfStudentInLocation(testCase.Ctx, db, "1", "2")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, productLocationRepo, studentEnrollmentStatusHistoryRepo)
		})
	}
}

func TestStudentProductService_GetRecurringProductsOfStudentInLocationForLOA(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
		ProductSettingRepo *mockRepositories.MockProductSettingRepo
	)
	expectedResp := []entities.StudentProduct{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
			},
			StudentID: pgtype.Text{
				String: constant.StudentID,
			},
			ProductID: pgtype.Text{
				String: constant.ProductID,
			},
		},
	}
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when get ignore student product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get ignore student product have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when getting active recurring products of student in location",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "get active student product have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when check isPauseable",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						ProductID: pgtype.Text{
							String: "product_id",
						},
					},
				}, nil)
				ProductSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{}, constant.ErrDefault)
			},
		},
		{
			Name:         "Success case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: expectedResp,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetActiveRecurringProductsOfStudentInLocation", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.StudentProduct{
					{
						ProductID: pgtype.Text{
							String: "product_id",
						},
					},
				}, nil)
				ProductSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{IsPausable: pgtype.Bool{Bool: true}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			ProductSettingRepo = new(mockRepositories.MockProductSettingRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				StudentProductRepo: studentProductRepo,
				ProductSettingRepo: ProductSettingRepo,
			}
			resp, err := s.GetRecurringProductsOfStudentInLocationForLOA(testCase.Ctx, db, "1", "2")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedStudentProducts := testCase.ExpectedResp.([]entities.StudentProduct)
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, len(expectedStudentProducts), len(resp))
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo, ProductSettingRepo)
		})
	}
}
