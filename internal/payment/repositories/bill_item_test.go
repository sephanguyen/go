package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func BillingItemRepoWithSqlMock() (*BillItemRepo, *testutil.MockDB) {
	r := &BillItemRepo{}
	return r, testutil.NewMockDB()
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestBillItemRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.BillItem{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("Create billingItem success", func(t *testing.T) {
		r, mockDB := BillingItemRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)
		_, err := r.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert billingItem fail", func(t *testing.T) {
		r, mockDB := BillingItemRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)
		_, err := r.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingItem: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillItemRepo_SetNonLatestBillItemByStudentProductID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.BillItem{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := BillingItemRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.SetNonLatestBillItemByStudentProductID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("err set none latest bill item fail", func(t *testing.T) {
		repo, mockDB := BillingItemRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := repo.SetNonLatestBillItemByStudentProductID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err set none latest bill item: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after set none latest bill item record", func(t *testing.T) {
		repo, mockDB := BillingItemRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := repo.SetNonLatestBillItemByStudentProductID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err set none latest bill item: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillItemRepo_GetLatestBillItemByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	billItemRepoWithSqlMock, mockDB := BillingItemRepoWithSqlMock()
	db := mockDB.DB
	row := mockDB.Row

	studentProductID := "student_product_id_1"
	billItem := &entities.BillItem{}
	fields, _ := billItem.FieldMap()
	scanFields := database.GetScanFields(billItem, fields)

	testCases := []utils.TestCase{
		{
			Name: constant.HappyCase,
			Ctx:  nil,
			Req:  studentProductID,
			ExpectedResp: &entities.BillItem{
				StudentProductID: pgtype.Text{
					String: studentProductID,
				},
				IsLatestBillItem: pgtype.Bool{
					Bool:   true,
					Status: pgtype.Present,
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentProductID)
				db.On("QueryRow").Once().Return(row)
				row.On("Scan", scanFields...).Once().Return(nil)
			},
		},
		{
			Name:         "Failed case: Error when scan",
			Ctx:          nil,
			Req:          mock.Anything,
			ExpectedResp: &entities.BillItem{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentProductID)
				db.On("QueryRow").Once().Return(row)
				row.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			req := (testCase.Req).(string)
			_, err := billItemRepoWithSqlMock.GetLatestBillItemByStudentProductID(ctx, db, req)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestBillItemRepo_GetBillItemByStudentProductIDAndPeriodID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: constant.FailCaseErrorRow,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentProductID,
				constant.BillingSchedulePeriodID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				entity := &entities.BillItem{}
				fields, values := entity.FieldMap()
				mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				entity := &entities.BillItem{}
				fields, values := entity.FieldMap()
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)
			billingItem, err := billingItemRepoWithSqlMock.GetBillItemByStudentProductIDAndPeriodID(ctx, mockDB.DB, constant.StudentProductID, constant.BillingSchedulePeriodID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NotNil(t, billingItem)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_UpdateReviewFlagByOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when execute",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				true,
			},
			ExpectedErr: fmt.Errorf("err update bill item: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, constant.FailCommandTag, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: "Fail case: Error when rows affected is 0",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				true,
			},
			ExpectedErr: fmt.Errorf("updating review flag for bill item by order id %v have %d RowsAffected", constant.OrderID, 0),
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, constant.FailCommandTag, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				true,
			},
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, constant.SuccessCommandTag, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderIDReq := testCase.Req.([]interface{})[0].(string)
			isReviewReq := testCase.Req.([]interface{})[1].(bool)
			err := billingItemRepoWithSqlMock.UpdateReviewFlagByOrderID(testCase.Ctx, mockDB.DB, orderIDReq, isReviewReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_UpdateBillingStatusByOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when execute",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			},
			ExpectedErr: fmt.Errorf("err void bill item: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, constant.FailCommandTag, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: "Fail case: Error when rows affected is 0",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, constant.FailCommandTag, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			},
			Setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, constant.SuccessCommandTag, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderIDReq := testCase.Req.([]interface{})[0].(string)
			statusReq := testCase.Req.([]interface{})[1].(string)
			err := billingItemRepoWithSqlMock.VoidBillItemByOrderID(testCase.Ctx, mockDB.DB, orderIDReq, statusReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: constant.FailCaseErrorRow,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				int32(1),
				pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				mockDB.Row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				int32(1),
				pb.BillingStatus_BILLING_STATUS_BILLED.String(),
			},
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				mockDB.Row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			billItemSequenceNumberReq := testCase.Req.([]interface{})[0].(int32)
			statusReq := testCase.Req.([]interface{})[1].(string)
			orderID, err := billingItemRepoWithSqlMock.UpdateBillingStatusByBillItemSequenceNumberAndReturnOrderID(testCase.Ctx, mockDB.DB, billItemSequenceNumberReq, statusReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, orderID)
			}
		})
	}
}

func TestBillItemRepo_GetRecurringBillItemsForScheduledGenerationOfNextBillItems(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	expectedBillItems := []*entities.BillItem{
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 1,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 2,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)
				order := &entities.BillItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedBillItems)).Return(true)

				billItem := &entities.BillItem{}
				fields, _ := billItem.FieldMap()
				scanFields := database.GetScanFields(billItem, fields)

				for _, expectedBillItem := range expectedBillItems {
					refExpectedBillItem := expectedBillItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedBillItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedBillItem.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedBillItem.ProductID.String
						args[3].(*pgtype.Text).String = refExpectedBillItem.StudentProductID.String
						args[4].(*pgtype.Text).String = refExpectedBillItem.BillType.String
						args[5].(*pgtype.Text).String = refExpectedBillItem.BillStatus.String
						args[6].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillDate.Time
						args[7].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillFrom.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillTo.Time
						args[9].(*pgtype.Text).String = refExpectedBillItem.BillSchedulePeriodID.String
						args[10].(*pgtype.Text).String = refExpectedBillItem.ProductDescription.String
						args[11].(*pgtype.Int4).Int = refExpectedBillItem.ProductPricing.Int
						args[12].(*pgtype.Text).String = refExpectedBillItem.DiscountAmountType.String
						args[13].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmountValue.Exp
						args[14].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmount.Exp
						args[15].(*pgtype.Numeric).Exp = refExpectedBillItem.RawDiscountAmount.Exp
						args[16].(*pgtype.Text).String = refExpectedBillItem.TaxID.String
						args[17].(*pgtype.Text).String = refExpectedBillItem.TaxCategory.String
						args[18].(*pgtype.Int4).Int = refExpectedBillItem.TaxPercentage.Int
						args[19].(*pgtype.Numeric).Exp = refExpectedBillItem.TaxAmount.Exp
						args[20].(*pgtype.Numeric).Exp = refExpectedBillItem.FinalPrice.Exp
						args[21].(*pgtype.Timestamptz).Time = refExpectedBillItem.UpdatedAt.Time
						args[22].(*pgtype.Timestamptz).Time = refExpectedBillItem.CreatedAt.Time
						args[23].(*pgtype.Int4).Int = refExpectedBillItem.BillItemSequenceNumber.Int
						args[24].(*pgtype.JSONB).Bytes = refExpectedBillItem.BillingItemDescription.Bytes
						args[25].(*pgtype.Text).String = refExpectedBillItem.ResourcePath.String
						args[26].(*pgtype.Text).String = refExpectedBillItem.BillApprovalStatus.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			_, err := billingItemRepoWithSqlMock.GetRecurringBillItemsForScheduledGenerationOfNextBillItems(testCase.Ctx, mockDB.DB)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_GetBillItemByOrderIDAndPaging(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	expectedBillItems := []*entities.BillItem{
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 1,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 2,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.OrderID,
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf(constant.RowScanError, constant.ErrDefault),
			Req: []interface{}{
				constant.OrderID,
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)
				order := &entities.BillItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedBillItems)).Return(true)

				billItem := &entities.BillItem{}
				fields, _ := billItem.FieldMap()
				scanFields := database.GetScanFields(billItem, fields)

				for _, expectedBillItem := range expectedBillItems {
					refExpectedBillItem := expectedBillItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedBillItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedBillItem.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedBillItem.ProductID.String
						args[3].(*pgtype.Text).String = refExpectedBillItem.StudentProductID.String
						args[4].(*pgtype.Text).String = refExpectedBillItem.BillType.String
						args[5].(*pgtype.Text).String = refExpectedBillItem.BillStatus.String
						args[6].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillDate.Time
						args[7].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillFrom.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillTo.Time
						args[9].(*pgtype.Text).String = refExpectedBillItem.BillSchedulePeriodID.String
						args[10].(*pgtype.Text).String = refExpectedBillItem.ProductDescription.String
						args[11].(*pgtype.Int4).Int = refExpectedBillItem.ProductPricing.Int
						args[12].(*pgtype.Text).String = refExpectedBillItem.DiscountAmountType.String
						args[13].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmountValue.Exp
						args[14].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmount.Exp
						args[15].(*pgtype.Numeric).Exp = refExpectedBillItem.RawDiscountAmount.Exp
						args[16].(*pgtype.Text).String = refExpectedBillItem.TaxID.String
						args[17].(*pgtype.Text).String = refExpectedBillItem.TaxCategory.String
						args[18].(*pgtype.Int4).Int = refExpectedBillItem.TaxPercentage.Int
						args[19].(*pgtype.Numeric).Exp = refExpectedBillItem.TaxAmount.Exp
						args[20].(*pgtype.Numeric).Exp = refExpectedBillItem.FinalPrice.Exp
						args[21].(*pgtype.Timestamptz).Time = refExpectedBillItem.UpdatedAt.Time
						args[22].(*pgtype.Timestamptz).Time = refExpectedBillItem.CreatedAt.Time
						args[23].(*pgtype.Int4).Int = refExpectedBillItem.BillItemSequenceNumber.Int
						args[24].(*pgtype.JSONB).Bytes = refExpectedBillItem.BillingItemDescription.Bytes
						args[25].(*pgtype.Text).String = refExpectedBillItem.ResourcePath.String
						args[26].(*pgtype.Text).String = refExpectedBillItem.BillApprovalStatus.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderID := testCase.Req.([]interface{})[0].(string)
			from := testCase.Req.([]interface{})[1].(int64)
			limit := testCase.Req.([]interface{})[2].(int64)

			_, err := billingItemRepoWithSqlMock.GetBillItemByOrderIDAndPaging(testCase.Ctx, mockDB.DB, orderID, from, limit)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_GetBillItemByStudentIDAndLocationIDsPaging(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	expectedBillItems := []*entities.BillItem{
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 1,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 2,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.OrderID,
				[]string{"location_1", "location_2"},
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf(constant.RowScanError, constant.ErrDefault),
			Req: []interface{}{
				constant.OrderID,
				[]string{"location_1", "location_2"},
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)
				order := &entities.BillItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				[]string{},
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedBillItems)).Return(true)

				billItem := &entities.BillItem{}
				fields, _ := billItem.FieldMap()
				scanFields := database.GetScanFields(billItem, fields)

				for _, expectedBillItem := range expectedBillItems {
					refExpectedBillItem := expectedBillItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedBillItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedBillItem.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedBillItem.ProductID.String
						args[3].(*pgtype.Text).String = refExpectedBillItem.StudentProductID.String
						args[4].(*pgtype.Text).String = refExpectedBillItem.BillType.String
						args[5].(*pgtype.Text).String = refExpectedBillItem.BillStatus.String
						args[6].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillDate.Time
						args[7].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillFrom.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillTo.Time
						args[9].(*pgtype.Text).String = refExpectedBillItem.BillSchedulePeriodID.String
						args[10].(*pgtype.Text).String = refExpectedBillItem.ProductDescription.String
						args[11].(*pgtype.Int4).Int = refExpectedBillItem.ProductPricing.Int
						args[12].(*pgtype.Text).String = refExpectedBillItem.DiscountAmountType.String
						args[13].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmountValue.Exp
						args[14].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmount.Exp
						args[15].(*pgtype.Numeric).Exp = refExpectedBillItem.RawDiscountAmount.Exp
						args[16].(*pgtype.Text).String = refExpectedBillItem.TaxID.String
						args[17].(*pgtype.Text).String = refExpectedBillItem.TaxCategory.String
						args[18].(*pgtype.Int4).Int = refExpectedBillItem.TaxPercentage.Int
						args[19].(*pgtype.Numeric).Exp = refExpectedBillItem.TaxAmount.Exp
						args[20].(*pgtype.Numeric).Exp = refExpectedBillItem.FinalPrice.Exp
						args[21].(*pgtype.Timestamptz).Time = refExpectedBillItem.UpdatedAt.Time
						args[22].(*pgtype.Timestamptz).Time = refExpectedBillItem.CreatedAt.Time
						args[23].(*pgtype.Int4).Int = refExpectedBillItem.BillItemSequenceNumber.Int
						args[24].(*pgtype.JSONB).Bytes = refExpectedBillItem.BillingItemDescription.Bytes
						args[25].(*pgtype.Text).String = refExpectedBillItem.ResourcePath.String
						args[26].(*pgtype.Text).String = refExpectedBillItem.BillApprovalStatus.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.OrderID,
				[]string{"location_1", "location_2"},
				int64(0),
				int64(2),
			},
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedBillItems)).Return(true)

				billItem := &entities.BillItem{}
				fields, _ := billItem.FieldMap()
				scanFields := database.GetScanFields(billItem, fields)

				for _, expectedBillItem := range expectedBillItems {
					refExpectedBillItem := expectedBillItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedBillItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedBillItem.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedBillItem.ProductID.String
						args[3].(*pgtype.Text).String = refExpectedBillItem.StudentProductID.String
						args[4].(*pgtype.Text).String = refExpectedBillItem.BillType.String
						args[5].(*pgtype.Text).String = refExpectedBillItem.BillStatus.String
						args[6].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillDate.Time
						args[7].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillFrom.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillTo.Time
						args[9].(*pgtype.Text).String = refExpectedBillItem.BillSchedulePeriodID.String
						args[10].(*pgtype.Text).String = refExpectedBillItem.ProductDescription.String
						args[11].(*pgtype.Int4).Int = refExpectedBillItem.ProductPricing.Int
						args[12].(*pgtype.Text).String = refExpectedBillItem.DiscountAmountType.String
						args[13].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmountValue.Exp
						args[14].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmount.Exp
						args[15].(*pgtype.Numeric).Exp = refExpectedBillItem.RawDiscountAmount.Exp
						args[16].(*pgtype.Text).String = refExpectedBillItem.TaxID.String
						args[17].(*pgtype.Text).String = refExpectedBillItem.TaxCategory.String
						args[18].(*pgtype.Int4).Int = refExpectedBillItem.TaxPercentage.Int
						args[19].(*pgtype.Numeric).Exp = refExpectedBillItem.TaxAmount.Exp
						args[20].(*pgtype.Numeric).Exp = refExpectedBillItem.FinalPrice.Exp
						args[21].(*pgtype.Timestamptz).Time = refExpectedBillItem.UpdatedAt.Time
						args[22].(*pgtype.Timestamptz).Time = refExpectedBillItem.CreatedAt.Time
						args[23].(*pgtype.Int4).Int = refExpectedBillItem.BillItemSequenceNumber.Int
						args[24].(*pgtype.JSONB).Bytes = refExpectedBillItem.BillingItemDescription.Bytes
						args[25].(*pgtype.Text).String = refExpectedBillItem.ResourcePath.String
						args[26].(*pgtype.Text).String = refExpectedBillItem.BillApprovalStatus.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			studentID := testCase.Req.([]interface{})[0].(string)
			locationIDs := testCase.Req.([]interface{})[1].([]string)
			from := testCase.Req.([]interface{})[2].(int64)
			limit := testCase.Req.([]interface{})[3].(int64)

			_, err := billingItemRepoWithSqlMock.GetBillItemByStudentIDAndLocationIDsPaging(testCase.Ctx, mockDB.DB, studentID, locationIDs, from, limit)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_CountBillItemByOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  constant.OrderID,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderID := testCase.Req.(string)

			_, err := billingItemRepoWithSqlMock.CountBillItemByOrderID(testCase.Ctx, mockDB.DB, orderID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_CountBillItemByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         []string{"lcoation_1", "location_2"},
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  []string{"lcoation_1", "location_2"},
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  []string{},
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			studentID := constant.StudentID
			locationIDs := testCase.Req.([]string)
			_, err := billingItemRepoWithSqlMock.CountBillItemByStudentIDAndLocationIDs(testCase.Ctx, mockDB.DB, studentID, locationIDs)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_GetBillItemInfoByOrderIDAndUniqueByProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	expectedBillItems := []*entities.BillItem{
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 1,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 2,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf(constant.RowScanError, constant.ErrDefault),
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)
				order := &entities.BillItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  constant.OrderID,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedBillItems)).Return(true)

				billItem := &entities.BillItem{}
				fields, _ := billItem.FieldMap()
				scanFields := database.GetScanFields(billItem, fields)

				for _, expectedBillItem := range expectedBillItems {
					refExpectedBillItem := expectedBillItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedBillItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedBillItem.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedBillItem.ProductID.String
						args[3].(*pgtype.Text).String = refExpectedBillItem.StudentProductID.String
						args[4].(*pgtype.Text).String = refExpectedBillItem.BillType.String
						args[5].(*pgtype.Text).String = refExpectedBillItem.BillStatus.String
						args[6].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillDate.Time
						args[7].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillFrom.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillTo.Time
						args[9].(*pgtype.Text).String = refExpectedBillItem.BillSchedulePeriodID.String
						args[10].(*pgtype.Text).String = refExpectedBillItem.ProductDescription.String
						args[11].(*pgtype.Int4).Int = refExpectedBillItem.ProductPricing.Int
						args[12].(*pgtype.Text).String = refExpectedBillItem.DiscountAmountType.String
						args[13].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmountValue.Exp
						args[14].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmount.Exp
						args[15].(*pgtype.Numeric).Exp = refExpectedBillItem.RawDiscountAmount.Exp
						args[16].(*pgtype.Text).String = refExpectedBillItem.TaxID.String
						args[17].(*pgtype.Text).String = refExpectedBillItem.TaxCategory.String
						args[18].(*pgtype.Int4).Int = refExpectedBillItem.TaxPercentage.Int
						args[19].(*pgtype.Numeric).Exp = refExpectedBillItem.TaxAmount.Exp
						args[20].(*pgtype.Numeric).Exp = refExpectedBillItem.FinalPrice.Exp
						args[21].(*pgtype.Timestamptz).Time = refExpectedBillItem.UpdatedAt.Time
						args[22].(*pgtype.Timestamptz).Time = refExpectedBillItem.CreatedAt.Time
						args[23].(*pgtype.Int4).Int = refExpectedBillItem.BillItemSequenceNumber.Int
						args[24].(*pgtype.JSONB).Bytes = refExpectedBillItem.BillingItemDescription.Bytes
						args[25].(*pgtype.Text).String = refExpectedBillItem.ResourcePath.String
						args[26].(*pgtype.Text).String = refExpectedBillItem.BillApprovalStatus.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderID := testCase.Req.(string)

			billItems, err := billingItemRepoWithSqlMock.GetBillItemInfoByOrderIDAndUniqueByProductID(testCase.Ctx, mockDB.DB, orderID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, len(billItems), len(expectedBillItems))
			}
		})
	}
}

