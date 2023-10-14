package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// TeacherRepo provides method to work with teachers entity
type TeacherRepo struct{}

// Retrieve returns user list by ids. If not specific fields, return all fields
func (r *TeacherRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities_bob.Teacher, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Retrieve")
	defer span.End()

	t := &entities_bob.Teacher{}
	teacherFields := database.GetFieldNames(t)
	u := &entities_bob.User{}
	userFields := database.GetFieldNames(u)

	selectFields := make([]string, 0, len(teacherFields)+len(userFields))
	for _, f := range teacherFields {
		selectFields = append(selectFields, t.TableName()+"."+f)
	}

	for _, f := range userFields {
		selectFields = append(selectFields, u.TableName()+"."+f)
	}

	selectStmt := fmt.Sprintf("SELECT %s FROM teachers JOIN users ON teacher_id=user_id WHERE teacher_id=ANY($1) AND teachers.deleted_at IS NULL;",
		strings.Join(selectFields, ","),
	)

	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := make([]entities_bob.Teacher, 0, len(ids.Elements))
	for rows.Next() {
		t := entities_bob.Teacher{}
		scanFields := append(database.GetScanFields(&t, teacherFields), database.GetScanFields(&t.User, userFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		teachers = append(teachers, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}

func (r *TeacherRepo) IsInSchool(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray, schoolID pgtype.Int4) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.IsInSchool")
	defer span.End()

	e := &entities_bob.Teacher{}
	query := fmt.Sprintf("SELECT $1 = ANY(school_ids) AS is_in_school FROM %s WHERE teacher_id = ANY($2)", e.TableName())

	rows, err := db.Query(ctx, query, &schoolID, &teacherIDs)
	if err != nil {
		return false, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var b bool
		if err := rows.Scan(&b); err != nil {
			return false, errors.Wrap(err, "rows.Scan")
		}
		if b == false {
			return false, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, errors.Wrap(err, "rows.Err")
	}

	return true, nil
}

func (r *TeacherRepo) FindRegardlessDeletion(ctx context.Context, db database.QueryExecer, teacherID pgtype.Text) (*entities_bob.Teacher, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Find")
	defer span.End()

	teacher := &entities_bob.Teacher{}
	fields := database.GetFieldNames(teacher)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE teacher_id = $1", strings.Join(fields, ","), teacher.TableName())
	row := db.QueryRow(ctx, query, &teacherID)
	if err := row.Scan(database.GetScanFields(teacher, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return teacher, nil
}

func (r *TeacherRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.Teacher, error) {
	teachers, err := r.Retrieve(ctx, db, database.TextArray([]string{id.String}))
	if err != nil {
		return nil, err
	}

	if len(teachers) == 0 {
		return nil, pgx.ErrNoRows
	}

	return &teachers[0], nil
}

// Create teacher and user records, to be called in txn
func (r *TeacherRepo) Create(ctx context.Context, db database.QueryExecer, t *entities_bob.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.Create")
	defer span.End()

	now := timeutil.Now()
	t.CreatedAt.Set(now)
	t.UpdatedAt = t.CreatedAt

	t.User.CreatedAt = t.CreatedAt
	t.User.UpdatedAt = t.CreatedAt
	t.User.Group = database.Text(entities_bob.UserGroupTeacher)
	if t.User.ResourcePath.Status == pgtype.Null {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		t.User.ResourcePath.Set(resourcePath)
	}

	t.User.ID = t.ID
	if _, err := database.Insert(ctx, &t.User, db.Exec); err != nil {
		return err
	}

	cmdTag, err := database.Insert(ctx, t, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	group := &entities_bob.UserGroup{}
	err = multierr.Combine(
		group.UserID.Set(t.ID.String),
		group.GroupID.Set(entities_bob.UserGroupTeacher),
		group.IsOrigin.Set(true),
		group.Status.Set(entities_bob.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
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

// Get get teacher that has all school ids
func (r *TeacherRepo) GetTeacherHasSchoolIDs(ctx context.Context, db database.QueryExecer, teacherID string, schoolIDs []int32) (*entities_bob.Teacher, error) {
	teacher := &entities_bob.Teacher{}
	teacherFields := database.GetFieldNames(teacher)
	int32Array := pgtype.Int4Array{}
	int32Array.Set(schoolIDs)

	query := fmt.Sprintf("select %s from %s where teacher_id = $1 and school_ids @> $2 and deleted_at IS NULL", strings.Join(teacherFields, ","), teacher.TableName())
	err := db.QueryRow(ctx, query, teacherID, int32Array).Scan(database.GetScanFields(teacher, teacherFields)...)
	if err != nil {
		return nil, err
	}

	return teacher, nil
}

func (r *TeacherRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, teachers []*entities_bob.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities_bob.Teacher) {
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

	for _, u := range teachers {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(teachers); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("teacher not inserted")
		}
	}

	return nil
}

func (r *TeacherRepo) SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	sql := `UPDATE teachers SET deleted_at = NOW(), updated_at = NOW() WHERE teacher_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &ids)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *TeacherRepo) Update(ctx context.Context, db database.QueryExecer, s *entities_bob.Teacher) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Update")
	defer span.End()

	now := time.Now()
	s.UpdatedAt.Set(now)

	cmdTag, err := database.Update(ctx, s, db.Exec, "teacher_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update teacher")
	}

	return nil
}
