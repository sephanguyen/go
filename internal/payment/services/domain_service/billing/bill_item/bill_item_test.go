package service

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
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
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	FailCaseLengthError                   = "Fail case: Error length pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseSlotError                     = "Fail case: Error slot pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseMapSlotError                  = "Fail case: Error map slot pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseWeightError                   = "Fail case: Error weight pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseMapWeightError                = "Fail case: Error map weight pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseCreateError                   = "Fail case: Error Create pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseMultiCreateError              = "Fail case: Error MultiCreate billItemCourse pre-condition when generateBillingDescriptionForPackageAndCreateData"
	FailCaseSettingNonLatestBillItemError = "Fail case: Error when setting non latest bill item for student product entities"
	FailCaseGetCountError                 = "Fail case: getting count of bill item by order id"
	MissingSlotError                      = "missing slot in course info course_id_1 of bill item"
	InconsistentCourseInfoError           = "inconsistency course info course_id_1 between order item and bill item"
	MissingWeightError                    = "missing weight in course info course_id_1 of bill item"
)

func TestBillItemService_CreateNewBillItemForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	mapErrorCourseSlotInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseSlotInfoInOrder["course_id_1"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}
	mapErrorCourseSlotInfoInOrder["course_id_2"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}

	mapErrorCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}
	mapErrorCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}

	mapCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(1),
	}
	mapCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(2),
	}

	mapCourseSlottInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapCourseSlottInfoInOrder["course_id_1"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(1),
	}
	mapCourseSlottInfoInOrder["course_id_2"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(2),
	}

	discountName := "discount_name"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when generateBillingDescriptionForMaterialAndCreateData: recurring",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error get material repo when generateBillingDescriptionForMaterialAndCreateData one time",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error create when generateBillingDescriptionForMaterialAndCreateData one time",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{}, nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error create when generateBillingDescriptionForFeeAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_FEE,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseLengthError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingCourseInfoBillItem),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingSlotError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseMapSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Slot:     wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Slot:     wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_SLOT,
					MapCourseInfo: mapErrorCourseSlotInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingWeightError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseMapWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapErrorCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseMultiCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case weigth",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
							TaxItem: &pb.TaxBillItem{
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
								TaxPercentage: float32(2),
								TaxId:         "tax_id_1",
								TaxAmount:     float32(3),
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "discount_id_1",
								DiscountAmountValue: float32(12),
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
								DiscountAmount:      float32(23),
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Happy case slot",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId:   "course_id_1",
									Slot:       wrapperspb.Int32(1),
									CourseName: "course_name_1",
								},
								{
									CourseId:   "course_id_2",
									Slot:       wrapperspb.Int32(2),
									CourseName: "course_name_1",
								},
							},
							TaxItem: &pb.TaxBillItem{
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
								TaxPercentage: float32(2),
								TaxId:         "tax_id_1",
								TaxAmount:     float32(3),
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "discount_id_1",
								DiscountAmountValue: float32(12),
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
								DiscountAmount:      float32(23),
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK,
					MapCourseInfo: mapCourseSlottInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)

			err := s.CreateNewBillItemForOneTimeBilling(testCase.Ctx, db, req, discountName)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_CreateUpdateBillItemForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	oldBillItem := entities.BillItem{}

	mapErrorCourseSlotInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseSlotInfoInOrder["course_id_1"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}
	mapErrorCourseSlotInfoInOrder["course_id_2"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}

	mapErrorCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}
	mapErrorCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}

	mapCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(1),
	}
	mapCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(2),
	}

	discountName := "discount_name"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when generating bill item entities for update without adjustment price",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "generating bill item entities for update without adjustment price"),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseSettingNonLatestBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error GetByIDForUpdate when generateBillingDescriptionForMaterialAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{}, constant.ErrDefault)

			},
		},
		{
			Name:        "Fail case: Error Create when generateBillingDescriptionForMaterialAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{
					CustomBillingDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
				}, nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, constant.ErrDefault)

			},
		},
		{
			Name:        "Fail case: Error Create when generateBillingDescriptionForFeeAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_FEE,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseLengthError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingCourseInfoBillItem),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingSlotError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseMapSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Slot:     wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Slot:     wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_SLOT,
					MapCourseInfo: mapErrorCourseSlotInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingWeightError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseMapWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapErrorCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseMultiCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
							TaxItem: &pb.TaxBillItem{
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
								TaxPercentage: float32(2),
								TaxId:         "tax_id_1",
								TaxAmount:     float32(3),
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "discount_id_1",
								DiscountAmountValue: float32(12),
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
								DiscountAmount:      float32(23),
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			req := testCase.Req.(utils.OrderItemData)
			err := s.CreateUpdateBillItemForOneTimeBilling(testCase.Ctx, db, oldBillItem, req, discountName)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_CreateCancelBillItemForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseSettingNonLatestBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.BillItem{
				Price: pgtype.Numeric{
					Int: big.NewInt(100),
				},
				FinalPrice: pgtype.Numeric{
					Int: big.NewInt(100),
				},
				StudentProductID: pgtype.Text{
					String: "student_product_id",
				},
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when creating billing item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.BillItem{
				Price: pgtype.Numeric{
					Int: big.NewInt(100),
				},
				FinalPrice: pgtype.Numeric{
					Int: big.NewInt(100),
				},
				StudentProductID: pgtype.Text{
					String: "student_product_id",
				},
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: entities.BillItem{
				Price: pgtype.Numeric{
					Int: big.NewInt(100),
				},
				FinalPrice: pgtype.Numeric{
					Int: big.NewInt(100),
				},
				StudentProductID: pgtype.Text{
					String: "student_product_id",
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			oldBillItem := testCase.Req.(entities.BillItem)
			err := s.CreateCancelBillItemForOneTimeBilling(testCase.Ctx, db, oldBillItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetOldBillItemForUpdateOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseSettingNonLatestBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         utils.OrderItemData{},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         utils.OrderItemData{},

			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			orderItemData := testCase.Req.(utils.OrderItemData)
			_, err := s.GetOldBillItemForUpdateOneTimeBilling(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetMapOldBillingItemForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when getting old bill item by student product id and period id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					StudentProductId: wrapperspb.String("student_product_id"),
				},
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					StudentProductId: wrapperspb.String("student_product_id"),
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			orderItemData := testCase.Req.(utils.OrderItemData)
			mapPeriodInfo := make(map[string]entities.BillingSchedulePeriod, 1)
			mapPeriodInfo["period_id"] = entities.BillingSchedulePeriod{
				BillingSchedulePeriodID: pgtype.Text{
					String: "billing_schedule_period_id",
				},
			}
			_, err := s.GetMapOldBillingItemForRecurringBilling(testCase.Ctx, db, orderItemData, mapPeriodInfo)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_CreateCustomBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when creating bill item for custom order have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			order := entities.Order{
				OrderID: pgtype.Text{
					String: "order_id",
				},
				StudentID: pgtype.Text{
					String: "student_id",
				},
				LocationID: pgtype.Text{
					String: "location_id",
				},
			}
			customBillItem := pb.CustomBillingItem{
				TaxItem: &pb.TaxBillItem{
					TaxId:         "tax_id",
					TaxPercentage: float32(2),
					TaxAmount:     float32(2),
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
				},
				Name:  "custom_billing",
				Price: float32(3),
			}
			err := s.CreateCustomBillItem(testCase.Ctx, db, &customBillItem, order, "location_name")

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_CreateCustomBillItemWithAccountCategoryID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db              *mockDb.Ext
		billItemRepo    *mockRepositories.MockBillItemRepo
		materialRepo    *mockRepositories.MockMaterialRepo
		billItemCourse  *mockRepositories.MockBillItemCourseRepo
		accountCategory *mockRepositories.MockBillItemAccountCategoryRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when creating bill item for custom order with account category have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				accountCategory.On("CreateMultiple", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				accountCategory.On("CreateMultiple", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			accountCategory = &mockRepositories.MockBillItemAccountCategoryRepo{}
			s := &BillItemService{
				BillItemRepo:                billItemRepo,
				MaterialRepo:                materialRepo,
				BillItemCourseRepo:          billItemCourse,
				BillItemAccountCategoryRepo: accountCategory,
			}
			testCase.Setup(testCase.Ctx)
			order := entities.Order{
				OrderID: pgtype.Text{
					String: "order_id",
				},
				StudentID: pgtype.Text{
					String: "student_id",
				},
				LocationID: pgtype.Text{
					String: "location_id",
				},
			}
			customBillItem := pb.CustomBillingItem{
				TaxItem: &pb.TaxBillItem{
					TaxId:         "tax_id",
					TaxPercentage: float32(2),
					TaxAmount:     float32(2),
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
				},
				Name:               "custom_billing",
				Price:              float32(3),
				AccountCategoryIds: []string{"1", "2"},
			}
			err := s.CreateCustomBillItem(testCase.Ctx, db, &customBillItem, order, "location_name")

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_UpdateReviewFlagForBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when updating review flag in billing item service",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					StudentProductId: wrapperspb.String("student_product_id"),
				},
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("UpdateReviewFlagByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					StudentProductId: wrapperspb.String("student_product_id"),
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("UpdateReviewFlagByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)

			err := s.UpdateReviewFlagForBillItem(testCase.Ctx, db, "", true)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_VoidBillItemByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when void bill item by id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				billItemRepo.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				billItemRepo.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)

			orderIDReq := testCase.Req.(string)
			err := s.VoidBillItemByOrderID(testCase.Ctx, db, orderIDReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_UpdateBillItemStatusAndReturnOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: invalid bill item sequence number",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.InvalidArgument, "invalid bill item sequence number"),
			Req:         int32(-1),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: invalid bill item sequence number",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         int32(2),
			Setup: func(ctx context.Context) {
				billItemRepo.On("UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         int32(2),
			Setup: func(ctx context.Context) {
				billItemRepo.On("UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			billItemSequenceNumber := testCase.Req.(int32)
			_, err := s.UpdateBillItemStatusAndReturnOrderID(testCase.Ctx, db, billItemSequenceNumber, "")

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetRecurringBillItemsForScheduledGenerationOfNextBillItems(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: retrieving recurring bill itemsr",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetRecurringBillItemsForScheduledGenerationOfNextBillItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetRecurringBillItemsForScheduledGenerationOfNextBillItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			_, err := s.GetRecurringBillItemsForScheduledGenerationOfNextBillItems(testCase.Ctx, db)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetBillItemDescriptionByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseGetCountError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("CountBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int(2), constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: converting billItemEntity to billItemDescription",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("CountBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int(2), nil)
				billItemRepo.On("GetBillItemByOrderIDAndPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("CountBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int(2), nil)
				billItemRepo.On("GetBillItemByOrderIDAndPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			_, _, err := s.GetBillItemDescriptionsByOrderIDWithPaging(testCase.Ctx, db, "", 1, 2)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetBillItemDescriptionByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseGetCountError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("CountBillItemByStudentIDAndLocationIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int(2), constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: converting billItemEntity to billItemDescription",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("CountBillItemByStudentIDAndLocationIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int(2), nil)
				billItemRepo.On("GetBillItemByStudentIDAndLocationIDsPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("CountBillItemByStudentIDAndLocationIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int(2), nil)
				billItemRepo.On("GetBillItemByStudentIDAndLocationIDsPaging", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			_, _, err := s.GetBillItemDescriptionByStudentIDAndLocationIDs(testCase.Ctx, db, "", []string{}, 1, 2)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetBillItemInfoByOrderIDAndUniqueByProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseGetCountError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetBillItemInfoByOrderIDAndUniqueByProductID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetBillItemInfoByOrderIDAndUniqueByProductID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			_, err := s.GetBillItemInfoByOrderIDAndUniqueByProductID(testCase.Ctx, db, "")

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetFirstBillItemsByOrderIDAndProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: getting all first bill item distinct by order id and product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			_, _, err := s.GetFirstBillItemsByOrderIDAndProductID(testCase.Ctx, db, "", 1, 2)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_GetLatestBillItemByStudentProductIDForStudentBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db             *mockDb.Ext
		billItemRepo   *mockRepositories.MockBillItemRepo
		materialRepo   *mockRepositories.MockMaterialRepo
		billItemCourse *mockRepositories.MockBillItemCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: getting two latest bill item by student product id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLatestBillItemByStudentProductIDForStudentBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLatestBillItemByStudentProductIDForStudentBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			s := &BillItemService{
				BillItemRepo:       billItemRepo,
				MaterialRepo:       materialRepo,
				BillItemCourseRepo: billItemCourse,
			}
			testCase.Setup(testCase.Ctx)
			_, err := s.GetLatestBillItemByStudentProductIDForStudentBilling(testCase.Ctx, db, "")

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_CreateNewBillItemForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                   *mockDb.Ext
		billItemRepo         *mockRepositories.MockBillItemRepo
		materialRepo         *mockRepositories.MockMaterialRepo
		billItemCourse       *mockRepositories.MockBillItemCourseRepo
		upcomingBillItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	mapErrorCourseSlotInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseSlotInfoInOrder["course_id_1"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}
	mapErrorCourseSlotInfoInOrder["course_id_2"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}

	mapErrorCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}
	mapErrorCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}

	mapCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(1),
	}
	mapCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(2),
	}

	discountName := "discount_name"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when generateBillingDescriptionForMaterialAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error get material repo when generateBillingDescriptionForMaterialAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{}, constant.ErrDefault)
				// billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error create when generateBillingDescriptionForMaterialAndCreateData one time",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{}, nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error create when generateBillingDescriptionForFeeAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_FEE,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseLengthError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingCourseInfoBillItem),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingSlotError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseMapSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Slot:     wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Slot:     wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_SLOT,
					MapCourseInfo: mapErrorCourseSlotInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingWeightError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseMapWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapErrorCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseMultiCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
					StartDate:  timestamppb.Now(),
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:  "product_id_1",
							Price:      1000,
							FinalPrice: 2000,
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
							TaxItem: &pb.TaxBillItem{
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
								TaxPercentage: float32(2),
								TaxId:         "tax_id_1",
								TaxAmount:     float32(3),
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "discount_id_1",
								DiscountAmountValue: float32(12),
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
								DiscountAmount:      float32(23),
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			upcomingBillItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &BillItemService{
				BillItemRepo:         billItemRepo,
				MaterialRepo:         materialRepo,
				BillItemCourseRepo:   billItemCourse,
				UpcomingBillItemRepo: upcomingBillItemRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)
			proRatedBillItem := utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
				},
			}
			proRatedPrice := entities.ProductPrice{}
			ratioOfProRatedBillingItem := entities.BillingRatio{}
			normalBillItem := []utils.BillingItemData{
				{
					BillingItem: &pb.BillingItem{
						ProductId:               "product_id",
						Price:                   float32(1000),
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
						AdjustmentPrice:         wrapperspb.Float(23),
					},
				},
			}
			mapPeriodInfo := make(map[string]entities.BillingSchedulePeriod, 1)

			mapPeriodInfo["period_id"] = entities.BillingSchedulePeriod{
				BillingSchedulePeriodID: pgtype.Text{
					String: "BillingSchedulePeriodID",
				},
			}
			err := s.CreateNewBillItemForRecurringBilling(testCase.Ctx, db, req, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, discountName)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse)
		})
	}
}

func TestBillItemService_CreateUpdateBillItemForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                      *mockDb.Ext
		billItemRepo            *mockRepositories.MockBillItemRepo
		materialRepo            *mockRepositories.MockMaterialRepo
		billItemCourse          *mockRepositories.MockBillItemCourseRepo
		upcomingBillingItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	mapErrorCourseSlotInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseSlotInfoInOrder["course_id_1"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}
	mapErrorCourseSlotInfoInOrder["course_id_2"] = &pb.CourseItem{
		Slot: wrapperspb.Int32(3),
	}

	mapErrorCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapErrorCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}
	mapErrorCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(3),
	}

	mapCourseWeightInfoInOrder := make(map[string]*pb.CourseItem, 2)
	mapCourseWeightInfoInOrder["course_id_1"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(1),
	}
	mapCourseWeightInfoInOrder["course_id_2"] = &pb.CourseItem{
		Weight: wrapperspb.Int32(2),
	}

	discountName := "discount_name"
	testcases := []utils.TestCase{
		{
			Name:        FailCaseSettingNonLatestBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "generating bill item entities for update without adjustment price"),
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
			},

			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Fail case: Error GetByIDForUpdate when generateBillingDescriptionForMaterialAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{}, constant.ErrDefault)
				// upcomingBillingItemRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error Create when generateBillingDescriptionForMaterialAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_MATERIAL,
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Material{
					CustomBillingDate: pgtype.Timestamptz{
						Time:   time.Now(),
						Status: pgtype.Present,
					},
				}, nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, constant.ErrDefault)

			},
		},
		{
			Name:        "Fail case: Error Create when generateBillingDescriptionForFeeAndCreateData",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_FEE,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseLengthError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingCourseInfoBillItem),
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        FailCaseSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingSlotError),
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_SLOT,
				},
			},

			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        FailCaseMapSlotError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Slot:     wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Slot:     wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_SLOT,
					MapCourseInfo: mapErrorCourseSlotInfoInOrder,
				},
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        FailCaseWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, MissingWeightError),
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
								},
								{
									CourseId: "course_id_1",
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
				},
			},

			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        FailCaseMapWeightError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, InconsistentCourseInfoError),
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapErrorCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        FailCaseCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseMultiCreateError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:  constant.ProductID,
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
							TaxItem: &pb.TaxBillItem{
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
								TaxPercentage: float32(2),
								TaxId:         "tax_id_1",
								TaxAmount:     float32(3),
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "discount_id_1",
								DiscountAmountValue: float32(12),
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
								DiscountAmount:      float32(23),
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        constant.HappyCase + " 1",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
				IsUpcoming: true,
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId:  constant.ProductID,
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
				},
				Order: entities.Order{
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
							CourseItems: []*pb.CourseItem{
								{
									CourseId: "course_id_1",
									Weight:   wrapperspb.Int32(1),
								},
								{
									CourseId: "course_id_2",
									Weight:   wrapperspb.Int32(2),
								},
							},
							TaxItem: &pb.TaxBillItem{
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
								TaxPercentage: float32(2),
								TaxId:         "tax_id_1",
								TaxAmount:     float32(3),
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "discount_id_1",
								DiscountAmountValue: float32(12),
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
								DiscountAmount:      float32(23),
							},
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName: "location_name_1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageInfo: utils.PackageInfo{
					QuantityType:  pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT,
					MapCourseInfo: mapCourseWeightInfoInOrder,
				},
			},

			Setup: func(ctx context.Context) {
				upcomingBillingItemRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
				billItemCourse.On("MultiCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			upcomingBillingItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &BillItemService{
				BillItemRepo:         billItemRepo,
				MaterialRepo:         materialRepo,
				BillItemCourseRepo:   billItemCourse,
				UpcomingBillItemRepo: upcomingBillingItemRepo,
			}
			testCase.Setup(testCase.Ctx)
			req := testCase.Req.(utils.OrderItemData)
			proRatedBillItem := testCase.ExpectedResp.(utils.BillingItemData)
			proRatedPrice := entities.ProductPrice{}
			ratioOfProRatedBillingItem := entities.BillingRatio{}
			normalBillItem := []utils.BillingItemData{
				{
					BillingItem: &pb.BillingItem{
						ProductId:               "product_id",
						Price:                   float32(1000),
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
						AdjustmentPrice:         wrapperspb.Float(23),
					},
				},
			}
			mapPeriodInfo := make(map[string]entities.BillingSchedulePeriod, 1)
			mapOldBillingItem := make(map[string]entities.BillItem, 1)
			mapPeriodInfo["period_id"] = entities.BillingSchedulePeriod{
				BillingSchedulePeriodID: pgtype.Text{
					String: "BillingSchedulePeriodID",
				},
			}
			err := s.CreateUpdateBillItemForRecurringBilling(testCase.Ctx, db, req, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem, discountName)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse, upcomingBillingItemRepo)
		})
	}
}

