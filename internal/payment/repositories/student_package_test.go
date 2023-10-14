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

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StudentPackageRepoWithSqlMock() (*StudentPackageRepo, *testutil.MockDB) {
	studentPackageRepo := &StudentPackageRepo{}
	return studentPackageRepo, testutil.NewMockDB()
}

func TestStudentPackageRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.StudentPackages{}
	_, fieldMap := mockEntities.FieldMap()
	tag := pgconn.CommandTag{1}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run(constant.HappyCase, func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(tag, nil)
		err := studentProductRepoWithSqlMock.Insert(ctx, mockDB.DB, &entities.StudentPackages{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Insert student product fail", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(tag, pgx.ErrTxClosed)

		err := studentProductRepoWithSqlMock.Insert(ctx, mockDB.DB, &mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), pgx.ErrTxClosed.Error()))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.StudentPackages{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := studentPackageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student package fail", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentPackageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), fmt.Errorf("update student package have error: %w", pgx.ErrTxClosed).Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update student package", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := studentPackageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, "update student package have no row affected", err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var studentPackageID string = "1"
	studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t,
			mock.Anything,
			mock.Anything,
			studentPackageID,
		)
		entity := &entities.StudentPackages{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		period, err := studentPackageRepoWithSqlMock.GetByID(ctx, mockDB.DB, studentPackageID)
		assert.Nil(t, err)
		assert.NotNil(t, period)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			studentPackageID,
		)
		e := &entities.StudentPackages{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := studentPackageRepoWithSqlMock.GetByID(ctx, mockDB.DB, studentPackageID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)
	})
}

func TestStudentPackageRepo_GetStudentPackageForUpsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run(constant.HappyCase, func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
		row := mockDB.Row
		mockDB.DB.On("QueryRow",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(row)
		row.On("Scan", mock.Anything).Return(nil)
		entity := &entities.StudentPackages{
			LocationIDs: pgtype.TextArray{
				Elements: []pgtype.Text{{}},
			},
		}
		period, err := studentPackageRepoWithSqlMock.GetStudentPackageForUpsert(ctx, mockDB.DB, entity)
		assert.Nil(t, err)
		assert.NotNil(t, period)
	})
	t.Run("err case", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
		row := mockDB.Row
		mockDB.DB.On("QueryRow",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(row)
		row.On("Scan", mock.Anything).Return(constant.ErrDefault)
		entity := &entities.StudentPackages{
			LocationIDs: pgtype.TextArray{
				Elements: []pgtype.Text{{}},
			},
		}
		_, err := studentPackageRepoWithSqlMock.GetStudentPackageForUpsert(ctx, mockDB.DB, entity)
		assert.NotNil(t, err)
	})
	t.Run("happy case with zero row", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
		row := mockDB.Row
		mockDB.DB.On("QueryRow",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(row)
		row.On("Scan", mock.Anything).Return(pgx.ErrNoRows)
		entity := &entities.StudentPackages{
			LocationIDs: pgtype.TextArray{
				Elements: []pgtype.Text{{}},
			},
		}
		_, err := studentPackageRepoWithSqlMock.GetStudentPackageForUpsert(ctx, mockDB.DB, entity)
		assert.Nil(t, err)
	})
}

func TestStudentPackageRepo_Upsert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.StudentPackages{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := studentPackageRepoWithSqlMock.Upsert(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student package fail", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentPackageRepoWithSqlMock.Upsert(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error when upsert student package: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update student package", func(t *testing.T) {
		studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := studentPackageRepoWithSqlMock.Upsert(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, "upsert student package have no row affected", err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageRepo_SoftDeleteByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
	db := mockDB.DB

	now := time.Now()
	studentPackageIDs := []string{"student_package_id_1"}
	mockEntity := &entities.StudentPackages{}
	stmt := fmt.Sprintf(`UPDATE %s SET deleted_at = $1, updated_at = now() 
                         WHERE student_package_id = ANY($2) 
                           AND deleted_at IS NULL`, mockEntity.TableName())
	args := []interface{}{
		mock.Anything,
		stmt,
		now,
		database.TextArray(studentPackageIDs),
	}

	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  nil,
			Req: []interface{}{
				now,
				studentPackageIDs,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name: "Failed case: Error when exec",
			Req: []interface{}{
				now,
				studentPackageIDs,
			},
			ExpectedErr: fmt.Errorf("err db.Exec StudentPackageRepo.SoftDeleteByIDs: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			deletedAtReq := (testCase.Req.([]interface{})[0]).(time.Time)
			studentPackageIDsReq := (testCase.Req.([]interface{})[1]).([]string)
			err := studentPackageRepoWithSqlMock.SoftDeleteByIDs(ctx, db, studentPackageIDsReq, deletedAtReq)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentPackageRepo_UpdateTimeByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
	db := mockDB.DB

	args := []interface{}{
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}

	testCases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         nil,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:        "Failed case: Error when exec",
			ExpectedErr: fmt.Errorf("update time student package have error: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			err := studentPackageRepoWithSqlMock.UpdateTimeByID(ctx, db, "1", time.Now())
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentPackageRepo_CancelByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
	db := mockDB.DB

	args := []interface{}{
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}

	testCases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         nil,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:        "Failed case: Error when exec",
			ExpectedErr: fmt.Errorf("cancel student package have error: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			err := studentPackageRepoWithSqlMock.CancelByID(ctx, db, "1")
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentPackageRepo_GetStudentPackagesForCronjob(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	upcomingStudentPackageRepoWithSqlMock, mockDB := StudentPackageRepoWithSqlMock()
	db := mockDB.DB
	rows := mockDB.Rows

	mockEntity := &entities.StudentPackages{}
	fields, _ := mockEntity.FieldMap()
	scanFields := database.GetScanFields(mockEntity, fields)
	args := []interface{}{
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}

	expectedStudentPackages := []entities.StudentPackages{
		{
			ID:          pgtype.Text{},
			StudentID:   pgtype.Text{},
			PackageID:   pgtype.Text{},
			StartAt:     pgtype.Timestamptz{},
			EndAt:       pgtype.Timestamptz{},
			Properties:  pgtype.JSONB{},
			IsActive:    pgtype.Bool{},
			LocationIDs: pgtype.TextArray{},
			CreatedAt:   pgtype.Timestamptz{},
			UpdatedAt:   pgtype.Timestamptz{},
			DeletedAt:   pgtype.Timestamptz{},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  nil,
			Req: []interface{}{
				3,
			},
			ExpectedResp: expectedStudentPackages,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedStudentPackages)).Run(func(args mock.Arguments) {
					rows.On("Scan", scanFields...).Once().Return(nil)
				}).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "Failed case: Error when query",
			Req: []interface{}{
				3,
			},
			ExpectedResp: expectedStudentPackages,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, constant.ErrDefault)
				rows.On("Next").Times(len(expectedStudentPackages)).Run(func(args mock.Arguments) {
					rows.On("Scan", scanFields...).Once().Return(nil)
				}).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "Failed case: Error scan fields",
			Req: []interface{}{
				3,
			},
			ExpectedResp: expectedStudentPackages,
			ExpectedErr:  status.Errorf(codes.Internal, "err when scan student package"),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(status.Errorf(codes.Internal, "err when scan student package"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			day := testCase.Req.([]interface{})[0].(int)
			_, err := upcomingStudentPackageRepoWithSqlMock.GetStudentPackagesForCronjobByDay(ctx, db, day)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}