func TestBillItemRepo_GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	expectedBillItems := []*entities.BillItem{
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 1,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			BillItemSequenceNumber: pgtype.Int4{
				Int: 2,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			BillType: pgtype.Text{
				String: "",
			},
			BillStatus: pgtype.Text{
				String: "",
			},
			BillDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillFrom: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillTo: pgtype.Timestamptz{
				Time: time.Now(),
			},
			BillingItemDescription: pgtype.JSONB{},
			BillSchedulePeriodID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			ProductDescription: pgtype.Text{
				String: "",
			},
			ProductPricing: pgtype.Int4{
				Int: 1,
			},
			DiscountAmountType: pgtype.Text{
				String: "",
			},
			DiscountAmountValue: pgtype.Numeric{},
			DiscountAmount:      pgtype.Numeric{},
			TaxID: pgtype.Text{
				String: "1",
			},
			TaxCategory: pgtype.Text{
				String: "",
			},
			TaxAmount: pgtype.Numeric{},
			TaxPercentage: pgtype.Int4{
				Int: 1,
			},
			FinalPrice: pgtype.Numeric{},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			BillApprovalStatus: pgtype.Text{
				String: "",
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf(constant.RowScanError, constant.ErrDefault),
			Req:         constant.OrderID,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)
				order := &entities.BillItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  constant.OrderID,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedBillItems)).Return(true)

				billItem := &entities.BillItem{}
				fields, _ := billItem.FieldMap()
				scanFields := database.GetScanFields(billItem, fields)

				for _, expectedBillItem := range expectedBillItems {
					refExpectedBillItem := expectedBillItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedBillItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedBillItem.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedBillItem.ProductID.String
						args[3].(*pgtype.Text).String = refExpectedBillItem.StudentProductID.String
						args[4].(*pgtype.Text).String = refExpectedBillItem.BillType.String
						args[5].(*pgtype.Text).String = refExpectedBillItem.BillStatus.String
						args[6].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillDate.Time
						args[7].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillFrom.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedBillItem.BillTo.Time
						args[9].(*pgtype.Text).String = refExpectedBillItem.BillSchedulePeriodID.String
						args[10].(*pgtype.Text).String = refExpectedBillItem.ProductDescription.String
						args[11].(*pgtype.Int4).Int = refExpectedBillItem.ProductPricing.Int
						args[12].(*pgtype.Text).String = refExpectedBillItem.DiscountAmountType.String
						args[13].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmountValue.Exp
						args[14].(*pgtype.Numeric).Exp = refExpectedBillItem.DiscountAmount.Exp
						args[15].(*pgtype.Numeric).Exp = refExpectedBillItem.RawDiscountAmount.Exp
						args[16].(*pgtype.Text).String = refExpectedBillItem.TaxID.String
						args[17].(*pgtype.Text).String = refExpectedBillItem.TaxCategory.String
						args[18].(*pgtype.Int4).Int = refExpectedBillItem.TaxPercentage.Int
						args[19].(*pgtype.Numeric).Exp = refExpectedBillItem.TaxAmount.Exp
						args[20].(*pgtype.Numeric).Exp = refExpectedBillItem.FinalPrice.Exp
						args[21].(*pgtype.Timestamptz).Time = refExpectedBillItem.UpdatedAt.Time
						args[22].(*pgtype.Timestamptz).Time = refExpectedBillItem.CreatedAt.Time
						args[23].(*pgtype.Int4).Int = refExpectedBillItem.BillItemSequenceNumber.Int
						args[24].(*pgtype.JSONB).Bytes = refExpectedBillItem.BillingItemDescription.Bytes
						args[25].(*pgtype.Text).String = refExpectedBillItem.ResourcePath.String
						args[26].(*pgtype.Text).String = refExpectedBillItem.BillApprovalStatus.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			orderID := testCase.Req.(string)

			billItems, err := billingItemRepoWithSqlMock.GetAllFirstBillItemDistinctByOrderIDAndUniqueByProductID(testCase.Ctx, mockDB.DB, orderID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, len(billItems), len(expectedBillItems))
			}
		})
	}
}

