package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentDiscountTrackerRepoWithSqlMock() (*StudentDiscountTrackerRepo, *testutil.MockDB) {
	studentDiscountTrackerRepo := &StudentDiscountTrackerRepo{}
	return studentDiscountTrackerRepo, testutil.NewMockDB()
}

func TestStudentDiscountTrackerRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.StudentDiscountTracker{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discount, err := mockStudentDiscountTrackerRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, discount)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.StudentDiscountTracker{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := mockStudentDiscountTrackerRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)

	})
}

func TestStudentDiscountTrackerRepo_GetActiveTrackingByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := entities.StudentDiscountTracker{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		studentDiscountTrackers, err := mockStudentDiscountTrackerRepo.GetActiveTrackingByStudentIDs(ctx, mockDB.DB, []string{mock.Anything})
		assert.Nil(t, err)
		assert.Equal(t, e, studentDiscountTrackers[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := entities.StudentDiscountTracker{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		studentDiscountTrackers, err := mockStudentDiscountTrackerRepo.GetActiveTrackingByStudentIDs(ctx, mockDB.DB, []string{mock.Anything})
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, 0, len(studentDiscountTrackers))
	})
}

func TestStudentDiscountTrackerRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.StudentDiscountTracker{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := mockStudentDiscountTrackerRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert student discount tracker fail", func(t *testing.T) {
		mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := mockStudentDiscountTrackerRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentDiscountTracker: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert student discount tracker", func(t *testing.T) {
		mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := mockStudentDiscountTrackerRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentDiscountTracker: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestStudentDiscountTrackerRepo_UpdateTrackingDurationByStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := entities.StudentProduct{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := mockStudentDiscountTrackerRepo.UpdateTrackingDurationByStudentProduct(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("err case", func(t *testing.T) {
		mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := mockStudentDiscountTrackerRepo.UpdateTrackingDurationByStudentProduct(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update student discount tracker: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("err case: no rows affected", func(t *testing.T) {
		mockStudentDiscountTrackerRepo, mockDB := StudentDiscountTrackerRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := mockStudentDiscountTrackerRepo.UpdateTrackingDurationByStudentProduct(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "0 RowsAffected")

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
