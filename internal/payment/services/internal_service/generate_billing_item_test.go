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
	mockServices "github.com/manabie-com/backend/mock/payment/services/internal_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateBillItems(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		tx                      *mockDb.Tx
		upcomingBillItemService *mockServices.IUpcomingBillItemServiceForInternalService
		billItemService         *mockServices.IBillItemServiceForInternalService
		billingScheduleService  *mockServices.IBillingScheduleForInternalService
		studentProductService   *mockServices.IStudentProductServiceForInternalService
		discountService         *mockServices.IDiscountServiceForInternalService
		taxService              *mockServices.ITaxServiceForInternalService
		priceService            *mockServices.IPriceServiceForInternalService
		studentService          *mockServices.IStudentForInternalService
	)
	upcomingBillItems := []entities.UpcomingBillItem{
		{
			BillingSchedulePeriodID: pgtype.Text{
				String: "schedulePeriodID",
				Status: pgtype.Present,
			},
			BillingDate: pgtype.Timestamptz{
				Time:   time.Now(),
				Status: pgtype.Present,
			},
			IsGenerated: pgtype.Bool{
				Bool:   false,
				Status: pgtype.Present,
			},
			TaxID: pgtype.Text{
				String: "tax-id",
				Status: pgtype.Present,
			},
		},
	}

	studentProduct := entities.StudentProduct{
		StudentProductID: pgtype.Text{
			String: "studentProductId",
			Status: pgtype.Present,
		},
		ProductStatus: pgtype.Text{
			String: pb.StudentProductStatus_ORDERED.String(),
			Status: pgtype.Present,
		},
		StudentProductLabel: pgtype.Text{
			String: pb.StudentProductLabel_CREATED.String(),
			Status: pgtype.Present,
		},
		RootStudentProductID: pgtype.Text{
			String: "root-student-product-id-1",
			Status: pgtype.Present,
		},
	}

	billItems := []entities.BillItem{
		{
			StudentProductID: pgtype.Text{
				String: "studentProductId",
				Status: pgtype.Present,
			},
			OrderID: pgtype.Text{
				String: "orderId",
				Status: pgtype.Present,
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        "Failure during retrieving upcoming bill items to generate next bill items",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         &pb.GenerateBillingItemsRequest{},
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return([]entities.UpcomingBillItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failure when get student product and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when get bill item and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]entities.BillItem{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when get bill item null and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: fmt.Errorf("bill Item not found with order_id: { 0} and product_id: { 0}"),
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]entities.BillItem{}, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when get bill schedule period and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when get bill period and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when assign discount price and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)

				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when assign tax price and roll back",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)

				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				taxService.On("GetTaxByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Tax{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when get product prices by product id and price type",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)

				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				taxService.On("GetTaxByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Tax{}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when get student product by student product id for update",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Once().Return(studentProduct, nil)

				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				taxService.On("GetTaxByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Tax{}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.StudentProduct{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure when check is enrolled in org by student id and time",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Once().Return(studentProduct, nil)

				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
				taxService.On("GetTaxByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Tax{}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.StudentProduct{}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure while creating upcoming bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							String: "SchedulePeriodID",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountID: pgtype.Text{
						String: "discount-id",
						Status: pgtype.Present,
					},
				}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
				taxService.On("GetTaxByID", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: "tax-id",
						Status: pgtype.Present,
					},
				}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
				priceService.On("CalculatorBillItemPrice", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemService.On("CreateUpcomingBillItems", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "Failure while creating new upcoming bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  0,
				Failed:     1,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							String: "SchedulePeriodID",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountID: pgtype.Text{
						String: "discount-id",
						Status: pgtype.Present,
					},
				}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
				taxService.On("GetTaxByID", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: "tax-id",
						Status: pgtype.Present,
					},
				}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
				priceService.On("CalculatorBillItemPrice", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemService.On("CreateUpcomingBillItems", ctx, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("CreateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				upcomingBillItemService.On("AddExecuteNoteForCurrentUpcomingBillItem", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.GenerateBillingItemsRequest{},
			ExpectedResp: pb.GenerateBillingItemsResponse{
				Successful: true,
				Successed:  1,
				Failed:     0,
			},
			Setup: func(ctx context.Context) {
				upcomingBillItemService.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return(upcomingBillItems, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				studentProductService.On("GetStudentProductByStudentProductIDForUpdate", ctx, mock.Anything, mock.Anything).Return(studentProduct, nil)
				billItemService.On("GetRecurringBillItemsByOrderIDAndProductID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(billItems, nil)
				billingScheduleService.On("GetBillingSchedulePeriodByID", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingSchedulePeriodID: pgtype.Text{
					String: "1",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetLatestBillingSchedulePeriod", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{BillingScheduleID: pgtype.Text{
					String: "2",
					Status: pgtype.Present,
				}}, nil)
				billingScheduleService.On("GetNextBillingSchedulePeriod", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.BillingSchedulePeriod{
						BillingSchedulePeriodID: pgtype.Text{
							String: "SchedulePeriodID",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{},
						BillingScheduleID: pgtype.Text{
							String: "ScheduleID",
							Status: pgtype.Present,
						},
						StartDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -15),
							Status: pgtype.Present,
						},
						EndDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 15),
							Status: pgtype.Present,
						},
						BillingDate: pgtype.Timestamptz{
							Time:   time.Now(),
							Status: pgtype.Present,
						},
					},
					nil,
				)
				discountService.On("VerifyDiscountForGenerateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(entities.Discount{
					DiscountID: pgtype.Text{
						String: "discount-id",
						Status: pgtype.Present,
					},
				}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
				taxService.On("GetTaxByID", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: "tax-id",
						Status: pgtype.Present,
					},
				}, nil)
				priceService.On("GetProductPricesByProductIDAndPriceType", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
				priceService.On("CalculatorBillItemPrice", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				billItemService.On("CreateUpcomingBillItems", ctx, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("CreateUpcomingBillItem", ctx, mock.Anything, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				upcomingBillItemService.On("UpdateCurrentUpcomingBillItemStatus", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			upcomingBillItemService = new(mockServices.IUpcomingBillItemServiceForInternalService)
			billItemService = new(mockServices.IBillItemServiceForInternalService)
			studentProductService = new(mockServices.IStudentProductServiceForInternalService)
			billingScheduleService = new(mockServices.IBillingScheduleForInternalService)
			discountService = new(mockServices.IDiscountServiceForInternalService)
			taxService = new(mockServices.ITaxServiceForInternalService)
			priceService = new(mockServices.IPriceServiceForInternalService)
			studentService = new(mockServices.IStudentForInternalService)

			testCase.Setup(testCase.Ctx)

			s := &InternalService{
				DB:                      db,
				billItemService:         billItemService,
				studentProductService:   studentProductService,
				billingScheduleService:  billingScheduleService,
				upcomingBillItemService: upcomingBillItemService,
				taxService:              taxService,
				discountService:         discountService,
				productService:          nil,
				priceService:            priceService,
				studentService:          studentService,
			}

			resp, err := s.GenerateBillingItems(testCase.Ctx, testCase.Req.(*pb.GenerateBillingItemsRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.NotNil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, billItemService, upcomingBillItemService, studentProductService, priceService)
		})
	}

}
