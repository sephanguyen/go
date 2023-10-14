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

type UserAccessPathRepo struct{}

func (r *UserAccessPathRepo) Upsert(ctx context.Context, db database.QueryExecer, userAccessPaths []*entity.UserAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "UserAccessPathRepo.Upsert")
	defer span.End()

	if len(userAccessPaths) == 0 {
		return nil
	}

	var userIDs pgtype.TextArray

	for _, v := range userAccessPaths {
		userIDs = database.AppendText(userIDs, v.UserID)
	}

	batch := &pgx.Batch{}
	now := time.Now()
	batch.Queue(`UPDATE user_access_paths SET deleted_at = $1 WHERE user_id = ANY($2)`, now, userIDs)
	if err := r.queueUpsert(ctx, batch, userAccessPaths); err != nil {
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

func (r *UserAccessPathRepo) FindLocationIDsFromUserID(ctx context.Context, db database.QueryExecer, userID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserAccessPathRepo.FindParentIDsFromStudentID")
	defer span.End()

	locationIDs := []string{}

	query := fmt.Sprintf(`
			SELECT location_id FROM %s 
			WHERE user_id = $1 AND deleted_at is null
		`,
		(&entity.UserAccessPath{}).TableName())

	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		locationID := ""
		if err := rows.Scan(&locationID); err != nil {
			return nil, err
		}

		locationIDs = append(locationIDs, locationID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return locationIDs, nil
}

func (r *UserAccessPathRepo) Delete(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "UserAccessPathRepo.Delete")
	defer span.End()

	query := "UPDATE user_access_paths SET deleted_at = now(), updated_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &userIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("cannot delete user_access_path")
	}
	return nil
}

func (r *UserAccessPathRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, userAccessPaths []*entity.UserAccessPath) error {
	queue := func(b *pgx.Batch, userAccessPath *entity.UserAccessPath) {
		fieldNames := database.GetFieldNames(userAccessPath)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT user_access_paths_pk 
			DO UPDATE SET created_at = $4, updated_at = $5, access_path = $3, deleted_at = NULL`,
			userAccessPath.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		b.Queue(stmt, database.GetScanFields(userAccessPath, fieldNames)...)
	}

	now := time.Now()
	for _, uapEnt := range userAccessPaths {
		if uapEnt.LocationID.Status != pgtype.Present {
			continue
		}

		if uapEnt.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := uapEnt.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			uapEnt.CreatedAt.Set(now),
			uapEnt.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, uapEnt)
	}

	return nil
}