func TestBillItemService_CreateCancelBillItemForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                      *mockDb.Ext
		billItemRepo            *mockRepositories.MockBillItemRepo
		materialRepo            *mockRepositories.MockMaterialRepo
		billItemCourse          *mockRepositories.MockBillItemCourseRepo
		upcomingBillingItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseSettingNonLatestBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId:       &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:        constant.ProductID,
					StudentProductId: &wrapperspb.StringValue{Value: "student_product_id_1"},
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
					ProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_FEE,
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when creating billing item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId:       &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:        constant.ProductID,
					StudentProductId: &wrapperspb.StringValue{Value: "student_product_id_1"},
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
					ProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_FEE,
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillingItemRepo.On("RemoveOldUpcomingBillItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", time.Now(), nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			ExpectedResp: utils.BillingItemData{
				BillingItem: &pb.BillingItem{
					ProductId:               "product_id",
					Price:                   float32(1000),
					BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
					AdjustmentPrice:         wrapperspb.Float(23),
				},
			},
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId:       &wrapperspb.StringValue{Value: constant.DiscountID},
					ProductId:        constant.ProductID,
					StudentProductId: &wrapperspb.StringValue{Value: "student_product_id_1"},
				},
				Order: entities.Order{
					StudentID: pgtype.Text{
						String: "student_id",
						Status: pgtype.Present,
					},
					OrderID: pgtype.Text{
						String: "order_id_1",
						Status: pgtype.Present,
					},
					LocationID: pgtype.Text{
						String: "location_id_1",
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId:       "product_id_1",
							Price:           1000,
							FinalPrice:      2000,
							AdjustmentPrice: wrapperspb.Float(23),
						},
					},
				},
				ProductInfo: entities.Product{
					Name: pgtype.Text{
						String: "product_name_1",
						Status: pgtype.Present,
					},
				},
				StudentProduct: entities.StudentProduct{
					StudentProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
					ProductID: pgtype.Text{
						String: "student_product_id_1",
						Status: pgtype.Present,
					},
				},
				LocationName:     "location_name_1",
				ProductType:      pb.ProductType_PRODUCT_TYPE_FEE,
				IsOneTimeProduct: true,
			},

			Setup: func(ctx context.Context) {
				billItemRepo.On("SetNonLatestBillItemByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillingItemRepo.On("RemoveOldUpcomingBillItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", time.Now(), nil)
				billItemRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.Int4{Int: 3}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = &mockRepositories.MockBillItemRepo{}
			materialRepo = &mockRepositories.MockMaterialRepo{}
			billItemCourse = &mockRepositories.MockBillItemCourseRepo{}
			upcomingBillingItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &BillItemService{
				BillItemRepo:         billItemRepo,
				MaterialRepo:         materialRepo,
				BillItemCourseRepo:   billItemCourse,
				UpcomingBillItemRepo: upcomingBillingItemRepo,
			}
			testCase.Setup(testCase.Ctx)
			req := testCase.Req.(utils.OrderItemData)
			proRatedBillItem := testCase.ExpectedResp.(utils.BillingItemData)
			ratioOfProRatedBillingItem := entities.BillingRatio{}
			normalBillItem := []utils.BillingItemData{
				{
					BillingItem: &pb.BillingItem{
						ProductId:               "product_id",
						Price:                   float32(1000),
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "BillingSchedulePeriodID"},
						AdjustmentPrice:         wrapperspb.Float(23),
					},
				},
			}
			mapPeriodInfo := make(map[string]entities.BillingSchedulePeriod, 1)
			mapOldBillingItem := make(map[string]entities.BillItem, 1)
			mapPeriodInfo["period_id"] = entities.BillingSchedulePeriod{
				BillingSchedulePeriodID: pgtype.Text{
					String: "BillingSchedulePeriodID",
				},
			}
			err := s.CreateCancelBillItemForRecurringBilling(testCase.Ctx, db, req, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, materialRepo, billItemCourse, upcomingBillingItemRepo)
		})
	}
}

