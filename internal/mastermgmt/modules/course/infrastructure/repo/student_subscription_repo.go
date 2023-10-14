package repo

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type StudentSubscriptionRepo struct{}

func (s *StudentSubscriptionRepo) GetLocationActiveStudentSubscriptions(ctx context.Context, db database.Ext, courseIDs []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.GetLocationActiveStudentSubscriptions")
	defer span.End()
	result := make(map[string][]string)
	for _, courseID := range courseIDs {
		query := fmt.Sprintf(`
			WITH course_location as (SELECT properties->'can_do_quiz'->>0 as course_id, unnest(location_ids) as location_id
				FROM student_packages s
				WHERE properties->'can_do_quiz' @> '["%s"]'
				and end_at > now()
				and s.deleted_at is null and s.is_active = true)
				select distinct l.location_id from course_location cl join locations l 
				on cl.location_id = l.location_id and l.deleted_at is null
		`, courseID)
		rows, err := db.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var locationIDs []string
		for rows.Next() {
			var locationID string
			if err := rows.Scan(&locationID); err != nil {
				return nil, fmt.Errorf("rows.Scan: %w", err)
			}
			locationIDs = append(locationIDs, locationID)
		}
		result[courseID] = locationIDs
	}

	return result, nil
}
