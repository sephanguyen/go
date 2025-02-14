// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type MockStudentPackageAccessPathRepo struct {
	mock.Mock
}

func (r *MockStudentPackageAccessPathRepo) CheckExistStudentPackageAccessPath(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockStudentPackageAccessPathRepo) DeleteMulti(arg1 context.Context, arg2 database.QueryExecer, arg3 []entities.StudentPackageAccessPath) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageAccessPathRepo) GetByStudentIDAndCourseID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) (entities.StudentPackageAccessPath, error) {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Get(0).(entities.StudentPackageAccessPath), args.Error(1)
}

func (r *MockStudentPackageAccessPathRepo) GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) (map[string]entities.StudentPackageAccessPath, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(map[string]entities.StudentPackageAccessPath), args.Error(1)
}

func (r *MockStudentPackageAccessPathRepo) Insert(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.StudentPackageAccessPath) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageAccessPathRepo) InsertMulti(arg1 context.Context, arg2 database.QueryExecer, arg3 []entities.StudentPackageAccessPath) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageAccessPathRepo) RevertByStudentIDAndCourseID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockStudentPackageAccessPathRepo) SoftDeleteByStudentPackageIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string, arg4 time.Time) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockStudentPackageAccessPathRepo) Update(arg1 context.Context, arg2 database.QueryExecer, arg3 *entities.StudentPackageAccessPath) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
