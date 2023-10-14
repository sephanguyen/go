package database

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSelect(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sql := "SELECT name FROM users WHERE id = $1"
	db := &mock_database.QueryExecer{}
	mockRows := &mock_database.Rows{}

	t.Run("err when query", func(t *testing.T) {
		db.On("Query", ctx, sql, "user_id").Once().Return(nil, pgx.ErrTxClosed)

		rows := Select(ctx, db, sql, "user_id")
		assert.Equal(t, pgx.ErrTxClosed, errors.Unwrap(rows.err))
		assert.Nil(t, rows.pgxRows)
	})

	t.Run("success", func(t *testing.T) {
		db.On("Query", ctx, sql, "user_id").Once().Return(mockRows, nil)

		rows := Select(ctx, db, sql, "user_id")
		assert.Nil(t, rows.err)
		assert.Equal(t, mockRows, rows.pgxRows)
	})
}

type DummyUserEntity struct {
	ID   string
	Name string
}

func (r *DummyUserEntity) FieldMap() ([]string, []interface{}) {
	return []string{"id", "name"}, []interface{}{&r.ID, &r.Name}
}

func (r *DummyUserEntity) TableName() string {
	return "users"
}

type DummyUserEntities []*DummyUserEntity

func (r *DummyUserEntities) Add() Entity {
	e := &DummyUserEntity{}
	*r = append(*r, e)

	return e
}

