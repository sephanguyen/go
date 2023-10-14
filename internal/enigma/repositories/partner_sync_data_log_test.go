package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PartnerSyncDataLogRepoWithSqlMock() (*PartnerSyncDataLogRepo, *testutil.MockDB) {
	r := &PartnerSyncDataLogRepo{}
	return r, testutil.NewMockDB()
}

func TestPartnerSyncDataLogRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PartnerSyncDataLogRepoWithSqlMock()

	t.Run("err create", func(t *testing.T) {
		e := &entities.PartnerSyncDataLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.PartnerSyncDataLog{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestPartnerSyncDataLogRepo_GetBySignature(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PartnerSyncDataLogRepoWithSqlMock()
	signature := "signature-hash"
	e := &entities.PartnerSyncDataLog{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &signature)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		log, err := r.GetBySignature(ctx, mockDB.DB, signature)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, log)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &signature)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetBySignature(ctx, mockDB.DB, signature)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestPartnerSyncDataLogRepo_UpdateTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PartnerSyncDataLogRepoWithSqlMock()

	logID := "mock-log-id"
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, logID)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := repo.UpdateTime(ctx, mockDB.DB, logID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("no row affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, logID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := repo.UpdateTime(ctx, mockDB.DB, logID)
		assert.Equal(t, err, fmt.Errorf("no rows affected"))
	})

	t.Run("success", func(t *testing.T) {
		fields := []string{"updated_at"}

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, logID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.UpdateTime(ctx, mockDB.DB, logID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedFields(t, fields...)
	})
}
