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

type MockStudentSubmissionRepo struct {
	mock.Mock
}

func (r *MockStudentSubmissionRepo) CountSubmissions(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray, arg4 *pgtype.TextArray) (int, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(int), args.Error(1)
}

func (r *MockStudentSubmissionRepo) Create(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.StudentSubmission) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentSubmissionRepo) FindByID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entities.StudentSubmission, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StudentSubmission), args.Error(1)
}

func (r *MockStudentSubmissionRepo) List(arg1 context.Context, arg2 database.QueryExecer, arg3 *repositories.StudentSubmissionFilter) ([]*entities.StudentSubmission, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentSubmission), args.Error(1)
}

func (r *MockStudentSubmissionRepo) ListLatestScore(arg1 context.Context, arg2 database.QueryExecer, arg3 *repositories.StudentSubmissionFilter) ([]*repositories.SubmissionScore, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.SubmissionScore), args.Error(1)
}
