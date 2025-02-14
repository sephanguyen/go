// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_database

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	pgconn "github.com/jackc/pgconn"

	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Ext is an autogenerated mock type for the Ext type
type Ext struct {
	mock.Mock
}

// Begin provides a mock function with given fields: ctx
func (_m *Ext) Begin(ctx context.Context) (pgx.Tx, error) {
	ret := _m.Called(ctx)

	var r0 pgx.Tx
	if rf, ok := ret.Get(0).(func(context.Context) pgx.Tx); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.Tx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Exec provides a mock function with given fields: ctx, sql, args
func (_m *Ext) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, sql)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 pgconn.CommandTag
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) pgconn.CommandTag); ok {
		r0 = rf(ctx, sql, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgconn.CommandTag)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(ctx, sql, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Query provides a mock function with given fields: ctx, query, args
func (_m *Ext) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 pgx.Rows
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) pgx.Rows); ok {
		r0 = rf(ctx, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.Rows)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(ctx, query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryRow provides a mock function with given fields: ctx, query, args
func (_m *Ext) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 pgx.Row
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) pgx.Row); ok {
		r0 = rf(ctx, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.Row)
		}
	}

	return r0
}

// SendBatch provides a mock function with given fields: ctx, b
func (_m *Ext) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	ret := _m.Called(ctx, b)

	var r0 pgx.BatchResults
	if rf, ok := ret.Get(0).(func(context.Context, *pgx.Batch) pgx.BatchResults); ok {
		r0 = rf(ctx, b)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.BatchResults)
		}
	}

	return r0
}


func (_m *Ext) Stat() *pgxpool.Stat {
	r0 := &pgxpool.Stat{}

	return r0
}
type mockConstructorTestingTNewExt interface {
	mock.TestingT
	Cleanup(func())
}

// NewExt creates a new instance of Ext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewExt(t mockConstructorTestingTNewExt) *Ext {
	mock := &Ext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
