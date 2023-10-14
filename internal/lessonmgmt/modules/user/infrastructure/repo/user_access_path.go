package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type UserAccessPathRepo struct {
	UserAccessPathRepo service.UserAccessPathRepo
}

func (u *UserAccessPathRepo) GetLocationAssignedByUserID(ctx context.Context, db database.QueryExecer, userID []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserAccessPathRepo.GetLocationAssignedByUserID")
	defer span.End()

	query := "SELECT user_id,location_id FROM user_access_paths WHERE user_id = ANY($1) AND deleted_at IS NULL"

	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[string][]string{}
	for rows.Next() {
		var userID, locationID pgtype.Text
		if err = rows.Scan(&userID, &locationID); err != nil {
			return nil, err
		}
		res[userID.String] = append(res[userID.String], locationID.String)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (u *UserAccessPathRepo) Create(ctx context.Context, db database.QueryExecer, userAccessPaths []*domain.UserAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "UserAccessPathRepo.Create")
	defer span.End()
	batch := &pgx.Batch{}
	userAccessPathEntity, err := NewUserAccessPath(userAccessPaths)
	if err != nil {
		return err
	}
	err = u.queueCreate(batch, userAccessPathEntity)
	if err != nil {
		return err
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

func (u *UserAccessPathRepo) queueCreate(batch *pgx.Batch, userAccessPaths []*UserAccessPath) error {
	queue := func(b *pgx.Batch, userAccessPath *UserAccessPath) {
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
	for _, uap := range userAccessPaths {
		if err := uap.PreUpsert(); err != nil {
			return err
		}
		queue(batch, uap)
	}

	return nil
}
