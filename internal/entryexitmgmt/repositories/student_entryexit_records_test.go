package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentEntryExitRecordsRepoWithSqlMock() (*StudentEntryExitRecordsRepo, *testutil.MockDB) {
	repo := &StudentEntryExitRecordsRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentEntryExitRecordsRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentEntryExitRecords{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert student_entryexit_records failed", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentEntryExitRecordsRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after student_entryexit_records inserted", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentEntryExitRecordsRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentEntryExitRecordsRepo_GetLatestRecordByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()
	mockE := &entities.StudentEntryExitRecords{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("select student_entryexit_records failed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := repo.GetLatestRecordByID(ctx, mockDB.DB, string(mock.AnythingOfType("string")))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err GetLatestRecordByID StudentEntryExitRecordsRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		latestRecord, err := repo.GetLatestRecordByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, mockE, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("No rows affected after student_entryexit_records select", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.GetLatestRecordByID(ctx, mockDB.DB, string(mock.AnythingOfType("string")))

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err GetLatestRecordByID StudentEntryExitRecordsRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentEntryExitRecordsRepo_SoftDeleteByID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	entryExitID := database.Int4(1)
	mockE := &entities.StudentEntryExitRecords{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByID(ctx, mockDB.DB, entryExitID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("soft delete entry exit record fail", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteByID(ctx, mockDB.DB, entryExitID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete StudentEntryExitRecordsRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after soft deleting entry exit record", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByID(ctx, mockDB.DB, entryExitID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete StudentEntryExitRecordsRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentEntryExitRecordsRepo_Update(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentEntryExitRecords{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update entry exit record fail", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update StudentEntryExitRecordsRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after updating entry exit record", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update StudentEntryExitRecordsRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentEntryExitRecordsRepo_RetrieveRecordsByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := pgtype.Text{}
	_ = studentID.Set(uuid.NewString())
	selectFields := []string{"entryexit_id", "entry_at", "exit_at"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)
	repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()
	filterAll := RetrieveEntryExitRecordFilter{
		StudentID:    studentID,
		RecordFilter: eepb.RecordFilter_ALL,
		Limit:        pgtype.Int8{Int: 0},
	}
	filterLastMonth := RetrieveEntryExitRecordFilter{
		StudentID:    studentID,
		RecordFilter: eepb.RecordFilter_LAST_MONTH,
		Limit:        pgtype.Int8{Int: 0},
	}
	filterThisMonth := RetrieveEntryExitRecordFilter{
		StudentID:    studentID,
		RecordFilter: eepb.RecordFilter_THIS_MONTH,
		Limit:        pgtype.Int8{Int: 0},
	}
	filterThisYear := RetrieveEntryExitRecordFilter{
		StudentID:    studentID,
		RecordFilter: eepb.RecordFilter_THIS_YEAR,
		Limit:        pgtype.Int8{Int: 0},
	}
	t.Run("failed to select entry exit records", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		entryExitRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, filterAll)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err retrieve records StudentEntryExitRecordsRepo: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, entryExitRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("No rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		entryExitRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, filterAll)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err retrieve records StudentEntryExitRecordsRepo: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, entryExitRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success with select all", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		e := &entities.StudentEntryExitRecords{}
		_ = e.ID.Set(12)
		value := database.GetScanFields(e, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		entryExitRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, filterAll)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudentEntryExitRecords{
			{ID: database.Int4(12)},
		}, entryExitRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
	t.Run("success with select last month", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		e := &entities.StudentEntryExitRecords{}

		value := database.GetScanFields(e, selectFields)
		_ = e.ID.Set(1122)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		entryExitRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, filterLastMonth)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudentEntryExitRecords{
			{ID: database.Int4(1122)},
		}, entryExitRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
	t.Run("success with select this month", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		e := &entities.StudentEntryExitRecords{}
		_ = e.ID.Set(1155)

		value := database.GetScanFields(e, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		entryExitRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, filterThisMonth)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudentEntryExitRecords{
			{ID: database.Int4(1155)},
		}, entryExitRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
	t.Run("success with select this year", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		e := &entities.StudentEntryExitRecords{}
		_ = e.ID.Set(7)

		value := database.GetScanFields(e, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		entryExitRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, filterThisYear)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudentEntryExitRecords{
			{ID: database.Int4(7)},
		}, entryExitRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestStudentEntryExitRecordsRepo_LockAdvisoryByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.AnythingOfType("*bool")).Return(nil)

		lockAcquired, err := repo.LockAdvisoryByStudentID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, lockAcquired)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("tx closed", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.AnythingOfType("*bool")).Once().Return(pgx.ErrTxClosed)

		_, err := repo.LockAdvisoryByStudentID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err LockAdvisoryByStudentID StudentEntryExitRecordsRepo: %w - studentID: %s", pgx.ErrTxClosed, mock.Anything).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentEntryExitRecordsRepo_UnlockAdvisoryByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")})

		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := repo.UnLockAdvisoryByStudentID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("tx closed", func(t *testing.T) {
		repo, mockDB := StudentEntryExitRecordsRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")})

		mockDB.MockExecArgs(t, cmdTag, pgx.ErrTxClosed, args...)

		err := repo.UnLockAdvisoryByStudentID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err UnLockAdvisoryByStudentID StudentEntryExitRecordsRepo: %w - studentID: %s", pgx.ErrTxClosed, mock.Anything).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
