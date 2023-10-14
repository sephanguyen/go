package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type CourseClassRepo struct{}

func (r *CourseClassRepo) FindByCourseID(ctx context.Context, db database.Ext, courseID pgtype.Text, isAll bool) (map[pgtype.Int4]*entities_bob.CourseClass, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindBySchoolAndID")
	defer span.End()

	e := new(entities_bob.CourseClass)
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_id = $1", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL AND status = 'COURSE_CLASS_STATUS_ACTIVE'"
	}
	classes := map[pgtype.Int4]*entities_bob.CourseClass{}
	rows, err := db.Query(ctx, query, &courseID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities_bob.CourseClass)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		classes[c.ClassID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return classes, nil

}
func (r *CourseClassRepo) FindByCourseIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray, isAll bool) ([]*entities_bob.CourseClass, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindBySchoolAndID")
	defer span.End()

	e := new(entities_bob.CourseClass)
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_id = ANY($1)", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL AND status = 'COURSE_CLASS_STATUS_ACTIVE'"
	}
	classes := []*entities_bob.CourseClass{}
	rows, err := db.Query(ctx, query, &courseIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities_bob.CourseClass)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		classes = append(classes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return classes, nil

}

func (r *CourseClassRepo) UpsertV2(ctx context.Context, db database.Ext, cc []*entities_bob.CourseClass) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseClassRepo.UpsertV2")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities_bob.CourseClass) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT courses_classes_pk DO UPDATE SET status = $3, updated_at = $4, deleted_at = $6", t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		t.CreatedAt.Set(now)
		t.UpdatedAt.Set(now)

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course class not inserted")
		}
	}
	return nil
}

func (r *CourseClassRepo) SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseClassRepo.SoftDelete")
	defer span.End()

	query := "UPDATE courses_classes SET status = 'COURSE_CLASS_STATUS_INACTIVE', deleted_at = now(), updated_at = now() WHERE course_id = ANY($1) AND deleted_at IS NULL AND status != 'COURSE_CLASS_STATUS_INACTIVE'"
	cmdTag, err := db.Exec(ctx, query, &courseIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete course class")
	}

	return nil
}

func (r *CourseClassRepo) FindByClassIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapByClassID map[pgtype.Int4]pgtype.TextArray, err error) {
	query := `SELECT class_id, ARRAY_AGG(course_id)
	FROM courses_classes
	WHERE class_id = ANY($1) AND status = $2
	GROUP BY class_id`

	var pgStatus pgtype.Text
	_ = pgStatus.Set(entities_bob.CourseClassStatusActive)

	rows, err := db.Query(ctx, query, &ids, &pgStatus)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	defer rows.Close()

	result := make(map[pgtype.Int4]pgtype.TextArray)
	for rows.Next() {
		var (
			classID   pgtype.Int4
			courseIDs pgtype.TextArray
		)

		if err := rows.Scan(&classID, &courseIDs); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result[classID] = courseIDs
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}

func (r *CourseClassRepo) SoftDeleteClass(ctx context.Context, db database.QueryExecer, classID pgtype.Int4) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseClassRepo.SoftDelete")
	defer span.End()

	query := "UPDATE courses_classes SET status = 'COURSE_CLASS_STATUS_INACTIVE', deleted_at = now(), updated_at = now() WHERE class_id = $1"
	_, err := db.Exec(ctx, query, &classID)
	if err != nil {
		return err
	}

	return nil
}
