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
	"go.uber.org/multierr"
)

type GrantedRoleAccessPathRepo struct{}

func (r *GrantedRoleAccessPathRepo) Upsert(ctx context.Context, db database.QueryExecer, grantedRoleAccessPaths []*entity.GrantedRoleAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "GrantedRoleAccessPathRepo.Upsert")
	defer span.End()

	var gtrAccessPathIDs []string
	for _, gtrAccessPath := range grantedRoleAccessPaths {
		if gtrAccessPath.GrantedRoleID.String != "" {
			gtrAccessPathIDs = append(gtrAccessPathIDs, gtrAccessPath.GrantedRoleID.String)
		}
	}

	batch := &pgx.Batch{}
	if len(gtrAccessPathIDs) > 0 {
		batch.Queue(`UPDATE granted_role_access_path SET deleted_at = $1 WHERE granted_role_id = ANY($2)`, time.Now(), database.TextArray(gtrAccessPathIDs))
	}

	if err := r.queueUpsert(ctx, batch, grantedRoleAccessPaths); err != nil {
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

func (r *GrantedRoleAccessPathRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, grantedRoleAccessPaths []*entity.GrantedRoleAccessPath) error {
	queue := func(b *pgx.Batch, grantedRoleAccessPath *entity.GrantedRoleAccessPath) {
		fieldNames := database.GetFieldNames(grantedRoleAccessPath)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT pk__granted_role_access_path 
			DO UPDATE SET created_at = $3, updated_at = $4, deleted_at = NULL`,
			grantedRoleAccessPath.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		b.Queue(stmt, database.GetScanFields(grantedRoleAccessPath, fieldNames)...)
	}

	now := time.Now()
	for _, grantedRoleAccessPathEnt := range grantedRoleAccessPaths {
		if grantedRoleAccessPathEnt.LocationID.Status != pgtype.Present {
			continue
		}

		if grantedRoleAccessPathEnt.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := grantedRoleAccessPathEnt.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			grantedRoleAccessPathEnt.CreatedAt.Set(now),
			grantedRoleAccessPathEnt.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, grantedRoleAccessPathEnt)
	}

	return nil
}
