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

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	FailCaseCheckScheduleError        = "Fail case: Error when check schedule return pro rated item and map period info"
	FailCaseGetOldBillItemMapError    = "Fail case: Error when get map old billing item for recurring billing"
	FailCaseCancelPriceError          = "Fail case: Error when price for cancel recurring billing"
	FailCaseCreateCancelBillItemError = "Fail case: Error when create cancel bill item for recurring billing"
)

func TestBillingForRecurringProductService_CreateBillItemForOrderCreate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		DiscountService        *billingService.IDiscountServiceForRecurringBilling
		TaxService             *billingService.ITaxServiceForRecurringBilling
		PriceService           *billingService.IPriceServiceForRecurringBilling
		BillItemService        *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCheckScheduleError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when discount for recurring billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when tax for recurring billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when price for recurring billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when create new bill item for recurring billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				BillItemService.On("CreateNewBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				BillItemService.On("CreateNewBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			TaxService = &billingService.ITaxServiceForRecurringBilling{}
			PriceService = &billingService.IPriceServiceForRecurringBilling{}
			BillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			s := &BillingServiceForRecurringProduct{
				DiscountService:        DiscountService,
				TaxService:             TaxService,
				PriceService:           PriceService,
				BillItemService:        BillItemService,
				BillingScheduleService: BillingScheduleService,
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

func TestBillingForRecurringProductService_CreateBillItemForOrderUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		DiscountService        *billingService.IDiscountServiceForRecurringBilling
		TaxService             *billingService.ITaxServiceForRecurringBilling
		PriceService           *billingService.IPriceServiceForRecurringBilling
		BillItemService        *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCheckScheduleError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseGetOldBillItemMapError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when discount for recurring billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForUpdateRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				BillItemService.On("CreateUpdateBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when tax for recurring billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				PriceService.On("IsValidPriceForUpdateRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				BillItemService.On("CreateUpdateBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when price for recurring billing is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForUpdateRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when create update bill item for recurring billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForUpdateRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				BillItemService.On("CreateUpdateBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				DiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				TaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				PriceService.On("IsValidPriceForUpdateRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				BillItemService.On("CreateUpdateBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			TaxService = &billingService.ITaxServiceForRecurringBilling{}
			PriceService = &billingService.IPriceServiceForRecurringBilling{}
			BillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			s := &BillingServiceForRecurringProduct{
				DiscountService:        DiscountService,
				TaxService:             TaxService,
				PriceService:           PriceService,
				BillItemService:        BillItemService,
				BillingScheduleService: BillingScheduleService,
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

func TestBillingForRecurringProductService_CreateBillItemForOrderWithdrawal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		DiscountService        *billingService.IDiscountServiceForRecurringBilling
		TaxService             *billingService.ITaxServiceForRecurringBilling
		PriceService           *billingService.IPriceServiceForRecurringBilling
		BillItemService        *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCheckScheduleError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseGetOldBillItemMapError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCancelPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseCreateCancelBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			TaxService = &billingService.ITaxServiceForRecurringBilling{}
			PriceService = &billingService.IPriceServiceForRecurringBilling{}
			BillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			s := &BillingServiceForRecurringProduct{
				DiscountService:        DiscountService,
				TaxService:             TaxService,
				PriceService:           PriceService,
				BillItemService:        BillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := s.CreateBillItemForOrderWithdrawal(testCase.Ctx, db, orderItemDataReq)

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

func TestBillingForRecurringProductService_CreateBillItemForOrderGraduate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		DiscountService        *billingService.IDiscountServiceForRecurringBilling
		TaxService             *billingService.ITaxServiceForRecurringBilling
		PriceService           *billingService.IPriceServiceForRecurringBilling
		BillItemService        *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCheckScheduleError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseGetOldBillItemMapError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCancelPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseCreateCancelBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			TaxService = &billingService.ITaxServiceForRecurringBilling{}
			PriceService = &billingService.IPriceServiceForRecurringBilling{}
			BillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			s := &BillingServiceForRecurringProduct{
				DiscountService:        DiscountService,
				TaxService:             TaxService,
				PriceService:           PriceService,
				BillItemService:        BillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			testCase.Setup(testCase.Ctx)

			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := s.CreateBillItemForOrderGraduate(testCase.Ctx, db, orderItemDataReq)

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

func TestBillingForRecurringProductService_CreateBillItemForOrderCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		DiscountService        *billingService.IDiscountServiceForRecurringBilling
		TaxService             *billingService.ITaxServiceForRecurringBilling
		PriceService           *billingService.IPriceServiceForRecurringBilling
		BillItemService        *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseCheckScheduleError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseGetOldBillItemMapError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseCancelPriceError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        FailCaseCreateCancelBillItemError,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
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
				BillItems: []utils.BillingItemData{},
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				BillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				PriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			DiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			TaxService = &billingService.ITaxServiceForRecurringBilling{}
			PriceService = &billingService.IPriceServiceForRecurringBilling{}
			BillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			s := &BillingServiceForRecurringProduct{
				DiscountService:        DiscountService,
				TaxService:             TaxService,
				PriceService:           PriceService,
				BillItemService:        BillItemService,
				BillingScheduleService: BillingScheduleService,
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
