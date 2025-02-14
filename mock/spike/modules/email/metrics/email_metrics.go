// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_metrics

import (
	metrics "github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	mock "github.com/stretchr/testify/mock"

	prometheus "github.com/prometheus/client_golang/prometheus"
)

// EmailMetrics is an autogenerated mock type for the EmailMetrics type
type EmailMetrics struct {
	mock.Mock
}

// GetCollectors provides a mock function with given fields:
func (_m *EmailMetrics) GetCollectors() []prometheus.Collector {
	ret := _m.Called()

	var r0 []prometheus.Collector
	if rf, ok := ret.Get(0).(func() []prometheus.Collector); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]prometheus.Collector)
		}
	}

	return r0
}

// InitCounterValue provides a mock function with given fields:
func (_m *EmailMetrics) InitCounterValue() {
	_m.Called()
}

// RecordEmailEvents provides a mock function with given fields: event, num
func (_m *EmailMetrics) RecordEmailEvents(event metrics.EmailEventMetricType, num float64) {
	_m.Called(event, num)
}

type mockConstructorTestingTNewEmailMetrics interface {
	mock.TestingT
	Cleanup(func())
}

// NewEmailMetrics creates a new instance of EmailMetrics. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEmailMetrics(t mockConstructorTestingTNewEmailMetrics) *EmailMetrics {
	mock := &EmailMetrics{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
