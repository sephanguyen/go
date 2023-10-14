package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/support"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type UserGroupMembersRepo struct {
}

func (r *UserGroupMembersRepo) FindUserIDsByUserGroupID(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupMembersRepo.FindByUserGroupID")
	defer span.End()

	c := new(entities.UserGroupMember)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_group_id = $1", strings.Join(fields, ","), c.TableName())

	rows, err := db.Query(ctx, selectStmt, &userGroupID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var userGroupMemberIDs []string
	for rows.Next() {
		var (
			userID, userGroupID string
		)
		err := rows.Scan(&userID, &userGroupID)
		if err != nil {
			return nil, err
		}
		userGroupMemberIDs = append(userGroupMemberIDs, userID)
	}
	return userGroupMemberIDs, nil
}
