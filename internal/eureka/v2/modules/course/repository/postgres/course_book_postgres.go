package postgres

import (
	"context"
	"fmt"
	"time"

	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type CourseBookRepo struct {
	DB database.Ext
}

func (repo *CourseBookRepo) Upsert(ctx context.Context, courseBooks []*dto.CourseBookDto) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.Upsert")
	defer span.End()

	upsertCourseBookQuery := `INSERT INTO %s (%s) VALUES %s ON CONFLICT ON CONSTRAINT courses_books_pk DO UPDATE SET updated_at = excluded.updated_at, deleted_at = excluded.deleted_at`

	now := time.Now()
	for _, courseBook := range courseBooks {
		err := multierr.Combine(
			courseBook.CreatedAt.Set(now),
			courseBook.UpdatedAt.Set(now),
			courseBook.DeletedAt.Set(nil),
		)
		if err != nil {
			return errors.NewConversionError("CourseBookRepo.Upsert", err)
		}
	}
	err := eureka_db.BulkUpsert(ctx, repo.DB, upsertCourseBookQuery, courseBooks)
	if err != nil {
		return errors.NewDBError("eureka_db.BulkUpsertCourseBook", err)
	}
	return nil
}

func (repo *CourseBookRepo) RetrieveAssociatedBook(ctx context.Context, bookID string) ([]*dto.CourseBookDto, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseBookRepo.RetrieveAssociatedBook")
	defer span.End()

	e := &dto.CourseBookDto{}
	query := fmt.Sprintf(`SELECT b.book_id,cb.course_id,cb.created_at,cb.updated_at, cb.deleted_at FROM books b LEFT JOIN %s cb on b.book_id = cb.book_id WHERE b.book_id = $1 AND b.deleted_at IS null ORDER BY b.created_at DESC`, e.TableName())

	rows, err := repo.DB.Query(ctx, query, &bookID)
	if err != nil {
		return nil, errors.NewDBError("eureka_db.Query", err)
	}
	defer rows.Close()

	var records []*dto.CourseBookDto
	for rows.Next() {
		fmt.Println("have element")
		c := &dto.CourseBookDto{}
		if err := rows.Scan(&c.BookID, &c.CourseID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt); err != nil {
			return nil, errors.NewDBError("rows.Scan", err)
		}
		records = append(records, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewDBError("rows.Err", err)
	}

	return records, nil
}
