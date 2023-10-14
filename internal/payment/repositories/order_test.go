package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrderRepoWithSqlMock() (*OrderRepo, *testutil.MockDB) {
	orderRepo := &OrderRepo{}
	return orderRepo, testutil.NewMockDB()
}
func TestOrderRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.Order{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)
		err := orderRepoWithSqlMock.Create(ctx, mockDB.DB, &entities.Order{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert order fail", func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		err := orderRepoWithSqlMock.Create(ctx, mockDB.DB, &mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, pgx.ErrTxClosed.Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUpdateIsReviewFlagByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.SuccessCommandTag, nil)
		err := orderRepoWithSqlMock.UpdateIsReviewFlagByOrderID(ctx, mockDB.DB, "1", true, 0)
		assert.Nil(t, err)
	})
	t.Run("Update order fail", func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.FailCommandTag, constant.ErrDefault)
		err := orderRepoWithSqlMock.UpdateIsReviewFlagByOrderID(ctx, mockDB.DB, "1", true, 0)
		assert.NotNil(t, err)
	})
	t.Run("Update order fail with no row", func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.FailCommandTag, nil)
		err := orderRepoWithSqlMock.UpdateIsReviewFlagByOrderID(ctx, mockDB.DB, "1", true, 0)
		assert.NotNil(t, err)
	})
}

func TestUpdateOrderStatusByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.SuccessCommandTag, nil)
		err := orderRepoWithSqlMock.UpdateOrderStatusByOrderID(ctx, mockDB.DB, "1", "true")
		assert.Nil(t, err)
	})
	t.Run("Update order fail", func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.FailCommandTag, constant.ErrDefault)
		err := orderRepoWithSqlMock.UpdateOrderStatusByOrderID(ctx, mockDB.DB, "1", "true")
		assert.NotNil(t, err)
	})
	t.Run("Update order fail with no row", func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.FailCommandTag, nil)
		err := orderRepoWithSqlMock.UpdateOrderStatusByOrderID(ctx, mockDB.DB, "1", "true")
		assert.NotNil(t, err)
	})
}

func TestOrderRepoGetOrderIDForUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.Order{}
	fieldName, _ := mockEntities.FieldMap()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		scanFields := database.GetScanFields(&entities.Order{}, fieldName)
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
		mockDB.Row.On("Scan", scanFields...).Return(nil)
		_, err := orderRepoWithSqlMock.GetOrderByIDForUpdate(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		scanFields := database.GetScanFields(&entities.Order{}, fieldName)
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
		mockDB.Row.On("Scan", scanFields...).Return(constant.ErrDefault)
		_, err := orderRepoWithSqlMock.GetOrderByIDForUpdate(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestOrderRepoGetOrderTypeByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)
		_, err := orderRepoWithSqlMock.GetOrderTypeByOrderID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(constant.ErrDefault)
		_, err := orderRepoWithSqlMock.GetOrderTypeByOrderID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestOrderRepoGetAll(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		order := &entities.Order{}
		fields, _ := order.FieldMap()
		scanFields := database.GetScanFields(order, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.GetAll(ctx, mockDB.DB)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		order := &entities.Order{}
		fields, _ := order.FieldMap()
		scanFields := database.GetScanFields(order, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.GetAll(ctx, mockDB.DB)
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestGetOrderByStudentIDAndLocationIDsPaging(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		order := &entities.Order{}
		fields, _ := order.FieldMap()
		scanFields := database.GetScanFields(order, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.GetOrderByStudentIDAndLocationIDsPaging(ctx, mockDB.DB, "1", []string{"location_1", "location_2"}, int64(1), int64(10))
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		order := &entities.Order{}
		fields, _ := order.FieldMap()
		scanFields := database.GetScanFields(order, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.GetOrderByStudentIDAndLocationIDsPaging(ctx, mockDB.DB, "1", []string{}, int64(1), int64(10))
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		order := &entities.Order{}
		fields, _ := order.FieldMap()
		scanFields := database.GetScanFields(order, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.GetOrderByStudentIDAndLocationIDsPaging(ctx, mockDB.DB, "1", []string{"location_1", "location_2"}, int64(1), int64(10))
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestCountOrderByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.CountOrderByStudentIDAndLocationIDs(ctx, mockDB.DB, "1", []string{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := orderRepoWithSqlMock.CountOrderByStudentIDAndLocationIDs(ctx, mockDB.DB, "1", []string{"locaiton_1, location_2"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestOrderRepo_buildGetListOfOrdersWithFilterQuery(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		orderRepoWithSqlMock *OrderRepo
		studentName                = constant.StudentName
		orderStatus                = pb.OrderStatus_ORDER_STATUS_INVOICED.String()
		orderTypes                 = []string{pb.OrderType_ORDER_TYPE_NEW.String()}
		orderIDs                   = []string{constant.OrderID}
		locationIDs                = []string{constant.LocationID}
		createdFrom                = time.Now().AddDate(0, 0, -1)
		createdTo                  = time.Now().AddDate(0, 0, 1)
		isReviewed                 = false
		isStudentNotEnrolled       = true
		limit                int64 = 10
		offset               int64 = 5
	)
	table := entities.Order{}
	fieldNames, _ := table.FieldMap()
	orderFieldNamesWithPrefix := sliceutils.Map(fieldNames, func(fieldName string) string {
		return fmt.Sprintf("o.%s", fieldName)
	})
	isStudentNotEnrolledQuery := fmt.Sprintf(`JOIN (SELECT sesh.student_id FROM student_enrollment_status_history sesh 
WHERE sesh.student_id NOT IN(
	SELECT DISTINCT ON  (sesh.student_id, sesh.location_id)
	sesh.student_id 
	FROM student_enrollment_status_history sesh
	WHERE 
		(sesh.enrollment_status IN (
		'STUDENT_ENROLLMENT_STATUS_ENROLLED',
		'STUDENT_ENROLLMENT_STATUS_LOA')
		AND now() >= sesh.start_date)
	OR 
		(sesh.enrollment_status IN (
		'STUDENT_ENROLLMENT_STATUS_WITHDRAWN',
		'STUDENT_ENROLLMENT_STATUS_GRADUATED',
		'STUDENT_ENROLLMENT_STATUS_LOA')
		AND now() < sesh.start_date)
	ORDER BY sesh.student_id, sesh.location_id, sesh.start_date DESC
) 
GROUP BY sesh.student_id) as shbes ON shbes.student_id = o.student_id`)

	expectedQueryWithFullFilter := fmt.Sprintf(`
			SELECT %s FROM "%s" o %s
			WHERE student_full_name ~* '.*%s.*' AND o.order_status = $1 AND o.order_type = ANY($2) AND o.order_id = ANY($3) AND o.location_id = ANY($4) AND o.is_reviewed = false AND o.created_at >= $5 AND o.created_at <= $6 ORDER BY o.created_at DESC OFFSET $7 LIMIT $8`, strings.Join(orderFieldNamesWithPrefix, ","), table.TableName(), isStudentNotEnrolledQuery, studentName)
	expectedQueryWithoutPaging := fmt.Sprintf(`
			SELECT %s FROM "%s" o %s
			WHERE student_full_name ~* '.*%s.*' AND o.order_status = $1 AND o.order_type = ANY($2) AND o.order_id = ANY($3) AND o.location_id = ANY($4) AND o.is_reviewed = false AND o.created_at >= $5 AND o.created_at <= $6 ORDER BY o.created_at DESC`, strings.Join(orderFieldNamesWithPrefix, ","), table.TableName(), isStudentNotEnrolledQuery, studentName)
	expectedQueryWithoutFilter := fmt.Sprintf(`
			SELECT %s FROM "%s" o 
			WHERE student_full_name ~* '.*.*' ORDER BY o.created_at DESC`, strings.Join(orderFieldNamesWithPrefix, ","), table.TableName())
	testcases := []utils.TestCase{
		{
			Name: "Happy case (without paging)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName:          studentName,
				OrderStatus:          orderStatus,
				OrderTypes:           orderTypes,
				OrderIDs:             orderIDs,
				LocationIDs:          locationIDs,
				CreatedFrom:          createdFrom,
				CreatedTo:            createdTo,
				IsReviewed:           &isReviewed,
				IsStudentNotEnrolled: isStudentNotEnrolled,
			},
			ExpectedResp: expectedQueryWithoutPaging,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Happy case (full of filter fields)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName:          studentName,
				OrderStatus:          orderStatus,
				OrderTypes:           orderTypes,
				OrderIDs:             orderIDs,
				LocationIDs:          locationIDs,
				CreatedFrom:          createdFrom,
				CreatedTo:            createdTo,
				IsReviewed:           &isReviewed,
				IsStudentNotEnrolled: isStudentNotEnrolled,
				Limit:                &limit,
				Offset:               &offset,
			},
			ExpectedResp: expectedQueryWithFullFilter,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:         "Happy case (without filter fields)",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          OrderListFilter{},
			ExpectedResp: expectedQueryWithoutFilter,
			Setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			req := testCase.Req.(OrderListFilter)
			query, _ := orderRepoWithSqlMock.buildGetListOfOrdersWithFilterQuery(req)
			assert.Equal(t, testCase.ExpectedResp.(string), query)
		})
	}
}

func TestOrderRepo_GetOrdersByFilter(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		orderRepoWithSqlMock *OrderRepo
		mockDB               *testutil.MockDB

		studentName                 = constant.StudentName
		orderStatus                 = pb.OrderStatus_ORDER_STATUS_INVOICED.String()
		orderTypes                  = []string{pb.OrderType_ORDER_TYPE_NEW.String()}
		orderIDs                    = []string{constant.OrderID}
		locationIDs                 = []string{constant.LocationID}
		createdFrom                 = time.Now().AddDate(0, 0, -1)
		createdTo                   = time.Now().AddDate(0, 0, 1)
		isReviewed                  = false
		isOnlyStudentEnrolled       = false
		limit                 int64 = 10
		offset                int64 = 0
	)

	expectedOrders := []*entities.Order{
		{
			OrderID: pgtype.Text{
				String: constant.OrderID,
			},
			StudentID: pgtype.Text{
				String: constant.StudentID,
			},
			StudentFullName: pgtype.Text{
				String: constant.StudentName,
			},
			LocationID: pgtype.Text{
				String: constant.LocationID,
			},
			OrderStatus: pgtype.Text{
				String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
			},
			OrderType: pgtype.Text{
				String: pb.OrderType_ORDER_TYPE_NEW.String(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			IsReviewed: pgtype.Bool{
				Bool: false,
			},
			Background: pgtype.Text{
				String: "",
			},
			FutureMeasures: pgtype.Text{
				String: "",
			},
		},
	}

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when query",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName:          studentName,
				OrderStatus:          orderStatus,
				OrderTypes:           orderTypes,
				OrderIDs:             orderIDs,
				LocationIDs:          locationIDs,
				CreatedFrom:          createdFrom,
				CreatedTo:            createdTo,
				IsReviewed:           &isReviewed,
				IsStudentNotEnrolled: isOnlyStudentEnrolled,
				Limit:                &limit,
				Offset:               &offset,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			Name: "Fail case: Error when scan rows",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName: studentName,
				OrderStatus: orderStatus,
				OrderTypes:  orderTypes,
				OrderIDs:    orderIDs,
				LocationIDs: locationIDs,
				CreatedFrom: createdFrom,
				CreatedTo:   createdTo,
				IsReviewed:  &isReviewed,
				Limit:       &limit,
				Offset:      &offset,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)
				order := &entities.Order{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName: studentName,
				OrderStatus: orderStatus,
				OrderTypes:  orderTypes,
				OrderIDs:    orderIDs,
				LocationIDs: locationIDs,
				CreatedFrom: createdFrom,
				CreatedTo:   createdTo,
				IsReviewed:  &isReviewed,
				Limit:       &limit,
				Offset:      &offset,
			},
			Setup: func(ctx context.Context) {
				order := &entities.Order{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)

				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.Rows.On("Next").Once().Return(true)

				for _, expectedOrder := range expectedOrders {
					refExpectedOrder := expectedOrder
					mockDB.Rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedOrder.OrderID.String
						args[1].(*pgtype.Text).String = refExpectedOrder.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedOrder.StudentFullName.String
						args[3].(*pgtype.Text).String = refExpectedOrder.LocationID.String
						args[4].(*pgtype.Text).String = refExpectedOrder.OrderComment.String
						args[5].(*pgtype.Text).String = refExpectedOrder.OrderStatus.String
						args[6].(*pgtype.Text).String = refExpectedOrder.OrderType.String
						args[7].(*pgtype.Timestamptz).Time = refExpectedOrder.UpdatedAt.Time
						args[8].(*pgtype.Timestamptz).Time = refExpectedOrder.CreatedAt.Time
						args[9].(*pgtype.Bool).Bool = refExpectedOrder.IsReviewed.Bool
						args[10].(*pgtype.Timestamptz).Time = refExpectedOrder.WithdrawalEffectiveDate.Time
						args[11].(*pgtype.Timestamptz).Time = refExpectedOrder.LOAStartDate.Time
						args[12].(*pgtype.Timestamptz).Time = refExpectedOrder.LOAEndDate.Time
						args[13].(*pgtype.Text).String = refExpectedOrder.Background.String
						args[14].(*pgtype.Text).String = refExpectedOrder.FutureMeasures.String
						args[15].(*pgtype.Int4).Int = refExpectedOrder.OrderSequenceNumber.Int
						args[16].(*pgtype.Int4).Int = refExpectedOrder.VersionNumber.Int
						args[17].(*pgtype.Text).String = refExpectedOrder.ResourcePath.String
					}).Return(nil)
				}
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderRepoWithSqlMock, mockDB = OrderRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(OrderListFilter)

			_, err := orderRepoWithSqlMock.GetOrdersByFilter(testCase.Ctx, mockDB.DB, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderRepo_GetOrderStatsByFilter(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		orderRepoWithSqlMock *OrderRepo
		mockDB               *testutil.MockDB

		studentName       = constant.StudentName
		orderStatus       = pb.OrderStatus_ORDER_STATUS_INVOICED.String()
		orderTypes        = []string{pb.OrderType_ORDER_TYPE_NEW.String()}
		orderIDs          = []string{constant.OrderID}
		locationIDs       = []string{constant.LocationID}
		createdFrom       = time.Now().AddDate(0, 0, -1)
		createdTo         = time.Now().AddDate(0, 0, 1)
		isReviewed        = false
		limit       int64 = 10
		offset      int64 = 1
	)

	expectedOrderStats := &entities.OrderStats{
		TotalItems: pgtype.Int8{
			Int: 2,
		},
		TotalOfSubmitted: pgtype.Int8{
			Int: 2,
		},
		TotalOfPending:      pgtype.Int8{},
		TotalOfRejected:     pgtype.Int8{},
		TotalOfVoided:       pgtype.Int8{},
		TotalOfInvoiced:     pgtype.Int8{},
		TotalOfNeedToReview: pgtype.Int8{},
	}
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when scan",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName: studentName,
				OrderStatus: orderStatus,
				OrderTypes:  orderTypes,
				OrderIDs:    orderIDs,
				LocationIDs: locationIDs,
				CreatedFrom: createdFrom,
				CreatedTo:   createdTo,
				IsReviewed:  &isReviewed,
				Limit:       &limit,
				Offset:      &offset,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderStats := &entities.OrderStats{}
				fields, values := orderStats.FieldOrderStatsMap()
				scanFields := utils.GetScanFields(fields, values, fields)

				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: OrderListFilter{
				StudentName: studentName,
				OrderStatus: orderStatus,
				OrderTypes:  orderTypes,
				OrderIDs:    orderIDs,
				LocationIDs: locationIDs,
				CreatedFrom: createdFrom,
				CreatedTo:   createdTo,
				IsReviewed:  &isReviewed,
				Limit:       &limit,
				Offset:      &offset,
			},
			ExpectedResp: expectedOrderStats,
			Setup: func(ctx context.Context) {
				orderStats := &entities.OrderStats{}
				fields, values := orderStats.FieldOrderStatsMap()
				scanFields := utils.GetScanFields(fields, values, fields)

				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
					refExpectedOrderStats := expectedOrderStats
					args[0].(*pgtype.Int8).Int = refExpectedOrderStats.TotalItems.Int
					args[1].(*pgtype.Int8).Int = refExpectedOrderStats.TotalOfSubmitted.Int
					args[2].(*pgtype.Int8).Int = refExpectedOrderStats.TotalOfPending.Int
					args[3].(*pgtype.Int8).Int = refExpectedOrderStats.TotalOfRejected.Int
					args[4].(*pgtype.Int8).Int = refExpectedOrderStats.TotalOfVoided.Int
					args[5].(*pgtype.Int8).Int = refExpectedOrderStats.TotalOfInvoiced.Int
					args[6].(*pgtype.Int8).Int = refExpectedOrderStats.TotalOfNeedToReview.Int
				}).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderRepoWithSqlMock, mockDB = OrderRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(OrderListFilter)
			orderStatsResp, err := orderRepoWithSqlMock.GetOrderStatsByFilter(testCase.Ctx, mockDB.DB, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedResp := testCase.ExpectedResp.(*entities.OrderStats)
				assert.Equal(t, expectedResp.TotalItems, orderStatsResp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfRejected, orderStatsResp.TotalOfRejected)
				assert.Equal(t, expectedResp.TotalOfInvoiced, orderStatsResp.TotalOfInvoiced)
				assert.Equal(t, expectedResp.TotalOfPending, orderStatsResp.TotalOfPending)
				assert.Equal(t, expectedResp.TotalOfSubmitted, orderStatsResp.TotalOfSubmitted)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderRepoGetOrderByStudentIDAndLocationIDForResume(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)
		_, err := orderRepoWithSqlMock.GetOrderByStudentIDAndLocationIDForResume(ctx, mockDB.DB, "1", "2")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		orderRepoWithSqlMock, mockDB := OrderRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(constant.ErrDefault)
		_, err := orderRepoWithSqlMock.GetOrderByStudentIDAndLocationIDForResume(ctx, mockDB.DB, "1", "2")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestOrderRepo_GetLatestOrderByStudentIDAndLocationIDAndOrderType(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		orderRepoWithSqlMock *OrderRepo
		mockDB               *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when query row",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentID,
				constant.LocationID,
				pb.OrderType_ORDER_TYPE_LOA.String(),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				_, fieldValues := (&entities.Order{}).FieldMap()
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentID,
				constant.LocationID,
				pb.OrderType_ORDER_TYPE_LOA.String(),
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				_, fieldValues := (&entities.Order{}).FieldMap()
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderRepoWithSqlMock, mockDB = OrderRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			studentID := testCase.Req.([]interface{})[0].(string)
			locationID := testCase.Req.([]interface{})[1].(string)
			orderType := testCase.Req.([]interface{})[2].(string)

			_, err := orderRepoWithSqlMock.GetLatestOrderByStudentIDAndLocationIDAndOrderType(testCase.Ctx, mockDB.DB, studentID, locationID, orderType)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderRepo_GetLatestOrderByStudentIDAndLocationID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		orderRepoWithSqlMock *OrderRepo
		mockDB               *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when query row",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentID,
				constant.LocationID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				_, fieldValues := (&entities.Order{}).FieldMap()
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.StudentID,
				constant.LocationID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				_, fieldValues := (&entities.Order{}).FieldMap()
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			orderRepoWithSqlMock, mockDB = OrderRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			studentID := testCase.Req.([]interface{})[0].(string)
			locationID := testCase.Req.([]interface{})[1].(string)

			_, err := orderRepoWithSqlMock.GetLatestOrderByStudentIDAndLocationID(testCase.Ctx, mockDB.DB, studentID, locationID)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}
