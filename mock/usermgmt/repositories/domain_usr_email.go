// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
)

type MockDomainUsrEmailRepo struct {
	mock.Mock
}

func (r *MockDomainUsrEmailRepo) CreateMultiple(arg1 context.Context, arg2 database.QueryExecer, arg3 entity.Users) (valueobj.HasUserIDs, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(valueobj.HasUserIDs), args.Error(1)
}

func (r *MockDomainUsrEmailRepo) UpdateEmail(arg1 context.Context, arg2 database.QueryExecer, arg3 entity.User) error {
	args := r.Called(arg1, arg2, arg3)
	return args.Error(0)
}
