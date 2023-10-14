package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type CourseAccessPathRepo struct{}

func (c *CourseAccessPathRepo) Upsert(ctx context.Context, db database.Ext, cc []*entities.CourseAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.CourseAccessPath) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT course_access_paths_pk DO UPDATE
		SET updated_at = now(), deleted_at = NULL`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		err := multierr.Combine(
			t.CreatedAt.Set(now),
			t.UpdatedAt.Set(now),
			t.DeletedAt.Set(nil),
		)

		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

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
			return fmt.Errorf("course access path not inserted")
		}
	}
	return nil
}

func (c *CourseAccessPathRepo) FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.FindByCourseIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND course_id = ANY($1)`
	b := &entities.CourseAccessPath{}
	fields, _ := b.FieldMap()

	caps := entities.CourseAccessPaths{}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &courseIDs).ScanAll(&caps)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := map[string][]string{}
	for _, v := range caps {
		result[v.CourseID.String] = append(result[v.CourseID.String], v.LocationID.String)
	}

	return result, nil
}

func (c *CourseAccessPathRepo) Delete(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseAccessPathRepo.Delete")
	defer span.End()

	query := "UPDATE course_access_paths SET deleted_at = now(), updated_at = now() WHERE course_id = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &courseIDs)
	if err != nil {
		return err
	}

	return nil
}
