package database

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestExecInTxWithRetry(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("Begin Error", func(t *testing.T) {
		t.Parallel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, pgx.ErrTxClosed)

		err := ExecInTxWithRetry(ctx, db, nil)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		db.AssertNumberOfCalls(t, "Begin", 1)
	})

	t.Run("Err retry begin with retryable error", func(t *testing.T) {
		t.Parallel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, nil)
		tx.On("Rollback", ctx).Once().Return(nil)
		db.On("Begin", ctx).Once().Return(tx, pgx.ErrTxClosed)

		err := ExecInTxWithRetry(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			return &pgconn.PgError{Code: "40001"}
		})
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		db.AssertNumberOfCalls(t, "Begin", 2)
		tx.AssertNumberOfCalls(t, "Rollback", 1)
	})

	t.Run("Err rollback", func(t *testing.T) {
		t.Parallel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, nil)
		tx.On("Rollback", ctx).Once().Return(pgx.ErrTxClosed)

		err := ExecInTxWithRetry(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			return fmt.Errorf("err exec")
		})
		assert.Equal(t, fmt.Errorf("err exec"), err)

		db.AssertNumberOfCalls(t, "Begin", 1)
		tx.AssertNumberOfCalls(t, "Rollback", 1)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, nil)
		tx.On("Commit", ctx).Once().Return(nil)

		err := ExecInTxWithRetry(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			return nil
		})
		assert.Equal(t, nil, err)
		db.AssertNumberOfCalls(t, "Begin", 1)
		tx.AssertNumberOfCalls(t, "Commit", 1)
	})
}

func TestExecInTxWithContextDeadline(t *testing.T) {
	t.Parallel()
	t.Run("Begin Error", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, pgx.ErrTxClosed)

		err := ExecInTxWithContextDeadline(ctx, db, nil)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		db.AssertNumberOfCalls(t, "Begin", 1)
	})

	t.Run("Err rollback", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, nil)
		tx.On("Rollback", ctx).Once().Return(pgx.ErrTxClosed)

		err := ExecInTxWithContextDeadline(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			return fmt.Errorf("err exec")
		})
		assert.Equal(t, fmt.Errorf("err exec"), err)

		db.AssertNumberOfCalls(t, "Begin", 1)
		tx.AssertNumberOfCalls(t, "Rollback", 1)
	})

	t.Run("Err when context is deadline", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, nil)
		tx.On("Rollback", ctx).Once().Return(nil)

		err := ExecInTxWithContextDeadline(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			time.Sleep(10 * time.Second)
			return nil
		})
		assert.Equal(t, fmt.Errorf("the transaction is rolled back because context deadline was not met"), err)

		db.AssertNumberOfCalls(t, "Begin", 1)
		tx.AssertNumberOfCalls(t, "Rollback", 1)
	})

	t.Run("Happy case", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}

		db.On("Begin", ctx).Once().Return(tx, nil)
		tx.On("Commit", ctx).Once().Return(nil)

		err := ExecInTxWithContextDeadline(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			return nil
		})
		assert.Equal(t, nil, err)

		db.AssertNumberOfCalls(t, "Begin", 1)
		tx.AssertNumberOfCalls(t, "Rollback", 0)
		tx.AssertNumberOfCalls(t, "Commit", 1)
	})
}
