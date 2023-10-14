package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ClassDoAccountRepoWithSqlMock() (*ClassDoAccountRepo, *testutil.MockDB) {
	r := &ClassDoAccountRepo{}
	return r, testutil.NewMockDB()
}

func TestClassDoAccountRepo_UpsertClassDoAccounts(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockRepo, mockDB := ClassDoAccountRepoWithSqlMock()
	now := time.Now()
	mockData := domain.ClassDoAccounts{
		{
			ClassDoID:     "id-1",
			ClassDoEmail:  "user-1@email.com",
			ClassDoAPIKey: "APIKEY1234567",
			CreatedAt:     now,
			UpdatedAt:     now,
			DeletedAt:     nil,
		},
		{
			ClassDoID:     "id-2",
			ClassDoEmail:  "user-2@email.com",
			ClassDoAPIKey: "APIKEY1234567",
			CreatedAt:     now,
			UpdatedAt:     now,
			DeletedAt:     &now,
		},
		{
			ClassDoID:     "id-3",
			ClassDoEmail:  "user-3@email.com",
			ClassDoAPIKey: "APIKEY1234567",
			CreatedAt:     now,
			UpdatedAt:     now,
			DeletedAt:     nil,
		},
	}

	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(pgconn.CommandTag([]byte(`1`)), errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)

		err := mockRepo.UpsertClassDoAccounts(ctx, mockDB.DB, mockData)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(pgconn.CommandTag([]byte(`1`)), nil)
		batchResults.On("Close").Once().Return(nil)

		err := mockRepo.UpsertClassDoAccounts(ctx, mockDB.DB, mockData)
		assert.NoError(t, err)
	})
}

func TestClassDoAccountRepo_GetAllClassDoAccounts(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockRepo, mockDB := ClassDoAccountRepoWithSqlMock()
	classDoAccount := &ClassDoAccount{}
	fields, values := classDoAccount.FieldMap()

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)
		accounts, err := mockRepo.GetAllClassDoAccounts(ctx, mockDB.DB)

		assert.NoError(t, err)
		assert.NotNil(t, accounts)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)

		accounts, err := mockRepo.GetAllClassDoAccounts(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, accounts)

	})
}
