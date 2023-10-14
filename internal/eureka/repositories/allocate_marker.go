package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type AllocateMarkerRepo struct {
}

const allocateMarkerBulkUpsertStmtTpl = `INSERT INTO %s (%s)
VALUES %s ON CONFLICT ON CONSTRAINT pk_allocate_marker DO UPDATE 
SET
teacher_id = excluded.teacher_id, 
created_by = excluded.created_by,
updated_at = NOW();`

func (r *AllocateMarkerRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.AllocateMarker) error {
	err := dbeureka.BulkUpsert(ctx, db, allocateMarkerBulkUpsertStmtTpl, items)
	if err != nil {
		return fmt.Errorf("database.BulkUpsert error: %s", err.Error())
	}
	return nil
}

func (r *AllocateMarkerRepo) GetTeacherID(ctx context.Context, db database.QueryExecer, args *StudyPlanItemIdentity) (pgtype.Text, error) {
	ctx, span := interceptors.StartSpan(ctx, "AllocateMarkerRepo.GetTeacherID")
	defer span.End()

	var result pgtype.Text

	stmt := `
    SELECT teacher_id
      FROM allocate_marker
     WHERE deleted_at IS NULL
       AND student_id = $1::TEXT
       AND study_plan_id = $2::TEXT
       AND learning_material_id = $3::TEXT
	`

	if err := database.Select(ctx, db, stmt, args.StudentID, args.StudyPlanID, args.LearningMaterialID).ScanFields(&result); err != nil {
		return result, err
	}

	return result, nil
}

func (r *AllocateMarkerRepo) GetAllocateTeacherByCourseAccess(ctx context.Context, db database.QueryExecer, locationIds pgtype.TextArray) ([]*entities.AllocateTeacherItem, error) {
	stmt := `
	with tmp_markers as ( with tmp_users as ( select ugm.user_id from user_group_member ugm 
		join user_group ug using(user_group_id)
		join granted_role gr using (user_group_id)
		join role r using (role_id)
		where r.role_name = any(array['School Admin','Teacher','HQ Staff','Centre Manager','Centre Lead','Centre Staff']) 
		and gr.deleted_at is null 
		and ug.deleted_at is null
		and ugm.deleted_at is null
		and r.deleted_at is null group by ugm.user_id )
		select u.user_id,u.name from tmp_users join user_access_paths uap on uap.user_id = tmp_users.user_id
		join users u ON u.user_id = uap.user_id 
		where ($1::TEXT[] is null or uap.location_id = ANY($1::TEXT[])) and u.deleted_at is null and uap.deleted_at is null
		group by u.user_id,u.name)	
		select tmp_markers.user_id, tmp_markers.name, count(*) filter (where am.learning_material_id is not null) from allocate_marker am right join tmp_markers on tmp_markers.user_id = am.teacher_id 
		group by tmp_markers.user_id, tmp_markers.name
		order by tmp_markers.name
	`

	rows, err := db.Query(ctx, stmt, &locationIds)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	allocateTeachers := make([]*entities.AllocateTeacherItem, 0)
	for rows.Next() {
		item := &entities.AllocateTeacherItem{}
		if err := rows.Scan(&item.TeacherID, &item.TeacherName, &item.NumberAssignedSubmission); err != nil {
			return nil, fmt.Errorf("allocateTeachers.Scan: %w", err)
		}
		allocateTeachers = append(allocateTeachers, item)
	}

	return allocateTeachers, nil
}
