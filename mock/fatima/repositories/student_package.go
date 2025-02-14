// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockStudentPackageRepo struct {
	mock.Mock
}

func (r *MockStudentPackageRepo) BulkInsert(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entities.StudentPackage) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageRepo) CurrentPackage(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) ([]*entities.StudentPackage, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentPackage), args.Error(1)
}

func (r *MockStudentPackageRepo) Get(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) (*entities.StudentPackage, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StudentPackage), args.Error(1)
}

func (r *MockStudentPackageRepo) GetByCourseIDAndLocationIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.TextArray) ([]*entities.StudentPackage, error) {
	args := r.Called(arg1, arg2, arg3, arg4)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentPackage), args.Error(1)
}

func (r *MockStudentPackageRepo) GetByStudentIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) ([]*entities.StudentPackage, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StudentPackage), args.Error(1)
}

func (r *MockStudentPackageRepo) GetByStudentPackageIDAndStudentIDAndCourseID(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text, arg4 pgtype.Text, arg5 pgtype.Text) (*entities.StudentPackage, error) {
	args := r.Called(arg1, arg2, arg3, arg4, arg5)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StudentPackage), args.Error(1)
}

func (r *MockStudentPackageRepo) Insert(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.StudentPackage) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageRepo) SoftDelete(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.Text) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageRepo) SoftDeleteByIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageRepo) Update(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.StudentPackage) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
