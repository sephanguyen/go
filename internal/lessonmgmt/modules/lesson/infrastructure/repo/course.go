package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type CourseRepo struct{}

func (r *CourseRepo) UpdateEndDateByCourseIDs(ctx context.Context, db database.Ext, courseIDs []string, endDate time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.UpdateEndDateByCourseIDs")
	defer span.End()

	query := `
		UPDATE courses
		SET end_date = $2, updated_at = $3
		WHERE course_id = ANY($1)
			AND deleted_at IS NULL`
	_, err := db.Exec(ctx, query, &courseIDs, &endDate, time.Now())
	if err != nil {
		return fmt.Errorf("db.Exec: %s", err)
	}
	return nil
}

func (r *CourseRepo) ExportAllCoursesWithTeachingTimeValue(ctx context.Context, db database.QueryExecer, exportCols []exporter.ExportColumnMap) ([]byte, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.ExportAllCoursesWithTeachingTimeValue")
	defer span.End()

	courseTeaching := &CourseTeachingTimeToExport{}
	fields, _ := courseTeaching.FieldMap()

	query := `SELECT c.course_id, c.name, COALESCE(ctt.preparation_time,-1), COALESCE(ctt.break_time,-1), c.created_at, c.updated_at, c.deleted_at
		FROM courses c
		LEFT JOIN course_teaching_time ctt ON c.course_id = ctt.course_id
		WHERE c.deleted_at IS NULL
		AND ctt.deleted_at IS NULL
		ORDER BY name ASC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all courses with teaching time info: db.Query")
	}
	defer rows.Close()

	allCourses := []*CourseTeachingTimeToExport{}
	for rows.Next() {
		item := &CourseTeachingTimeToExport{}
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		allCourses = append(allCourses, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to export all course with teaching time info rows.Err")
	}

	exportable := sliceutils.Map(allCourses, func(d *CourseTeachingTimeToExport) database.Entity {
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("ExportBatch: %w", err)
	}
	// replace -1 as null value
	results := make([][]string, len(str))
	for index, row := range str {
		newRow := strings.Split(strings.ReplaceAll(strings.Join(row, ".."), "-1", ""), "..")
		results[index] = newRow
	}
	return exporter.ToCSV(results), nil
}

func (r *CourseRepo) CheckCourseIDs(ctx context.Context, db database.QueryExecer, ids []string) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.CheckCourseIDs")
	defer span.End()

	course := &Course{}
	fields, _ := course.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE course_id = ANY($1) AND deleted_at is null`,
		strings.Join(fields, ","),
		course.TableName(),
	)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	courses := make([]*Course, 0, len(ids))
	for rows.Next() {
		c := &Course{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return errors.Wrap(err, "rows.Scan")
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error on fetching Courses by IDs: %w", err)
	}

	if len(ids) != len(courses) {
		return fmt.Errorf("received Course IDs %v but only found %v", ids, courses)
	}

	return nil
}

func (r *CourseRepo) RegisterCourseTeachingTime(ctx context.Context, db database.QueryExecer, courses domain.Courses) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.RegisterCourseTeachingTime")
	defer span.End()

	courseDTOs := make([]*CourseTeachingTime, 0, len(courses))
	for _, c := range courses {
		courseDTO, err := NewCourseTeachingTimeFromEntity(c)
		if err != nil {
			return err
		}
		courseDTOs = append(courseDTOs, courseDTO)
	}

	b := &pgx.Batch{}
	for _, c := range courseDTOs {
		fields, args := c.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT course_id_pk DO 
		UPDATE SET preparation_time = $2, break_time = $3, updated_at = $5, deleted_at = $6`,
			c.TableName(),
			strings.Join(fields, ","),
			placeHolders)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course teaching time not updated")
		}
	}
	return nil
}
