package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ZoomAccountRepoWithSqlMock() (*ZoomAccountRepo, *testutil.MockDB) {
	r := &ZoomAccountRepo{}
	return r, testutil.NewMockDB()
}

func TestZoomAccountRepo_GetZoomAccountByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	z, mockDB := ZoomAccountRepoWithSqlMock()
	db := mockDB.DB
	row := mockDB.Row
	zoomAccountID := "zoom-account-id"
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &zoomAccountID)

		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(puddle.ErrClosedPool)
		details, err := z.GetZoomAccountByID(ctx, mockDB.DB, zoomAccountID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, details)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &zoomAccountID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		details, err := z.GetZoomAccountByID(ctx, mockDB.DB, zoomAccountID)
		assert.NoError(t, err)
		assert.NotNil(t, details)
	})
}

func TestZoomAccountRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	z, mockDB := ZoomAccountRepoWithSqlMock()
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := z.Upsert(ctx, mockDB.DB, domain.ZoomAccounts{&domain.ZoomAccount{ID: "123"}})

		assert.NotNil(t, err)
	})
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := z.Upsert(ctx, mockDB.DB, domain.ZoomAccounts{&domain.ZoomAccount{ID: "123"}})
		assert.NoError(t, err)
	})
}

func TestZoomAccountRepo_GetAllZoomAccount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	z, mockDB := ZoomAccountRepoWithSqlMock()
	zoomAccount := &ZoomAccount{}
	fields, values := zoomAccount.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)
		dateInfos, err := z.GetAllZoomAccount(ctx, mockDB.DB)

		assert.NoError(t, err)
		assert.NotNil(t, dateInfos)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		dateInfos, err := z.GetAllZoomAccount(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, dateInfos)

	})
}
