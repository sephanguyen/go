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

func TestBillingService_CreateBillItemForOrderCreate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext

		OneTimeDiscountService *billingService.IDiscountServiceForOneTimeBilling
		OneTimeTaxService      *billingService.ITaxServiceForOneTimeBilling
		OneTimePriceService    *billingService.IPriceServiceForOneTimeBilling
		OneTimeBillItemService *billingService.IBillItemServiceForOneTimeBilling

		RecurringDiscountService *billingService.IDiscountServiceForRecurringBilling
		RecurringTaxService      *billingService.ITaxServiceForRecurringBilling
		RecurringPriceService    *billingService.IPriceServiceForRecurringBilling
		RecurringBillItemService *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService   *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Happy case (isOneTimeProduct == true)",
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
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				OneTimeDiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				OneTimeTaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				OneTimePriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				OneTimeBillItemService.On("CreateNewBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Happy case (isOneTimeProduct == false)",
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
				IsOneTimeProduct: false,
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				RecurringDiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringTaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringPriceService.On("IsValidPriceForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				RecurringBillItemService.On("CreateNewBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			OneTimeDiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			OneTimeTaxService = &billingService.ITaxServiceForOneTimeBilling{}
			OneTimePriceService = &billingService.IPriceServiceForOneTimeBilling{}
			OneTimeBillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			oneTimeProductBilling := &BillingServiceForOneTimeProduct{
				DiscountService: OneTimeDiscountService,
				TaxService:      OneTimeTaxService,
				PriceService:    OneTimePriceService,
				BillItemService: OneTimeBillItemService,
			}

			RecurringDiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			RecurringTaxService = &billingService.ITaxServiceForRecurringBilling{}
			RecurringPriceService = &billingService.IPriceServiceForRecurringBilling{}
			RecurringBillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			recurringProductBilling := &BillingServiceForRecurringProduct{
				DiscountService:        RecurringDiscountService,
				TaxService:             RecurringTaxService,
				PriceService:           RecurringPriceService,
				BillItemService:        RecurringBillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			billingService := BillingService{
				OneTimeProductBilling:   oneTimeProductBilling,
				RecurringProductBilling: recurringProductBilling,
			}
			testCase.Setup(testCase.Ctx)
			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := billingService.CreateBillItemForOrderCreate(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, OneTimeDiscountService, OneTimeTaxService, OneTimePriceService, OneTimeBillItemService)
		})
	}
}

