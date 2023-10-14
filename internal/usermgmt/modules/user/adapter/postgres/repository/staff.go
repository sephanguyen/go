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

type StaffRepo struct{}

func (r *StaffRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, staffs []*entity.Staff) error {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.CreateMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, user *entity.Staff) {
		fields, values := user.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		// with `CONFLICT DO NOTHING`, avoid synced teachers, school admins (LT-14306)
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
			user.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)
		batch.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	for _, user := range staffs {
		user.CreatedAt = database.Timestamptz(now)
		user.UpdatedAt = database.Timestamptz(now)
		queueFn(batch, user)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range staffs {
		if _, err := batchResults.Exec(); err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (r *StaffRepo) Update(ctx context.Context, db database.QueryExecer, staff *entity.Staff) (*entity.Staff, error) {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.Update")
	defer span.End()
	now := time.Now()

	if err := multierr.Combine(
		staff.UpdatedAt.Set(now),
		staff.LegacyUser.UpdatedAt.Set(now),
	); err != nil {
		return nil, fmt.Errorf("err set entity: %w", err)
	}

	// update user
	cmdTag, err := database.Update(ctx, &staff.LegacyUser, db.Exec, "user_id")
	if err != nil {
		return nil, err
	}
	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("cannot update user")
	}

	// update staff
	cmdTag, err = database.Update(ctx, staff, db.Exec, "staff_id")
	if err != nil {
		return nil, err
	}
	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("cannot update staff")
	}

	return staff, nil
}

func (r *StaffRepo) UpdateStaffOnly(ctx context.Context, db database.QueryExecer, staff *entity.Staff) error {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.UpdateStaffOnly")
	defer span.End()
	now := time.Now()

	if err := multierr.Combine(
		staff.UpdatedAt.Set(now),
		staff.LegacyUser.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	// update user
	cmdTag, err := database.Update(ctx, staff, db.Exec, "staff_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update staff")
	}
	return nil
}

func (r *StaffRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]entity.Staff, error) {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.Retrieve")
	defer span.End()

	staff := &entity.Staff{}
	staffFields := database.GetFieldNames(staff)
	user := &entity.LegacyUser{}
	userFields := database.GetFieldNames(user)

	selectFields := make([]string, 0, len(staffFields)+len(userFields))
	for _, staffField := range staffFields {
		selectFields = append(selectFields, staff.TableName()+"."+staffField)
	}

	for _, userField := range userFields {
		selectFields = append(selectFields, user.TableName()+"."+userField)
	}

	selectStmt := fmt.Sprintf(
		`
		SELECT %s FROM
		staff JOIN users
		  ON staff_id=user_id
		WHERE
		  staff_id = ANY($1) AND
		  staff.deleted_at IS NULL
		`,
		strings.Join(selectFields, ", "),
	)
	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	staffs := make([]entity.Staff, 0, len(ids.Elements))
	for rows.Next() {
		staff := entity.Staff{}
		scanFields := append(database.GetScanFields(&staff, staffFields), database.GetScanFields(&staff.LegacyUser, userFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		staffs = append(staffs, staff)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return staffs, nil
}

func (r *StaffRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Staff, error) {
	staffs, err := r.Retrieve(ctx, db, database.TextArray([]string{id.String}))
	if err != nil {
		return nil, err
	}

	if len(staffs) == 0 {
		return nil, pgx.ErrNoRows
	}

	return &staffs[0], nil
}

func (r *StaffRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.SoftDelete")
	defer span.End()

	sql := `UPDATE staff SET deleted_at = NOW(), updated_at = NOW() WHERE staff_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &studentIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r *StaffRepo) Create(ctx context.Context, db database.QueryExecer, staff *entity.Staff) error {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		staff.CreatedAt.Set(now),
		staff.UpdatedAt.Set(now),

		staff.LegacyUser.ID.Set(staff.ID.String),
		staff.LegacyUser.CreatedAt.Set(now),
		staff.LegacyUser.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	if staff.LegacyUser.ResourcePath.Status == pgtype.Null {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		if err := multierr.Combine(
			staff.LegacyUser.ResourcePath.Set(resourcePath),
			staff.ResourcePath.Set(resourcePath),
		); err != nil {
			return err
		}
	}

	if _, err := database.Insert(ctx, &staff.LegacyUser, db.Exec); err != nil {
		return fmt.Errorf("err insert user: %w", err)
	}

	cmdTag, err := database.Insert(ctx, staff, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert staff: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	group := &entity.UserGroup{}
	err = multierr.Combine(
		group.UserID.Set(staff.ID.String),
		group.GroupID.Set(staff.LegacyUser.Group.String),
		group.IsOrigin.Set(true),
		group.Status.Set(entity.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
		group.ResourcePath.Set(staff.LegacyUser.ResourcePath),
	)
	if err != nil {
		return fmt.Errorf("err set UserGroup: %w", err)
	}

	cmdTag, err = database.Insert(ctx, group, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert UserGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	return nil
}

func (r *StaffRepo) Find(ctx context.Context, db database.QueryExecer, staffID pgtype.Text) (*entity.Staff, error) {
	ctx, span := interceptors.StartSpan(ctx, "StaffRepo.Find")
	defer span.End()

	staff := &entity.Staff{}
	fields := database.GetFieldNames(staff)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE staff_id = $1", strings.Join(fields, ","), staff.TableName())
	row := db.QueryRow(ctx, query, &staffID)
	if err := row.Scan(database.GetScanFields(staff, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return staff, nil
}
