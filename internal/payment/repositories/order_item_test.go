package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrderItemRepoWithSqlMock() (*OrderItemRepo, *testutil.MockDB) {
	orderItemRepo := &OrderItemRepo{}
	return orderItemRepo, testutil.NewMockDB()
}

func TestOrderItemRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	orderItemRepoWithSqlMock, mockDB := OrderItemRepoWithSqlMock()
	db := mockDB.DB
	mockEntity := &entities.BillItem{}
	_, fieldMap := mockEntity.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	testCases := []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Req:          entities.OrderItem{},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:         "Failed case: Error when insert",
			Req:          entities.OrderItem{},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("err insert OrderItem: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Return(constant.SuccessCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderItemRepoWithSqlMock, mockDB = OrderItemRepoWithSqlMock()
			db = mockDB.DB
			testCase.Setup(ctx)
			req := (testCase.Req).(entities.OrderItem)
			err := orderItemRepoWithSqlMock.Create(ctx, db, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestGetStudentProductIDsForVoidOrderByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("Get order item success", func(t *testing.T) {
		orderItemRepoWithSqlMock, mockDB := OrderItemRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		orderItem := &entities.OrderItem{}
		fields, _ := orderItem.FieldMap()
		scanFields := database.GetScanFields(orderItem, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderItemRepoWithSqlMock.GetStudentProductIDsForVoidOrderByOrderID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Get order item fail", func(t *testing.T) {
		orderItemRepoWithSqlMock, mockDB := OrderItemRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		orderItem := &entities.OrderItem{}
		fields, _ := orderItem.FieldMap()
		scanFields := database.GetScanFields(orderItem, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := orderItemRepoWithSqlMock.GetStudentProductIDsForVoidOrderByOrderID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestGetAllByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("Get order item success", func(t *testing.T) {
		orderItemRepoWithSqlMock, mockDB := OrderItemRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		orderItem := &entities.OrderItem{}
		fields, _ := orderItem.FieldMap()
		scanFields := database.GetScanFields(orderItem, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderItemRepoWithSqlMock.GetAllByOrderID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Get order item fail", func(t *testing.T) {
		orderItemRepoWithSqlMock, mockDB := OrderItemRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		orderItem := &entities.OrderItem{}
		fields, _ := orderItem.FieldMap()
		scanFields := database.GetScanFields(orderItem, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := orderItemRepoWithSqlMock.GetAllByOrderID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestOrderItemRepo_CountOrderItemsByOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		orderItemRepoWithSqlMock *OrderItemRepo
		mockDB                   *testutil.MockDB
	)

	testCases := []utils.TestCase{
		{
			Name:        constant.HappyCase,
			Req:         constant.OrderID,
			ExpectedErr: fmt.Errorf("row.Scan: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Row.On("Scan", mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:         constant.HappyCase,
			Req:          constant.OrderID,
			ExpectedResp: 5,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Row.On("Scan", mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderItemRepoWithSqlMock, mockDB = OrderItemRepoWithSqlMock()
			testCase.Setup(ctx)
			req := (testCase.Req).(string)
			_, err := orderItemRepoWithSqlMock.CountOrderItemsByOrderID(ctx, mockDB.DB, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())

			} else {
				assert.Nil(t, err)
				assert.Equal(t, testCase.ExpectedErr, err)
			}

		})
	}
}

func TestOrderItemRepo_GetOrderItemsByOrderIDWithPaging(t *testing.T) {
	t.Parallel()

	r, mockDB := OrderItemRepoWithSqlMock()
	testcases := buildOrderItemGetOrderItemsByOrderIDWithPagingTestcases(mockDB)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, testcase := range testcases {
		refTestcase := testcase
		req := refTestcase.Req.(testcaseGetOrderItemsByOrderIDWithPagingReq)
		t.Run(refTestcase.Name, func(t *testing.T) {
			refTestcase.Setup(ctx)
			expectedOrderItems := refTestcase.ExpectedResp.([]*entities.OrderItem)
			orderItems, err := r.GetOrderItemsByOrderIDWithPaging(ctx, mockDB.DB, req.orderID, req.offset, req.limit)
			if refTestcase.ExpectedErr != nil {
				assert.Equal(t, refTestcase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, len(expectedOrderItems), len(orderItems))
			for idx, expectedOrderItem := range expectedOrderItems {
				assert.Equal(t, expectedOrderItem.OrderID, orderItems[idx].OrderID)
				assert.Equal(t, expectedOrderItem.ProductID, orderItems[idx].ProductID)
				assert.Equal(t, expectedOrderItem.DiscountID, orderItems[idx].DiscountID)
				assert.Equal(t, expectedOrderItem.StartDate, orderItems[idx].StartDate)
				assert.Equal(t, expectedOrderItem.StudentProductID, orderItems[idx].StudentProductID)
				assert.Equal(t, expectedOrderItem.CreatedAt, orderItems[idx].CreatedAt)
				assert.Equal(t, expectedOrderItem.ResourcePath, orderItems[idx].ResourcePath)
			}
		})
	}
	mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
}

type testcaseGetOrderItemsByOrderIDWithPagingReq struct {
	orderID string `json:"order_id"`
	limit   int64  `json:"limit"`
	offset  int64  `json:"offset"`
}

func buildOrderItemGetOrderItemsByOrderIDWithPagingTestcases(mockDB *testutil.MockDB) []utils.TestCase {
	mockE := &entities.OrderItem{}
	fieldNames, _ := mockE.FieldMap()

	var req testcaseGetOrderItemsByOrderIDWithPagingReq
	req.orderID = "order_id_1"
	req.limit = 1
	req.offset = 3
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		strings.Join(fieldNames, ","),
		mockE.TableName(),
	)
	args := []interface{}{
		mock.Anything,
		stmt,
		req.orderID,
		req.limit,
		req.offset,
	}

	expectedOrderItems := []*entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "1",
			},
			StartDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
		},
		{
			OrderID: pgtype.Text{
				String: "order_id_1",
			},
			ProductID: pgtype.Text{
				String: "2",
			},
			DiscountID: pgtype.Text{
				String: "1",
			},
			StartDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
		},
	}

	return []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Req:          req,
			ExpectedResp: expectedOrderItems,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedOrderItems)).Return(true)

				order := &entities.OrderItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)

				for _, expectedOrderItem := range expectedOrderItems {
					refExpectedOrderItem := expectedOrderItem
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedOrderItem.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedOrderItem.ProductID.String
						args[2].(*pgtype.Text).String = refExpectedOrderItem.OrderItemID.String
						args[3].(*pgtype.Text).String = refExpectedOrderItem.DiscountID.String
						args[4].(*pgtype.Timestamptz).Time = refExpectedOrderItem.StartDate.Time
						args[5].(*pgtype.Timestamptz).Time = refExpectedOrderItem.CreatedAt.Time
						args[6].(*pgtype.Text).String = refExpectedOrderItem.StudentProductID.String
						args[7].(*pgtype.Text).String = refExpectedOrderItem.ResourcePath.String

					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:         "empty case",
			Req:          req,
			ExpectedResp: []*entities.OrderItem{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:         "query failed case",
			Req:          req,
			ExpectedResp: []*entities.OrderItem{},
			ExpectedErr:  errors.New("error query"),
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, errors.New("error query"))
			},
		},
		{
			Name:         "scan failed case",
			Req:          req,
			ExpectedResp: []*entities.OrderItem{},
			ExpectedErr:  fmt.Errorf("row.Scan: %w", errors.New("error scan")),
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				order := &entities.OrderItem{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				rows.On("Scan", scanFields...).Once().Return(errors.New("error scan"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}
}

func TestOrderItemRepo_GetOrderItemsByOrderIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockOrderItemRepo, mockDB := OrderItemRepoWithSqlMock()
	orderIDs := []string{"1", "2", "3"}
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			orderIDs,
		)
		e := &entities.OrderItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockOrderItemRepo.GetOrderItemsByOrderIDs(ctx, mockDB.DB, orderIDs)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)

	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			orderIDs,
		)
		e := &entities.OrderItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockOrderItemRepo.GetOrderItemsByOrderIDs(ctx, mockDB.DB, orderIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, orderIDs)
		billItems, err := mockOrderItemRepo.GetOrderItemsByOrderIDs(ctx, mockDB.DB, orderIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}

func TestOrderItemRepo_GetOrderItemsByProductIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockOrderItemRepo, mockDB := OrderItemRepoWithSqlMock()
	productIDs := []string{"1", "2", "3"}
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			productIDs,
		)
		e := &entities.OrderItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billItems, err := mockOrderItemRepo.GetOrderItemsByProductIDs(ctx, mockDB.DB, productIDs)
		assert.Nil(t, err)
		assert.NotNil(t, billItems)

	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			productIDs,
		)
		e := &entities.OrderItem{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billItems, err := mockOrderItemRepo.GetOrderItemsByProductIDs(ctx, mockDB.DB, productIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, productIDs)
		billItems, err := mockOrderItemRepo.GetOrderItemsByProductIDs(ctx, mockDB.DB, productIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billItems)
	})
}

func TestOrderItemRepo_GetOrderItemByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockOrderItemRepo, mockDB := OrderItemRepoWithSqlMock()
	t.Run("Success", func(t *testing.T) {
		e := &entities.OrderItem{}
		fields, values := e.FieldMap()
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		mockDB.MockRowScanFields(nil, fields, values)

		billItems, err := mockOrderItemRepo.GetOrderItemByStudentProductID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)
		assert.NotNil(t, billItems)

	})
	t.Run("err case scan row", func(t *testing.T) {
		e := &entities.OrderItem{}
		fields, values := e.FieldMap()
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		_, err := mockOrderItemRepo.GetOrderItemByStudentProductID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

	})
}
