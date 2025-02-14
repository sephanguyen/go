// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type MockLessonRepoImpl struct {
	mock.Mock
}

func (r *MockLessonRepoImpl) FindAllLessonsByIDsIgnoreDeletedAtCondition(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entity.Lesson, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Lesson), args.Error(1)
}

func (r *MockLessonRepoImpl) FindLessonsByIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entity.Lesson, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Lesson), args.Error(1)
}
