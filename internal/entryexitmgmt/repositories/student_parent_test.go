package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	userEntities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	pgx "github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentParentRepoWithSqlMock() (*StudentParentRepo, *testutil.MockDB) {
	repo := &StudentParentRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentEntryExitRecordsRepo_GetParentIDsByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockStudentParentRepo, mockDB := StudentParentRepoWithSqlMock()
	mockE := &userEntities.StudentParent{}
	fields, fieldMap := mockE.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("get student parents failed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := mockStudentParentRepo.GetParentIDsByStudentID(ctx, mockDB.DB, string(mock.AnythingOfType("string")))

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err GetParentIDsByStudentID StudentParentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})
		latestRecord, err := mockStudentParentRepo.GetParentIDsByStudentID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, []string([]string{""}), latestRecord)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("No rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := mockStudentParentRepo.GetParentIDsByStudentID(ctx, mockDB.DB, string(mock.AnythingOfType("string")))

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err GetParentIDsByStudentID StudentParentRepo: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
