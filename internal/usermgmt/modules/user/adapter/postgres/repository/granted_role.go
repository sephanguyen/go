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

type GrantedRoleRepo struct{}

func (r *GrantedRoleRepo) Create(ctx context.Context, db database.QueryExecer, grantedRole *entity.GrantedRole) error {
	ctx, span := interceptors.StartSpan(ctx, "GrantedRoleRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		grantedRole.UpdatedAt.Set(now),
		grantedRole.CreatedAt.Set(now),
		grantedRole.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("err set grantedrole: %w", err)
	}

	cmdTag, err := database.Insert(ctx, grantedRole, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert grantedrole: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot upsert grantedrole")
	}
	return nil
}

func (r *GrantedRoleRepo) LinkGrantedRoleToAccessPath(ctx context.Context, db database.QueryExecer, grantedRole *entity.GrantedRole, locationIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "GrantedRoleRepo.LinkGrantedRoleToAccessPath")
	defer span.End()

	queueFn := func(batch *pgx.Batch, grap *entity.GrantedRoleAccessPath) {
		fields, values := grap.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
			grap.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	now := time.Now()
	batch := &pgx.Batch{}

	for _, locationID := range locationIDs {
		grantedRoleAccessPath := &entity.GrantedRoleAccessPath{}
		if err := multierr.Combine(
			grantedRoleAccessPath.GrantedRoleID.Set(grantedRole.GrantedRoleID),
			grantedRoleAccessPath.LocationID.Set(locationID),
			grantedRoleAccessPath.CreatedAt.Set(now),
			grantedRoleAccessPath.UpdatedAt.Set(now),
			grantedRoleAccessPath.DeletedAt.Set(nil),
			grantedRoleAccessPath.ResourcePath.Set(grantedRole.ResourcePath),
		); err != nil {
			return fmt.Errorf("err set grantedRoleAccessPath: %w", err)
		}
		queueFn(batch, grantedRoleAccessPath)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for index := 0; index < len(locationIDs); index++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("cannot upsert grantedRoleAccessPath")
		}
	}

	return nil
}

func (r *GrantedRoleRepo) GetByUserGroup(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]*entity.GrantedRole, error) {
	ctx, span := interceptors.StartSpan(ctx, "GrantedRoleRepo.GetByUserGroup")
	defer span.End()

	grantedRole := &entity.GrantedRole{}
	fields := database.GetFieldNames(grantedRole)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_group_id = $1", strings.Join(fields, ","), grantedRole.TableName())

	rows, err := db.Query(ctx, stmt, &userGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grantedRoles := make([]*entity.GrantedRole, 0)
	for rows.Next() {
		grantedRole := &entity.GrantedRole{}
		if err := rows.Scan(database.GetScanFields(grantedRole, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		grantedRoles = append(grantedRoles, grantedRole)
	}

	return grantedRoles, nil
}

func (r *GrantedRoleRepo) Upsert(ctx context.Context, db database.QueryExecer, grantedRoles []*entity.GrantedRole) error {
	ctx, span := interceptors.StartSpan(ctx, "GrantedRoleRepo.Upsert")
	defer span.End()

	batch := &pgx.Batch{}
	if err := r.queueUpsert(ctx, batch, grantedRoles); err != nil {
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

func (r *GrantedRoleRepo) SoftDelete(ctx context.Context, db database.QueryExecer, grantedRoleIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "GrantedRoleRepo.SoftDelete")
	defer span.End()

	query := `UPDATE granted_role SET deleted_at = now(), updated_at = now() WHERE granted_role_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, query, &grantedRoleIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r *GrantedRoleRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, grantedRoles []*entity.GrantedRole) error {
	queue := func(batch *pgx.Batch, grantedRole *entity.GrantedRole) {
		fieldNames := database.GetFieldNames(grantedRole)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT pk__granted_role 
			DO UPDATE SET created_at = $4, updated_at = $5, deleted_at = NULL`,
			grantedRole.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		batch.Queue(stmt, database.GetScanFields(grantedRole, fieldNames)...)
	}

	now := time.Now()
	for _, grantedRoleEnt := range grantedRoles {
		if grantedRoleEnt.RoleID.Status != pgtype.Present {
			continue
		}

		if grantedRoleEnt.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := grantedRoleEnt.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			grantedRoleEnt.CreatedAt.Set(now),
			grantedRoleEnt.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, grantedRoleEnt)
	}

	return nil
}
