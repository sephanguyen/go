package services

import (
	"context"

	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/mock"
)

type MockCourseLocationScheduleService struct {
	mock.Mock
}

func (r *MockCourseLocationScheduleService) ImportCourseLocationSchedule(arg1 context.Context, arg2 *lpb.ImportCourseLocationScheduleRequest) (*lpb.ImportCourseLocationScheduleResponse, error) {
	args := r.Called(arg1, arg2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.ImportCourseLocationScheduleResponse), args.Error(1)
}

func (r *MockCourseLocationScheduleService) ExportCourseLocationSchedule(arg1 context.Context) (*lpb.ExportCourseLocationScheduleResponse, error) {
	args := r.Called(arg1)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.ExportCourseLocationScheduleResponse), args.Error(1)
}