func TestBillingService_CreateBillItemForOrderUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext

		OneTimeDiscountService *billingService.IDiscountServiceForOneTimeBilling
		OneTimeTaxService      *billingService.ITaxServiceForOneTimeBilling
		OneTimePriceService    *billingService.IPriceServiceForOneTimeBilling
		OneTimeBillItemService *billingService.IBillItemServiceForOneTimeBilling

		RecurringDiscountService *billingService.IDiscountServiceForRecurringBilling
		RecurringTaxService      *billingService.ITaxServiceForRecurringBilling
		RecurringPriceService    *billingService.IPriceServiceForRecurringBilling
		RecurringBillItemService *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService   *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Happy case (IsOneTimeProduct == true)",
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
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				OneTimeBillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				OneTimeDiscountService.On("IsValidDiscountForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				OneTimeTaxService.On("IsValidTaxForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				OneTimePriceService.On("IsValidPriceForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				OneTimePriceService.On("IsValidAdjustmentPriceForOneTimeBilling", mock.Anything, mock.Anything).Return(nil)
				OneTimeBillItemService.On("CreateUpdateBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Happy case (IsOneTimeProduct == false)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: false,
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				RecurringBillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				RecurringDiscountService.On("IsValidDiscountForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringTaxService.On("IsValidTaxForRecurringBilling", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringPriceService.On("IsValidPriceForUpdateRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				RecurringBillItemService.On("CreateUpdateBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			OneTimeDiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			OneTimeTaxService = &billingService.ITaxServiceForOneTimeBilling{}
			OneTimePriceService = &billingService.IPriceServiceForOneTimeBilling{}
			OneTimeBillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			oneTimeProductBilling := &BillingServiceForOneTimeProduct{
				DiscountService: OneTimeDiscountService,
				TaxService:      OneTimeTaxService,
				PriceService:    OneTimePriceService,
				BillItemService: OneTimeBillItemService,
			}

			RecurringDiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			RecurringTaxService = &billingService.ITaxServiceForRecurringBilling{}
			RecurringPriceService = &billingService.IPriceServiceForRecurringBilling{}
			RecurringBillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			recurringProductBilling := &BillingServiceForRecurringProduct{
				DiscountService:        RecurringDiscountService,
				TaxService:             RecurringTaxService,
				PriceService:           RecurringPriceService,
				BillItemService:        RecurringBillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			billingService := BillingService{
				OneTimeProductBilling:   oneTimeProductBilling,
				RecurringProductBilling: recurringProductBilling,
			}
			testCase.Setup(testCase.Ctx)
			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := billingService.CreateBillItemForOrderUpdate(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, OneTimeDiscountService, OneTimeTaxService, OneTimePriceService, OneTimeBillItemService)
		})
	}
}

func TestBillingService_CreateBillItemForOrderWithdrawal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext

		OneTimeDiscountService *billingService.IDiscountServiceForOneTimeBilling
		OneTimeTaxService      *billingService.ITaxServiceForOneTimeBilling
		OneTimePriceService    *billingService.IPriceServiceForOneTimeBilling
		OneTimeBillItemService *billingService.IBillItemServiceForOneTimeBilling

		RecurringDiscountService *billingService.IDiscountServiceForRecurringBilling
		RecurringTaxService      *billingService.ITaxServiceForRecurringBilling
		RecurringPriceService    *billingService.IPriceServiceForRecurringBilling
		RecurringBillItemService *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService   *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when withdraw bill item for one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: true,
			},
			ExpectedErr: status.Errorf(codes.InvalidArgument, "we can't withdraw bill item for one time product"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: false,
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				RecurringBillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				RecurringPriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringBillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			OneTimeDiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			OneTimeTaxService = &billingService.ITaxServiceForOneTimeBilling{}
			OneTimePriceService = &billingService.IPriceServiceForOneTimeBilling{}
			OneTimeBillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			oneTimeProductBilling := &BillingServiceForOneTimeProduct{
				DiscountService: OneTimeDiscountService,
				TaxService:      OneTimeTaxService,
				PriceService:    OneTimePriceService,
				BillItemService: OneTimeBillItemService,
			}

			RecurringDiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			RecurringTaxService = &billingService.ITaxServiceForRecurringBilling{}
			RecurringPriceService = &billingService.IPriceServiceForRecurringBilling{}
			RecurringBillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			recurringProductBilling := &BillingServiceForRecurringProduct{
				DiscountService:        RecurringDiscountService,
				TaxService:             RecurringTaxService,
				PriceService:           RecurringPriceService,
				BillItemService:        RecurringBillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			billingService := BillingService{
				OneTimeProductBilling:   oneTimeProductBilling,
				RecurringProductBilling: recurringProductBilling,
			}
			testCase.Setup(testCase.Ctx)
			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := billingService.CreateBillItemForOrderWithdrawal(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, OneTimeDiscountService, OneTimeTaxService, OneTimePriceService, OneTimeBillItemService)
		})
	}
}

