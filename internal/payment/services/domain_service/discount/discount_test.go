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
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	FailCaseCheckCorrectDiscountError = "Fail case: Error when check correct discount info between discount entities and bill item"
)

func TestDiscountService_getDiscountAndCheckProductDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
		userDiscountTagRepo *mockRepositories.MockUserDiscountTagRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get discount by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         utils.OrderItemData{OrderItem: &pb.OrderItem{DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID}}},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get product discount by product id and discount id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal,
				"Product %v and discount %v have non-association",
				constant.ProductID,
				constant.DiscountID,
			),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID}},
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get user discount tag by id and discount tag id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal,
				"Product %v and discount %v have non-association",
				constant.ProductID,
				constant.DiscountID,
			),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID}},
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountTagID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				}, nil)
				userDiscountTagRepo.On("GetDiscountTagByUserIDAndDiscountTagID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.UserDiscountTag{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID}},
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			userDiscountTagRepo = &mockRepositories.MockUserDiscountTagRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
				userDiscountTagRepo: userDiscountTagRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)

			_, err := s.getDiscountAndCheckProductDiscount(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_checkCorrectDiscountAmountBetweenDiscountEntitiesAndProRatingBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when discount amount is wrong (billing_ratio_numerator == 0)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 1, float32(0))},
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{DiscountAmount: 1},
					},
				},
				entities.Discount{},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 0,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case (billing_ratio_numerator == 0)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{DiscountAmount: 0},
					},
				},
				entities.Discount{},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 0,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when discount amount is wrong (discount_amount_type == DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 25, 20)},
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      25,
							DiscountAmountValue: 20,
						},
						Price: float32(100),
					},
				},
				entities.Discount{
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 4,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case(discount_amount_type == DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
						Price: float32(100),
					},
				},
				entities.Discount{
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 4,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when discount amount is wrong (discount_amount_type != DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 20, 5)},
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
						Price: float32(100),
					},
				},
				entities.Discount{
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 4,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case (discount_amount_type != DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      5,
							DiscountAmountValue: 20,
						},
						Price: float32(100),
					},
				},
				entities.Discount{
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 1,
					},
					BillingRatioDenominator: pgtype.Int4{
						Int: 4,
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
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			billItem := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			discountEntities := testCase.Req.([]interface{})[1].(entities.Discount)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)

			err := s.checkCorrectDiscountAmountBetweenDiscountEntitiesAndProRatingBillItem(billItem, discountEntities, ratioOfProRatedBillingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_checkCorrectDiscountInfoBetweenDiscountEntitiesAndBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when discount is not available",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.DiscountIsNotAvailable),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount: 1,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when discount type is changed",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with id %v change discount type from %s to %s",
				constant.ProductID,
				pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
				pb.DiscountType_DISCOUNT_TYPE_COMBO.String()),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount: 1,
							DiscountType:   pb.DiscountType_DISCOUNT_TYPE_REGULAR,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_COMBO.String(),
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when discount amount type is changed",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with id %v change discount amount type from %s to %s",
				constant.ProductID,
				pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:     1,
							DiscountType:       pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when discount amount value is changed",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with id %v change discount amount value from %s to %s",
				constant.ProductID,
				big.NewInt(120),
				fmt.Sprintf("%v", 100)),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      1,
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: float32(100),
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(120),
						Status: pgtype.Present,
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
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      1,
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: float32(120),
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(120),
						Status: pgtype.Present,
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
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			billItem := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			discountEntities := testCase.Req.([]interface{})[1].(entities.Discount)

			err := s.checkCorrectDiscountInfoBetweenDiscountEntitiesAndBillItem(billItem, discountEntities)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_checkCorrectDiscountAmountBetweenDiscountEntitiesAndBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when discount is wrong (discount_amount_type == DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 10, 20)},
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      10,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when discount is wrong (discount_amount_type != DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 10, 20)},
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount: 10,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case (discount_amount_type == DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case (discount_amount_type != DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountAmount: 20,
						},
					},
				},
				entities.Discount{
					AvailableFrom: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
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
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			billItem := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			discountEntities := testCase.Req.([]interface{})[1].(entities.Discount)

			err := s.checkCorrectDiscountAmountBetweenDiscountEntitiesAndBillItem(billItem, discountEntities)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_isDiscountOfNormalBillItemsValid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when discount between bill item and order item is consistency",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.InconsistentDiscountID),
			Req: []interface{}{
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price: float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
							},
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: "constant.DiscountID_1"},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseCheckCorrectDiscountError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.DiscountIsNotAvailable),
			Req: []interface{}{
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price:     float32(100),
							ProductId: constant.ProductID,
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      1,
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: float32(120),
							},
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: constant.DiscountID},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Null,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(120),
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when check correct discount amount between discount entities and bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 100, 20)},
			),
			Req: []interface{}{
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price:     float32(100),
							ProductId: constant.ProductID,
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      100,
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: float32(20),
							},
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: constant.DiscountID},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
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
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price:     float32(100),
							ProductId: constant.ProductID,
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: float32(20),
							},
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: constant.DiscountID},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
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
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			normalBillItems := testCase.Req.([]interface{})[0].([]utils.BillingItemData)
			discountEntities := testCase.Req.([]interface{})[1].(entities.Discount)

			err := s.isDiscountOfNormalBillItemsValid(normalBillItems, discountEntities)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_isRecurringValidDurationValid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when bill item should not have a discount because order item does not have",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This bill item should not have a discount because order item does not have a discount id"),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: nil,
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
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
			Name: "Fail case: Error when get discount and check product discount",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal,
				"Error when get discount of product %v with error %s",
				constant.ProductID,
				constant.ErrDefault),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: recurring_valid_duration status is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{Status: pgtype.Null},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when maximum discount is reached",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Maximum discount is reached"),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    0,
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
				OrderItem: &pb.OrderItem{
					DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: "Happy case (discount_item of order_item, bill_item is empty)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					DiscountId: nil,
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: nil,
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
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)

			_, err := s.isRecurringValidDurationValid(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_isDiscountOfProRatingBillItemsValid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when discount between bill item and order item is consistency",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.InconsistentDiscountID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: "constant.DiscountID_1"},
				},
				entities.BillingRatio{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseCheckCorrectDiscountError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.DiscountIsNotAvailable),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price:     float32(100),
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      1,
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: float32(120),
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: constant.DiscountID},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Null,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(120),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when check correct discount amount between discount entities and pro rating bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 100, 0)},
			),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price:     float32(100),
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      100,
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: float32(20),
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: constant.DiscountID},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 0,
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
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price:     float32(100),
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      0,
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: float32(20),
						},
					},
				},
				entities.Discount{
					DiscountID: pgtype.Text{String: constant.DiscountID},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				},
				entities.BillingRatio{
					BillingRatioNumerator: pgtype.Int4{
						Int: 0,
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
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			proRatingBillItem := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			discountEntities := testCase.Req.([]interface{})[1].(entities.Discount)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)

			err := s.isDiscountOfProRatingBillItemsValid(proRatingBillItem, discountEntities, ratioOfProRatedBillingItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_IsValidDiscountForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo

		discountName = "constant.Discount name"
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when recurring duration is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This bill item should not have a discount because order item does not have a discount id"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						DiscountId: nil,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								DiscountItem: &pb.DiscountBillItem{
									DiscountId: constant.DiscountID,
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when assign discount name have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "assigning discount name have error "),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price: float32(100),
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountAmountValue: 20,
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when assign discount name have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "assigning discount name have error "),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price: float32(100),
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountAmountValue: 20,
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when discount of normal bill items is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.InconsistentDiscountID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price: float32(100),
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountAmountValue: 20,
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price: float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: "constant.DiscountID",
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when discount of pro rating bill items is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.InconsistentDiscountID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "constant.DiscountID",
							DiscountAmount:      20,
							DiscountAmountValue: 20,
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
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price: float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							},
							ProductId: constant.ProductID,
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price:     float32(100),
						ProductId: constant.ProductID,
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: float32(20),
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
				[]utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							Price: float32(100),
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountAmount:      20,
								DiscountAmountValue: 20,
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							},
							ProductId: constant.ProductID,
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: "Happy case (discount status is invalid)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price: float32(100),
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountAmountValue: 20,
								},
							},
						},
					},
				},
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						Price: float32(100),
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountAmount:      20,
							DiscountAmountValue: 20,
						},
					},
				},
				entities.BillingRatio{},
				[]utils.BillingItemData{
					{},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Null,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			proRatingBillItem := testCase.Req.([]interface{})[1].(utils.BillingItemData)
			ratioOfProRatedBillingItem := testCase.Req.([]interface{})[2].(entities.BillingRatio)
			normalBillItem := testCase.Req.([]interface{})[3].([]utils.BillingItemData)
			discountName := testCase.Req.([]interface{})[4].(*string)

			err := s.IsValidDiscountForRecurringBilling(testCase.Ctx, db, orderItemData, proRatingBillItem, ratioOfProRatedBillingItem, normalBillItem, discountName)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_IsValidDiscountForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		discountRepo        *mockRepositories.MockDiscountRepo
		productDiscountRepo *mockRepositories.MockProductDiscountRepo

		discountName = "constant.Discount name"
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get discount and check product discount",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when discount is not available",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.DiscountIsNotAvailable),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Status: pgtype.Null,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: FailCaseCheckCorrectDiscountError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with id %v change discount type from %s to %s",
				constant.ProductID,
				pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
				pb.DiscountType_DISCOUNT_TYPE_COMBO.String()),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_COMBO.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: "Fail case: Error when check correct discount amount between discount entities and bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DiscountAmountsAreNotEqual,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.DiscountAmountsAreNotEqualDebugMsg, 10, 20)},
			),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      10,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:     float32(100),
								ProductId: constant.ProductID,
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          constant.DiscountID,
									DiscountAmount:      20,
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: float32(20),
								},
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Discount{
					RecurringValidDuration: pgtype.Int4{
						Int:    2,
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					DiscountID: pgtype.Text{
						String: constant.DiscountID,
						Status: pgtype.Present,
					},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().Add(-1 * time.Hour),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().Add(1 * time.Hour),
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					},
					DiscountAmountValue: pgtype.Numeric{
						Int:    big.NewInt(20),
						Status: pgtype.Present,
					},
				}, nil)
				productDiscountRepo.On("GetByProductIDAndDiscountID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductDiscount{}, nil)
			},
		},
		{
			Name: "Happy case: When discount item of bill item is nil",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderItemData{
					ProductInfo: entities.Product{ProductID: pgtype.Text{String: constant.ProductID}},
					OrderItem: &pb.OrderItem{
						DiscountId: &wrapperspb.StringValue{Value: constant.DiscountID},
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								Price:        float32(100),
								ProductId:    constant.ProductID,
								DiscountItem: nil,
							},
						},
					},
				},
				&discountName,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountRepo = &mockRepositories.MockDiscountRepo{}
			productDiscountRepo = &mockRepositories.MockProductDiscountRepo{}
			s := &DiscountService{
				discountRepo:        discountRepo,
				productDiscountRepo: productDiscountRepo,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			discountNameReq := testCase.Req.([]interface{})[1].(*string)

			err := s.IsValidDiscountForOneTimeBilling(testCase.Ctx, db, orderItemDataReq, discountNameReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo, productDiscountRepo)
		})
	}
}

