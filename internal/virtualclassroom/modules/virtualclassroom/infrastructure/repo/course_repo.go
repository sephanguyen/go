package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type CourseRepo struct{}

func (c *CourseRepo) GetValidCoursesByCourseIDsAndStatus(ctx context.Context, db database.QueryExecer, courseIDs []string, status domain.CourseStatus) (domain.Courses, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetValidCoursesByCourseIDsAndStatus")
	defer span.End()

	course := &Course{}
	fields, values := course.FieldMap()

	baseQuery := fmt.Sprintf(` SELECT DISTINCT c.%s
		FROM %s c
		LEFT JOIN courses_academic_years ca ON c.course_id = ca.course_id
		LEFT JOIN academic_years ay ON ay.academic_year_id = ca.academic_year_id`,
		strings.Join(fields, ",c."),
		course.TableName(),
	)

	whereClause := ` WHERE c.course_id = ANY($1) 
		AND c.deleted_at IS NULL 
		AND l.deleted_at is NULL 
		AND c.status != 'COURSE_STATUS_INACTIVE'
		AND (ay.status IS NULL OR ay.status = 'ACADEMIC_YEAR_STATUS_ACTIVE') `

	if len(status) > 0 {
		baseQuery += ` LEFT JOIN lessons AS l ON c.course_id = l.course_id `

		switch status {
		case domain.StatusActive:
			whereClause += ` AND c.end_date >= now() AND l.lesson_id IS NOT NULL `
		case domain.StatusCompleted:
			whereClause += ` AND c.end_date < now() AND l.lesson_id IS NOT NULL `
		case domain.StatusOnGoing:
			whereClause += ` AND l.lesson_id IS NULL `
		}
	}

	query := baseQuery + whereClause
	rows, err := db.Query(ctx, query, &courseIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	courses := []*domain.Course{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		courses = append(courses, course.ToCourseDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return courses, nil
}