func TestStudentProductService_GetExportStudentBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		tx           *mockDb.Tx
		billItemRepo *mockRepositories.MockBillItemRepo
		userRepo     *mockRepositories.MockUserRepo
		billItemResp []*entities.BillItem
	)
	slotValue := int32(2)
	discountName := "discount_name"
	studentIDs := []string{"student_1", "student_2"}
	billItemResp = []*entities.BillItem{
		{
			StudentID: pgtype.Text{
				String: studentIDs[0],
				Status: pgtype.Present,
			},
			ProductID: pgtype.Text{
				String: "product_1",
				Status: pgtype.Present,
			},
			LocationID: pgtype.Text{
				String: "location",
				Status: pgtype.Present,
			},
			BillStatus: pgtype.Text{
				String: pb.BillingStatus_BILLING_STATUS_BILLED.String(),
				Status: pgtype.Present,
			},
			BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
				ProductID:    "10",
				ProductName:  "Product 1",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
				DiscountName: &discountName,
				CourseItems: []*entities.CourseItem{
					{
						CourseID:   idutil.ULIDNow(),
						CourseName: "course_name_1",
						Slot:       &slotValue,
					},
					{
						CourseID:   idutil.ULIDNow(),
						CourseName: "course_name_1",
						Slot:       &slotValue,
					},
				},
				BillingRatioNumerator:   &slotValue,
				BillingRatioDenominator: &slotValue,
			}),
			AdjustmentPrice: database.Numeric(200),
			BillType: pgtype.Text{
				String: pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(),
				Status: pgtype.Present,
			},
			CreatedAt: pgtype.Timestamptz{
				Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
			},
			DiscountAmount: database.Numeric(200),
			TaxAmount:      database.Numeric(200),
			FinalPrice:     database.Numeric(2000),
		},
		{
			StudentID: pgtype.Text{
				String: studentIDs[1],
				Status: pgtype.Present,
			},
			ProductID: pgtype.Text{
				String: "product_2",
				Status: pgtype.Present,
			},
			LocationID: pgtype.Text{
				String: "location",
				Status: pgtype.Present,
			},
			BillStatus: pgtype.Text{
				String: pb.BillingStatus_BILLING_STATUS_BILLED.String(),
				Status: pgtype.Present,
			},
			BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
				ProductID:    "10",
				ProductName:  "Product 2",
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
				DiscountName: &discountName,
			}),
			CreatedAt: pgtype.Timestamptz{
				Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
			},
			DiscountAmount: database.Numeric(200),
			TaxAmount:      database.Numeric(200),
			FinalPrice:     database.Numeric(2000),
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        "error get billing item for export student billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when get billing item for export student billing: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetExportStudentBilling", ctx, tx, []string{}).Return([]*entities.BillItem{}, []string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "error get user for export student billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when get user for export student billing: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetExportStudentBilling", ctx, tx, []string{}).Return(billItemResp, studentIDs, nil)
				userRepo.On("GetStudentsByIDs", ctx, tx, studentIDs).Return([]entities.User{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetExportStudentBilling", ctx, tx, []string{}).Return(billItemResp, studentIDs, nil)
				userRepo.On("GetStudentsByIDs", ctx, tx, studentIDs).Return([]entities.User{
					{
						UserID: pgtype.Text{
							String: "student_1",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{
							String: "student_name_1",
							Status: pgtype.Present,
						},
					},
					{
						UserID: pgtype.Text{
							String: "student_2",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{
							String: "student_name_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			tx = new(mockDb.Tx)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			userRepo = new(mockRepositories.MockUserRepo)

			testCase.Setup(testCase.Ctx)
			s := &BillItemService{
				BillItemRepo: billItemRepo,
				UserRepo:     userRepo,
			}
			exportStudentBillingResp, err := s.GetExportStudentBilling(testCase.Ctx, tx, []string{})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, 2, len(exportStudentBillingResp))
			}
			mock.AssertExpectationsForObjects(t, tx, billItemRepo, userRepo)
		})
	}
}

func TestStudentProductService_BuildMapBillItemWithProductIDByOrderIDAndProductIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		tx           *mockDb.Tx
		billItemRepo *mockRepositories.MockBillItemRepo
		billItemResp []entities.BillItem
	)

	productIDs := []string{"1", "2"}

	billItemResp = []entities.BillItem{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
				Status: pgtype.Present,
			},
			ProductID: pgtype.Text{
				String: productIDs[0],
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
			ProductID: pgtype.Text{
				String: productIDs[1],
				Status: pgtype.Present,
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get bill items by order id and product ids",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				[]string{"1", "2"},
			},
			ExpectedErr: status.Errorf(codes.Internal, "Error when get bill items by order ID and product IDs: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetByOrderIDAndProductIDs", ctx, tx, constant.OrderID, productIDs).Return([]entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:         constant.HappyCase,
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: billItemResp,
			Req: []interface{}{
				constant.OrderID,
				[]string{"1", "2"},
			},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetByOrderIDAndProductIDs", ctx, tx, constant.OrderID, productIDs).Return(billItemResp, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			tx = new(mockDb.Tx)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillItemService{
				BillItemRepo: billItemRepo,
			}
			orderIDReq := testCase.Req.([]interface{})[0].(string)
			productIDsReq := testCase.Req.([]interface{})[1].([]string)
			mapProductIDAndBillItemResp, err := s.BuildMapBillItemWithProductIDByOrderIDAndProductIDs(testCase.Ctx, tx, orderIDReq, productIDsReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedResp := testCase.ExpectedResp.([]entities.BillItem)
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, len(expectedResp), len(mapProductIDAndBillItemResp))
			}
			mock.AssertExpectationsForObjects(t, tx, billItemRepo)
		})
	}
}

