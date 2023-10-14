package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PartnerConvenienceStoreRepoWithSqlMock() (*PartnerConvenienceStoreRepo, *testutil.MockDB) {
	repo := &PartnerConvenienceStoreRepo{}
	return repo, testutil.NewMockDB()
}

func TestPartnerConvenienceStoreRepo_FindOne(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PartnerConvenienceStoreRepoWithSqlMock()
	mockE := &entities.PartnerConvenienceStore{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.FindOne(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.Equal(t, mockE, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		e, err := repo.FindOne(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindOne(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
