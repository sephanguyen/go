package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type UserGroupV2Repo struct{}

func (r *UserGroupV2Repo) FindByIDs(ctx context.Context, db database.QueryExecer, userGroupIDs []string) ([]*entity.UserGroupV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.FindByIDs")
	defer span.End()

	userGroup := &entity.UserGroupV2{}
	userGroupFields := database.GetFieldNames(userGroup)
	queryStmt := `SELECT %s FROM %s WHERE user_group_id = ANY($1)`
	query := fmt.Sprintf(queryStmt, strings.Join(userGroupFields, ","), userGroup.TableName())

	rows, err := db.Query(ctx, query, &userGroupIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userGroups := make([]*entity.UserGroupV2, 0, len(userGroupIDs))
	for rows.Next() {
		userGroup := &entity.UserGroupV2{}
		scanFields := database.GetScanFields(userGroup, userGroupFields)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		userGroups = append(userGroups, userGroup)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userGroups, nil
}

func (r *UserGroupV2Repo) Create(ctx context.Context, db database.QueryExecer, userGroup *entity.UserGroupV2) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		userGroup.UpdatedAt.Set(now),
		userGroup.CreatedAt.Set(now),
		userGroup.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("err set usergroup: %w", err)
	}

	cmdTag, err := database.Insert(ctx, userGroup, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert usergroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot upsert usergroup")
	}
	return nil
}

func (r *UserGroupV2Repo) Find(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) (*entity.UserGroupV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.Find")
	defer span.End()

	userGroup := &entity.UserGroupV2{}
	fields := database.GetFieldNames(userGroup)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_group_id = $1", strings.Join(fields, ","), userGroup.TableName())
	row := db.QueryRow(ctx, query, &userGroupID)
	if err := row.Scan(database.GetScanFields(userGroup, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return userGroup, nil
}

func (r *UserGroupV2Repo) Update(ctx context.Context, db database.QueryExecer, userGroup *entity.UserGroupV2) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.Update")
	defer span.End()

	cmdTag, err := database.Update(ctx, userGroup, db.Exec, "user_group_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update student")
	}

	return nil
}

func (r *UserGroupV2Repo) FindUserGroupAndRoleByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (map[string][]*entity.Role, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.FindUserGroupAndRoleByUserID")
	defer span.End()

	query := `
		SELECT ug.*, r.*
		FROM user_group_member ugm

		INNER JOIN user_group ug ON
			ugm.user_group_id = ug.user_group_id AND
			ug.deleted_at IS NULL

		INNER JOIN granted_role gr ON
			ugm.user_group_id = gr.user_group_id AND
			gr.deleted_at IS NULL

		INNER JOIN role r ON
			r.role_id = gr.role_id AND
			r.deleted_at IS NULL

		WHERE ugm.user_id = $1 AND
		      ugm.deleted_at IS NULL
	`
	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	defer rows.Close()

	mapUserGroupRoles := make(map[string][]*entity.Role)
	for rows.Next() {
		role := entity.Role{}
		roleFields := database.GetFieldNames(&role)

		userGroup := entity.UserGroupV2{}
		userGroupFields := database.GetFieldNames(&userGroup)

		scanFields := append(database.GetScanFields(&userGroup, userGroupFields), database.GetScanFields(&role, roleFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		mapUserGroupRoles[userGroup.UserGroupName.String] = append(mapUserGroupRoles[userGroup.UserGroupName.String], &role)
	}

	return mapUserGroupRoles, nil
}

// FindUserGroupByRoleName: only find system user_group by role name for now
func (r *UserGroupV2Repo) FindUserGroupByRoleName(ctx context.Context, db database.QueryExecer, roleName string) (*entity.UserGroupV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.FindUserGroupByRoleName")
	defer span.End()

	userGroup := &entity.UserGroupV2{}
	grantedRole := &entity.GrantedRole{}
	fields := database.GetFieldNames(userGroup)
	userGroupFields := make([]string, len(fields))
	for index := range fields {
		userGroupFields[index] = fmt.Sprintf("ug.%s", fields[index])
	}
	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			%s ug
		INNER JOIN %s gr
			ON ug.user_group_id = gr.user_group_id
		INNER JOIN role r
			ON gr.role_id = r.role_id
		WHERE
			ug.is_system = true AND r.role_name = $1 AND r.deleted_at IS NULL;
	`, strings.Join(userGroupFields, ", "), userGroup.TableName(), grantedRole.TableName())
	if err := db.
		QueryRow(ctx, query, roleName).
		Scan(database.GetScanFields(userGroup, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return userGroup, nil
}

func (r *UserGroupV2Repo) FindAndMapUserGroupAndRolesByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (map[entity.UserGroupV2][]*entity.Role, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.FindAndMapUserGroupAndRolesByUserID")
	defer span.End()

	userGroup := &entity.UserGroupV2{}
	role := &entity.Role{}

	userGroupFields := database.GetFieldNames(userGroup)
	roleFields := database.GetFieldNames(role)

	selectFields := make([]string, 0, len(userGroupFields)+len(roleFields))
	for _, userGroupField := range userGroupFields {
		selectFields = append(selectFields, fmt.Sprintf("ug.%s", userGroupField))
	}
	for _, roleField := range roleFields {
		selectFields = append(selectFields, fmt.Sprintf("r.%s", roleField))
	}

	query := fmt.Sprintf(`
	SELECT %s
	FROM user_group_member ugm

	INNER JOIN %s ug ON
		ugm.user_group_id = ug.user_group_id AND
		ug.deleted_at IS NULL

	INNER JOIN granted_role gr ON
		ugm.user_group_id = gr.user_group_id AND
		gr.deleted_at IS NULL

	INNER JOIN %s r ON
		r.role_id = gr.role_id AND
		r.deleted_at IS NULL

	WHERE ugm.user_id = $1 AND
		  ugm.deleted_at IS NULL
`, strings.Join(selectFields, ", "), userGroup.TableName(), role.TableName())

	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	defer rows.Close()

	mapUserGroupAndRoles := make(map[entity.UserGroupV2][]*entity.Role)
	for rows.Next() {
		role := entity.Role{}

		userGroup := entity.UserGroupV2{}

		scanFields := append(database.GetScanFields(&userGroup, userGroupFields), database.GetScanFields(&role, roleFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		mapUserGroupAndRoles[userGroup] = append(mapUserGroupAndRoles[userGroup], &role)
	}

	return mapUserGroupAndRoles, nil
}
