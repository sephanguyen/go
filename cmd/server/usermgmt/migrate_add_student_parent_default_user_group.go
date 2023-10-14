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
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type findListUsers interface {
	findPaginatingMapRoleWithUsers(ctx context.Context, dbPool database.QueryExecer) (map[string][]*entity.LegacyUser, error)
}
type findLearnerHasNoUserGroup struct {
	lastUserID string
	limit      int
}

func (f *findLearnerHasNoUserGroup) findPaginatingMapRoleWithUsers(ctx context.Context, dbPool database.QueryExecer) (map[string][]*entity.LegacyUser, error) {
	fieldName, _ := new(entity.LegacyUser).FieldMap()
	query := fmt.Sprintf(`
	WITH learners_had_user_group(user_id) AS (
	  SELECT
	    ugm.user_id

	  FROM
	    user_group_member ugm

	  INNER JOIN granted_role gr
	    ON ugm.user_group_id = gr.user_group_id AND
	       gr.deleted_at IS NULL

	  INNER JOIN role r
	    ON gr.role_id = r.role_id AND
	       r.deleted_at IS NULL

	  where
	    r.role_name = ANY (ARRAY['Student'::text, 'Parent'::text]) AND
	    ugm.deleted_at IS NULL
	)

	SELECT %s from users
	where
	  user_id NOT IN (SELECT user_id FROM learners_had_user_group) AND
	  user_group = ANY(ARRAY['USER_GROUP_STUDENT'::text, 'USER_GROUP_PARENT'::text]) AND
	  user_id > $1
	  ORDER BY user_id ASC LIMIT $2;
	`, strings.Join(fieldName, ", "))
	rows, err := dbPool.Query(ctx, query, f.lastUserID, f.limit)
	if err != nil {
		return nil, errors.Wrap(err, "query get remainning users failed")
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "rows get remainning users failed")
	}

	// separate result into user types group
	var user *entity.LegacyUser
	mapUserWithRole := map[string][]*entity.LegacyUser{}
	mapUserWithRole[constant.RoleParent] = []*entity.LegacyUser{}
	mapUserWithRole[constant.RoleStudent] = []*entity.LegacyUser{}
	for rows.Next() {
		user = new(entity.LegacyUser)
		_, userAttr := user.FieldMap()

		if err := rows.Scan(userAttr...); err != nil {
			return nil, fmt.Errorf("failed to scan an orgs row: %s", err)
		}

		// get default first role
		userRole := constant.MapLegacyUserGroupWithRoles[user.Group.String][0]
		if _, ok := mapUserWithRole[userRole]; !ok {
			return nil, fmt.Errorf("query get user with wrong group %s: %s", userRole, err)
		}
		mapUserWithRole[userRole] = append(mapUserWithRole[userRole], user)
	}

	// re-assign last user id for cursor pagination purpose
	if user != nil {
		f.lastUserID = user.ID.String
	}
	return mapUserWithRole, nil
}

func RunMigrationAddDefaultUserGroupForStudentParent(ctx context.Context, c *configurations.Config) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----Migration Add Default UserGroup For Student & Parent-----")
	defer zLogger.Sugar().Sync()

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	// run migration for each organization
	var orgID, orgName interface{}
	orgQuery := "SELECT organization_id, name FROM organizations"
	rows, err := dbPool.Query(ctx, orgQuery)
	if err != nil {
		zLogger.Fatal("Get orgs failed")
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&orgID, &orgName); err != nil {
			zLogger.Sugar().Infof("failed to scan an orgs row: %s", err)
			continue
		}

		id, ok := orgID.(string)
		if !ok {
			continue
		}
		ctx = auth.InjectFakeJwtToken(ctx, id)

		userGroupRepo := new(repository.UserGroupV2Repo)
		parentUserGroup, err := userGroupRepo.FindUserGroupByRoleName(ctx, dbPool, constant.RoleParent)
		if err != nil {
			zLogger.Fatal(errors.Wrap(err, "can not get parent user group").Error())
			return
		}
		studentUserGroup, err := userGroupRepo.FindUserGroupByRoleName(ctx, dbPool, constant.RoleStudent)
		if err != nil {
			zLogger.Fatal(errors.Wrap(err, "can not get student user group").Error())
			return
		}

		mapRoleWithUserGroups := map[string]*entity.UserGroupV2{
			constant.RoleStudent: studentUserGroup,
			constant.RoleParent:  parentUserGroup,
		}

		findListUsers := &findLearnerHasNoUserGroup{
			lastUserID: "",
			limit:      1000,
		}

		if err := runMigrationAddDefaultUserGroupForUsers(ctx, dbPool, mapRoleWithUserGroups, findListUsers); err != nil {
			zLogger.Sugar().Infof("-----Migration run failed for %s: %w-----", orgName, err)
			continue
		}

		zLogger.Sugar().Infof("-----Done migration for %s-----", orgName)
	}
}

func runMigrationAddDefaultUserGroupForUsers(ctx context.Context, dbPool *pgxpool.Pool, userGroups map[string]*entity.UserGroupV2, findGroupUsers findListUsers) error {
	for {
		mapUserWithRole, err := findGroupUsers.findPaginatingMapRoleWithUsers(ctx, dbPool)
		if err != nil {
			return errors.Wrap(err, "findAGroupUser")
		}

		totalQueriedUsers := 0
		for _, users := range mapUserWithRole {
			totalQueriedUsers += len(users)
		}
		if totalQueriedUsers == 0 {
			// run migration done
			return nil
		}

		for groupKey, userGroup := range userGroups {
			// assign found user groups to each type of user
			if err := assignUserGroupToUser(ctx, dbPool, userGroup, mapUserWithRole[groupKey]); err != nil {
				zLogger.Sugar().Infof("assignUserGroupToUser: %w", err.Error())
			}
		}
	}
}

func assignUserGroupToUser(ctx context.Context, db database.QueryExecer, userGroup *entity.UserGroupV2, users []*entity.LegacyUser) error {
	if len(users) == 0 {
		return nil
	}

	userGroupsMemberRepo := new(repository.UserGroupsMemberRepo)
	if err := userGroupsMemberRepo.AssignWithUserGroup(ctx, db, users, userGroup.UserGroupID); err != nil {
		return fmt.Errorf("can not assign %s user group to users", userGroup.UserGroupName.String)
	}
	return nil
}
