package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type UserRepo struct{}

func (u *UserRepo) GetStaffsByLocationIDsAndNameOrEmail(ctx context.Context, db database.QueryExecer, locationIDs, filteredTeacherIDs []string, keyword string, limit int) ([]*dto.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetStaffsByLocations")
	defer span.End()
	whereClause := ""
	limitClause := ""
	arg := []interface{}{
		locationIDs,
	}
	countArg := 1
	if len(keyword) > 0 {
		whereClause = fmt.Sprintf(`%s and ( nospace(ubi."name") ilike nospace('%%%s%%') or ubi."email" ilike '%%%s%%' )`, whereClause, keyword, keyword)
	}
	if len(filteredTeacherIDs) > 0 {
		countArg++
		arg = append(arg, filteredTeacherIDs)
		whereClause = fmt.Sprintf(`%s and ubi.user_id <> all ($%d) `, whereClause, countArg)
	}
	if limit > 0 {
		countArg++
		arg = append(arg, limit)
		limitClause = fmt.Sprintf("limit $%d", countArg)
	}

	query := fmt.Sprintf(`with us as (select ubi.user_id, ubi."name", ubi.email from staff s 
						 join user_access_paths uap on s.staff_id = uap.user_id 
					 	 join user_basic_info ubi on ubi.user_id = s.staff_id
				 		 where uap.location_id = ANY($1)
						 %s
						 and s.deleted_at is NULL
						 and uap.deleted_at is NULL
						 and ubi.deleted_at is NULL ),
				  ugrl as (select ugm.user_id, gr.granted_role_id 
						   from granted_role gr join user_group_member ugm on ugm.user_group_id = gr.user_group_id
						   join role r on r.role_id = gr.role_id 
						   where gr.deleted_at is NULL and ugm.deleted_at is NULL and r.deleted_at is null and r.role_name = 'Teacher' ),
				  gra as (select grap.granted_role_id from granted_role_access_path grap join locations l on l.location_id = grap.location_id
						  where grap.location_id = ANY( SELECT regexp_split_to_table(access_path, '/') FROM locations WHERE location_id = ANY($1))
						  and grap.deleted_at is NULL )
			 select distinct us.user_id, us."name", us.email
		     from ( ugrl join gra on ugrl.granted_role_id = gra.granted_role_id ) join us on us.user_id = ugrl.user_id %s`, whereClause, limitClause)
	rows, err := db.Query(ctx, query, arg...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	var resp []*dto.User
	for rows.Next() {
		user := &User{}
		_, values := user.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		resp = append(resp, user.ConvertDTO())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	if len(resp) > limit && limit != 0 {
		return nil, fmt.Errorf("expect limit: %d but got %d in response", limit, len(resp))
	}
	return resp, nil
}

func (u *UserRepo) GetStaffsByLocationAndWorkingStatus(ctx context.Context, db database.QueryExecer, locationID string, workingStatus []string, useUserBasicInfoTable bool) ([]*dto.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetStaffsByLocations")
	defer span.End()
	table := "users"
	tablePrefix := "u"
	if useUserBasicInfoTable {
		table = "user_basic_info"
		tablePrefix = "ubi"
	}
	joinUserQuery := fmt.Sprintf("left join %s %s on %s.user_id = s.staff_id", table, tablePrefix, tablePrefix)
	query := fmt.Sprintf(`with us as (select %s.user_id, %s."name", %s.email from staff s 
						 left join user_access_paths uap on s.staff_id = uap.user_id 
				 		 %s
				 		 where uap.location_id = $1
				 		 and s.working_status = ANY($2)
						 and s.deleted_at is NULL
						 and uap.deleted_at is NULL
						 and %s.deleted_at is NULL ),
				  ugrl as (select ugm.user_id, gr.granted_role_id
						   from granted_role gr join user_group_member ugm on ugm.user_group_id = gr.user_group_id
						   join role r on r.role_id = gr.role_id 
						   where gr.deleted_at is NULL and ugm.deleted_at is NULL and r.deleted_at is null and r.role_name = 'Teacher' ),
				  gra as (select grap.granted_role_id from granted_role_access_path grap join locations l on l.location_id = grap.location_id
						  where grap.location_id = ANY( SELECT regexp_split_to_table(access_path, '/') FROM locations WHERE location_id = $1)
						  and grap.deleted_at is NULL )
			 select distinct us.user_id, us."name", us.email
		     from ( ugrl join gra on ugrl.granted_role_id = gra.granted_role_id ) join us on us.user_id = ugrl.user_id`, tablePrefix, tablePrefix, tablePrefix, joinUserQuery, tablePrefix)

	rows, err := db.Query(ctx, query, locationID, workingStatus)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	resp := []*dto.User{}
	for rows.Next() {
		user := &User{}
		_, values := user.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		resp = append(resp, user.ConvertDTO())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return resp, nil
}

func (u *UserRepo) GetStudentCurrentGradeByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string, useUserBasicInfoTable bool) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetStudentCurrentGradeByUserIDs")
	defer span.End()

	query := ""
	if useUserBasicInfoTable {
		query = `SELECT ubi.user_id, ubi.grade_id
				 FROM user_basic_info ubi
				 WHERE ubi.user_id = ANY($1)
				 AND ubi.deleted_at IS NULL`
	} else {
		query = `SELECT s.student_id,
						CASE 
							WHEN s.grade_id IS NOT NULL THEN g."name" 
							WHEN s.current_grade IS NOT NULL THEN s.current_grade::text
							ELSE NULL
						END as "student_grade"
			FROM students s
			LEFT JOIN grade g ON g.grade_id = s.grade_id
			WHERE s.student_id = ANY($1)
			AND g.deleted_at IS NULL
			AND s.deleted_at IS NULL`
	}

	rows, err := db.Query(ctx, query, userIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	studentGradeMap := make(map[string]string, len(userIDs))
	for rows.Next() {
		var (
			userID  pgtype.Text
			gradeID pgtype.Text
		)

		if err = rows.Scan(&userID, &gradeID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		studentGradeMap[userID.String] = gradeID.String
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return studentGradeMap, nil
}
