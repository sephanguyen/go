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

type UserPhoneNumberRepo struct{}

func (r *UserPhoneNumberRepo) Upsert(ctx context.Context, db database.QueryExecer, userPhoneNumbers []*entity.UserPhoneNumber) error {
	ctx, span := interceptors.StartSpan(ctx, "UserPhoneNumberRepo.Upsert")

	if len(userPhoneNumbers) == 0 {
		return nil
	}

	defer span.End()

	batch := &pgx.Batch{}
	if err := r.queueUpsert(ctx, batch, userPhoneNumbers); err != nil {
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

func (r *UserPhoneNumberRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, userPhoneNumbers []*entity.UserPhoneNumber) error {
	queue := func(b *pgx.Batch, userPhoneNumber *entity.UserPhoneNumber) {
		fieldNames := database.GetFieldNames(userPhoneNumber)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT user_phone_number__pk 
			DO UPDATE SET phone_number = excluded.phone_number, type = excluded.type, updated_at = now() 
			`,
			userPhoneNumber.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		b.Queue(stmt, database.GetScanFields(userPhoneNumber, fieldNames)...)
	}

	now := time.Now()
	for _, userPhoneNumber := range userPhoneNumbers {
		if userPhoneNumber.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := userPhoneNumber.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			userPhoneNumber.CreatedAt.Set(now),
			userPhoneNumber.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, userPhoneNumber)
	}

	return nil
}

func (r *UserPhoneNumberRepo) FindByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserPhoneNumber, error) {
	ctx, span := interceptors.StartSpan(ctx, "UseUserPhoneNumberRepo.FindByUserID")
	defer span.End()

	userPhoneNumber := &entity.UserPhoneNumber{}
	userPhoneNumbers := entity.UserPhoneNumbers{}

	fields, _ := userPhoneNumber.FieldMap()

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL",
		strings.Join(fields, ","), userPhoneNumber.TableName())

	err := database.Select(ctx, db, sql, &userID).ScanAll(&userPhoneNumbers)
	if err != nil {
		return nil, err
	}

	return userPhoneNumbers, nil
}

func (r *UserPhoneNumberRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error {
	sql := `UPDATE user_phone_number SET deleted_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &userIDs)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}
