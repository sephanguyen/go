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

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	Separator = " "
)

type findUserInListIDs struct {
	lastUserID    string
	limit         int
	userIDs       []string
	usersWithRole string
}

func (f *findUserInListIDs) findPaginatingMapRoleWithUsers(ctx context.Context, dbPool database.QueryExecer) (map[string][]*entity.LegacyUser, error) {
	var tableName, columnIDName string

	switch f.usersWithRole {
	case constant.RoleTeacher:
		tableName = new(entity.Teacher).TableName()
		columnIDName = "teacher_id"
	case constant.RoleSchoolAdmin:
		tableName = new(entity.SchoolAdmin).TableName()
		columnIDName = "school_admin_id"
	}

	query := `
	  SELECT %[1]s, resource_path from %[2]s
	  WHERE
	    resource_path IS NOT NULL AND
	    %[1]s > $1
	    %[3]s
	    ORDER BY %[1]s ASC LIMIT $2;
	`
	queryArgs := []interface{}{f.lastUserID, f.limit}
	conditionRangeUser := ""
	// if user ids was passed, we will find all, unless just find specified users
	if len(f.userIDs) > 0 {
		conditionRangeUser = fmt.Sprintf("AND %s = ANY($3)", columnIDName)
		queryArgs = append(queryArgs, database.TextArray(f.userIDs))
	}

	query = fmt.Sprintf(query, columnIDName, tableName, conditionRangeUser)
	rows, err := dbPool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err, "query get remainning users failed")
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows get remainning users failed")
	}

	// separate result into user type roles
	var user *entity.LegacyUser
	mapUserWithRole := map[string][]*entity.LegacyUser{}
	mapUserWithRole[f.usersWithRole] = []*entity.LegacyUser{}
	for rows.Next() {
		user = new(entity.LegacyUser)
		var userID, resourcePath string

		if err := rows.Scan(&userID, &resourcePath); err != nil {
			return nil, fmt.Errorf("failed to scan an orgs row: %s", err)
		}

		user.ID = database.Text(userID)
		user.ResourcePath = database.Text(resourcePath)

		mapUserWithRole[f.usersWithRole] = append(mapUserWithRole[f.usersWithRole], user)
	}

	// re-assign last user id for cursor pagination purpose
	if user != nil {
		f.lastUserID = user.ID.String
	}
	return mapUserWithRole, nil
}

func RunMigrationAssignUsergroupToSpecificStaff(ctx context.Context, c *configurations.Config, organizationID, userGroupID, userIDsSequence string) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----Migration Assign Usergroup To Specify Staff-----")
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

	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	userIDs := ExtractSliceFromSequenceElements(userIDsSequence, Separator)
	userGroupRepo := new(repository.UserGroupV2Repo)
	roleRepo := new(repository.RoleRepo)
	userGroup, err := userGroupRepo.Find(ctx, dbPool, database.Text(userGroupID))
	if err != nil {
		zLogger.Fatal(errors.Wrapf(err, "userGroupRepo.Find: %s", userGroupID).Error())
	}

	roles, err := roleRepo.FindBelongedRoles(ctx, dbPool, userGroup.UserGroupID)
	if err != nil {
		zLogger.Fatal(errors.Wrapf(err, "userGroupRepo.FindBelongedRoles: %s", userGroup.UserGroupID.String).Error())
	}

	// find role support for migration
	var usersWithRole string
	mapRoleWithUserGroup := map[string]*entity.UserGroupV2{}
	for _, role := range roles {
		switch role.RoleName.String {
		case constant.RoleTeacher:
			usersWithRole = constant.RoleTeacher
		case constant.RoleSchoolAdmin:
			usersWithRole = constant.RoleSchoolAdmin
		default:
			continue
		}
		break
	}

	if usersWithRole == "" {
		zLogger.Sugar().Fatalf("can not find role of user group %s support for migration", userGroup.UserGroupID)
		return
	}
	mapRoleWithUserGroup[usersWithRole] = userGroup

	findListUsers := &findUserInListIDs{
		lastUserID:    "",
		limit:         1000,
		userIDs:       userIDs,
		usersWithRole: usersWithRole,
	}

	if err := runMigrationAddDefaultUserGroupForUsers(ctx, dbPool, mapRoleWithUserGroup, findListUsers); err != nil {
		zLogger.Sugar().Infof("-----Migration run failed: %s-----", err.Error())
	}

	zLogger.Sugar().Infof("-----Done migration-----")
}

func ExtractSliceFromSequenceElements(sequence, separateChar string) []string {
	sliceElements := []string{}
	for _, ele := range strings.Split(sequence, separateChar) {
		ele = strings.TrimSpace(ele)
		if ele != "" {
			sliceElements = append(sliceElements, ele)
		}
	}
	return sliceElements
}
