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

type CourseStudentAccessPathRepo struct {
}

const courseStudentAccessPathRepoBulkUpsertStmtTpl = `INSERT INTO %s AS cs (%s) VALUES (%s)
ON CONFLICT (course_student_id, location_id)
DO UPDATE SET
	deleted_at = NULL,
	updated_at = NOW()`

func (p *CourseStudentAccessPathRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudentsAccessPath) error {
	b := &pgx.Batch{}
	e := &entities.CourseStudentsAccessPath{}
	currentTime := timeutil.Now().UTC()

	for _, item := range items {
		fieldNames, value := item.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(courseStudentAccessPathRepoBulkUpsertStmtTpl, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

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

func (p *CourseStudentAccessPathRepo) DeleteLatestCourseStudentAccessPathsByCourseStudentIDs(ctx context.Context, db database.QueryExecer, courseStudentIDs pgtype.TextArray) error {
	csap := &entities.CourseStudentsAccessPath{}
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE course_student_id = ANY($1) AND deleted_at IS NULL", csap.TableName())
	_, err := db.Exec(ctx, query, &courseStudentIDs)
	if err != nil {
		return err
	}
	return nil
}

func (p *CourseStudentAccessPathRepo) GetByLocationsAndStudents(ctx context.Context, db database.QueryExecer, locationIDs, studentIDs pgtype.TextArray) ([]*entities.CourseStudentsAccessPath, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentAccessPathRepo.GetByLocationsAndStudents")
	defer span.End()

	var courseStudentsAccessPath entities.CourseStudentsAccessPath
	var courseStudentsAccessPaths entities.CourseStudentsAccessPaths

	fields, _ := courseStudentsAccessPath.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT %s
    FROM %s
    WHERE location_id = ANY($1::_TEXT)
	AND student_id = ANY($2::_TEXT)
	AND deleted_at IS NULL`, strings.Join(fields, ", "), courseStudentsAccessPath.TableName())

	err := database.Select(ctx, db, stmt, &locationIDs, &studentIDs).ScanAll(&courseStudentsAccessPaths)

	if err != nil {
		return nil, err
	}

	return courseStudentsAccessPaths, nil
}

func (p *CourseStudentAccessPathRepo) GetByLocationsStudentsAndCourse(ctx context.Context, db database.QueryExecer, locationIDs, studentIDs, courseIDs pgtype.TextArray) ([]*entities.CourseStudentsAccessPath, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentAccessPathRepo.GetByLocationsStudentsAndCourse")
	defer span.End()

	var courseStudentsAccessPath entities.CourseStudentsAccessPath
	var courseStudentsAccessPaths entities.CourseStudentsAccessPaths

	fields, _ := courseStudentsAccessPath.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT %s
    FROM %s
    WHERE location_id = ANY($1::_TEXT)
	AND student_id = ANY($2::_TEXT)
	AND course_id = ANY($3::_TEXT)
	AND deleted_at IS NULL`, strings.Join(fields, ", "), courseStudentsAccessPath.TableName())

	err := database.Select(ctx, db, stmt, &locationIDs, &studentIDs, &courseIDs).ScanAll(&courseStudentsAccessPaths)

	if err != nil {
		return nil, err
	}

	return courseStudentsAccessPaths, nil
}

func (p *CourseStudentAccessPathRepo) GetByLocationsAndCourse(ctx context.Context, db database.QueryExecer, locationIDs, courseIDs pgtype.TextArray) ([]*entities.CourseStudentsAccessPath, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentAccessPathRepo.GetByLocationsAndCourse")
	defer span.End()

	var courseStudentsAccessPath entities.CourseStudentsAccessPath
	var courseStudentsAccessPaths entities.CourseStudentsAccessPaths

	fields, _ := courseStudentsAccessPath.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT %s FROM %s
    WHERE location_id = ANY($1::_TEXT)
	AND course_id = ANY($2::_TEXT)
	AND deleted_at IS NULL`, strings.Join(fields, ", "), courseStudentsAccessPath.TableName())

	err := database.Select(ctx, db, stmt, &locationIDs, &courseIDs).ScanAll(&courseStudentsAccessPaths)

	if err != nil {
		return nil, err
	}

	return courseStudentsAccessPaths, nil
}
