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
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// SchoolAdminRepo works with school_admin_id
type SchoolAdminRepo struct {
}

func (r *SchoolAdminRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entity.SchoolAdmin) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdmin.CreateMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, u *entity.SchoolAdmin) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			u.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	for _, u := range schoolAdmins {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		queueFn(batch, u)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(schoolAdmins); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("schoolAdmin not inserted")
		}
	}

	return nil
}

func (r *SchoolAdminRepo) Upsert(ctx context.Context, db database.QueryExecer, schoolAdmin *entity.SchoolAdmin) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdminRepo.Upsert")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		schoolAdmin.CreatedAt.Set(now),
		schoolAdmin.UpdatedAt.Set(now),
		schoolAdmin.DeletedAt.Set(nil),
	); err != nil {
		return err
	}

	fieldNames, fields := schoolAdmin.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(
		`
		   INSERT INTO %s (%s)
		   VALUES (%s)
		      ON CONFLICT ON CONSTRAINT school_admins_pk
		      DO UPDATE SET updated_at = $3, deleted_at = NULL
		`,
		schoolAdmin.TableName(),
		strings.Join(fieldNames, ","),
		placeHolder,
	)
	cmdTag, err := db.Exec(ctx, query, fields...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot upsert schoolAdmin %s", schoolAdmin.SchoolAdminID.String)
	}
	return nil
}

func (r *SchoolAdminRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entity.SchoolAdmin) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdminRepo.UpsertMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, u *entity.SchoolAdmin) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			`
				INSERT INTO %s (%s) VALUES (%s)
				ON CONFLICT ON CONSTRAINT school_admins_pk
				DO update set updated_at = now(), deleted_at = null
			`,
			u.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	for _, schoolAdmin := range schoolAdmins {
		if err := multierr.Combine(
			schoolAdmin.UpdatedAt.Set(now),
			schoolAdmin.CreatedAt.Set(now),
			schoolAdmin.DeletedAt.Set(nil),
		); err != nil {
			return errors.Wrap(err, "multierr.Combine")
		}
		queueFn(batch, schoolAdmin)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range schoolAdmins {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %v", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("school admin is not upserted")
		}
	}

	return nil
}

func (r *SchoolAdminRepo) Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entity.SchoolAdmin, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdminRepo.Get")
	defer span.End()

	schoolAdmin := &entity.SchoolAdmin{}
	fields := database.GetFieldNames(schoolAdmin)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE school_admin_id = $1", strings.Join(fields, ","), schoolAdmin.TableName())
	if err := database.Select(ctx, db, query, &schoolAdminID).ScanOne(schoolAdmin); err != nil {
		return nil, err
	}

	return schoolAdmin, nil
}

func (r *SchoolAdminRepo) SoftDelete(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdminRepo.SoftDelete")
	defer span.End()

	schoolAdminIDs := []string{schoolAdminID.String}
	return r.SoftDeleteMultiple(ctx, db, database.TextArray(schoolAdminIDs))
}

func (r *SchoolAdminRepo) SoftDeleteMultiple(ctx context.Context, db database.QueryExecer, schoolAdminIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdminRepo.SoftDeleteMultiple")
	defer span.End()

	query := `UPDATE school_admins SET deleted_at = now(), updated_at = now() WHERE school_admin_id = any($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, query, &schoolAdminIDs)
	if err != nil {
		return err
	}

	return nil
}