func TestDiscountService_GetDiscountsByDiscountIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		discountRepo *mockRepositories.MockDiscountRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when getting discounts by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Success case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Discount{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountRepo = new(mockRepositories.MockDiscountRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				discountRepo: discountRepo,
			}
			_, err := s.GetDiscountsByDiscountIDs(testCase.Ctx, db, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo)
		})
	}
}

func TestDiscountService_VerifyDiscountForGenerateUpcomingBillItem(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		discountRepo *mockRepositories.MockDiscountRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "success case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				discountRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountID: pgtype.Text{
						String: "discount-id",
						Status: pgtype.Present,
					},
					Name: pgtype.Text{
						String: "discount-name",
						Status: pgtype.Present,
					},
					DiscountType: pgtype.Text{
						String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
						Status: pgtype.Present,
					},
					DiscountAmountType: pgtype.Text{
						String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
						Status: pgtype.Present,
					},
					AvailableFrom: pgtype.Timestamptz{
						Time:   time.Now().AddDate(0, -1, 0),
						Status: pgtype.Present,
					},
					AvailableUntil: pgtype.Timestamptz{
						Time:   time.Now().AddDate(0, 1, 0),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountRepo = new(mockRepositories.MockDiscountRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				discountRepo: discountRepo,
			}
			_, err := s.VerifyDiscountForGenerateUpcomingBillItem(testCase.Ctx, db, []entities.BillItem{
				{
					DiscountID: pgtype.Text{
						String: "discount-id",
						Status: pgtype.Present,
					},
				},
			})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo)
		})
	}
}

func TestDiscountService_CalculatorDiscountPrice(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		discountRepo *mockRepositories.MockDiscountRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "success case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				// nothing to do
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountRepo = new(mockRepositories.MockDiscountRepo)
			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				discountRepo: discountRepo,
			}
			discount := entities.Discount{
				DiscountID: pgtype.Text{
					String: "discount_id",
					Status: pgtype.Present,
				},
				DiscountType: pgtype.Text{
					String: pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
					Status: pgtype.Present,
				},
				DiscountAmountType: pgtype.Text{
					String: pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
					Status: pgtype.Present,
				},
				DiscountAmountValue: pgtype.Numeric{
					Int:    big.NewInt(10),
					Status: pgtype.Present,
				},
				AvailableFrom: pgtype.Timestamptz{
					Time:   time.Now().AddDate(0, -1, 0),
					Status: pgtype.Present,
				},
				AvailableUntil: pgtype.Timestamptz{
					Time:   time.Now().AddDate(0, 1, 0),
					Status: pgtype.Present,
				},
			}
			price := float32(100)
			billItem := entities.BillItem{}
			_, err := s.CalculatorDiscountPrice(discount, price, &billItem)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo)
		})
	}
}
