package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/constants"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct{}

// GetUserGroupByUserID returns empty string in case of error
func (u *UserRepo) GetUserGroupByUserID(ctx context.Context, db database.QueryExecer, id string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUserGroupByUserID")
	defer span.End()
	return getCurrentUserGroup(ctx, db, id, u.retrieve)
}

func getCurrentUserGroup(ctx context.Context, db database.QueryExecer, id string, retrieveFn func(context.Context, database.QueryExecer, pgtype.TextArray, ...string) ([]*User, error)) (string, error) {
	users, err := retrieveFn(ctx, db, database.TextArray([]string{id}), "user_group")
	if err != nil {
		return "", errors.Wrap(err, "getCurrentUserGroup.retrieveFn")
	}

	if len(users) == 0 {
		return "", ErrUserNotFound
	}

	if len(users) != 1 {
		return "", errors.Errorf("expecting only 1 user returned, got %d", len(users))
	}

	return users[0].Group.String, nil
}

// retrieve returns user list by ids. If not specific fields, return all fields
func (u *UserRepo) retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*User, error) {
	filter := &UserFindFilter{}
	err := multierr.Combine(
		filter.UserGroup.Set(nil),
		filter.Email.Set(nil),
		filter.Phone.Set(nil),
		filter.IDs.Set(nil),
	)
	if err != nil {
		return nil, err
	}
	filter.IDs = ids
	return u.find(ctx, db, filter, fields...)
}

func (u *UserRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*User, error) {
	return u.retrieve(ctx, db, ids, fields...)
}

// find returns null if no filter provided
func (u *UserRepo) find(ctx context.Context, db database.QueryExecer, filter *UserFindFilter, fields ...string) ([]*User, error) {
	userDTO := &User{}
	if len(fields) == 0 {
		fields = database.GetFieldNames(userDTO)
	}

	var args []interface{}
	args = append(args, &filter.IDs.Elements)
	args = append(args, &filter.Email)
	args = append(args, &filter.Phone)
	args = append(args, &filter.UserGroup)
	query := fmt.Sprintf("SELECT %s FROM %s "+
		"WHERE deleted_at IS NULL AND ($1::text[] IS NULL OR user_id = ANY($1)) "+
		"AND ($2::text IS NULL OR email = $2) "+
		"AND($3::text IS NULL OR phone_number = $3) "+
		"AND ($4::text IS NULL OR user_group = $4)", strings.Join(fields, ","), userDTO.TableName())

	users := Users{}
	err := database.Select(ctx, db, query, args...).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (u *UserRepo) GetStudentCurrentGradeByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetStudentCurrentGradeByUserIDs")
	defer span.End()

	query := `SELECT ubi.user_id, ubi.grade_id
				FROM user_basic_info ubi
				WHERE ubi.user_id = ANY($1)
				AND ubi.deleted_at IS NULL`

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

		if err := rows.Scan(&userID, &gradeID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		studentGradeMap[userID.String] = gradeID.String
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return studentGradeMap, nil
}

func (u *UserRepo) GetUserByUserID(ctx context.Context, db database.QueryExecer, userID string) (*domain.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUserByUserID")
	defer span.End()

	userDTO := &User{}
	fields := database.GetFieldNames(userDTO)

	query := fmt.Sprintf(`
	SELECT u.%s FROM %s AS u
	WHERE u.user_id = $1
	AND u.deleted_at IS NULL`, strings.Join(fields, ","), userDTO.TableName())

	err := database.Select(ctx, db, query, &userID).ScanOne(userDTO)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	userDomain := userDTO.ToUserEntity()
	return userDomain, nil
}

func (u *UserRepo) FindByResourcePath(ctx context.Context, db database.QueryExecer, resourcePath string, limit int, offSet int) (domain.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.FindByResourcePath")
	defer span.End()

	userDTO := &User{}
	values := Users{}

	fields := database.GetFieldNames(userDTO)

	query := fmt.Sprintf(`
	SELECT u.%s FROM %s AS u
	WHERE u.deleted_at IS NULL
	AND u.resource_path = $1
	ORDER BY u.user_id
	LIMIT $2 OFFSET $3`, strings.Join(fields, ","), userDTO.TableName())

	err := database.Select(ctx, db, query, &resourcePath, &limit, &offSet).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	usersDTO := make(domain.Users, 0, len(values))
	for _, v := range values {
		userDTO := v.ToUserEntity()
		usersDTO = append(usersDTO, userDTO)
	}
	return usersDTO, nil
}

func (u *UserRepo) GetStudentsManyReferenceByNameOrEmail(ctx context.Context, db database.QueryExecer, keyword string, limit, offset uint32) (domain.Students, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetStudentsManyReferenceByNameOrEmail")
	defer span.End()
	whereClause := fmt.Sprintf("where ubi.deleted_at is null and ubi.user_role = '%s'", constants.UserRoleStudent)
	if len(keyword) != 0 {
		whereClause = fmt.Sprintf(`%s and ( nospace(ubi."name") ilike nospace('%%%s%%') or ubi."email" ilike '%%%s%%' )`, whereClause, keyword, keyword)
	}
	query := fmt.Sprintf("select ubi.user_id, ubi.name, ubi.email from user_basic_info ubi %s order by ubi.created_at desc limit $1 offset $2", whereClause)
	rows, err := db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	defer rows.Close()
	res := domain.Students{}
	for rows.Next() {
		student := &Student{}
		_, v := student.FieldMap()

		if err = rows.Scan(v...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		res = append(res, student.ToStudentEntity())
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return res, nil
}
