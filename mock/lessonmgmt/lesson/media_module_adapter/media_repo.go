// Code generated by mockgen. DO NOT EDIT.
package mock_media_module

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type MockMediaRepo struct {
	mock.Mock
}

func (r *MockMediaRepo) CreateMedia(arg1 context.Context, arg2 database.QueryExecer, arg3 *domain.Media) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockMediaRepo) DeleteMedias(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (r *MockMediaRepo) RetrieveMediasByIDs(arg1 context.Context, arg2 database.QueryExecer, arg3 []string) (domain.Medias, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(domain.Medias), args.Error(1)
}
