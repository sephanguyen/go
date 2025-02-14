// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	esapi "github.com/elastic/go-elasticsearch/v7/esapi"

	mock "github.com/stretchr/testify/mock"

	op "github.com/manabie-com/backend/internal/payment/search/op"

	search "github.com/manabie-com/backend/internal/payment/search"
)

// Engine is an autogenerated mock type for the Engine type
type Engine struct {
	mock.Mock
}

// CheckIndexExists provides a mock function with given fields: index
func (_m *Engine) CheckIndexExists(index string) (bool, error) {
	ret := _m.Called(index)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (bool, error)); ok {
		return rf(index)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(index)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(index)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CountValue provides a mock function with given fields: ctx, tableName, columnName, condition
func (_m *Engine) CountValue(ctx context.Context, tableName string, columnName string, condition op.Condition) (uint32, error) {
	ret := _m.Called(ctx, tableName, columnName, condition)

	var r0 uint32
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, op.Condition) (uint32, error)); ok {
		return rf(ctx, tableName, columnName, condition)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, op.Condition) uint32); ok {
		r0 = rf(ctx, tableName, columnName, condition)
	} else {
		r0 = ret.Get(0).(uint32)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, op.Condition) error); ok {
		r1 = rf(ctx, tableName, columnName, condition)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateIndex provides a mock function with given fields: index, body
func (_m *Engine) CreateIndex(index string, body io.Reader) (*esapi.Response, error) {
	ret := _m.Called(index, body)

	var r0 *esapi.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader) (*esapi.Response, error)); ok {
		return rf(index, body)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader) *esapi.Response); ok {
		r0 = rf(index, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*esapi.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader) error); ok {
		r1 = rf(index, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteIndex provides a mock function with given fields: index
func (_m *Engine) DeleteIndex(index string) (*esapi.Response, error) {
	ret := _m.Called(index)

	var r0 *esapi.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*esapi.Response, error)); ok {
		return rf(index)
	}
	if rf, ok := ret.Get(0).(func(string) *esapi.Response); ok {
		r0 = rf(index)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*esapi.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(index)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: ctx, tableName, funcRecv, pagingParam, sortParams
func (_m *Engine) GetAll(ctx context.Context, tableName string, funcRecv func([]byte) (interface{}, error), pagingParam search.PagingParam, sortParams ...search.SortParam) ([]interface{}, error) {
	_va := make([]interface{}, len(sortParams))
	for _i := range sortParams {
		_va[_i] = sortParams[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, tableName, funcRecv, pagingParam)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, func([]byte) (interface{}, error), search.PagingParam, ...search.SortParam) ([]interface{}, error)); ok {
		return rf(ctx, tableName, funcRecv, pagingParam, sortParams...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, func([]byte) (interface{}, error), search.PagingParam, ...search.SortParam) []interface{}); ok {
		r0 = rf(ctx, tableName, funcRecv, pagingParam, sortParams...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, func([]byte) (interface{}, error), search.PagingParam, ...search.SortParam) error); ok {
		r1 = rf(ctx, tableName, funcRecv, pagingParam, sortParams...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Insert provides a mock function with given fields: ctx, tableName, contents
func (_m *Engine) Insert(ctx context.Context, tableName string, contents []search.InsertionContent) (int, error) {
	ret := _m.Called(ctx, tableName, contents)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []search.InsertionContent) (int, error)); ok {
		return rf(ctx, tableName, contents)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []search.InsertionContent) int); ok {
		r0 = rf(ctx, tableName, contents)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []search.InsertionContent) error); ok {
		r1 = rf(ctx, tableName, contents)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Search provides a mock function with given fields: ctx, tableName, condition, funcRecv, pagingParam, sortParams
func (_m *Engine) Search(ctx context.Context, tableName string, condition op.Condition, funcRecv func([]byte) (interface{}, error), pagingParam search.PagingParam, sortParams ...search.SortParam) ([]interface{}, error) {
	_va := make([]interface{}, len(sortParams))
	for _i := range sortParams {
		_va[_i] = sortParams[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, tableName, condition, funcRecv, pagingParam)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, op.Condition, func([]byte) (interface{}, error), search.PagingParam, ...search.SortParam) ([]interface{}, error)); ok {
		return rf(ctx, tableName, condition, funcRecv, pagingParam, sortParams...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, op.Condition, func([]byte) (interface{}, error), search.PagingParam, ...search.SortParam) []interface{}); ok {
		r0 = rf(ctx, tableName, condition, funcRecv, pagingParam, sortParams...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, op.Condition, func([]byte) (interface{}, error), search.PagingParam, ...search.SortParam) error); ok {
		r1 = rf(ctx, tableName, condition, funcRecv, pagingParam, sortParams...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchWithoutPaging provides a mock function with given fields: ctx, tableName, condition, funcRecv, sortParams
func (_m *Engine) SearchWithoutPaging(ctx context.Context, tableName string, condition op.Condition, funcRecv func([]byte) (interface{}, error), sortParams ...search.SortParam) ([]interface{}, error) {
	_va := make([]interface{}, len(sortParams))
	for _i := range sortParams {
		_va[_i] = sortParams[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, tableName, condition, funcRecv)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, op.Condition, func([]byte) (interface{}, error), ...search.SortParam) ([]interface{}, error)); ok {
		return rf(ctx, tableName, condition, funcRecv, sortParams...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, op.Condition, func([]byte) (interface{}, error), ...search.SortParam) []interface{}); ok {
		r0 = rf(ctx, tableName, condition, funcRecv, sortParams...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, op.Condition, func([]byte) (interface{}, error), ...search.SortParam) error); ok {
		r1 = rf(ctx, tableName, condition, funcRecv, sortParams...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewEngine interface {
	mock.TestingT
	Cleanup(func())
}

// NewEngine creates a new instance of Engine. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEngine(t mockConstructorTestingTNewEngine) *Engine {
	mock := &Engine{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
