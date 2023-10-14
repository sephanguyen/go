package repo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type CourseClassRepo struct{}

func (r *CourseClassRepo) FindActiveCourseClassByID(ctx context.Context, db database.QueryExecer, classIDs []int32) (map[int32][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseClassRepo.FindActiveCourseClassByID")
	defer span.End()

	query := ` SELECT course_id, class_id
		FROM courses_classes
		WHERE class_id = ANY($1) 
		AND status = $2 `

	rows, err := db.Query(ctx, query, database.Int4Array(classIDs), string(domain.CourseClassStatusActive))
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var (
		courseID pgtype.Text
		classID  pgtype.Int4
	)
	result := make(map[int32][]string)
	for rows.Next() {
		if err := rows.Scan(&courseID, &classID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result[classID.Int] = append(result[classID.Int], courseID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}
