package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type CourseRepo struct{}

func (c *CourseRepo) GetMapCourseByIDs(ctx context.Context, db database.Ext, ids []string) (map[string]*Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetMapCourseByIDs")
	defer span.End()
	courseDTO := &Course{}

	fields := database.GetFieldNames(courseDTO)
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE course_id = ANY($1)
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		courseDTO.TableName(),
	)

	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapCourse := make(map[string]*Course)
	for rows.Next() {
		course := &Course{}
		_, value := course.FieldMap()
		if err = rows.Scan(value...); err != nil {
			return nil, err
		}
		mapCourse[course.CourseID.String] = course
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return mapCourse, nil
}
