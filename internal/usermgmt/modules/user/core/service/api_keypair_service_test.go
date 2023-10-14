package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	domainAPIKeypairRepo := new(mock_repositories.MockDomainAPIKeypairRepo)
	userGroupRepo := new(mock_repositories.MockDomainUserGroupRepo)
	userGroupMemberRepo := new(mock_repositories.MockDomainUserGroupMemberRepo)

	service := APIKeyPairService{
		DB:                   db,
		DomainAPIKeypairRepo: domainAPIKeypairRepo,
		UserGroupRepo:        userGroupRepo,
		UserGroupMemberRepo:  userGroupMemberRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req:  &valueobj.RandomHasUserID{},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				domainAPIKeypairRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleOpenAPI).Once().Return(entity.UserGroupWillBeDelegated{}, nil)
				userGroupMemberRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", ctx).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := auth.InjectFakeJwtToken(testCase.ctx, fmt.Sprint(constants.ManabieSchool))
			if testCase.setup != nil {
				testCase.setup(ctx)
			}
			err := service.GenerateKey(ctx, testCase.req.(*valueobj.RandomHasUserID))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
