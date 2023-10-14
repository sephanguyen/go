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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillingForCustomOrderService_CreateBillItemForCustomOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db              *mockDb.Ext
		TaxService      *billingService.ITaxServiceForCustomOrder
		BillItemService *billingService.IBillItemServiceForCustomOrder
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when tax for custom order is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					CustomBillingItems: []*pb.CustomBillingItem{
						{
							Name:  constant.ProductName,
							Price: 1000,
						},
					},
				},
				entities.Order{},
				constant.LocationName,
			},
			Setup: func(ctx context.Context) {
				TaxService.On("IsValidTaxForCustomOrder", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				BillItemService.On("CreateCustomBillItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "Fail case: Error when create custom bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					CustomBillingItems: []*pb.CustomBillingItem{
						{
							Name:  constant.ProductName,
							Price: 1000,
						},
					},
				},
				entities.Order{},
				constant.LocationName,
			},
			Setup: func(ctx context.Context) {
				TaxService.On("IsValidTaxForCustomOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCustomBillItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					CustomBillingItems: []*pb.CustomBillingItem{
						{
							Name:  constant.ProductName,
							Price: 1000,
						},
					},
				},
				entities.Order{},
				constant.LocationName,
			},
			Setup: func(ctx context.Context) {
				TaxService.On("IsValidTaxForCustomOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				BillItemService.On("CreateCustomBillItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			TaxService = &billingService.ITaxServiceForCustomOrder{}
			BillItemService = &billingService.IBillItemServiceForCustomOrder{}
			s := &BillingServiceForCustomOrder{
				TaxService:      TaxService,
				BillItemService: BillItemService,
			}
			testCase.Setup(testCase.Ctx)

			createCustomBillingReq := testCase.Req.([]interface{})[0].(*pb.CreateCustomBillingRequest)
			orderReq := testCase.Req.([]interface{})[1].(entities.Order)
			locationNameReq := testCase.Req.([]interface{})[2].(string)

			err := s.CreateBillItemForCustomOrder(testCase.Ctx, db, createCustomBillingReq, orderReq, locationNameReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, TaxService, TaxService, BillItemService)
		})
	}
}
