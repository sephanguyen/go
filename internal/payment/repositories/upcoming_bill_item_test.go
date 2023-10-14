package repositories

import (
	"context"
	"fmt"
	"github.com/manabie-com/backend/internal/payment/utils"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UpcomingBillItemRepoWithSqlMock() (*UpcomingBillItemRepo, *testutil.MockDB) {
	repo := &UpcomingBillItemRepo{}
	return repo, testutil.NewMockDB()
}

func TestUpcomingBillItemRepo_GetAllUpcomingStudentCourseByUpcomingStudentPackageID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	upcomingBillItemRepoWithSqlMock, mockDB := UpcomingBillItemRepoWithSqlMock()
	upcomingBillItem := &entities.UpcomingBillItem{}
	t.Run("Success", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		fields, _ := upcomingBillItem.FieldMap()
		scanFields := database.GetScanFields(upcomingBillItem, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := upcomingBillItemRepoWithSqlMock.GetUpcomingBillItemsForGenerate(ctx, mockDB.DB)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("err case", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		fields, _ := upcomingBillItem.FieldMap()
		scanFields := database.GetScanFields(upcomingBillItem, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := upcomingBillItemRepoWithSqlMock.GetUpcomingBillItemsForGenerate(ctx, mockDB.DB)
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUpcomingBillItemRepo_RemoveOldUpcomingBillItem(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	upcomingBillItemRepoWithSqlMock, mockDB := UpcomingBillItemRepoWithSqlMock()

	var (
		billingSchedulePeriodID string
		billingDate             time.Time
	)
	t.Run("Success", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Scan", &billingDate, &billingSchedulePeriodID).Once().Return(nil)
		_, _, err := upcomingBillItemRepoWithSqlMock.RemoveOldUpcomingBillItem(ctx, mockDB.DB, "1", "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("err case", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Scan", &billingDate, &billingSchedulePeriodID).Once().Return(constant.ErrDefault)
		_, _, err := upcomingBillItemRepoWithSqlMock.RemoveOldUpcomingBillItem(ctx, mockDB.DB, "1", "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUpcomingBillItemRepo_VoidUpcomingBillItemByOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUpcomingBillItemRepo, mockDB := UpcomingBillItemRepoWithSqlMock()
	db := mockDB.DB

	mockEntity := &entities.UpcomingBillItem{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now(), updated_at = now() 
						WHERE order_id = $1 AND is_generated = false
						AND deleted_at IS NULL;`, mockEntity.TableName())
	testCases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         nil,
			Req:         constant.OrderID,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, sql, constant.OrderID).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:        "Failed case: Error when exec",
			Req:         constant.OrderID,
			ExpectedErr: fmt.Errorf("err db.Exec UpcomingBillItemRepo.VoidUpcomingBillItemsByOrderID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, sql, constant.OrderID).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			orderID := testCase.Req.(string)
			err := mockUpcomingBillItemRepo.VoidUpcomingBillItemsByOrderID(ctx, db, orderID)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestUpcomingBillItemRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.UpcomingBillItem{}
	_, fieldMap := mockE.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("Success", func(t *testing.T) {
		r, mockDB := UpcomingBillItemRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		err := r.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert", func(t *testing.T) {
		r, mockDB := UpcomingBillItemRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)
		err := r.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert upcomingBillItem: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestUpcomingBillItemRepo_GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	upcomingBillItemRepoWithSqlMock, mockDB := UpcomingBillItemRepoWithSqlMock()
	upcomingBillItem := &entities.UpcomingBillItem{}
	t.Run("Success", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		fields, _ := upcomingBillItem.FieldMap()
		scanFields := database.GetScanFields(upcomingBillItem, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := upcomingBillItemRepoWithSqlMock.GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(ctx, mockDB.DB, mock.Anything, mock.Anything, mock.Anything)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("err case", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		fields, _ := upcomingBillItem.FieldMap()
		scanFields := database.GetScanFields(upcomingBillItem, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := upcomingBillItemRepoWithSqlMock.GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(ctx, mockDB.DB, mock.Anything, mock.Anything, mock.Anything)
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
