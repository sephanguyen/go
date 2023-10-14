package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type CourseAccessPathRepo struct{}

func (r *CourseAccessPathRepo) GetCourseAccessPathByUCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapCourseAccess map[string]interface{}, err error) {
	var (
		valueMap interface{}
	)
	mapCourseAccess = make(map[string]interface{})
	stmt :=
		`
		SELECT course_id,location_id
		FROM 
			%s
		WHERE 
			course_id = ANY($1) AND deleted_at IS NULL
		`
	stmt = fmt.Sprintf(
		stmt,
		(&entities.CourseAccessPath{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, courseIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var courseID, locationID string
		err = rows.Scan(&courseID, &locationID)
		if err != nil {
			err = fmt.Errorf("row.Scan: %w", err)
			return
		}
		key := fmt.Sprintf("%v_%v", locationID, courseID)
		mapCourseAccess[key] = valueMap
	}
	return
}