func TestBillItemRepo_GetLatestBillItemByStudentProductIDForStudentBilling(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         constant.StudentProductID,
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				entity := &entities.BillItem{}
				fields, values := entity.FieldMap()
				mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         constant.StudentProductID,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				entity := &entities.BillItem{}
				fields, values := entity.FieldMap()
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)
			req := testCase.Req.(string)
			billingItem, err := billingItemRepoWithSqlMock.GetLatestBillItemByStudentProductIDForStudentBilling(ctx, mockDB.DB, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NotNil(t, billingItem)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_GetBillingItemsThatNeedToBeBilled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockBillItemRepo, mockDB := BillingItemRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetBillingItemsThatNeedToBeBilled(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)

	})
	t.Run("err when scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetBillingItemsThatNeedToBeBilled(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)

	})
	t.Run("err when query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything)
		billItems, err := mockBillItemRepo.GetBillingItemsThatNeedToBeBilled(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}

func TestBillItemRepo_GetExportStudentBilling(t *testing.T) {
	t.Parallel()
	r, mockDB := BillingItemRepoWithSqlMock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run("Error when get data", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, mock.Anything)
		_, _, err := r.GetExportStudentBilling(ctx, mockDB.DB, []string{})
		require.NotNil(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
	t.Run("Error when scan data", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(fmt.Errorf("error something"), fields, dst)
		_, _, err := r.GetExportStudentBilling(ctx, mockDB.DB, []string{})
		require.NotNil(t, err)
		assert.Equal(t, "row.Scan: error something", err.Error())
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(nil, fields, dst)
		_, _, err := r.GetExportStudentBilling(ctx, mockDB.DB, []string{"location_1", "location_2"})
		require.Nil(t, err)
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(nil, fields, dst)
		_, _, err := r.GetExportStudentBilling(ctx, mockDB.DB, []string{})
		require.Nil(t, err)
	})
}

func TestBillItemRepo_GetByOrderIDAndProductIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockBillItemRepo, mockDB := BillingItemRepoWithSqlMock()
	const orderID = "order_id"
	productIDs := []string{"10", "20", "30"}
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			productIDs, orderID,
		)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetByOrderIDAndProductIDs(ctx, mockDB.DB, orderID, productIDs)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)

	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			productIDs, orderID,
		)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetByOrderIDAndProductIDs(ctx, mockDB.DB, orderID, productIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, productIDs, orderID)
		billItems, err := mockBillItemRepo.GetByOrderIDAndProductIDs(ctx, mockDB.DB, orderID, productIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}

