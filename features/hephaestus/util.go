package hephaestus

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

type ITable interface {
	TableName() string
	CreateTableInSourceAndSink(ctx context.Context, sourceDB, sinkDB database.QueryExecer) error
	GenerateSampleRecordsInSource(ctx context.Context, sourceDB database.QueryExecer) ([]string, error)
	GetSampleRecordsInSink(ctx context.Context, sink database.QueryExecer, ids []string) ([]any, error)
}

type TableImpl struct {
	Name string
}

func (t *TableImpl) TableName() string {
	return t.Name
}

func nextTable(tableName string) ITable {
	return newTestDebeziumRecord(tableName)
}

type testDebeziumBase struct {
	ID pgtype.Text
	A  pgtype.Text
	B  pgtype.Text
	C  pgtype.Int4
	D  pgtype.Timestamptz
}

type testDebeziumRecord struct {
	testDebeziumBase
	TableImpl
}

func newTestDebeziumRecord(tableName string) *testDebeziumRecord {
	return &testDebeziumRecord{
		TableImpl: TableImpl{
			Name: tableName,
		},
	}
}

func (e *testDebeziumRecord) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "a", "b", "c", "d"}
	values = []interface{}{&e.ID, &e.A, &e.B, &e.C, &e.D}
	return
}

func (e *testDebeziumRecord) TableName() string {
	return e.Name
}

func (e *testDebeziumRecord) CreateTableInSourceAndSink(ctx context.Context, sourceDB, sinkDB database.QueryExecer) error {
	tableName := e.TableName()
	createTableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY, a text, b text, c integer, d timestamptz)`, tableName)
	_, err := sourceDB.Exec(ctx, createTableSQL)
	if err != nil {
		return err
	}

	getTablePublicationSQL := `SELECT COUNT(*) FROM pg_publication_tables WHERE pubname=$1 AND tablename=$2`
	var cnt pgtype.Int8
	err = sourceDB.QueryRow(ctx, getTablePublicationSQL, database.Text("debezium_publication"), database.Text(tableName)).Scan(&cnt)
	if err != nil {
		return err
	}
	if cnt.Int == 0 {
		addTableToPublicationSQL := fmt.Sprintf(`ALTER PUBLICATION debezium_publication ADD TABLE public.%s`, tableName)
		_, err = sourceDB.Exec(ctx, addTableToPublicationSQL)
		if err != nil {
			return err
		}
	}

	_, err = sinkDB.Exec(ctx, createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func (e *testDebeziumRecord) GenerateSampleRecordsInSource(ctx context.Context, sourceDB database.QueryExecer) ([]string, error) {
	sql := fmt.Sprintf(`INSERT INTO %s(id, a, b, c, d) VALUES($1, $2, $3, $4, $5)`, e.TableName())
	ids := make([]string, 0)

	for i := 0; i < 10; i++ {
		e := testDebeziumRecord{}
		now := time.Now()
		e.ID = database.Text(idutil.ULIDNow())
		e.A = database.Text(fmt.Sprintf("A - %s", now))
		e.B = database.Text(fmt.Sprintf("B - %s", now))
		e.C = database.Int4(100)
		e.D = database.Timestamptz(now)
		_, err := sourceDB.Exec(ctx, sql, e.ID, e.A, e.B, e.C, e.D)
		if err != nil {
			return nil, err
		}
		ids = append(ids, e.ID.String)
	}
	return ids, nil
}

func (e *testDebeziumRecord) GetSampleRecordsInSink(ctx context.Context, sinkDB database.QueryExecer, ids []string) ([]any, error) {
	es := testDebeziumRecords{}
	query := `SELECT %s FROM %s WHERE id = ANY($1)`
	fields, _ := e.FieldMap()
	err := database.Select(ctx, sinkDB, fmt.Sprintf(query, strings.Join(fields, ", "), e.TableName()), database.TextArray(ids)).ScanAll(&es)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	res := make([]any, 0, len(ids))
	for _, e := range es {
		res = append(res, e)
	}
	return res, nil
}

type testDebeziumRecords []*testDebeziumRecord

func (u *testDebeziumRecords) Add() database.Entity {
	e := &testDebeziumRecord{}
	*u = append(*u, e)

	return e
}

func (s *suite) cleanUpKafkaConnectorDir() {
	_ = os.RemoveAll(s.SourceConnectorDir)
	_ = os.RemoveAll(s.SinkConnectorDir)
}
