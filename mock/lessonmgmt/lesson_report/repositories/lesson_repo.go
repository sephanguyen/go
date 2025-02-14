// Code generated by mockgen. DO NOT EDIT.
package mock_lesson_report

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type MockLessonRepo struct {
	mock.Mock
}

func (r *MockLessonRepo) FindByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*domain.Lesson, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Lesson), args.Error(1)
}

func (r *MockLessonRepo) GetLearnerIDsOfLesson(arg1 context.Context, arg2 database.QueryExecer, arg3 string) ([]string, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
