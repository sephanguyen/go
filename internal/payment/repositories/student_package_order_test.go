package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func StudentPackageOrderRepoWithSqlMock() (*StudentPackageOrderRepo, *testutil.MockDB) {
	studentPackageOrderRepo := &StudentPackageOrderRepo{}
	return studentPackageOrderRepo, testutil.NewMockDB()
}

func TestStudentPackageOrderRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.StudentPackageOrder{}
	_, fieldMap := mockEntities.FieldMap()
	tag := pgconn.CommandTag{1}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run(constant.HappyCase, func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		err := studentProductRepoWithSqlMock.Create(ctx, mockDB.DB, entities.StudentPackageOrder{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert student product fail", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(tag, pgx.ErrTxClosed)

		err := studentProductRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), pgx.ErrTxClosed.Error()))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageOrderRepo_UpdateExecuteError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mockDb.Ext{}
	upcomingStudentPackageRepo := &StudentPackageOrderRepo{}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: entities.StudentPackageOrder{
				ExecutedError: pgtype.Text{String: constant.ErrDefault.Error()},
			},
			ExpectedErr: fmt.Errorf("err db.Exec studentPackageOrderRepo.UpdateExecuteError: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  ctx,
			Req: entities.StudentPackageOrder{
				ExecutedError: pgtype.Text{String: constant.ErrDefault.Error()},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			upcomingStudentPackageReq := testCase.Req.(entities.StudentPackageOrder)
			err := upcomingStudentPackageRepo.UpdateExecuteError(testCase.Ctx, db, upcomingStudentPackageReq)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_UpdateExecuteStatus(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mockDb.Ext{}
	upcomingStudentPackageRepo := &StudentPackageOrderRepo{}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: entities.StudentPackageOrder{
				ExecutedError: pgtype.Text{String: constant.ErrDefault.Error()},
			},
			ExpectedErr: fmt.Errorf("err db.Exec studentPackageOrderRepo.UpdateExecuteStatus: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  ctx,
			Req: entities.StudentPackageOrder{
				ExecutedError: pgtype.Text{String: constant.ErrDefault.Error()},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			upcomingStudentPackageReq := testCase.Req.(entities.StudentPackageOrder)
			err := upcomingStudentPackageRepo.UpdateExecuteStatus(testCase.Ctx, db, upcomingStudentPackageReq)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_GetByStudentPackageOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageOrderRepo, mockDB := StudentPackageOrderRepoWithSqlMock()
	db := mockDB.DB

	mockEntity := &entities.StudentPackageOrder{}
	_, fieldValues := mockEntity.FieldMap()
	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when scan",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
			},
			ExpectedErr: fmt.Errorf("row.Scan studentPackageOrderRepo.GetByStudentPackageOrderID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			studentPackageOrderID := testCase.Req.([]interface{})[0].(string)
			_, err := studentPackageOrderRepo.GetByStudentPackageOrderID(testCase.Ctx, db, studentPackageOrderID)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_GetByStudentPackageIDAndOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageOrderRepo, mockDB := StudentPackageOrderRepoWithSqlMock()
	db := mockDB.DB

	mockEntity := &entities.StudentPackageOrder{}
	_, fieldValues := mockEntity.FieldMap()
	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when scan",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				constant.OrderID,
			},
			ExpectedErr: fmt.Errorf("row.Scan studentPackageOrderRepo.GetByStudentPackageIDAndOrderID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				constant.OrderID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			studentPackageID := testCase.Req.([]interface{})[0].(string)
			orderID := testCase.Req.([]interface{})[1].(string)
			_, err := studentPackageOrderRepo.GetByStudentPackageIDAndOrderID(testCase.Ctx, db, studentPackageID, orderID)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_Upsert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.StudentPackageOrder{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := studentPackageRepoWithSqlMock.Upsert(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student package fail", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentPackageRepoWithSqlMock.Upsert(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error when upsert student package order studentPackageOrderRepo.Upsert: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update student package", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := studentPackageRepoWithSqlMock.Upsert(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, "error when upsert student package order with no affected row", err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageOrderRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.StudentPackageOrder{}
	_, fieldMap := mockEntities.FieldMap()
	tag := pgconn.CommandTag{1}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run(constant.HappyCase, func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		err := studentProductRepoWithSqlMock.Update(ctx, mockDB.DB, entities.StudentPackageOrder{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert student product fail", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageOrderRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(tag, pgx.ErrTxClosed)

		err := studentProductRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), pgx.ErrTxClosed.Error()))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageOrderRepo_RevertByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mockDb.Ext{}
	studentPackageOrderRepo := &StudentPackageOrderRepo{}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedErr: fmt.Errorf("err db.Exec studentPackageOrderRepo.RevertByID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			id := testCase.Req.([]interface{})[0].(string)
			err := studentPackageOrderRepo.RevertByID(testCase.Ctx, db, id)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_SoftDeleteByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mockDb.Ext{}
	studentPackageOrderRepo := &StudentPackageOrderRepo{}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: []interface{}{
				constant.UpcomingStudentPackageID,
			},
			ExpectedErr: fmt.Errorf("err db.Exec studentPackageOrderRepo.SoftDeleteByID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.UpcomingStudentPackageID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			id := testCase.Req.([]interface{})[0].(string)
			err := studentPackageOrderRepo.SoftDeleteByID(testCase.Ctx, db, id)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_ResetCurrentPosition(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mockDb.Ext{}
	studentPackageOrderRepo := &StudentPackageOrderRepo{}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: []interface{}{
				constant.UpcomingStudentPackageID,
			},
			ExpectedErr: fmt.Errorf("err db.Exec studentPackageOrderRepo.ResetCurrentPosition: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.UpcomingStudentPackageID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			id := testCase.Req.([]interface{})[0].(string)
			err := studentPackageOrderRepo.ResetCurrentPosition(testCase.Ctx, db, id)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_SetCurrentStudentPackageByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mockDb.Ext{}
	studentPackageOrderRepo := &StudentPackageOrderRepo{}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
				true,
			},
			ExpectedErr: fmt.Errorf("err db.Exec studentPackageOrderRepo.SetCurrentStudentPackageByID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
				true,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.FailCommandTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			id := testCase.Req.([]interface{})[0].(string)
			isCurrent := testCase.Req.([]interface{})[1].(bool)
			err := studentPackageOrderRepo.SetCurrentStudentPackageByID(testCase.Ctx, db, id, isCurrent)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageOrderRepo_GetStudentPackageOrdersByStudentPackageID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageOrderRepo, db := StudentPackageOrderRepoWithSqlMock()
	mockEntity := &entities.StudentPackageOrder{}
	_, fieldValues := mockEntity.FieldMap()
	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when exec",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
			},
			ExpectedErr: fmt.Errorf("error when query studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(db.Rows, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: Error when scan",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
			},
			ExpectedErr: fmt.Errorf(constant.RowScanError, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(db.Rows, nil)
				db.Rows.On("Next").Once().Return(true)
				db.Rows.On("Scan", fieldValues...).Return(constant.ErrDefault)
				db.Rows.On("Close").Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
			},
			ExpectedErr: fmt.Errorf(constant.RowScanError, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(db.Rows, nil)
				db.Rows.On("Next").Once().Return(true)
				db.Rows.On("Scan", fieldValues...).Return(nil)
				db.Rows.On("Close").Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			id := testCase.Req.([]interface{})[0].(string)
			_, err := studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID(testCase.Ctx, db.DB, id)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}