func TestBillItemService_GetMapPresentAndFutureBillItemInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		billItemRepo *mockRepositories.MockBillItemRepo
	)
	mapBillItems := map[string]*entities.BillItem(map[string]*entities.BillItem{})
	mapBillItems[constant.StudentProductID] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: constant.StudentProductID,
			Status: pgtype.Present,
		},
	}
	mapBillItems["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
	}
	billItem := []*entities.BillItem{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
		},
	}

	billItemUpdateOrder := []*entities.BillItem{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         &entities.BillItem{},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetPresentAndFutureBillItemsByStudentProductIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItem, nil)
			},
		},
		{
			Name:        "happy case when have update order",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         &entities.BillItem{},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetPresentAndFutureBillItemsByStudentProductIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItemUpdateOrder, nil)
			},
		},
		{
			Name:        "error case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         &entities.BillItem{},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetPresentAndFutureBillItemsByStudentProductIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillItemService{
				BillItemRepo: billItemRepo,
			}
			billItemResp, err := s.GetMapPresentAndFutureBillItemInfo(testCase.Ctx, db, []string{}, "")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, mapBillItems, billItemResp)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo)
		})
	}
}

func TestBillItemService_GetMapPastBillItemInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		billItemRepo *mockRepositories.MockBillItemRepo
	)
	mapBillItems := map[string]*entities.BillItem(map[string]*entities.BillItem{})
	mapBillItems[constant.StudentProductID] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: constant.StudentProductID,
			Status: pgtype.Present,
		},
	}
	mapBillItems["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
	}
	billItem := []*entities.BillItem{
		{
			StudentProductID: pgtype.Text{
				String: constant.StudentProductID,
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
		},
	}
	testcases := []utils.TestCase{
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         &entities.BillItem{},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetPastBillItemsByStudentProductIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItem, nil)
			},
		},
		{
			Name:        "error case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         &entities.BillItem{},
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetPastBillItemsByStudentProductIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.BillItem{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillItemService{
				BillItemRepo: billItemRepo,
			}
			billItemResp, err := s.GetMapPastBillItemInfo(testCase.Ctx, db, []string{}, "")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, mapBillItems, billItemResp)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo)
		})
	}
}

