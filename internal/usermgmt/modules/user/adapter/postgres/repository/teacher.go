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

// TeacherRepo provides method to work with teachers entity
type TeacherRepo struct{}

func (r *TeacherRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entity.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.CreateMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, user *entity.Teacher) {
		fields, values := user.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			user.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	for _, user := range teachers {
		_ = user.UpdatedAt.Set(now)
		_ = user.CreatedAt.Set(now)
		queueFn(batch, user)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(teachers); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("teacher not inserted")
		}
	}

	return nil
}

func (r *TeacherRepo) Upsert(ctx context.Context, db database.QueryExecer, teacher *entity.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Upsert")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		teacher.CreatedAt.Set(now),
		teacher.UpdatedAt.Set(now),
		teacher.DeletedAt.Set(nil),
	); err != nil {
		return err
	}

	fieldNames, fields := teacher.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(
		`
		   INSERT INTO %s (%s)
		   VALUES (%s)
		      ON CONFLICT ON CONSTRAINT teachers_pk
		      DO UPDATE SET updated_at = $3, deleted_at = NULL
		`,
		teacher.TableName(),
		strings.Join(fieldNames, ","),
		placeHolder,
	)
	cmdTag, err := db.Exec(ctx, query, fields...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot upsert teacher %s", teacher.ID.String)
	}
	return nil
}

func (r *TeacherRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, teachers []*entity.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.UpsertMultiple")
	defer span.End()

	queueFn := func(batch *pgx.Batch, teacher *entity.Teacher) {
		fields, values := teacher.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			`
				INSERT INTO %s (%s) VALUES (%s)
				ON CONFLICT ON CONSTRAINT teachers_pk
				DO update set updated_at = now(), deleted_at = null
			`,
			teacher.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	for _, teacher := range teachers {
		if err := multierr.Combine(
			teacher.UpdatedAt.Set(now),
			teacher.CreatedAt.Set(now),
			teacher.DeletedAt.Set(nil),
		); err != nil {
			return errors.Wrap(err, "teacher.ResourcePath.Set")
		}
		queueFn(batch, teacher)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range teachers {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %v", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("teacher not upserted")
		}
	}

	return nil
}

func (r *TeacherRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entity.Teacher, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Retrieve")
	defer span.End()

	teacher := &entity.Teacher{}
	teacherFields := database.GetFieldNames(teacher)
	user := &entity.LegacyUser{}
	userFields := database.GetFieldNames(user)

	selectFields := make([]string, 0, len(teacherFields)+len(userFields))
	for _, f := range teacherFields {
		selectFields = append(selectFields, teacher.TableName()+"."+f)
	}

	for _, f := range userFields {
		selectFields = append(selectFields, user.TableName()+"."+f)
	}

	selectStmt := fmt.Sprintf("SELECT %s FROM teachers JOIN users ON teacher_id=user_id WHERE teacher_id=ANY($1) AND teachers.deleted_at IS NULL;",
		strings.Join(selectFields, ","),
	)

	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := make([]entity.Teacher, 0, len(ids.Elements))
	for rows.Next() {
		teacher := entity.Teacher{}
		scanFields := append(database.GetScanFields(&teacher, teacherFields), database.GetScanFields(&teacher.LegacyUser, userFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		teachers = append(teachers, teacher)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}

func (r *TeacherRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Teacher, error) {
	teachers, err := r.Retrieve(ctx, db, database.TextArray([]string{id.String}))
	if err != nil {
		return nil, err
	}

	if len(teachers) == 0 {
		return nil, pgx.ErrNoRows
	}

	return &teachers[0], nil
}

func (r *TeacherRepo) SoftDelete(ctx context.Context, db database.QueryExecer, teacherID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.SoftDelete")
	defer span.End()

	teacherIDs := []string{teacherID.String}
	return r.SoftDeleteMultiple(ctx, db, database.TextArray(teacherIDs))
}

func (r *TeacherRepo) SoftDeleteMultiple(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.SoftDeleteMultiple")
	defer span.End()

	query := `UPDATE teachers SET deleted_at = now(), updated_at = now() WHERE teacher_id = any($1) AND deleted_at IS NULL`
	if _, err := db.Exec(ctx, query, &teacherIDs); err != nil {
		return err
	}

	return nil
}

// Create teacher to be called in txn
func (r *TeacherRepo) Create(ctx context.Context, db database.QueryExecer, teacher *entity.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		teacher.CreatedAt.Set(now),
		teacher.UpdatedAt.Set(now),
	); err != nil {
		return err
	}

	if teacher.ResourcePath.Status == pgtype.Null {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		if err := multierr.Combine(
			teacher.ResourcePath.Set(resourcePath),
		); err != nil {
			return err
		}
	}

	cmdTag, err := database.Insert(ctx, teacher, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert teacher: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	return nil
}

func (r *TeacherRepo) Find(ctx context.Context, db database.QueryExecer, teacherID pgtype.Text) (*entity.Teacher, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Find")
	defer span.End()

	teacher := &entity.Teacher{}
	fields := database.GetFieldNames(teacher)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE teacher_id = $1", strings.Join(fields, ","), teacher.TableName())
	row := db.QueryRow(ctx, query, &teacherID)
	if err := row.Scan(database.GetScanFields(teacher, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return teacher, nil
}

func (r *TeacherRepo) Update(ctx context.Context, db database.QueryExecer, teacher *entity.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Update")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		teacher.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	// update teacher
	cmdTag, err := database.Update(ctx, teacher, db.Exec, "teacher_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update teacher")
	}

	return nil
}
