// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	mock "github.com/stretchr/testify/mock"
)

// IPackageServiceForStudentBilling is an autogenerated mock type for the IPackageServiceForStudentBilling type
type IPackageServiceForStudentBilling struct {
	mock.Mock
}

type IPackageServiceForStudentBilling_Expecter struct {
	mock *mock.Mock
}

func (_m *IPackageServiceForStudentBilling) EXPECT() *IPackageServiceForStudentBilling_Expecter {
	return &IPackageServiceForStudentBilling_Expecter{mock: &_m.Mock}
}

// GetTotalAssociatedPackageWithCourseIDAndPackageID provides a mock function with given fields: ctx, db, packageID, courseIDs
func (_m *IPackageServiceForStudentBilling) GetTotalAssociatedPackageWithCourseIDAndPackageID(ctx context.Context, db database.Ext, packageID string, courseIDs []string) (int32, error) {
	ret := _m.Called(ctx, db, packageID, courseIDs)

	var r0 int32
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, string, []string) (int32, error)); ok {
		return rf(ctx, db, packageID, courseIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, string, []string) int32); ok {
		r0 = rf(ctx, db, packageID, courseIDs)
	} else {
		r0 = ret.Get(0).(int32)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.Ext, string, []string) error); ok {
		r1 = rf(ctx, db, packageID, courseIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTotalAssociatedPackageWithCourseIDAndPackageID'
type IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call struct {
	*mock.Call
}

// GetTotalAssociatedPackageWithCourseIDAndPackageID is a helper method to define mock.On call
//   - ctx context.Context
//   - db database.Ext
//   - packageID string
//   - courseIDs []string
func (_e *IPackageServiceForStudentBilling_Expecter) GetTotalAssociatedPackageWithCourseIDAndPackageID(ctx interface{}, db interface{}, packageID interface{}, courseIDs interface{}) *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call {
	return &IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call{Call: _e.mock.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, db, packageID, courseIDs)}
}

func (_c *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call) Run(run func(ctx context.Context, db database.Ext, packageID string, courseIDs []string)) *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(database.Ext), args[2].(string), args[3].([]string))
	})
	return _c
}

func (_c *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call) Return(total int32, err error) *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call {
	_c.Call.Return(total, err)
	return _c
}

func (_c *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call) RunAndReturn(run func(context.Context, database.Ext, string, []string) (int32, error)) *IPackageServiceForStudentBilling_GetTotalAssociatedPackageWithCourseIDAndPackageID_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewIPackageServiceForStudentBilling interface {
	mock.TestingT
	Cleanup(func())
}

// NewIPackageServiceForStudentBilling creates a new instance of IPackageServiceForStudentBilling. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIPackageServiceForStudentBilling(t mockConstructorTestingTNewIPackageServiceForStudentBilling) *IPackageServiceForStudentBilling {
	mock := &IPackageServiceForStudentBilling{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
