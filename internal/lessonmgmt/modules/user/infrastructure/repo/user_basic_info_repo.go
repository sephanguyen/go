package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type UserBasicInfoRepo struct{}

func (u *UserBasicInfoRepo) GetTeachersSameGrantedLocation(ctx context.Context, db database.QueryExecer, query domain.UserBasicInfoQuery) (domain.UsersBasicInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserBasicInfoRepo.GetTeachersSameGrantedLocation")
	defer span.End()
	queryLocation := ""
	paramsNum := 1
	args := []interface{}{}
	queryLocation = `and exists( 
						select 1 
						from granted_role_access_path grap
						where grap.deleted_at is null 
							and grap.granted_role_id = gr.granted_role_id 
							and exists (select 1 from lst_location ll where ll.child_location = grap.location_id  and ($1 = ''
							or ll.parent_location = $1))
						) `
	paramsNum++
	args = append(args, query.LocationID)

	stmtSelectLocation := `lst_location as (
		select
			   l.location_id as child_location, l1.location_id as parent_location
		   from
			   locations l
		   join locations l1 on
			   l1.access_path ~~ (l.access_path || '%'::text)
		   where
			l.deleted_at is null
		   and l1.deleted_at is null
	   )`

	selectStmtTeacher := fmt.Sprintf(`WITH %s, teacher AS ( select  user_id 
		from user_group_member ugm 
			join user_group ug on ug.user_group_id = ugm.user_group_id
			join granted_role gr on gr.user_group_id = ug.user_group_id
		where ugm.deleted_at is null
			and ug.deleted_at is null
			and gr.deleted_at is null
			and exists(
				select 1 from  role r
					where 
						r.role_id = gr.role_id
						and r.deleted_at is null
						and r.role_name = 'Teacher'
			)
			%s  
		)
		`, stmtSelectLocation, queryLocation)

	q := selectStmtTeacher + `SELECT ubi.user_id, ubi.name, ubi.first_name, ubi.last_name, ubi.full_name_phonetic, ubi.first_name_phonetic,
			ubi.last_name_phonetic, ubi.email, ubi.created_at, ubi.updated_at
			FROM user_basic_info ubi
			
			WHERE ubi.deleted_at IS null
			and exists (select 1 from teacher where teacher.user_id = ubi.user_id) 

	`
	if query.KeyWord != "" {
		q += fmt.Sprintf(` AND (lower(ubi."name") like lower(CONCAT('%%',$%d::text,'%%'))
					OR lower(ubi."full_name_phonetic") like lower(CONCAT('%%',$%d::text,'%%'))
					OR lower(ubi."email") like lower(CONCAT('%%',$%d::text,'%%'))
				)`, paramsNum, paramsNum, paramsNum)

		args = append(args, query.KeyWord)
		paramsNum++
	}

	q += fmt.Sprintf(" limit $%d offset $%d", paramsNum, paramsNum+1)
	args = append(args, query.Limit, query.Offset)
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []*domain.UserBasicInfo{}
	for rows.Next() {
		var (
			userID            pgtype.Text
			name              pgtype.Text
			firstName         pgtype.Text
			lastName          pgtype.Text
			fullNamePhonetic  pgtype.Text
			firstNamePhonetic pgtype.Text
			lastNamePhonetic  pgtype.Text
			email             pgtype.Text
			createdAt         pgtype.Timestamptz
			updatedAt         pgtype.Timestamptz
		)
		if err = rows.Scan(&userID, &name, &firstName,
			&lastName, &fullNamePhonetic, &firstNamePhonetic, &lastNamePhonetic,
			&email, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		rs := &domain.UserBasicInfo{
			UserID:            userID.String,
			FullName:          name.String,
			FirstName:         firstName.String,
			LastName:          lastName.String,
			FullNamePhonetic:  fullNamePhonetic.String,
			FirstNamePhonetic: firstNamePhonetic.String,
			LastNamePhonetic:  lastNamePhonetic.String,
			Email:             email.String,
			CreatedAt:         createdAt.Time,
			UpdatedAt:         updatedAt.Time,
		}
		results = append(results, rs)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (u *UserBasicInfoRepo) GetUser(ctx context.Context, db database.QueryExecer, userIDs []string) ([]*UserBasicInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.ListByGrantedLocation")
	defer span.End()

	fields := database.GetFieldNames(&UserBasicInfo{})
	selectStmt := fmt.Sprintf(`select %s from user_basic_info where user_id = ANY($1) and deleted_at is null `,
		strings.Join(fields, ", "))

	rows, err := db.Query(ctx, selectStmt, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userBasicInfo := make([]*UserBasicInfo, 0)
	for rows.Next() {
		u := new(UserBasicInfo)
		if err := rows.Scan(database.GetScanFields(u, fields)...); err != nil {
			return nil, err
		}
		userBasicInfo = append(userBasicInfo, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userBasicInfo, nil
}

func (u *UserBasicInfoRepo) UpdateEmailOfUsers(ctx context.Context, db database.QueryExecer, users domain.Users) error {
	ctx, span := interceptors.StartSpan(ctx, "UserBasicInfoRepo.UpdateEmailOfUsers")
	defer span.End()
	b := &pgx.Batch{}
	for _, user := range users {
		b.Queue(`UPDATE user_basic_info SET email = $2 WHERE user_id = $1`, user.ID, user.Email)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("user_basic_info is not update")
		}
	}
	return nil
}
