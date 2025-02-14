// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type MockLessonTeacherRepo struct {
	mock.Mock
}

func (r *MockLessonTeacherRepo) GetTeachersByLessonIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) (map[string]domain.LessonTeachers, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(map[string]domain.LessonTeachers), args.Error(1)
}

func (r *MockLessonTeacherRepo) GetTeachersWithNamesByLessonIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string, arg4 bool) (map[string]domain.LessonTeachers, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(map[string]domain.LessonTeachers), args.Error(1)
}
