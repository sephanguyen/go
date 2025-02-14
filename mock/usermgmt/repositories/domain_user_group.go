// Code generated by mockgen. DO NOT EDIT.
package mock_repositories

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type MockDomainUserGroupRepo struct {
	mock.Mock
}

func (r *MockDomainUserGroupRepo) FindUserGroupByRoleName(arg1 context.Context, arg2 database.QueryExecer, arg3 string) (entity.DomainUserGroup, error) {
	args := r.Called(arg1, arg2, arg3)
	return args.Get(0).(entity.DomainUserGroup), args.Error(1)
}
