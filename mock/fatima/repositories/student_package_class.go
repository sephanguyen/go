// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockStudentPackageClassRepo struct {
	mock.Mock
}

func (r *MockStudentPackageClassRepo) BulkUpsert(arg1 context.Context, arg2 database.QueryExecer, arg3 []*entities.StudentPackageClass) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockStudentPackageClassRepo) DeleteByStudentPackageIDAndCourseID(arg1 context.Context, arg2 database.QueryExecer, arg3 string, arg4 string) error {
	args := r.Called(arg1, arg2, arg3, arg4)
	return args.Error(0)
}

func (r *MockStudentPackageClassRepo) DeleteByStudentPackageIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 pgtype.TextArray) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
