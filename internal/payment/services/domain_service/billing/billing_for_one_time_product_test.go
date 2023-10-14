package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	billingService "github.com/manabie-com/backend/mock/payment/services/domain_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	FailCaseQuantityError = "Fail case: Error when quantity bill item is 1"
)

func TestBillingForOneTimeProductService_CreateBillItemForOrderCreate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db              *mockDb.Ext
		DiscountService *billingService.IDiscountServiceForOneTimeBilling
		TaxService      *billingService.ITaxServiceForOneTimeBilling
		PriceService    *billingService.IPriceServiceForOneTimeBilling
		BillItemService *billingService.IBillItemServiceForOneTimeBilling
	)

	testcases := []utils.TestCase{
		{
			Name: FailCaseQuantityError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"we can't create bill item for product %v because quantity bill item is %v",
				constant.ProductID,
				0),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when discount for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateNewBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when tax for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateNewBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when price for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				BillItemService.On("CreateNewBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when create new bill item for one time billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateNewBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateNewBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			TaxService = &billingService.ITaxServiceForOneTimeBilling{}
			PriceService = &billingService.IPriceServiceForOneTimeBilling{}
			BillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			s := &BillingServiceForOneTimeProduct{
				DiscountService: DiscountService,
				TaxService:      TaxService,
				PriceService:    PriceService,
				BillItemService: BillItemService,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := s.CreateBillItemForOrderCreate(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, DiscountService, TaxService, PriceService, BillItemService)
		})
	}
}

func TestBillingForOneTimeProductService_CreateBillItemForOrderUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db              *mockDb.Ext
		DiscountService *billingService.IDiscountServiceForOneTimeBilling
		TaxService      *billingService.ITaxServiceForOneTimeBilling
		PriceService    *billingService.IPriceServiceForOneTimeBilling
		BillItemService *billingService.IBillItemServiceForOneTimeBilling
	)

	testcases := []utils.TestCase{
		{
			Name: FailCaseQuantityError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"we can't update bill item for product %v because quantity bill item is %v",
				constant.ProductID,
				0),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when get old bill item for update one time billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when discount for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when tax for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when price for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				PriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when adjustment price for one time billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(constant.ErrDefault)
				BillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when create update bill item for one time billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			TaxService = &billingService.ITaxServiceForOneTimeBilling{}
			PriceService = &billingService.IPriceServiceForOneTimeBilling{}
			BillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			s := &BillingServiceForOneTimeProduct{
				DiscountService: DiscountService,
				TaxService:      TaxService,
				PriceService:    PriceService,
				BillItemService: BillItemService,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := s.CreateBillItemForOrderUpdate(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, DiscountService, TaxService, PriceService, BillItemService)
		})
	}
}

func TestBillingForOneTimeProductService_CreateBillItemForOrderCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db              *mockDb.Ext
		DiscountService *billingService.IDiscountServiceForOneTimeBilling
		TaxService      *billingService.ITaxServiceForOneTimeBilling
		PriceService    *billingService.IPriceServiceForOneTimeBilling
		BillItemService *billingService.IBillItemServiceForOneTimeBilling
	)

	testcases := []utils.TestCase{
		{
			Name: FailCaseQuantityError,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"we can't cancel bill item for product %v because quantity bill item is %v",
				constant.ProductID,
				0),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when get old bill item for update one time billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when create cancel bill item for one time billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				BillItemService.On("CreateCancelBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							ProductId: constant.ProductID,
							BillingSchedulePeriodId: &wrapperspb.StringValue{
								Value: constant.BillingSchedulePeriodID,
							},
							Price: 100,
							TaxItem: &pb.TaxBillItem{
								TaxId: constant.TaxID,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId: constant.DiscountID,
							},
							FinalPrice: 100,
							Quantity: &wrapperspb.Int32Value{
								Value: 1,
							},
						},
						IsUpcoming: false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				BillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				BillItemService.On("CreateCancelBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			TaxService = &billingService.ITaxServiceForOneTimeBilling{}
			PriceService = &billingService.IPriceServiceForOneTimeBilling{}
			BillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			s := &BillingServiceForOneTimeProduct{
				DiscountService: DiscountService,
				TaxService:      TaxService,
				PriceService:    PriceService,
				BillItemService: BillItemService,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := s.CreateBillItemForOrderCancel(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, DiscountService, TaxService, PriceService, BillItemService)
		})
	}
}
