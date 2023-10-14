package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type CourseClassRepo struct {
}

func (p *CourseClassRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseClass) error {
	b := &pgx.Batch{}
	e := &entities.CourseClass{}
	currentTime := timeutil.Now().UTC()

	for _, item := range items {
		fieldNames, value := item.FieldMap()
		const placeHolders = "$1, $2, $3, $4, $5, $6"

		query := fmt.Sprintf(`INSERT INTO %s AS cc (%s) VALUES (%s)
			ON CONFLICT (course_id, class_id)
			DO UPDATE SET deleted_at = NULL, updated_at = NOW()
			WHERE cc.deleted_at IS NOT NULL`,
			e.TableName(), strings.Join(fieldNames, ","), placeHolders)

		if item.CreatedAt.Status != pgtype.Present && item.UpdatedAt.Status != pgtype.Present {
			b.Queue(query, append(value[:3], currentTime, currentTime, nil)...)
		} else {
			b.Queue(query, value...)
		}
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

const courseClassRepoDeleteStmtTpl = `UPDATE course_classes SET deleted_at = NOW()
WHERE (course_id, class_id) IN (%s)
AND deleted_at IS NULL`

// Delete soft-deletes
func (p *CourseClassRepo) Delete(ctx context.Context, db database.QueryExecer, items []*entities.CourseClass) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseClassRepo.Delete")
	defer span.End()

	inCondition, args := database.CompositeKeysPlaceHolders(len(items), func(i int) []interface{} {
		return []interface{}{items[i].CourseID.String, items[i].ClassID.String}
	})

	_, err := db.Exec(ctx, fmt.Sprintf(courseClassRepoDeleteStmtTpl, inCondition), args...)
	if err != nil {
		return err
	}

	return nil
}

func (p *CourseClassRepo) DeleteClass(ctx context.Context, db database.QueryExecer, classID string) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseClassRepo.DeleteClass")
	defer span.End()

	courseClass := &entities.CourseClass{}

	query := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE class_id = $1 AND deleted_at IS NULL`, courseClass.TableName())

	cmd, err := db.Exec(ctx, query, classID)

	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any course class to delete: %w", pgx.ErrNoRows)
	}

	return nil
}

func (p *CourseClassRepo) FindClassIDByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) ([]string, error) {
	courseClass := &entities.CourseClass{}
	query := fmt.Sprintf(`SELECT class_id FROM %s WHERE deleted_at is NULL AND course_id = $1`, courseClass.TableName())
	rows, err := db.Query(ctx, query, &courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var classIDs []string
	for rows.Next() {
		var classID string
		if err := rows.Scan(&classID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		classIDs = append(classIDs, classID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return classIDs, nil
}
