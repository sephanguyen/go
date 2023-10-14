package entryexitmgmt

import (
	"context"
	"fmt"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type userOption func(ctx context.Context, db database.Ext, u *bob_entities.User, ugm *entity.UserGroupMember) error

func withID(id string) userOption {
	return func(ctx context.Context, db database.Ext, u *bob_entities.User, ugm *entity.UserGroupMember) error {
		u.ID = database.Text(id)
		return nil
	}
}

func withUserGroup(group string) userOption {
	return func(ctx context.Context, db database.Ext, u *bob_entities.User, ugm *entity.UserGroupMember) error {
		u.Group = database.Text(group)
		return nil
	}
}

func withResourcePath(resourcePath string) userOption {
	return func(ctx context.Context, db database.Ext, u *bob_entities.User, ugm *entity.UserGroupMember) error {
		u.ResourcePath = database.Text(resourcePath)
		return nil
	}
}

func withRole(roleName string) userOption {
	return func(ctx context.Context, db database.Ext, u *bob_entities.User, ugm *entity.UserGroupMember) error {
		stepState := StepStateFromContext(ctx)

		stmt := "SELECT user_group_id FROM user_group WHERE user_group_name = $1 AND resource_path = $2"

		var userGroupID string
		row := db.QueryRow(ctx, stmt, roleName, stepState.ResourcePath)

		err := row.Scan(&userGroupID)
		if err != nil {
			return fmt.Errorf("Error on fetching user group %v", err)
		}

		ugm.UserGroupID = database.Text(userGroupID)

		return nil
	}
}
