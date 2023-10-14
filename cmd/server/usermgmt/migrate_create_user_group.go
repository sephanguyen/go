package usermgmt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	whiteSpaceCharacter    = " "
	dashCharacter          = "-"
	verticalSlashCharacter = "|"
)

func RunMigrateCreateUserGroup(ctx context.Context, config *configurations.Config, userGroupName, roles, locationIDs, organizationID string) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", config.Common.Environment == "local")
	zLogger.Sugar().Info("-----START: Migration create user group-----")
	defer zLogger.Sugar().Sync()

	if err := validationArgs(userGroupName, roles, organizationID); err != nil {
		zLogger.Fatal(fmt.Sprintf("validationArgs failed: %s", err.Error()))
	}

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, config.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	organization, err := (&repository.OrganizationRepo{}).Find(ctx, dbPool, database.Text(strings.TrimSpace(organizationID)))
	if err != nil {
		zLogger.Fatal(fmt.Sprintf("find organization %s failed: %s", organizationID, err.Error()))
	}

	ctx = auth.InjectFakeJwtToken(ctx, organization.OrganizationID.String)
	if err := database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		locationOrg, err := (&repo.LocationRepo{}).GetLocationOrg(ctx, tx, organization.OrganizationID.String)
		if err != nil {
			return errors.Wrap(err, "get location org failed")
		}

		userGroupName = strings.ReplaceAll(userGroupName, verticalSlashCharacter, whiteSpaceCharacter)
		userGroup, err := toUserGroupEntity(userGroupName, organization.OrganizationID.String, locationOrg.LocationID)
		if err != nil {
			return errors.Wrap(err, "to user group entity")
		}
		if err := (&repository.UserGroupV2Repo{}).Create(ctx, tx, userGroup); err != nil {
			return errors.Wrap(err, "create user group failed")
		}

		mapRoleAndLocationIDs := separateRoleAndLocation(roles, locationIDs)
		return grantPermissionForUserGroup(ctx, tx, userGroup, mapRoleAndLocationIDs)
	}); err != nil {
		zLogger.Fatal(fmt.Sprintf("RunMigrateCreateUserGroup failed: %s", err.Error()))
	}

	zLogger.Sugar().Info("-----END: Migration create user group success-----")
}

func validationArgs(userGroupName, roles, organizationID string) error {
	switch {
	case userGroupName == "":
		return fmt.Errorf("userGroupName must not empty")
	case roles == "":
		return fmt.Errorf("roles must not empty")
	case organizationID == "":
		return fmt.Errorf("organizationID must not empty")
	}

	return nil
}

func grantPermissionForUserGroup(ctx context.Context, db pgx.Tx, userGroup *entity.UserGroupV2, mapRoleAndLocationIDs map[string][]string) error {
	for roleName, locationIDs := range mapRoleAndLocationIDs {
		role, err := (&repository.RoleRepo{}).GetByName(ctx, db, database.Text(roleName))
		if err != nil {
			return errors.Wrap(err, "get role by role name failed")
		}

		if len(locationIDs) > 0 {
			// validate locationIDs valid in our system
			if _, err := (&repo.LocationRepo{}).GetLocationsByLocationIDs(ctx, db, database.TextArray(locationIDs), false); err != nil {
				return errors.Wrap(err, "get locations by ids failed")
			}
		} else {
			// granted permission to org level if don't have locationIDs
			locationIDs = []string{userGroup.OrgLocationID.String}
		}

		grantedRole, err := toGrantedRoleEntity(role.RoleID.String, userGroup.UserGroupID.String, userGroup.ResourcePath.String)
		if err != nil {
			return errors.Wrap(err, "toGrantedRoleEntity failed")
		}

		grantedRoleRepo := &repository.GrantedRoleRepo{}
		if err = grantedRoleRepo.Create(ctx, db, grantedRole); err != nil {
			return errors.Wrap(err, "create granted role failed")
		}

		if err := grantedRoleRepo.LinkGrantedRoleToAccessPath(ctx, db, grantedRole, locationIDs); err != nil {
			return errors.Wrap(err, "link granted role to access path failed")
		}
	}

	return nil
}

func separateRoleAndLocation(roles, locationIDs string) map[string][]string {
	// separate list roles by "-"
	listRoles := strings.Split(roles, dashCharacter)
	// separate list locationIDs by "-"
	listLocationIDs := []string{}
	if locationIDs != "" {
		listLocationIDs = strings.Split(locationIDs, dashCharacter)
	}

	mapRoleAndLocation := make(map[string][]string)
	for idx, role := range listRoles {
		roleName := strings.ReplaceAll(role, verticalSlashCharacter, whiteSpaceCharacter)
		locationIDs := []string{}

		if len(listLocationIDs) > idx {
			locationIDs = strings.Split(strings.TrimSpace(listLocationIDs[idx]), verticalSlashCharacter)
		}

		mapRoleAndLocation[roleName] = locationIDs
	}
	return mapRoleAndLocation
}

func toUserGroupEntity(name, resourcePath, orgLocationID string) (*entity.UserGroupV2, error) {
	userGroup := &entity.UserGroupV2{}
	database.AllNullEntity(userGroup)
	if err := multierr.Combine(
		userGroup.UserGroupID.Set(idutil.ULIDNow()),
		userGroup.UserGroupName.Set(name),
		userGroup.ResourcePath.Set(resourcePath),
		userGroup.OrgLocationID.Set(orgLocationID),
		userGroup.IsSystem.Set(true),
	); err != nil {
		return nil, fmt.Errorf("set user group failed: %w", err)
	}

	return userGroup, nil
}

func toGrantedRoleEntity(roleID, userGroupID, resourcePath string) (*entity.GrantedRole, error) {
	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	if err := multierr.Combine(
		grantedRole.GrantedRoleID.Set(idutil.ULIDNow()),
		grantedRole.UserGroupID.Set(userGroupID),
		grantedRole.RoleID.Set(roleID),
		grantedRole.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, fmt.Errorf("set granted role failed: %w", err)
	}

	return grantedRole, nil
}