func TestBillItemRepo_GetPresentBillingByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockBillItemRepo, mockDB := BillingItemRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetPresentBillingByStudentProductID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)

	})
	t.Run("err when scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetPresentBillingByStudentProductID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)

	})
	t.Run("err when query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything)
		billItems, err := mockBillItemRepo.GetPresentBillingByStudentProductID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}

func TestBillItemRepo_GetUpcomingBillingByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		billingItemRepoWithSqlMock *BillItemRepo
		mockDB                     *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         constant.StudentProductID,
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				entity := &entities.BillItem{}
				fields, values := entity.FieldMap()
				mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         constant.StudentProductID,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				entity := &entities.BillItem{}
				fields, values := entity.FieldMap()
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			billingItemRepoWithSqlMock, mockDB = BillingItemRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)
			billingItem, err := billingItemRepoWithSqlMock.GetUpcomingBillingByStudentProductID(ctx, mockDB.DB, mock.Anything, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NotNil(t, billingItem)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestBillItemRepo_GetPastBillItemsByStudentProductIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockBillItemRepo, mockDB := BillingItemRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetPastBillItemsByStudentProductIDs(ctx, mockDB.DB, []string{}, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)
	})
	t.Run("err when scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetPastBillItemsByStudentProductIDs(ctx, mockDB.DB, []string{}, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}

func TestBillItemRepo_GetPresentAndFutureBillItemsByStudentProductIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockBillItemRepo, mockDB := BillingItemRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetPresentAndFutureBillItemsByStudentProductIDs(ctx, mockDB.DB, []string{}, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)
	})
	t.Run("err when scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockBillItemRepo.GetPresentAndFutureBillItemsByStudentProductIDs(ctx, mockDB.DB, []string{}, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}
