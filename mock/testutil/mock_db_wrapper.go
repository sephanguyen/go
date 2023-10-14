package testutil

import (
	"reflect"
	"testing"

	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	DB      *mock_database.Ext
	Rows    *mock_database.Rows
	Row     *mock_database.Row
	RawStmt *RawStmt
}

func NewMockDB() *MockDB {
	return &MockDB{
		DB:   &mock_database.Ext{},
		Rows: &mock_database.Rows{},
		Row:  &mock_database.Row{},
	}
}

func (m *MockDB) MockExecArgs(t *testing.T, cmd pgconn.CommandTag, err error, args ...interface{}) {
	m.DB.On("Exec", args...).Once().Run(func(args mock.Arguments) {
		sql := args.String(1)
		m.RawStmt = ParseSQL(t, sql)
	}).Return(cmd, err)
}

func (m *MockDB) MockQueryArgs(t *testing.T, err error, args ...interface{}) {
	rows := m.Rows
	if err != nil {
		rows = nil
	}

	m.DB.On("Query", args...).Once().Run(func(args mock.Arguments) {
		sql := args.String(1)
		m.RawStmt = ParseSQL(t, sql)
	}).Return(rows, err)
}

func (m *MockDB) MockQueryRowArgs(t *testing.T, args ...interface{}) {
	row := m.Row
	m.DB.On("QueryRow", args...).Once().Run(func(args mock.Arguments) {
		sql := args.String(1)
		m.RawStmt = ParseSQL(t, sql)
	}).Return(row)
}

func (m *MockDB) MockScanFields(err error, fields []string, values []interface{}) {
	m.mockScanFields(err, fields, values, true)
}

func (m *MockDB) MockRowScanFields(err error, fields []string, values []interface{}) {
	mockArgs := []interface{}{}
	for range fields {
		mockArgs = append(mockArgs, mock.Anything)
	}
	m.Row.On("Scan", mockArgs...).Once().Run(func(args mock.Arguments) {
		for i := range args {
			reflect.ValueOf(args[i]).Elem().Set(reflect.ValueOf(values[i]).Elem())
		}
	}).Return(err)
}

func (m *MockDB) MockScanArray(err error, fields []string, dst [][]interface{}) {
	if len(dst) == 0 {
		panic("missing dst")
	}

	closeRow := false
	for i, values := range dst {
		if i == len(dst)-1 {
			closeRow = true
		}
		m.mockScanFields(err, fields, values, closeRow)
	}
}

func (m *MockDB) mockScanFields(err error, fields []string, values []interface{}, withCloseRows bool) {
	mockArgs := []interface{}{}
	fieldDescriptions := make([]pgproto3.FieldDescription, 0, len(fields))
	for _, f := range fields {
		fieldDescriptions = append(fieldDescriptions, pgproto3.FieldDescription{Name: []byte(f)})
		mockArgs = append(mockArgs, mock.Anything)
	}

	m.Rows.On("Next").Once().Return(true)
	m.Rows.On("FieldDescriptions").Return(fieldDescriptions)
	m.Rows.On("Scan", mockArgs...).Once().Run(func(args mock.Arguments) {
		for i := range args {
			reflect.ValueOf(args[i]).Elem().Set(reflect.ValueOf(values[i]).Elem())
		}
	}).Return(err)

	if withCloseRows {
		m.Rows.On("Next").Once().Return(false)
		m.Rows.On("Close").Once().Return(true)
		m.Rows.On("Err").Once().Return(nil)
	}
}
