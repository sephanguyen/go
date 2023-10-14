package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type RoleRepo struct{}

func (r *RoleRepo) GetRolesByRoleIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) ([]*entity.Role, error) {
	ctx, span := interceptors.StartSpan(ctx, "RoleRepo.GetRolesByRoleIDs")
	defer span.End()
	role := &entity.Role{}

	fields := database.GetFieldNames(role)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE role_id = ANY ($1) AND deleted_at IS NULL", strings.Join(fields, ","), role.TableName())
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	roles := []*entity.Role{}
	for rows.Next() {
		rolePt := new(entity.Role)
		if err := rows.Scan(database.GetScanFields(rolePt, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		roles = append(roles, rolePt)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return roles, nil
}

// Deprecated: please use GetByUserGroupIDs in domain_repo.go
func (r *RoleRepo) GetRolesByUserGroupIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) (map[string][]*entity.Role, error) {
	ctx, span := interceptors.StartSpan(ctx, "RoleRepo.GetRolesByUserGroupIDs")
	defer span.End()

	query := `
	  SELECT
	    r.*, gr.user_group_id
	  FROM
	    role r

	  INNER JOIN granted_role gr
	    ON r.role_id = gr.role_id AND
	       gr.deleted_at IS NULL

	  WHERE
	    gr.user_group_id = any($1) AND
	    gr.deleted_at IS NULL
	`
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	defer rows.Close()

	mapUserGroupRoles := make(map[string][]*entity.Role)
	roleFields := database.GetFieldNames(new(entity.Role))
	for rows.Next() {
		role := entity.Role{}
		userGroupID := ""
		scanFields := append(database.GetScanFields(&role, roleFields), &userGroupID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}
		mapUserGroupRoles[userGroupID] = append(mapUserGroupRoles[userGroupID], &role)
	}

	return mapUserGroupRoles, nil
}

func (r *RoleRepo) Create(ctx context.Context, db database.Ext, role *entity.Role) error {
	ctx, span := interceptors.StartSpan(ctx, "RoleRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		role.CreatedAt.Set(now),
		role.UpdatedAt.Set(now),
		role.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("err set role: %w", err)
	}

	cmdTag, err := database.Insert(ctx, role, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert role: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("no row effected")
	}

	return nil
}

func (r *RoleRepo) UpsertPermission(ctx context.Context, db database.Ext, permissionRoles []*entity.PermissionRole) error {
	ctx, span := interceptors.StartSpan(ctx, "RoleRepo.UpsertPermission")
	defer span.End()

	batch := &pgx.Batch{}
	if err := r.queueUpsert(ctx, batch, permissionRoles); err != nil {
		return fmt.Errorf("queueUpsert error: %w", err)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func (r *RoleRepo) GetByName(ctx context.Context, db database.QueryExecer, name pgtype.Text) (*entity.Role, error) {
	ctx, span := interceptors.StartSpan(ctx, "RoleRepo.GetByName")
	defer span.End()
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	role := &entity.Role{}
	fields := database.GetFieldNames(role)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE role_name = $1 AND resource_path = $2", strings.Join(fields, ","), role.TableName())
	row := db.QueryRow(ctx, query, &name, &resourcePath)
	if err := row.Scan(database.GetScanFields(role, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return role, nil
}

func (r *RoleRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, permissionRoles []*entity.PermissionRole) error {
	queue := func(batch *pgx.Batch, permissionRole *entity.PermissionRole) {
		fieldNames := database.GetFieldNames(permissionRole)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT pk__permission_role 
			DO UPDATE SET created_at = $4, updated_at = $5, deleted_at = NULL`,
			permissionRole.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		batch.Queue(stmt, database.GetScanFields(permissionRole, fieldNames)...)
	}

	now := time.Now()
	for _, permissionRoleEnt := range permissionRoles {
		if permissionRoleEnt.PermissionID.Status != pgtype.Present {
			continue
		}

		if permissionRoleEnt.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := permissionRoleEnt.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			permissionRoleEnt.CreatedAt.Set(now),
			permissionRoleEnt.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, permissionRoleEnt)
	}
	return nil
}

func (r *RoleRepo) FindBelongedRoles(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]*entity.Role, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupV2Repo.FindBelongedRoles")
	defer span.End()

	fields, _ := new(entity.Role).FieldMap()
	stmt := fmt.Sprintf(
		`
		  SELECT r.%s
		  FROM role r

		  INNER JOIN granted_role g
		    ON g.role_id = r.role_id AND
		       g.deleted_at IS NULL AND
		       g.resource_path = r.resource_path

		  WHERE
		    g.user_group_id = $1
		`,
		strings.Join(fields, ", r."),
	)

	roles := entity.Roles{}
	if err := database.Select(ctx, db, stmt, &userGroupID).ScanAll(&roles); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return roles, nil
}