func TestRowsScanFields(t *testing.T) {
	t.Parallel()
	t.Run("dont handle if has error", func(t *testing.T) {
		t.Parallel()
		rows := &RowScanner{
			err: pgx.ErrTxClosed,
		}

		e := &DummyUserEntity{}
		err := rows.ScanFields(&e.ID, &e.Name)
		assert.Equal(t, pgx.ErrTxClosed, err)
	})

	t.Run("err no rows", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(false)
		mockRows.On("Err").Once().Return(nil)

		e := &DummyUserEntity{}
		err := rows.ScanFields(&e.ID, &e.Name)
		assert.Equal(t, pgx.ErrNoRows.Error(), err.Error())

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("rows err when can not get Next", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(false)
		mockRows.On("Err").Once().Return(puddle.ErrNotAvailable)

		e := &DummyUserEntity{}
		err := rows.ScanFields(&e.ID, &e.Name)
		assert.Equal(t, puddle.ErrNotAvailable, errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("err scan", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := &DummyUserEntity{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("Scan", &e.ID, &e.Name).Once().Return(errors.New("failed to scan"))

		err := rows.ScanFields(&e.ID, &e.Name)
		assert.Equal(t, errors.New("failed to scan"), errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Scan", 1)
	})

	t.Run("rows err", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := &DummyUserEntity{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("Scan", &e.ID, &e.Name).Once().Return(nil)
		mockRows.On("Err").Once().Return(puddle.ErrNotAvailable)

		err := rows.ScanFields(&e.ID, &e.Name)
		assert.Equal(t, puddle.ErrNotAvailable, errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Scan", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := &DummyUserEntity{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("Scan", &e.ID, &e.Name).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().SetString("userID")
			reflect.ValueOf(args[1]).Elem().SetString("userName")
		}).Return(nil)
		mockRows.On("Err").Once().Return(nil)

		err := rows.ScanFields(&e.ID, &e.Name)
		assert.Nil(t, err)

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Scan", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)

		assert.Equal(t, "userID", e.ID)
		assert.Equal(t, "userName", e.Name)
	})
}

func TestRowsScanOne(t *testing.T) {
	t.Parallel()
	t.Run("dont handle if has error", func(t *testing.T) {
		t.Parallel()
		rows := &RowScanner{
			err: pgx.ErrTxClosed,
		}

		e := &DummyUserEntity{}
		err := rows.ScanOne(e)
		assert.Equal(t, pgx.ErrTxClosed, err)
	})

	t.Run("err no rows", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(false)
		mockRows.On("Err").Once().Return(nil)

		e := &DummyUserEntity{}
		err := rows.ScanOne(e)
		assert.Equal(t, pgx.ErrNoRows.Error(), err.Error())

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("rows err when can not get Next", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(false)
		mockRows.On("Err").Once().Return(puddle.ErrNotAvailable)

		e := &DummyUserEntity{}
		err := rows.ScanOne(e)
		assert.Equal(t, puddle.ErrNotAvailable, errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("err scan", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := &DummyUserEntity{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("FieldDescriptions").Return([]pgproto3.FieldDescription{
			{Name: []byte("id")},
			{Name: []byte("name")},
		})
		mockRows.On("Scan", &e.ID, &e.Name).Once().Return(errors.New("failed to scan"))

		err := rows.ScanOne(e)
		assert.Equal(t, errors.New("failed to scan"), errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Scan", 1)
		mockRows.AssertNumberOfCalls(t, "FieldDescriptions", 2)
	})

	t.Run("rows err", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := &DummyUserEntity{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("FieldDescriptions").Return([]pgproto3.FieldDescription{
			{Name: []byte("id")},
			{Name: []byte("name")},
		})
		mockRows.On("Scan", &e.ID, &e.Name).Once().Return(nil)
		mockRows.On("Err").Once().Return(puddle.ErrNotAvailable)

		err := rows.ScanOne(e)
		assert.Equal(t, puddle.ErrNotAvailable, errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Scan", 1)
		mockRows.AssertNumberOfCalls(t, "FieldDescriptions", 2)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := &DummyUserEntity{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("FieldDescriptions").Return([]pgproto3.FieldDescription{
			{Name: []byte("id")},
			{Name: []byte("name")},
		})
		mockRows.On("Scan", &e.ID, &e.Name).Once().Run(func(args mock.Arguments) {
			reflect.ValueOf(args[0]).Elem().SetString("userID")
			reflect.ValueOf(args[1]).Elem().SetString("userName")
		}).Return(nil)
		mockRows.On("Err").Once().Return(nil)

		err := rows.ScanOne(e)
		assert.Nil(t, err)

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Scan", 1)
		mockRows.AssertNumberOfCalls(t, "FieldDescriptions", 2)
		mockRows.AssertNumberOfCalls(t, "Err", 1)

		assert.Equal(t, "userID", e.ID)
		assert.Equal(t, "userName", e.Name)
	})
}

func TestRowsScanAll(t *testing.T) {
	t.Parallel()
	t.Run("dont handle if has error", func(t *testing.T) {
		t.Parallel()
		rows := &RowScanner{
			err: pgx.ErrTxClosed,
		}

		e := DummyUserEntities{}
		err := rows.ScanAll(&e)
		assert.Equal(t, pgx.ErrTxClosed, err)
	})

	t.Run("err no rows", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(false)
		mockRows.On("Err").Once().Return(pgx.ErrNoRows)

		e := DummyUserEntities{}
		err := rows.ScanAll(&e)
		assert.Equal(t, pgx.ErrNoRows, errors.Unwrap(err))

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 1)
		mockRows.AssertNumberOfCalls(t, "Err", 1)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockRows := &mock_database.Rows{}
		rows := &RowScanner{
			pgxRows: mockRows,
		}
		e := DummyUserEntities{}

		mockRows.On("Close").Once().Return()
		mockRows.On("Next").Once().Return(true)
		mockRows.On("Next").Once().Return(true)
		mockRows.On("Next").Once().Return(false)
		mockRows.On("FieldDescriptions").Return([]pgproto3.FieldDescription{
			{Name: []byte("id")},
			{Name: []byte("name")},
		})

		count := 0
		mockRows.On("Scan", mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Twice().Run(func(args mock.Arguments) {
			count++
			reflect.ValueOf(args[0]).Elem().SetString(fmt.Sprintf("userID-%d", count))
			reflect.ValueOf(args[1]).Elem().SetString(fmt.Sprintf("userName-%d", count))
		}).Return(nil)

		mockRows.On("Err").Once().Return(nil)

		err := rows.ScanAll(&e)
		assert.Nil(t, err)

		mockRows.AssertNumberOfCalls(t, "Close", 1)
		mockRows.AssertNumberOfCalls(t, "Next", 3)
		mockRows.AssertNumberOfCalls(t, "Scan", 2)
		mockRows.AssertNumberOfCalls(t, "FieldDescriptions", 2)
		mockRows.AssertNumberOfCalls(t, "Err", 1)

		assert.Greater(t, len(e), 1)

		for i, v := range e {
			assert.Equal(t, fmt.Sprintf("userID-%d", i+1), v.ID)
			assert.Equal(t, fmt.Sprintf("userName-%d", i+1), v.Name)
		}
	})
}
