package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type CourseBookRepo struct{}

func (r *CourseBookRepo) FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByCourseIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND course_id = ANY($1)`
	b := &entities_bob.CoursesBooks{}
	fields, _ := b.FieldMap()

	books := entities_bob.CoursesBookss{}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &courseIDs).ScanAll(&books)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := map[string][]string{}
	for _, v := range books {
		result[v.CourseID.String] = append(result[v.CourseID.String], v.BookID.String)
	}

	return result, nil
}

func (rcv *CourseBookRepo) SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs, bookIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.SoftDelete")
	defer span.End()

	query := "UPDATE courses_books SET deleted_at = now(), updated_at = now() WHERE book_id = ANY($1) AND course_id = ANY($2) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &bookIDs, &courseIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete course book")
	}

	return nil
}

func (r *CourseBookRepo) Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.CoursesBooks) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities_bob.CoursesBooks) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT courses_books_pk DO UPDATE SET updated_at = $3, deleted_at = $5", t.TableName(), strings.Join(fieldNames, ","), placeHolders)
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
			return fmt.Errorf("course book not inserted")
		}
	}
	return nil
}

func (r *CourseBookRepo) FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByBookID")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = $1`
	b := &entities_bob.CoursesBooks{}
	fields, _ := b.FieldMap()

	books := entities_bob.CoursesBookss{}
	pgID := database.Text(bookID)
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &pgID).ScanAll(&books)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := []string{}
	for _, v := range books {
		result = append(result, v.CourseID.String)
	}

	return result, nil
}

func (r *CourseBookRepo) FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) ([]*entities_bob.CoursesBooks, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByBookIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = ANY($1)`
	b := &entities_bob.CoursesBooks{}
	fields, _ := b.FieldMap()

	coursesBooks := entities_bob.CoursesBookss{}
	pgIDs := database.TextArray(bookIDs)
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &pgIDs).ScanAll(&coursesBooks)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return coursesBooks, nil
}
