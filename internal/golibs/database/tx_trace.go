package database

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type TxTrace struct {
	Tx pgx.Tx
}

var _ pgx.Tx = (*TxTrace)(nil)

func (rcv *TxTrace) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.QueryFunc")
	defer span.End()

	return rcv.Tx.QueryFunc(ctx, sql, args, scans, f)
}

func (rcv *TxTrace) Begin(ctx context.Context) (pgx.Tx, error) {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.Begin")
	defer span.End()

	return rcv.Tx.Begin(ctx)
}

func (rcv *TxTrace) BeginFunc(ctx context.Context, f func(tx pgx.Tx) error) error {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.BeginFunc")
	defer span.End()

	return rcv.Tx.BeginFunc(ctx, f)
}

func (rcv *TxTrace) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.CopyFrom")
	defer span.End()

	return rcv.Tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (rcv *TxTrace) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.SendBatch")
	defer span.End()

	return rcv.Tx.SendBatch(ctx, b)
}

func (rcv *TxTrace) LargeObjects() pgx.LargeObjects {
	return rcv.Tx.LargeObjects()
}

func (rcv *TxTrace) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.Prepare")
	defer span.End()

	return rcv.Tx.Prepare(ctx, name, sql)
}

func (rcv *TxTrace) Conn() *pgx.Conn {
	return rcv.Tx.Conn()
}

func (rcv *TxTrace) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.Query")
	defer span.End()

	return rcv.Tx.Query(ctx, query, args...)
}

func (rcv *TxTrace) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.QueryRow")
	defer span.End()

	return rcv.Tx.QueryRow(ctx, query, args...)
}

func (rcv *TxTrace) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.Exec")
	defer span.End()

	return rcv.Tx.Exec(ctx, sql, args...)
}

func (rcv *TxTrace) Commit(ctx context.Context) error {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.Commit")
	defer span.End()

	return rcv.Tx.Commit(ctx)
}

func (rcv *TxTrace) Rollback(ctx context.Context) error {
	ctx, span := interceptors.StartSpan(ctx, "TxTrace.Rollback")
	defer span.End()

	return rcv.Tx.Rollback(ctx)
}
