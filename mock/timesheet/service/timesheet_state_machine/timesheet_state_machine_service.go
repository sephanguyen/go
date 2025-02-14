// Code generated by mockgen. DO NOT EDIT.
package mock_timesheet

import (
	"context"

	"github.com/stretchr/testify/mock"

	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type MockTimesheetStateMachineService struct {
	mock.Mock
}

func (r *MockTimesheetStateMachineService) ApproveTimesheet(arg1 context.Context, arg2 []string) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}

func (r *MockTimesheetStateMachineService) CancelApproveTimesheet(arg1 context.Context, arg2 string) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}

func (r *MockTimesheetStateMachineService) CancelSubmissionTimesheet(arg1 context.Context, arg2 string) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}

func (r *MockTimesheetStateMachineService) ConfirmTimesheet(arg1 context.Context, arg2 []string) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}

func (r *MockTimesheetStateMachineService) DeleteTimesheet(arg1 context.Context, arg2 string) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}

func (r *MockTimesheetStateMachineService) PublishLockLessonEvent(arg1 context.Context, arg2 *tpb.TimesheetLessonLockEvt) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}

func (r *MockTimesheetStateMachineService) SubmitTimesheet(arg1 context.Context, arg2 string) error {
	args := r.Called(arg1, arg2)
	return args.Error(0)
}
