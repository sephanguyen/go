package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
)

type TeacherRepo struct{}

func (r *TeacherRepo) ListByIDs(ctx context.Context, db database.QueryExecer, ids []string) (domain.Teachers, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.ListByIDs")
	defer span.End()

	t := &Teacher{}
	teacherFields := database.GetFieldNames(t)
	u := &UserBasicInfo{}
	userFields := database.GetFieldNames(u)

	selectFields := make([]string, 0, len(teacherFields)+len(userFields))
	for _, f := range teacherFields {
		selectFields = append(selectFields, t.TableName()+"."+f)
	}

	for _, f := range userFields {
		selectFields = append(selectFields, u.TableName()+"."+f)
	}

	selectStmt := fmt.Sprintf("SELECT %s FROM staff JOIN user_basic_info ON staff_id=user_id WHERE staff_id=ANY($1) AND staff.deleted_at IS NULL",
		strings.Join(selectFields, ","),
	)

	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := make(Teachers, 0, len(ids))
	for rows.Next() {
		t := Teacher{}
		scanFields := append(database.GetScanFields(&t, teacherFields), database.GetScanFields(&t.UserBasicInfo, userFields)...)
		if err = rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		teachers = append(teachers, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teachers.ToTeachersEntity(), nil
}

func (r *TeacherRepo) ListByGrantedLocation(ctx context.Context, db database.QueryExecer) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.ListByGrantedLocation")
	defer span.End()

	selectStmt := `select user_id,location_id from user_group_member ugm 
	join user_group ug on ug.user_group_id = ugm.user_group_id
	join granted_role gr on gr.user_group_id = ug.user_group_id 
	join granted_role_access_path grap on grap.granted_role_id = gr.granted_role_id  
	join role r on r.role_id = gr.role_id 
	where ugm.deleted_at is null 
	and ug.deleted_at is null 
	and gr.deleted_at is null 
	and grap.deleted_at is null  
	and r.deleted_at is null 
	and r.role_name = 'Teacher' `

	rows, err := db.Query(ctx, selectStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := make(map[string][]string, 0)
	for rows.Next() {
		var userID, locationID pgtype.Text
		if err = rows.Scan(&userID, &locationID); err != nil {
			return nil, err
		}
		teachers[userID.String] = append(teachers[userID.String], locationID.String)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}
