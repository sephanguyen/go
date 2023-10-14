package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type CourseBookRepo struct{}

func (r *CourseBookRepo) FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByCourseIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND course_id = ANY($1::_TEXT)`
	b := &entities.CoursesBooks{}
	fields, _ := b.FieldMap()

	books := entities.CoursesBookss{}
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

// nolint
func (rcv *CourseBookRepo) SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs, bookIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.SoftDelete")
	defer span.End()

	query := "UPDATE courses_books SET deleted_at = now() WHERE book_id = ANY($1::_TEXT) AND course_id = ANY($2::_TEXT) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &bookIDs, &courseIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete course book")
	}

	return nil
}

const bulkUpsertCourseBookStmTpl = `INSERT INTO %s (%s) 
VALUES %s 
ON CONFLICT ON CONSTRAINT courses_books_pk 
DO UPDATE SET 
	updated_at = excluded.updated_at, 
	deleted_at = excluded.deleted_at
`

func (r *CourseBookRepo) Upsert(ctx context.Context, db database.Ext, courseBooks []*entities.CoursesBooks) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.Upsert")
	defer span.End()

	now := time.Now()
	for _, courseBook := range courseBooks {
		err := multierr.Combine(
			courseBook.CreatedAt.Set(now),
			courseBook.UpdatedAt.Set(now),
		)
		if err != nil {
			return err
		}
	}
	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertCourseBookStmTpl, courseBooks)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertCourseBook error: %s", err.Error())
	}
	return nil
}

// nolint
func (r *CourseBookRepo) FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByBookID")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = $1::TEXT`
	b := &entities.CoursesBooks{}
	fields, _ := b.FieldMap()

	books := entities.CoursesBookss{}
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

func (r *CourseBookRepo) FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) ([]*entities.CoursesBooks, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByBookIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = ANY($1::_TEXT)`
	b := &entities.CoursesBooks{}
	fields, _ := b.FieldMap()

	coursesBooks := entities.CoursesBookss{}
	pgIDs := database.TextArray(bookIDs)
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &pgIDs).ScanAll(&coursesBooks)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return coursesBooks, nil
}

func (r *CourseBookRepo) FindByCourseIDAndBookID(ctx context.Context, db database.QueryExecer, bookID, courseID pgtype.Text) (*entities.CoursesBooks, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByCourseIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND book_id = $1::TEXT AND course_id = $2::TEXT`
	e := &entities.CoursesBooks{}
	fields, _ := e.FieldMap()
	var result entities.CoursesBooks
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), e.TableName()), &bookID, &courseID).ScanOne(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *CourseBookRepo) FindByCourseIDsV2(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.CoursesBooks, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.FindByCourseIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND ($1::TEXT[] IS NULL OR course_id = ANY($1::TEXT[]))`
	b := &entities.CoursesBooks{}
	fields, _ := b.FieldMap()

	books := entities.CoursesBookss{}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &courseIDs).ScanAll(&books)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return books, nil
}
