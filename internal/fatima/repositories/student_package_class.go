package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type StudentPackageClassRepo struct{}

const studentPackageClassRepoBulkUpsertStmtTpl = `INSERT INTO %s AS sp (%s) VALUES (%s)
ON CONFLICT (student_package_id, course_id, student_id, location_id, class_id)
DO UPDATE SET
	deleted_at = NULL,
	updated_at = NOW()`

func (r *StudentPackageClassRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudentPackageClass) error {
	batch := &pgx.Batch{}
	e := &entities.StudentPackageClass{}
	currentTime := timeutil.Now().UTC()

	for _, item := range items {
		if err := multierr.Combine(
			item.UpdatedAt.Set(currentTime),
			item.CreatedAt.Set(currentTime),
			item.DeletedAt.Set(nil),
		); err != nil {
			return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set DeletedAt.Set: %w", err)
		}
		fieldNames, value := item.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(studentPackageClassRepoBulkUpsertStmtTpl, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

		batch.Queue(query, value...)
	}
	result := db.SendBatch(ctx, batch)
	defer result.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

const studentPackageClassRepoDeleteByStudentPackageIDsStmtTpl = `
	UPDATE %s 
	SET deleted_at = NOW() 
	WHERE student_package_id = ANY($1) 
	AND deleted_at IS NULL`

func (r *StudentPackageClassRepo) DeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, spIDs pgtype.TextArray) error {
	spc := &entities.StudentPackageClass{}
	query := fmt.Sprintf(studentPackageClassRepoDeleteByStudentPackageIDsStmtTpl, spc.TableName())
	_, err := db.Exec(ctx, query, &spIDs)
	if err != nil {
		return err
	}
	return nil
}

const studentPackageClassRepoDeleteByStudentPackageIDAndCourseIDStmtTpl = `
	UPDATE %s 
	SET deleted_at = NOW() 
	WHERE student_package_id = $1 AND course_id = $2
	AND deleted_at IS NULL`

func (r *StudentPackageClassRepo) DeleteByStudentPackageIDAndCourseID(ctx context.Context, db database.QueryExecer, studentPackageID string, courseID string) error {
	spc := &entities.StudentPackageClass{}
	query := fmt.Sprintf(studentPackageClassRepoDeleteByStudentPackageIDAndCourseIDStmtTpl, spc.TableName())
	_, err := db.Exec(ctx, query, studentPackageID, courseID)
	if err != nil {
		return err
	}
	return nil
}
