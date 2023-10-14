package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
)

type CourseRepository struct{}

func (l *CourseRepository) GetByIDs(ctx context.Context, db database.Ext, id []string) ([]*domain.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepository.GetByIDs")
	defer span.End()

	fields := database.GetFieldNames(&domain.Course{})
	query := fmt.Sprintf(`
		SELECT %s FROM courses
		WHERE course_id = ANY($1)
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	row, err := db.Query(ctx, query, &id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	res := []*domain.Course{}
	for row.Next() {
		course := &domain.Course{}
		_, value := course.FieldMap()
		if err = row.Scan(value...); err != nil {
			return nil, err
		}
		res = append(res, course)
	}
	if err = row.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
