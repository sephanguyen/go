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

type UserAddressRepo struct{}

func (r *UserAddressRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error {
	sql := `UPDATE user_address SET deleted_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &userIDs)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *UserAddressRepo) Upsert(ctx context.Context, db database.QueryExecer, userAddresses []*entity.UserAddress) error {
	ctx, span := interceptors.StartSpan(ctx, "UserAddressRepo.Upsert")
	defer span.End()

	batch := &pgx.Batch{}
	if err := r.queueUpsert(ctx, batch, userAddresses); err != nil {
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

func (r *UserAddressRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, userAddresses []*entity.UserAddress) error {
	queue := func(b *pgx.Batch, userAddress *entity.UserAddress) {
		fieldNames := database.GetFieldNames(userAddress)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT user_address__pk 
			DO UPDATE SET user_id = EXCLUDED.user_id, address_type = EXCLUDED.address_type, postal_code = EXCLUDED.postal_code, 
            prefecture_id = EXCLUDED.prefecture_id, city = EXCLUDED.city, 
            first_street = EXCLUDED.first_street, second_street = EXCLUDED.second_street, 
            updated_at = now(), deleted_at = NULL`,
			userAddress.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		b.Queue(stmt, database.GetScanFields(userAddress, fieldNames)...)
	}

	now := time.Now()
	for _, userAddress := range userAddresses {
		if userAddress.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := userAddress.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			userAddress.CreatedAt.Set(now),
			userAddress.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, userAddress)
	}

	return nil
}

func (r *UserAddressRepo) GetByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserAddress, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserAddressRepo.GetByUserID")
	defer span.End()

	userAddress := &entity.UserAddress{}
	fields := database.GetFieldNames(userAddress)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), userAddress.TableName())

	rows, err := db.Query(ctx, stmt, &userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userAddresses := make([]*entity.UserAddress, 0)
	for rows.Next() {
		userAddress := &entity.UserAddress{}
		if err := rows.Scan(database.GetScanFields(userAddress, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		userAddresses = append(userAddresses, userAddress)
	}

	return userAddresses, nil
}