func TestBillingService_CreateBillItemForOrderGraduate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext

		OneTimeDiscountService *billingService.IDiscountServiceForOneTimeBilling
		OneTimeTaxService      *billingService.ITaxServiceForOneTimeBilling
		OneTimePriceService    *billingService.IPriceServiceForOneTimeBilling
		OneTimeBillItemService *billingService.IBillItemServiceForOneTimeBilling

		RecurringDiscountService *billingService.IDiscountServiceForRecurringBilling
		RecurringTaxService      *billingService.ITaxServiceForRecurringBilling
		RecurringPriceService    *billingService.IPriceServiceForRecurringBilling
		RecurringBillItemService *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService   *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when graduate bill item for one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: true,
			},
			ExpectedErr: status.Errorf(codes.InvalidArgument, "we can't graduate bill item for one time product"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: false,
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				RecurringBillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				RecurringPriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringBillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			OneTimeDiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			OneTimeTaxService = &billingService.ITaxServiceForOneTimeBilling{}
			OneTimePriceService = &billingService.IPriceServiceForOneTimeBilling{}
			OneTimeBillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			oneTimeProductBilling := &BillingServiceForOneTimeProduct{
				DiscountService: OneTimeDiscountService,
				TaxService:      OneTimeTaxService,
				PriceService:    OneTimePriceService,
				BillItemService: OneTimeBillItemService,
			}

			RecurringDiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			RecurringTaxService = &billingService.ITaxServiceForRecurringBilling{}
			RecurringPriceService = &billingService.IPriceServiceForRecurringBilling{}
			RecurringBillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			recurringProductBilling := &BillingServiceForRecurringProduct{
				DiscountService:        RecurringDiscountService,
				TaxService:             RecurringTaxService,
				PriceService:           RecurringPriceService,
				BillItemService:        RecurringBillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			billingService := BillingService{
				OneTimeProductBilling:   oneTimeProductBilling,
				RecurringProductBilling: recurringProductBilling,
			}
			testCase.Setup(testCase.Ctx)
			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := billingService.CreateBillItemForOrderGraduate(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, OneTimeDiscountService, OneTimeTaxService, OneTimePriceService, OneTimeBillItemService)
		})
	}
}

func TestBillingService_CreateBillItemForOrderLOA(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext

		OneTimeDiscountService *billingService.IDiscountServiceForOneTimeBilling
		OneTimeTaxService      *billingService.ITaxServiceForOneTimeBilling
		OneTimePriceService    *billingService.IPriceServiceForOneTimeBilling
		OneTimeBillItemService *billingService.IBillItemServiceForOneTimeBilling

		RecurringDiscountService *billingService.IDiscountServiceForRecurringBilling
		RecurringTaxService      *billingService.ITaxServiceForRecurringBilling
		RecurringPriceService    *billingService.IPriceServiceForRecurringBilling
		RecurringBillItemService *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService   *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when graduate bill item for one time product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: true,
			},
			ExpectedErr: status.Errorf(codes.InvalidArgument, "we can't pause bill item for one time product"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
				},
				BillItems:        []utils.BillingItemData{},
				IsOneTimeProduct: false,
			},
			Setup: func(ctx context.Context) {
				BillingScheduleService.On("CheckScheduleReturnProRatedItemAndMapPeriodInfo", mock.Anything, mock.Anything, mock.Anything).Return(utils.BillingItemData{}, entities.BillingRatio{}, []utils.BillingItemData{}, map[string]entities.BillingSchedulePeriod{}, nil)
				RecurringBillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				RecurringPriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringBillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			OneTimeDiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			OneTimeTaxService = &billingService.ITaxServiceForOneTimeBilling{}
			OneTimePriceService = &billingService.IPriceServiceForOneTimeBilling{}
			OneTimeBillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			oneTimeProductBilling := &BillingServiceForOneTimeProduct{
				DiscountService: OneTimeDiscountService,
				TaxService:      OneTimeTaxService,
				PriceService:    OneTimePriceService,
				BillItemService: OneTimeBillItemService,
			}

			RecurringDiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			RecurringTaxService = &billingService.ITaxServiceForRecurringBilling{}
			RecurringPriceService = &billingService.IPriceServiceForRecurringBilling{}
			RecurringBillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			recurringProductBilling := &BillingServiceForRecurringProduct{
				DiscountService:        RecurringDiscountService,
				TaxService:             RecurringTaxService,
				PriceService:           RecurringPriceService,
				BillItemService:        RecurringBillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			billingService := BillingService{
				OneTimeProductBilling:   oneTimeProductBilling,
				RecurringProductBilling: recurringProductBilling,
			}
			testCase.Setup(testCase.Ctx)
			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := billingService.CreateBillItemForOrderLOA(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, OneTimeDiscountService, OneTimeTaxService, OneTimePriceService, OneTimeBillItemService)
		})
	}
}

