package mock_repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"

	"github.com/stretchr/testify/mock"
)

type MockZoomAccountRepo struct {
	mock.Mock
}

func (m *MockZoomAccountRepo) GetZoomAccountByID(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (*domain.ZoomAccount, error) {
	args := m.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ZoomAccount), args.Error(1)
}

func (m *MockZoomAccountRepo) Upsert(arg1 context.Context, arg2 database.Ext, arg3 domain.ZoomAccounts) error {
	args := m.Called(arg1, arg2, arg3)
	return args.Error(0)
}

func (m *MockZoomAccountRepo) GetAllZoomAccount(arg1 context.Context, arg2 database.QueryExecer) ([]*repo.ZoomAccount, error) {
	args := m.Called(arg1, arg2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repo.ZoomAccount), args.Error(1)
}
