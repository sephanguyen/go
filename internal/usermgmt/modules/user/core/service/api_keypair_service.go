package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgx/v4"
)

type APIKeyPairService struct {
	DB database.Ext

	DomainAPIKeypairRepo interface {
		Create(ctx context.Context, db database.QueryExecer, apiKeyToCreate aggregate.DomainAPIKeypair) error
	}

	UserGroupRepo interface {
		FindUserGroupByRoleName(ctx context.Context, db libdatabase.QueryExecer, roleName string) (entity.DomainUserGroup, error)
	}

	UserGroupMemberRepo interface {
		CreateMultiple(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error
	}
}

func (s *APIKeyPairService) GenerateKey(ctx context.Context, userID valueobj.HasUserID) error {
	organizationID, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return fmt.Errorf("interceptors.OrganizationFromContext err: %v", err)
	}

	apiKeypair, err := entity.NewRandomDomainAPIKeypair()
	if err != nil {
		return fmt.Errorf("entity.NewRandomDomainAPIKeypair err: %v", err)
	}

	apiKeypairToDelegate := entity.APIKeyPairToDelegate{
		APIKeyPair:        apiKeypair,
		HasUserID:         userID,
		HasOrganizationID: organizationID,
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := s.DomainAPIKeypairRepo.Create(ctx, tx, aggregate.DomainAPIKeypair{
			DomainAPIKeypair: apiKeypairToDelegate,
		})
		if err != nil {
			return fmt.Errorf("s.DomainAPIKeypairRepo.Create err: %v", err)
		}

		openAPIUserGroup, err := s.UserGroupRepo.FindUserGroupByRoleName(ctx, tx, constant.RoleOpenAPI)
		if err != nil {
			return fmt.Errorf("service.UserGroupRepo.FindUserGroupByRoleName err: %v", err)
		}

		err = s.UserGroupMemberRepo.CreateMultiple(ctx, tx, entity.UserGroupMemberWillBeDelegated{
			HasUserGroupID:    openAPIUserGroup,
			HasUserID:         userID,
			HasOrganizationID: organizationID,
		})
		if err != nil {
			return fmt.Errorf("s.UserGroupMemberRepo.CreateMultiple err: %v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("database.ExecInTx err: %v", err)
	}

	return nil
}