func TestBillingService_CreateBillItemForOrderCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db *mockDb.Ext

		OneTimeDiscountService *billingService.IDiscountServiceForOneTimeBilling
		OneTimeTaxService      *billingService.ITaxServiceForOneTimeBilling
		OneTimePriceService    *billingService.IPriceServiceForOneTimeBilling
		OneTimeBillItemService *billingService.IBillItemServiceForOneTimeBilling

		RecurringDiscountService *billingService.IDiscountServiceForRecurringBilling
		RecurringTaxService      *billingService.ITaxServiceForRecurringBilling
		RecurringPriceService    *billingService.IPriceServiceForRecurringBilling
		RecurringBillItemService *billingService.IBillItemServiceForRecurringBilling
		BillingScheduleService   *billingService.IBillingScheduleServiceForRecurringBilling
	)

	testcases := []utils.TestCase{
		{
			Name: "Happy case (IsOneTimeProduct == true)",
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
				IsOneTimeProduct: true,
			},
			Setup: func(ctx context.Context) {
				OneTimeBillItemService.On("GetOldBillItemForUpdateOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				OneTimeBillItemService.On("CreateCancelBillItemForOneTimeBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Happy case (isOneTimeProduct == false)",
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
				RecurringBillItemService.On("GetMapOldBillingItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.BillItem{}, nil)
				RecurringPriceService.On("IsValidPriceForCancelRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				RecurringBillItemService.On("CreateCancelBillItemForRecurringBilling", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			OneTimeDiscountService = &billingService.IDiscountServiceForOneTimeBilling{}
			OneTimeTaxService = &billingService.ITaxServiceForOneTimeBilling{}
			OneTimePriceService = &billingService.IPriceServiceForOneTimeBilling{}
			OneTimeBillItemService = &billingService.IBillItemServiceForOneTimeBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			oneTimeProductBilling := &BillingServiceForOneTimeProduct{
				DiscountService: OneTimeDiscountService,
				TaxService:      OneTimeTaxService,
				PriceService:    OneTimePriceService,
				BillItemService: OneTimeBillItemService,
			}

			RecurringDiscountService = &billingService.IDiscountServiceForRecurringBilling{}
			RecurringTaxService = &billingService.ITaxServiceForRecurringBilling{}
			RecurringPriceService = &billingService.IPriceServiceForRecurringBilling{}
			RecurringBillItemService = &billingService.IBillItemServiceForRecurringBilling{}
			BillingScheduleService = &billingService.IBillingScheduleServiceForRecurringBilling{}
			recurringProductBilling := &BillingServiceForRecurringProduct{
				DiscountService:        RecurringDiscountService,
				TaxService:             RecurringTaxService,
				PriceService:           RecurringPriceService,
				BillItemService:        RecurringBillItemService,
				BillingScheduleService: BillingScheduleService,
			}
			billingService := BillingService{
				OneTimeProductBilling:   oneTimeProductBilling,
				RecurringProductBilling: recurringProductBilling,
			}
			testCase.Setup(testCase.Ctx)
			orderItemDataReq := testCase.Req.(utils.OrderItemData)

			err := billingService.CreateBillItemForOrderCancel(testCase.Ctx, db, orderItemDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, OneTimeDiscountService, OneTimeTaxService, OneTimePriceService, OneTimeBillItemService)
		})
	}
}
