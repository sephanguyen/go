package database

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type Tx interface {
	pgx.Tx
}

type TxHandler = func(ctx context.Context, tx Tx) error

func ExecInTx(ctx context.Context, db database.Ext, txHandler TxHandler) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Begin")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = errors.Wrap(tx.Commit(ctx), "tx.Commit")
	}()

	err = txHandler(ctx, tx)

	return err
}
