// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockQuizSetRepo struct {
	mock.Mock
}

func (r *MockQuizSetRepo) CountQuizOnLO(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) (map[string]int32, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(map[string]int32), args.Error(1)
}

func (r *MockQuizSetRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.QuizSet) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockQuizSetRepo) Delete(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockQuizSetRepo) GetQuizExternalIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Int8, arg5 pgtype.Int8) ([]string, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (r *MockQuizSetRepo) GetQuizSetByLoID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entities.QuizSet, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.QuizSet), args.Error(1)
}

func (r *MockQuizSetRepo) GetQuizSetsContainQuiz(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (entities.QuizSets, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.QuizSets), args.Error(1)
}

func (r *MockQuizSetRepo) GetQuizSetsOfLOContainQuiz(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Text) (entities.QuizSets, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(entities.QuizSets), args.Error(1)
}

func (r *MockQuizSetRepo) GetTotalQuiz(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) (map[string]int32, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(map[string]int32), args.Error(1)
}

func (r *MockQuizSetRepo) Search(arg1 context.Context, arg2 database.QueryExecer, arg3 repositories.QuizSetFilter) (entities.QuizSets, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entities.QuizSets), args.Error(1)
}
