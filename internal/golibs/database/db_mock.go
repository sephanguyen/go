package database

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type QueryExecerMock struct {
	sendBatch func(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func NewQueryExecerMock(sendBatch func(ctx context.Context, b *pgx.Batch) pgx.BatchResults) *QueryExecerMock {
	return &QueryExecerMock{
		sendBatch: sendBatch,
	}
}

func (q *QueryExecerMock) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return q.sendBatch(ctx, b)
}

type TxStarterMock struct {
	begin func(ctx context.Context) (pgx.Tx, error)
}

func NewTxStarterMock(begin func(ctx context.Context) (pgx.Tx, error)) *TxStarterMock {
	return &TxStarterMock{
		begin: begin,
	}
}

func (t *TxStarterMock) Begin(ctx context.Context) (pgx.Tx, error) {
	return t.begin(ctx)
}

type ExtMock struct {
	QueryExecerMock
	TxStarterMock
	query    func(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	queryRow func(ctx context.Context, query string, args ...interface{}) pgx.Row
	exec     func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

func (e ExtMock) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return e.query(ctx, query, args...)
}

func (e ExtMock) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return e.queryRow(ctx, query, args...)
}

func (e ExtMock) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return e.exec(ctx, sql, args...)
}

type TxMock struct {
	begin        func(ctx context.Context) (pgx.Tx, error)
	beginFunc    func(ctx context.Context, f func(pgx.Tx) error) error
	commit       func(ctx context.Context) error
	rollback     func(ctx context.Context) error
	copyFrom     func(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	sendBatch    func(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	largeObjects func() pgx.LargeObjects
	prepare      func(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	exec         func(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	query        func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	queryRow     func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	queryFunc    func(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
	conn         func() *pgx.Conn
}

func NewTxMock(
	begin func(ctx context.Context) (pgx.Tx, error),
	beginFunc func(ctx context.Context, f func(pgx.Tx) error) error,
	commit func(ctx context.Context) error,
	rollback func(ctx context.Context) error,
	copyFrom func(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error),
	sendBatch func(ctx context.Context, b *pgx.Batch) pgx.BatchResults,
	largeObjects func() pgx.LargeObjects,
	prepare func(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error),
	exec func(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error),
	query func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error),
	queryRow func(ctx context.Context, sql string, args ...interface{}) pgx.Row,
	conn func() *pgx.Conn,
) *TxMock {
	return &TxMock{
		begin:        begin,
		beginFunc:    beginFunc,
		commit:       commit,
		rollback:     rollback,
		copyFrom:     copyFrom,
		sendBatch:    sendBatch,
		largeObjects: largeObjects,
		prepare:      prepare,
		exec:         exec,
		query:        query,
		queryRow:     queryRow,
		conn:         conn,
	}
}

func (t TxMock) Begin(ctx context.Context) (pgx.Tx, error) {
	return t.begin(ctx)
}

func (t TxMock) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	return t.beginFunc(ctx, f)
}

func (t TxMock) Commit(ctx context.Context) error {
	return t.commit(ctx)
}

func (t TxMock) Rollback(ctx context.Context) error {
	return t.rollback(ctx)
}

func (t TxMock) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return t.copyFrom(ctx, tableName, columnNames, rowSrc)
}

func (t TxMock) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return t.sendBatch(ctx, b)
}

func (t TxMock) LargeObjects() pgx.LargeObjects {
	return t.largeObjects()
}

func (t TxMock) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return t.prepare(ctx, name, sql)
}

func (t TxMock) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	return t.exec(ctx, sql, arguments...)
}

func (t TxMock) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return t.query(ctx, sql, args...)
}

func (t TxMock) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return t.queryRow(ctx, sql, args...)
}

func (t TxMock) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	return t.queryFunc(ctx, sql, args, scans, f)
}

func (t TxMock) Conn() *pgx.Conn {
	return t.conn()
}
