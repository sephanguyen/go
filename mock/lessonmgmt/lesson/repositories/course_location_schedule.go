package mock_repositories

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	"github.com/stretchr/testify/mock"
)

type MockCourseLocationScheduleRepo struct {
	mock.Mock
}

func (r *MockCourseLocationScheduleRepo) UpsertMultiCourseLocationSchedule(arg1 context.Context, arg2 database.QueryExecer, arg3 []*domain.CourseLocationSchedule) *domain.ImportCourseLocationScheduleError {
	args := r.Called(arg1, arg2, arg3)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ImportCourseLocationScheduleError)
}

func (r *MockCourseLocationScheduleRepo) ExportCourseLocationSchedule(arg1 context.Context, arg2 database.QueryExecer) ([]*domain.CourseLocationSchedule, error) {
	args := r.Called(arg1, arg2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CourseLocationSchedule), args.Error(1)

}

func (r *MockCourseLocationScheduleRepo) GetAcademicWeekValid(arg1 context.Context, arg2 database.QueryExecer, arg3 []string, arg4 time.Time) (map[string]bool, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]bool), args.Error(1)
}

func (r *MockCourseLocationScheduleRepo) GetByCourseIDAndLocationID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) (*domain.CourseLocationSchedule, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CourseLocationSchedule), args.Error(1)
}
