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

type UserGroupRepo struct{}

func (r *UserGroupRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entity.UserGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.Upsert")
	defer span.End()

	now := time.Now()

	if err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("err set userGroup: %w", err)
	}

	fields, values := e.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	stmt := "INSERT INTO " + e.TableName() + " (" + strings.Join(fields, ",") + ") VALUES (" + placeHolders + ") ON CONFLICT (user_id,group_id) DO UPDATE SET updated_at = $6, is_origin = $3, status = $4;"

	cmd, err := db.Exec(ctx, stmt, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("cannot upsert userGroup")
	}

	return nil
}

func (r *UserGroupRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, userID, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.UpdateStatus")
	defer span.End()

	sql := `UPDATE users_groups SET updated_at = NOW(), status = $2 WHERE user_id = $1`
	cmdTag, err := db.Exec(ctx, sql, &userID, &status)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("not found any records")
	}

	return nil
}

func (r *UserGroupRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entity.UserGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entity.UserGroup) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			u.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	b := &pgx.Batch{}
	now := time.Now()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, u := range userGroups {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		if u.ResourcePath.Status == pgtype.Null {
			if err := u.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(userGroups); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("userGroup not inserted")
		}
	}

	return nil
}

func (r *UserGroupRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entity.UserGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.UpsertMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, userGroup *entity.UserGroup) {
		fields, values := userGroup.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			`
				INSERT INTO %s (%s) VALUES (%s)
				ON CONFLICT ON CONSTRAINT pk__users_groups
				DO update set status = $%d, is_origin = true, updated_at = now()
			`,
			userGroup.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			len(fields)+1,
		)

		batch.Queue(stmt, append(values, entity.UserGroupStatusActive)...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	for _, userGroup := range userGroups {
		if err := multierr.Combine(
			userGroup.CreatedAt.Set(now),
			userGroup.UpdatedAt.Set(now),
		); err != nil {
			return err
		}
		queueFn(batch, userGroup)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range userGroups {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %v", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("user group not inserted")
		}
	}

	return nil
}

func (r *UserGroupRepo) UpdateOrigin(ctx context.Context, db database.QueryExecer, userID pgtype.Text, isOrigin pgtype.Bool) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.UpdateOrigin")
	defer span.End()

	sql := `UPDATE users_groups SET is_origin = $1 WHERE user_id = $2`
	cmdTag, err := db.Exec(ctx, sql, &isOrigin, &userID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("not found any records")
	}

	return nil
}

func (r *UserGroupRepo) DeactivateMultiple(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, groupID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.DeactivateMultiple")
	defer span.End()

	sql := `
		UPDATE users_groups
		SET
			is_origin = FALSE,
			status = $1,
			updated_at = NOW()
		WHERE
			group_id = $2 AND
			user_id = ANY($3)
	`
	if _, err := db.Exec(ctx, sql, entity.UserGroupStatusInActive, groupID, userIDs); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}
