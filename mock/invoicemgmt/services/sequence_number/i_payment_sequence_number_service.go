// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_seqnumberservice

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/invoicemgmt/entities"

	mock "github.com/stretchr/testify/mock"
)

// IPaymentSequenceNumberService is an autogenerated mock type for the IPaymentSequenceNumberService type
type IPaymentSequenceNumberService struct {
	mock.Mock
}

// AssignSeqNumberToPayment provides a mock function with given fields: payment
func (_m *IPaymentSequenceNumberService) AssignSeqNumberToPayment(payment *entities.Payment) error {
	ret := _m.Called(payment)

	var r0 error
	if rf, ok := ret.Get(0).(func(*entities.Payment) error); ok {
		r0 = rf(payment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssignSeqNumberToPayments provides a mock function with given fields: payments
func (_m *IPaymentSequenceNumberService) AssignSeqNumberToPayments(payments []*entities.Payment) error {
	ret := _m.Called(payments)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*entities.Payment) error); ok {
		r0 = rf(payments)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InitLatestSeqNumber provides a mock function with given fields: ctx, db
func (_m *IPaymentSequenceNumberService) InitLatestSeqNumber(ctx context.Context, db database.QueryExecer) error {
	ret := _m.Called(ctx, db)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer) error); ok {
		r0 = rf(ctx, db)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InitLatestSeqNumberWithLock provides a mock function with given fields: ctx, db
func (_m *IPaymentSequenceNumberService) InitLatestSeqNumberWithLock(ctx context.Context, db database.QueryExecer) (func(), error) {
	ret := _m.Called(ctx, db)

	var r0 func()
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer) func()); ok {
		r0 = rf(ctx, db)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func())
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer) error); ok {
		r1 = rf(ctx, db)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIPaymentSequenceNumberService interface {
	mock.TestingT
	Cleanup(func())
}

// NewIPaymentSequenceNumberService creates a new instance of IPaymentSequenceNumberService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIPaymentSequenceNumberService(t mockConstructorTestingTNewIPaymentSequenceNumberService) *IPaymentSequenceNumberService {
	mock := &IPaymentSequenceNumberService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