func TestStudentProductService_GetUpcomingBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		tx           *mockDb.Tx
		billItemRepo *mockRepositories.MockBillItemRepo
		billItemResp *entities.BillItem
	)
	billItemResp = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: constant.StudentProductID,
			Status: pgtype.Present,
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        "error because of database",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error get upcoming billing by student product ID: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetUpcomingBillingByStudentProductID", ctx, tx, constant.StudentProductID, "student_id").Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetUpcomingBillingByStudentProductID", ctx, tx, constant.StudentProductID, "student_id").Return(billItemResp, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			tx = new(mockDb.Tx)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillItemService{
				BillItemRepo: billItemRepo,
			}
			billItem, err := s.GetUpcomingBilling(testCase.Ctx, tx, constant.StudentProductID, "student_id")

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, billItem, billItemResp)
			}
			mock.AssertExpectationsForObjects(t, tx, billItemRepo)
		})
	}
}

func TestBillItemService_CreateUpcomingBillItems(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		billItemRepo *mockRepositories.MockBillItemRepo
	)

	testCases := []utils.TestCase{
		{
			Name:         "happy case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(pgtype.Int4{
					Int:    4,
					Status: pgtype.Present,
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillItemService{
				BillItemRepo: billItemRepo,
			}
			billItem := entities.BillItem{
				BillType: pgtype.Text{
					String: pb.BillingItemType_ONE_TIME_MATERIAL.String(),
					Status: pgtype.Present,
				},
				BillStatus: pgtype.Text{
					String: pb.BillingStatus_BILLING_STATUS_PENDING.String(),
					Status: pgtype.Present,
				},
				ProductDescription: pgtype.Text{
					String: "productName",
					Status: pgtype.Present,
				},
			}
			err := s.CreateUpcomingBillItems(testCase.Ctx, db, &billItem)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
