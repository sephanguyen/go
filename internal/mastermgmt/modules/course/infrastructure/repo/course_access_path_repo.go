package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type CourseAccessPathRepo struct{}

func (c *CourseAccessPathRepo) Upsert(ctx context.Context, db database.Ext, cc []*domain.CourseAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *CourseAccessPath) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT course_access_paths_pk DO UPDATE
		SET updated_at = now(), deleted_at = NULL`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, t := range cc {
		courseAccessPath, err := NewCourseAccessPathFromEntity(t)
		if err != nil {
			return fmt.Errorf("NewCourseAccessPathFromEntity err: %w", err)
		}
		queue(b, courseAccessPath)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course access path not inserted")
		}
	}
	return nil
}

func (c *CourseAccessPathRepo) Delete(ctx context.Context, db database.QueryExecer, courseIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.Delete")
	defer span.End()

	query := "UPDATE course_access_paths SET deleted_at = now(), updated_at = now() WHERE course_id = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &courseIDs)
	if err != nil {
		return fmt.Errorf("CourseAccessPathRepo delete err: %w", err)
	}

	return nil
}

func (c *CourseAccessPathRepo) GetByCourseIDs(ctx context.Context, db database.Ext, courseIDs []string) ([]*domain.CourseAccessPath, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.GetByCourseIDs")
	defer span.End()
	t := &CourseAccessPath{}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_id = ANY ($1) AND deleted_at IS NULL", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, &courseIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*domain.CourseAccessPath
	for rows.Next() {
		p := new(CourseAccessPath)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p.ToCourseAccessPathEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (c *CourseAccessPathRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]*CourseAccessPath, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.GetAll")
	defer span.End()
	t := &CourseAccessPath{}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caps []*CourseAccessPath
	for rows.Next() {
		courseAP := new(CourseAccessPath)
		if err := rows.Scan(database.GetScanFields(courseAP, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		caps = append(caps, courseAP)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return caps, nil
}
