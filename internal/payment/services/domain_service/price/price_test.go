package service

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	mockServices "github.com/manabie-com/backend/mock/payment/services/domain_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	FailCaseValidateFinalPriceError                          = "Fail case: Error when validate final price"
	FailCaseCheckPriceNormalBillItemError                    = "Fail case: Error when check price for normal bill item"
	FailCaseCheckPriceProRatingBillItemError                 = "Fail case: Error when check price for pro rating bill item"
	FailCaseValidateAdjustmentPriceError                     = "Fail case: Error when validate adjustment price"
	FailCaseGetProductPriceByIDAndScheduleIDError            = "Fail case: Error when get product price by product id and billing schedule period id"
	FailCaseGetProductPriceByIDAndScheduleIDAndQuantityError = "Fail case: Error when get product price by product id and billing schedule period id and quantity"
)

func TestPriceService_validateFinalPrice(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when final price is incorrect",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 95, 90),
			Req: &pb.BillingItem{
				ProductId: constant.ProductID,
				Price:     100,
				DiscountItem: &pb.DiscountBillItem{
					DiscountAmount: 10,
				},
				FinalPrice: 95,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.BillingItem{
				ProductId: constant.ProductID,
				Price:     100,
				DiscountItem: &pb.DiscountBillItem{
					DiscountAmount: 10,
				},
				FinalPrice: 90,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			testCase.Setup(testCase.Ctx)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			req := testCase.Req.(*pb.BillingItem)
			err := validateFinalPrice(req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_calculateOriginalDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				float32(10.0), pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(), float32(10.0),
			},
			ExpectedResp: float32(1.0),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			testCase.Setup(testCase.Ctx)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			originalPriceReq := testCase.Req.([]interface{})[0].(float32)
			discountTypeReq := testCase.Req.([]interface{})[1].(string)
			discountAmountValueReq := testCase.Req.([]interface{})[2].(float32)
			originalDiscount := calculateOriginalDiscount(originalPriceReq, discountTypeReq, discountAmountValueReq)

			assert.Equal(t, testCase.ExpectedResp, originalDiscount)

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_validateAdjustmentPrice(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)
	price := pgtype.Numeric{}
	discountAmount := pgtype.Numeric{}
	discountAmountValue := pgtype.Numeric{}
	_ = price.Set(1000)
	_ = discountAmount.Set(50)
	_ = discountAmountValue.Set(100)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when missing adjustment price",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: nil,
					ProductId:       constant.ProductID,
				},
				entities.BillItem{},
				int32(1),
				int32(1),
			},
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when adjustment price for update of product is incorrect",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{
						Value: 5,
					},
					ProductId: constant.ProductID,
					Price:     float32(16),
				},
				entities.BillItem{
					Price: pgtype.Numeric{
						Int:    big.NewInt(10),
						Status: pgtype.Present,
					},
				},
				int32(1),
				int32(1),
			},
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Incorrect adjustment price for update of product %v actual = %v vs expect = %v",
				constant.ProductID,
				float32(5),
				float32(6),
			),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: No discount",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{
						Value: 6,
					},
					ProductId: constant.ProductID,
					Price:     float32(16),
				},
				entities.BillItem{
					Price: pgtype.Numeric{
						Int:    big.NewInt(10),
						Status: pgtype.Present,
					},
				},
				int32(1),
				int32(1),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: billingRatioNumerator/billingRatioDenominator = 0",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{
						Value: 0,
					},
					ProductId: constant.ProductID,
					Price:     float32(16),
				},
				entities.BillItem{
					Price: pgtype.Numeric{
						Int:    big.NewInt(10),
						Status: pgtype.Present,
					},
				},
				int32(0),
				int32(0),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: Have old discount amount",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{
						Value: -2,
					},
					ProductId: constant.ProductID,
					Price:     float32(6),
				},
				entities.BillItem{
					Price: pgtype.Numeric{
						Int:    big.NewInt(10),
						Status: pgtype.Present,
					},
					DiscountAmount: pgtype.Numeric{
						Status: pgtype.Present,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
						Status: pgtype.Present,
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				int32(1),
				int32(1),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: Have discounts for old billing item and new billing item ",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{
						Value: float32(8),
					},
					ProductId: constant.ProductID,
					Price:     float32(20),
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
						DiscountAmountValue: 20,
					},
				},
				entities.BillItem{
					Price: pgtype.Numeric{
						Int:    big.NewInt(10),
						Status: pgtype.Present,
					},
					DiscountAmount: pgtype.Numeric{
						Status: pgtype.Present,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
						Status: pgtype.Present,
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				int32(1),
				int32(1),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: for debug ",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{
						Value: float32(50),
					},
					ProductId: constant.ProductID,
					Price:     float32(500),
				},
				entities.BillItem{
					Price:          price,
					DiscountAmount: discountAmount,
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
						Status: pgtype.Present,
					},
					DiscountAmountValue: discountAmountValue,
					RawDiscountAmount:   discountAmount,
				},
				int32(2),
				int32(4),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			testCase.Setup(testCase.Ctx)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			billingItem := testCase.Req.([]interface{})[0].(*pb.BillingItem)
			oldBillItem := testCase.Req.([]interface{})[1].(entities.BillItem)
			billingRatioNumerator := testCase.Req.([]interface{})[2].(int32)
			billingRatioDenominator := testCase.Req.([]interface{})[3].(int32)
			err := validateAdjustmentPrice(billingItem, oldBillItem, billingRatioNumerator, billingRatioDenominator)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_validateAdjustmentPriceForCancelOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when missing adjustment price",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: nil,
					ProductId:       constant.ProductID,
				},
				entities.BillItem{},
				int32(1),
				int32(1),
			},
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when adjustment price for cancel of product is incorrect",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{Value: float32(15)},
					ProductId:       constant.ProductID,
				},
				entities.BillItem{
					FinalPrice: pgtype.Numeric{
						Int:    big.NewInt(12),
						Status: pgtype.Present,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(12),
						Status: pgtype.Present,
					},
				},
				int32(1),
				int32(1),
			},
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Incorrect adjustment price for cancel of product %v actual = %v vs expect = %v",
				constant.ProductID,
				15,
				-12,
			),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					AdjustmentPrice: &wrapperspb.FloatValue{Value: float32(-12)},
					ProductId:       constant.ProductID,
				},
				entities.BillItem{
					FinalPrice: pgtype.Numeric{
						Int:    big.NewInt(12),
						Status: pgtype.Present,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(12),
						Status: pgtype.Present,
					},
				},
				int32(1),
				int32(1),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			testCase.Setup(testCase.Ctx)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			billingItem := testCase.Req.([]interface{})[0].(*pb.BillingItem)
			oldBillItem := testCase.Req.([]interface{})[1].(entities.BillItem)
			billingRatioNumerator := testCase.Req.([]interface{})[2].(int32)
			billingRatioDenominator := testCase.Req.([]interface{})[3].(int32)
			err := validateAdjustmentPriceForCancelOrder(billingItem, oldBillItem, billingRatioNumerator, billingRatioDenominator)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkQuantityOfOrderItemAndBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when billing item quantity is empty",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:  nil,
						ProductId: constant.ProductID,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when quantity is incorrect",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:  wrapperspb.Int32(2),
						ProductId: constant.ProductID,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:  wrapperspb.Int32(1),
						ProductId: constant.ProductID,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			billingItemData := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			err := s.checkQuantityOfOrderItemAndBillItem(orderItemData, billingItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_getPriceOFQuantityProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when product price by product id and quantity",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
				constant.ProductID,
				constant.ErrDefault,
			),
			Req: []interface{}{
				utils.PackageInfo{
					Quantity: int32(1),
					Package:  entities.Package{PackageID: pgtype.Text{String: constant.ProductID}},
				},
				pb.ProductPriceType_DEFAULT_PRICE.String(),
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.PackageInfo{
					Quantity: int32(1),
					Package:  entities.Package{PackageID: pgtype.Text{String: constant.ProductID}},
				},
				pb.ProductPriceType_DEFAULT_PRICE.String(),
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			packageReq := testCase.Req.([]interface{})[0].(utils.PackageInfo)
			priceTypeReq := testCase.Req.([]interface{})[1].(string)
			_, err := s.getPricesOfQuantityProduct(testCase.Ctx, db, packageReq, priceTypeReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_getPricesOfNoneQuantityProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get product prices of product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
				constant.ProductID,
				constant.ErrDefault,
			),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{
					String: constant.ProductID,
				}},
				pb.ProductPriceType_DEFAULT_PRICE.String(),
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndPriceType", ctx, db, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Product{ProductID: pgtype.Text{
					String: constant.ProductID,
				}},
				pb.ProductPriceType_DEFAULT_PRICE.String(),
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndPriceType", ctx, db, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			productReq := testCase.Req.([]interface{})[0].(entities.Product)
			priceTypeReq := testCase.Req.([]interface{})[1].(string)
			_, err := s.getPricesOfNoneQuantityProduct(testCase.Ctx, db, productReq, priceTypeReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkPriceForNoneQuantityProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when price of product does not exist",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.ProductPriceNotExistOrUpdated,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Price of product with id %v does not exist in system (or just updated)", constant.ProductID),
				},
			),
			Req: []interface{}{
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     float32(1),
				},
				[]entities.ProductPrice{
					{
						Price: pgtype.Numeric{
							Int:    &big.Int{},
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseValidateFinalPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 95, 90),
			Req: []interface{}{
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     100,
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmount: 10,
					},
					FinalPrice: 95,
				},
				[]entities.ProductPrice{
					{
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     100,
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmount: 10,
					},
					FinalPrice: 90,
				},
				[]entities.ProductPrice{
					{
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			billItemReq := testCase.Req.([]interface{})[0].(*pb.BillingItem)
			productPricesReq := testCase.Req.([]interface{})[1].([]entities.ProductPrice)
			err := s.checkPriceForNoneQuantityProduct(billItemReq, productPricesReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_validatePriceForOneTimeNonQuantityProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get multi prices of none quantity product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
				constant.ProductID,
				constant.ErrDefault,
			),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
				},
				&pb.BillingItem{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndPriceType", ctx, db, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case:Error when check price for none quantity product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.ProductPriceNotExistOrUpdated,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Price of product with id %v does not exist in system (or just updated)", constant.ProductID),
				},
			),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     float32(1),
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndPriceType", ctx, db, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{
					Price: pgtype.Numeric{
						Int:    &big.Int{},
						Status: pgtype.Present,
					},
				}}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     100,
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmount: 10,
					},
					FinalPrice: 90,
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndPriceType", ctx, db, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					}}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			billingItem := testCase.Req.([]interface{})[1].(*pb.BillingItem)
			err := s.validatePriceForOneTimeNonQuantityProduct(testCase.Ctx, db, orderItemData, billingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_validatePriceForOneTimeQuantityProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when quantity is empty",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Quantity:  nil,
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when quantities is inconsistency",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal,
				"inconsistency quantity between order item and bill item of product %v, order item quantity %v vs bill item quantity %v",
				constant.ProductID,
				1,
				2,
			),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					PackageInfo: utils.PackageInfo{Quantity: 1},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Quantity:  wrapperspb.Int32(2),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when get price of quantity product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
				constant.ProductID,
				constant.ErrDefault,
			),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
						Package: entities.Package{
							PackageID: pgtype.Text{
								String: constant.ProductID,
							},
						},
					},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Quantity:  wrapperspb.Int32(1),
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseValidateFinalPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 95, 90),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
						Package: entities.Package{
							PackageID: pgtype.Text{
								String: constant.ProductID,
							},
						},
					},
				},
				&pb.BillingItem{
					Quantity:  wrapperspb.Int32(1),
					ProductId: constant.ProductID,
					Price:     100,
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmount: 10,
					},
					FinalPrice: 95,
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
						Package: entities.Package{
							PackageID: pgtype.Text{
								String: constant.ProductID,
							},
						},
					},
				},
				&pb.BillingItem{
					Quantity:  wrapperspb.Int32(1),
					ProductId: constant.ProductID,
					Price:     100,
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmount: 10,
					},
					FinalPrice: 90,
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
						Package: entities.Package{
							PackageID: pgtype.Text{
								String: constant.ProductID,
							},
						},
					},
				},
				&pb.BillingItem{
					Quantity:  wrapperspb.Int32(1),
					ProductId: constant.ProductID,
					Price:     100,
					DiscountItem: &pb.DiscountBillItem{
						DiscountAmount: 10,
					},
					FinalPrice: 90,
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			billingItem := testCase.Req.([]interface{})[1].(*pb.BillingItem)
			err := s.validatePriceForOneTimeQuantityProduct(testCase.Ctx, db, orderItemData, billingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_IsValidPriceForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when validate price for one time non quantity product (product_type != ProductType_PRODUCT_TYPE_PACKAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPrice,
				constant.ProductID,
				constant.ErrDefault,
			),
			Req: utils.OrderItemData{
				ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
				ProductInfo: entities.Product{ProductID: pgtype.Text{
					String: constant.ProductID,
				}},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndPriceType", ctx, db, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when validate price for one time quantity product (product_type != ProductType_PRODUCT_TYPE_PACKAGE)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: utils.OrderItemData{
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
				ProductInfo: entities.Product{ProductID: pgtype.Text{
					String: constant.ProductID,
				}},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							Quantity:  nil,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: utils.OrderItemData{
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
				ProductInfo: entities.Product{ProductID: pgtype.Text{
					String: constant.ProductID,
				}},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							Quantity:  nil,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)
			err := s.IsValidPriceForOneTimeBilling(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkPriceForProRatingBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when product price is incorrect",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice,
				constant.ProductID,
				float32(100),
				75,
			),
			Req: []interface{}{
				entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 3,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 4,
					},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     float32(100),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseValidateFinalPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 95, 100),
			Req: []interface{}{
				entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&pb.BillingItem{
					ProductId:  constant.ProductID,
					Price:      float32(100),
					FinalPrice: float32(95),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&pb.BillingItem{
					ProductId:  constant.ProductID,
					Price:      float32(100),
					FinalPrice: float32(100),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			productPrice := testCase.Req.([]interface{})[0].(entities.ProductPrice)
			ratio := testCase.Req.([]interface{})[1].(entities.BillingRatio)
			billItem := testCase.Req.([]interface{})[2].(*pb.BillingItem)

			err := s.checkPriceForProRatingBillItem(productPrice, ratio, billItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkPriceRecurringNoneQuantityBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when check price for pro rating billing item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 95, 100),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              95,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: "Fail case: Error when check price for pro rating billing item (validate final price)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice,
				constant.ProductID,
				90,
				100,
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   90,
						FinalPrice:              95,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when check price for pro rating billing item (validate final price)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting price of none quantity recurring product have have err %v", constant.ErrDefault),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCheckPriceNormalBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice, constant.ProductID, float32(100), float32(80)),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: float32(100),
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(80),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			proRatedBillItem := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[1].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[2].([]utils.BillingItemData)
			orderItemData := testCase.Req.([]interface{})[3].(utils.OrderItemData)
			_, err := s.checkPriceRecurringNoneQuantityBilling(testCase.Ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkPriceForNormalBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when product price is incorrect",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice, constant.ProductID, 100, 80),
			Req: []interface{}{
				entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(80),
						Status: pgtype.Present,
					},
				},
				&pb.BillingItem{
					ProductId: constant.ProductID,
					Price:     float32(100),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseValidateFinalPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 90, 80),
			Req: []interface{}{
				entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(80),
						Status: pgtype.Present,
					},
				},
				&pb.BillingItem{
					ProductId:  constant.ProductID,
					Price:      float32(80),
					FinalPrice: float32(90),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: FailCaseValidateFinalPriceError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(80),
						Status: pgtype.Present,
					},
				},
				&pb.BillingItem{
					ProductId:  constant.ProductID,
					Price:      float32(80),
					FinalPrice: float32(80),
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			productPrice := testCase.Req.([]interface{})[0].(entities.ProductPrice)
			billItem := testCase.Req.([]interface{})[1].(*pb.BillingItem)
			err := s.checkPriceForNormalBillItem(productPrice, billItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_IsValidAdjustmentPriceForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseValidateAdjustmentPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface {
			}{
				entities.BillItem{},
				utils.OrderItemData{
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:       constant.ProductID,
								AdjustmentPrice: nil,
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.BillItem{
					Price: pgtype.Numeric{
						Int:    big.NewInt(10),
						Status: pgtype.Present,
					},
					DiscountAmount: pgtype.Numeric{
						Status: pgtype.Present,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
						Status: pgtype.Present,
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				utils.OrderItemData{
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId: constant.ProductID,
								AdjustmentPrice: &wrapperspb.FloatValue{
									Value: float32(-8),
								},
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			oldBillItem := testCase.Req.([]interface{})[0].(entities.BillItem)
			orderItemData := testCase.Req.([]interface{})[1].(utils.OrderItemData)
			err := s.IsValidAdjustmentPriceForOneTimeBilling(oldBillItem, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkPriceRecurringQuantityBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when check price recurring quantity billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
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
					OrderItem:              nil,
					BillItems:              nil,
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:  nil,
						ProductId: constant.ProductID,
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDAndQuantityError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProRatingPriceOfNoneQuantity, constant.ErrDefault),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:                wrapperspb.Int32(1),
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Quantity:  wrapperspb.Int32(1),
							ProductId: constant.ProductID,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: FailCaseCheckPriceProRatingBillItemError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice,
				constant.ProductID,
				90,
				100,
			),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:                wrapperspb.Int32(1),
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(90),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Quantity:  wrapperspb.Int32(1),
							ProductId: constant.ProductID,
							Price:     float32(100),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseCheckPriceProRatingBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:                wrapperspb.Int32(1),
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(110),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
							Quantity:   nil,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDAndQuantityError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:                wrapperspb.Int32(1),
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(110),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
							Quantity:   &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
					Quantity: pgtype.Int4{
						Int:    1,
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCheckPriceNormalBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice, constant.ProductID, float32(100), float32(110)),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:                wrapperspb.Int32(1),
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(110),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
							Quantity:   &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
					Quantity: pgtype.Int4{
						Int:    1,
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Quantity:                wrapperspb.Int32(1),
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(110),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
							Quantity:   &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
					Quantity: pgtype.Int4{
						Int:    1,
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			proRatedBillItem := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[3].([]utils.BillingItemData)
			_, err := s.checkPriceRecurringQuantityBilling(testCase.Ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_IsValidPriceForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when check price recurring none quantity billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:               constant.ProductID,
								BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "1"},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when check price recurring quantity billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:               constant.ProductID,
								BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "1"},
								Quantity:                nil,
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: product_type = ProductType_PRODUCT_TYPE_PACKAGE",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{Quantity: 1},
					ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Quantity:                wrapperspb.Int32(1),
								ProductId:               constant.ProductID,
								BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
								Price:                   float32(100),
								FinalPrice:              float32(100),
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
						Quantity:                wrapperspb.Int32(1),
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
							Quantity:   &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
					Quantity: pgtype.Int4{
						Int:    1,
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
					Quantity: pgtype.Int4{
						Int:    1,
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: product_type != ProductType_PRODUCT_TYPE_PACKAGE",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
					ProductInfo: entities.Product{ProductID: pgtype.Text{
						String: constant.ProductID,
					}},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:               constant.ProductID,
								BillingSchedulePeriodId: &wrapperspb.StringValue{Value: "1"},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   100,
						FinalPrice:              100,
					},
				},
				entities.BillingRatio{
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			proRatedBillItem := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[3].([]utils.BillingItemData)
			_, err := s.IsValidPriceForRecurringBilling(testCase.Ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkUpdatePriceRecurringNoneQuantityBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(110),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
				map[string]entities.BillItem{
					"": {},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCheckPriceProRatingBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 110, 100),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
				map[string]entities.BillItem{
					"": {},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseValidateAdjustmentPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice:         nil,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(92),
						},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 20,
							},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
						DiscountAmount: pgtype.Numeric{
							Status: pgtype.Present,
						},
						DiscountAmountType: pgtype.Text{
							String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
							Status: pgtype.Present,
						},
						DiscountAmountValue: pgtype.Numeric{
							Int:    big.NewInt(20),
							Status: pgtype.Present,
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{},
					constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCheckPriceNormalBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 90, 100),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(92),
						},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(90),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 20,
							},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
						DiscountAmount: pgtype.Numeric{
							Status: pgtype.Present,
						},
						DiscountAmountType: pgtype.Text{
							String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
							Status: pgtype.Present,
						},
						DiscountAmountValue: pgtype.Numeric{
							Int:    big.NewInt(20),
							Status: pgtype.Present,
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				},
					nil)
			},
		},
		{
			Name:        FailCaseValidateAdjustmentPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(92),
						},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 20,
							},
							Quantity: nil,
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
						DiscountAmount: pgtype.Numeric{
							Status: pgtype.Present,
						},
						DiscountAmountType: pgtype.Text{
							String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
							Status: pgtype.Present,
						},
						DiscountAmountValue: pgtype.Numeric{
							Int:    big.NewInt(20),
							Status: pgtype.Present,
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			proRatedBillItem := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[1].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[3].([]utils.BillingItemData)
			mapOldBillingItem := testCase.Req.([]interface{})[4].(map[string]entities.BillItem)
			orderItemData := testCase.Req.([]interface{})[5].(utils.OrderItemData)

			_, err := s.checkUpdatePriceRecurringNoneQuantityBilling(testCase.Ctx, db, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapOldBillingItem, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_checkUpdatePriceRecurringQuantityBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when check quantity of order item and bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(110),
						FinalPrice:              float32(110),
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
				map[string]entities.BillItem{
					"": {},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDAndQuantityError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(110),
						Quantity: &wrapperspb.Int32Value{
							Value: 1,
						},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price:      float32(100),
							FinalPrice: float32(100),
						},
					},
				},
				map[string]entities.BillItem{
					"": {},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: FailCaseCheckPriceProRatingBillItemError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectProductPrice,
				constant.ProductID,
				100,
				120,
			),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(120),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseValidateAdjustmentPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice:         nil,
						Quantity:                &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when check quantity of order item and bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
							Quantity: nil,
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseGetProductPriceByIDAndScheduleIDAndQuantityError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCheckPriceNormalBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.IncorrectFinalPrice, constant.ProductID, 90, 100),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(90),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseValidateAdjustmentPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice:      float32(100),
							ProductId:       constant.ProductID,
							Price:           float32(100),
							AdjustmentPrice: nil,
							Quantity:        &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        FailCaseValidateAdjustmentPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice:      float32(100),
							ProductId:       constant.ProductID,
							Price:           float32(100),
							AdjustmentPrice: nil,
							Quantity:        &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			proRatedBillItem := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[4].([]utils.BillingItemData)
			mapOldBillingItem := testCase.Req.([]interface{})[5].(map[string]entities.BillItem)

			_, err := s.checkUpdatePriceRecurringQuantityBilling(testCase.Ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapOldBillingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_IsValidPriceForCancelRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo

		periodStartDate        = time.Now().Add(-1 * time.Hour)
		periodEndDate          = time.Now().Add(24 * time.Hour)
		orderItemEffectiveDate = timestamppb.Now()
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when validate adjustment price for cancel order (pro rated bill item)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice:         nil,
						Quantity:                &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
				map[string]entities.BillingSchedulePeriod{
					constant.BillingSchedulePeriodID: {
						BillingSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						Name:              pgtype.Text{},
						BillingScheduleID: pgtype.Text{},
						StartDate:         pgtype.Timestamptz{},
						EndDate:           pgtype.Timestamptz{},
						BillingDate:       pgtype.Timestamptz{},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when adjustment price is incorrect",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "adjustment price of this period %v should be 0 with start date: %v, end date: %v, effective date: %v, adjustment price: %v",
				constant.BillingSchedulePeriodID,
				periodStartDate.String(),
				periodEndDate.String(),
				orderItemEffectiveDate.AsTime(),
				float32(-100),
			),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					IsDisableProRatingFlag: true,
					OrderItem:              &pb.OrderItem{EffectiveDate: orderItemEffectiveDate},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(-100),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(-100),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
						FinalPrice: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
				map[string]entities.BillingSchedulePeriod{
					constant.BillingSchedulePeriodID: {
						BillingSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						Name:              pgtype.Text{},
						BillingScheduleID: pgtype.Text{},
						StartDate: pgtype.Timestamptz{
							Time: periodStartDate,
						},
						EndDate: pgtype.Timestamptz{
							Time: periodEndDate,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when validate adjustment price for cancel order (normal bill item)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.MissingAdjustmentPriceWhenUpdatingOrder, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					IsDisableProRatingFlag: false,
					OrderItem:              &pb.OrderItem{EffectiveDate: orderItemEffectiveDate},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(-100),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice:      float32(100),
							ProductId:       constant.ProductID,
							Price:           float32(100),
							AdjustmentPrice: nil,
							Quantity:        &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
						FinalPrice: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
				map[string]entities.BillingSchedulePeriod{
					constant.BillingSchedulePeriodID: {
						BillingSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						Name:              pgtype.Text{},
						BillingScheduleID: pgtype.Text{},
						StartDate: pgtype.Timestamptz{
							Time: periodStartDate,
						},
						EndDate: pgtype.Timestamptz{
							Time: periodEndDate,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					IsDisableProRatingFlag: false,
					OrderItem:              &pb.OrderItem{EffectiveDate: orderItemEffectiveDate},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(-100),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(-100),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
						FinalPrice: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
				map[string]entities.BillingSchedulePeriod{
					constant.BillingSchedulePeriodID: {
						BillingSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						Name:              pgtype.Text{},
						BillingScheduleID: pgtype.Text{},
						StartDate: pgtype.Timestamptz{
							Time: periodStartDate,
						},
						EndDate: pgtype.Timestamptz{
							Time: periodEndDate,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			proRatedBillItem := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[3].([]utils.BillingItemData)
			mapOldBillingItem := testCase.Req.([]interface{})[4].(map[string]entities.BillItem)
			mapPeriodInfo := testCase.Req.([]interface{})[5].(map[string]entities.BillingSchedulePeriod)
			err := s.IsValidPriceForCancelRecurringBilling(orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapOldBillingItem, mapPeriodInfo)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_IsValidPriceForUpdateRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when check update price recurring none quantity billing (product_type != ProductType_PRODUCT_TYPE_PACKAGE)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				},
				utils.BillingItemData{
					BillingItem: nil,
				},
				&entities.ProductPrice{},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(-100),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
						FinalPrice: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: product_type != ProductType_PRODUCT_TYPE_PACKAGE",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
					},
				},
				&entities.ProductPrice{},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when check update price recurring quantity billing (product_type == ProductType_PRODUCT_TYPE_PACKAGE)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, constant.ErrorWhenGettingProductPriceWithEmptyQuantity, constant.ProductID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
				},
				utils.BillingItemData{
					BillingItem: nil,
				},
				&entities.ProductPrice{},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(-100),
							},
							Quantity: nil,
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
						FinalPrice: pgtype.Numeric{
							Int:    big.NewInt(100),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case: product_type == ProductType_PRODUCT_TYPE_PACKAGE",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					PackageInfo: utils.PackageInfo{
						Quantity: 1,
					},
					ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId:               constant.ProductID,
						BillingSchedulePeriodId: &wrapperspb.StringValue{Value: constant.BillingSchedulePeriodID},
						Price:                   float32(100),
						FinalPrice:              float32(100),
						AdjustmentPrice: &wrapperspb.FloatValue{
							Value: float32(90),
						},
						Quantity: &wrapperspb.Int32Value{Value: 1},
					},
				},
				&entities.ProductPrice{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					BillingSchedulePeriodID: pgtype.Text{
						String: constant.BillingSchedulePeriodID,
					},
					Price: pgtype.Numeric{
						Int:    big.NewInt(110),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 1,
					},
				},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							FinalPrice: float32(100),
							ProductId:  constant.ProductID,
							Price:      float32(100),
							AdjustmentPrice: &wrapperspb.FloatValue{
								Value: float32(90),
							},
							Quantity: &wrapperspb.Int32Value{Value: 1},
						},
					},
				},
				map[string]entities.BillItem{
					constant.BillingSchedulePeriodID: {
						BillSchedulePeriodID: pgtype.Text{
							String: constant.BillingSchedulePeriodID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						BillingRatioNumerator: pgtype.Int4{
							Int: 1,
						},
						BillingRatioDenominator: pgtype.Int4{
							Int: 1,
						},
						Price: pgtype.Numeric{
							Int:    big.NewInt(10),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ProductPrice{
					Price: pgtype.Numeric{
						Int:    big.NewInt(100),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			s := &PriceService{
				productPriceRepo: productPriceRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			proRatedBillItem := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[3].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[4].([]utils.BillingItemData)
			mapOldBillingItem := testCase.Req.([]interface{})[5].(map[string]entities.BillItem)
			_, err := s.IsValidPriceForUpdateRecurringBilling(testCase.Ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapOldBillingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}

func TestPriceService_CalculatorBillItemPrice(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db               *mockDb.Ext
		productPriceRepo *mockRepositories.MockProductPriceRepo
		discountService  *mockServices.IDiscountServiceForProductPrice
		taxService       *mockServices.ITaxServiceForProductPrice
	)

	testcases := []utils.TestCase{
		{
			Name: "Happy case: product_type == ProductType_PRODUCT_TYPE_PACKAGE",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  nil,
			Setup: func(ctx context.Context) {
				discountService.On("CalculatorDiscountPrice", mock.Anything, mock.Anything, mock.Anything).Return(float32(100), nil)
				taxService.On("CalculatorTaxPrice", mock.Anything, mock.Anything, mock.Anything).Return(float32(110), nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{Price: pgtype.Numeric{
					Int:    big.NewInt(20),
					Status: pgtype.Present,
				}}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productPriceRepo = &mockRepositories.MockProductPriceRepo{}
			taxService = new(mockServices.ITaxServiceForProductPrice)
			discountService = new(mockServices.IDiscountServiceForProductPrice)
			s := &PriceService{
				productPriceRepo: productPriceRepo,
				DiscountService:  discountService,
				TaxService:       taxService,
			}
			testCase.Setup(testCase.Ctx)

			billItem := entities.BillItem{
				OrderID: pgtype.Text{
					String: "order-id",
					Status: pgtype.Present,
				},
			}
			upcomingBillItem := entities.UpcomingBillItem{
				OrderID: pgtype.Text{
					String: "order-id",
					Status: pgtype.Present,
				},
			}
			tax := entities.Tax{
				TaxID: pgtype.Text{String: "tax-id", Status: pgtype.Present},
				TaxPercentage: pgtype.Int4{
					Int:    10,
					Status: pgtype.Present,
				},
				TaxCategory: pgtype.Text{
					String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					Status: pgtype.Present,
				},
			}
			discount := entities.Discount{
				DiscountID: pgtype.Text{
					String: "discount-id",
					Status: pgtype.Present,
				},
				DiscountType: pgtype.Text{
					String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					Status: pgtype.Present,
				},
				DiscountAmountType: pgtype.Text{
					String: "discount-amount-type",
					Status: pgtype.Present,
				},
				DiscountAmountValue: pgtype.Numeric{
					Int:    big.NewInt(int64(10)),
					Exp:    -2,
					Status: pgtype.Present,
				},
			}
			priceType := pb.ProductPriceType_ENROLLED_PRICE.String()
			quantityType := pb.QuantityType_QUANTITY_TYPE_SLOT.String()
			courseSlot := int32(2)
			billItemDescription := entities.BillingItemDescription{
				ProductType:   pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
				QuantityType:  &quantityType,
				BillingPeriod: nil,
				BillingRatio:  nil,
				CourseItems: []*entities.CourseItem{
					{
						Weight: nil,
						Slot:   &courseSlot,
					},
					{
						Slot: &courseSlot,
					},
				},
			}
			billingSchedulePeriod := entities.BillingSchedulePeriod{
				BillingScheduleID: pgtype.Text{
					String: "billing-schedule-id",
					Status: pgtype.Present,
				},
			}
			err := s.CalculatorBillItemPrice(testCase.Ctx, db, &billItem, upcomingBillItem, tax, discount, priceType, &billItemDescription, billingSchedulePeriod)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}
