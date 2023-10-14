package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ErrUserNotFound use when no user found when we should expecting one
var ErrUserNotFound = errors.New("user not found")

// UserRepo stores
type UserRepo struct {
}

// UserFindFilter for filtering users in DB
type UserFindFilter struct {
	IDs       pgtype.TextArray
	Email     pgtype.Text
	Phone     pgtype.Text
	UserGroup pgtype.Text
}

// Find returns nill if no filter provided
func (repo *UserRepo) Find(ctx context.Context, db database.QueryExecer, filter *UserFindFilter, fields ...string) ([]*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Retrieve")
	defer span.End()

	user := &entity.LegacyUser{}
	if len(fields) == 0 {
		fields = database.GetFieldNames(user)
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
		"AND ($4::text IS NULL OR user_group = $4)", strings.Join(fields, ","), user.TableName())

	users := entity.LegacyUsers{}
	err := database.Select(ctx, db, query, args...).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

// Retrieve returns user list by ids. If not specific fields, return all fields
func (repo *UserRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entity.LegacyUser, error) {
	filter := &UserFindFilter{}
	if err := multierr.Combine(
		filter.UserGroup.Set(nil),
		filter.Email.Set(nil),
		filter.Phone.Set(nil),
		filter.IDs.Set(nil),
	); err != nil {
		return nil, err
	}
	filter.IDs = ids
	return repo.Find(ctx, db, filter, fields...)
}

// UserGroup returns empty string in case of error
func (repo *UserRepo) UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UserGroup")
	defer span.End()
	return getCurrentUserGroup(ctx, db, id, repo.Retrieve)
}

// ResourcePath returns the user's resource path (empty string in case of error)
func (repo *UserRepo) ResourcePath(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.ResourcePath")
	defer span.End()
	return getUserResourcePath(ctx, db, id, repo.Retrieve)
}

func getCurrentUserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text, retrieveFn func(context.Context, database.QueryExecer, pgtype.TextArray, ...string) ([]*entity.LegacyUser, error)) (string, error) {
	users, err := retrieveFn(ctx, db, database.TextArray([]string{id.String}), "user_group")
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

func getUserResourcePath(ctx context.Context, db database.QueryExecer, id pgtype.Text, retrieveFn func(context.Context, database.QueryExecer, pgtype.TextArray, ...string) ([]*entity.LegacyUser, error)) (string, error) {
	users, err := retrieveFn(ctx, db, database.TextArray([]string{id.String}), "resource_path")
	if err != nil {
		return "", errors.Wrap(err, "getUserResourcePath.retrieveFn")
	}

	if len(users) == 0 {
		return "", ErrUserNotFound
	}

	if len(users) != 1 {
		return "", errors.Errorf("expecting only 1 user returned, got %d", len(users))
	}

	return users[0].ResourcePath.String, nil
}

// UpdateLastLoginDate updates user "last_login_date" only
func (repo *UserRepo) UpdateLastLoginDate(ctx context.Context, db database.QueryExecer, user *entity.LegacyUser) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateLastLoginDate")
	defer span.End()

	fields := []string{
		"last_login_date",
	}

	cmdTag, err := database.UpdateFields(ctx, user, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update user last_login_date")
	}
	return nil
}

// Get similar to retrieve, get all fields of single user
func (repo *UserRepo) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error) {
	users, err := repo.Retrieve(ctx, db, database.TextArray([]string{id.String}))
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, pgx.ErrNoRows
	}

	return users[0], nil
}

func (repo *UserRepo) GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetByEmail")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	user := &entity.LegacyUser{}
	fields, _ := user.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (email = ANY($1)) AND resource_path = $2",
		strings.Join(fields, ","), user.TableName())

	users := entity.LegacyUsers{}
	err := database.Select(ctx, db, query, &emails, &resourcePath).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (repo *UserRepo) GetByEmailInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) ([]*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetByEmailInsensitiveCase")
	defer span.End()

	user := &entity.LegacyUser{}
	fields, _ := user.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (email = ANY($1))",
		strings.Join(fields, ","), user.TableName())

	// lowerCaseEmails := make([]string, 0, len(emails))
	// for _, email := range emails {
	// 	lowerCaseEmails = append(lowerCaseEmails, strings.ToLower(email))
	// }
	users := entity.LegacyUsers{}
	err := database.Select(ctx, db, query, database.TextArray(emails)).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (repo *UserRepo) GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetByEmail")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	user := &entity.LegacyUser{}
	fields, _ := user.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (phone_number = ANY($1)) AND resource_path = $2",
		strings.Join(fields, ","), user.TableName())

	users := entity.LegacyUsers{}
	err := database.Select(ctx, db, query, &phones, &resourcePath).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (repo *UserRepo) Create(ctx context.Context, db database.QueryExecer, user *entity.LegacyUser) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		user.UpdatedAt.Set(now),
		user.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if user.ResourcePath.Status == pgtype.Null {
		err := user.ResourcePath.Set(resourcePath)
		if err != nil {
			return err
		}
	}
	cmdTag, err := database.Insert(ctx, user, db.Exec)
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("user not inserted: %w", err)
	}

	return nil
}

func (repo *UserRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, user *entity.LegacyUser) {
		fields, values := user.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := "INSERT INTO " + user.TableName() + " (" + strings.Join(fields, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, values...)
	}

	b := &pgx.Batch{}
	now := time.Now()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, user := range users {
		_ = user.UpdatedAt.Set(now)
		_ = user.CreatedAt.Set(now)
		if user.ResourcePath.Status == pgtype.Null {
			if err := user.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		queueFn(b, user)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(users); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("user not inserted")
		}
	}

	return nil
}

// UpdateEmail updates user "email" only (with updated_at)
func (repo *UserRepo) UpdateEmail(ctx context.Context, db database.QueryExecer, user *entity.LegacyUser) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateEmail")
	defer span.End()

	fields := []string{
		"email",
		"login_email",
		"updated_at",
	}

	err := user.UpdatedAt.Set(time.Now())
	if err != nil {
		return err
	}

	cmdTag, err := database.UpdateFields(ctx, user, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update user email")
	}
	return nil
}

const listStmtTpl = `
	SELECT DISTINCT %s
	FROM %s
	LEFT JOIN user_access_paths AS uap
	USING(user_id)
	WHERE %[2]s.deleted_at is NULL AND user_id = ANY($1::_TEXT)
	AND (($2::TEXT IS NULL) OR (name LIKE $2 OR given_name LIKE $2 OR full_name_phonetic LIKE $2))
	AND ((ARRAY_LENGTH($3::TEXT[], 1) IS NULL) OR (location_id = ANY($3::TEXT[])))
	ORDER BY name LIMIT $4 OFFSET $5;
`

type SearchProfileFilter struct {
	Limit         uint
	OffsetInteger uint

	StudentIDs  pgtype.TextArray
	StudentName pgtype.Text
	LocationIDs pgtype.TextArray
}

// SearchProfile return array basic profile according StudentIds
func (repo *UserRepo) SearchProfile(ctx context.Context, db database.QueryExecer, filter *SearchProfileFilter) ([]*entity.LegacyUser, error) {
	e := &entity.LegacyUser{}
	result := make(entity.LegacyUsers, 0, len(filter.StudentIDs.Elements))
	fields, _ := e.FieldMap()

	// This code is for preventing ambiguous error in Postgres
	for i := 0; i < len(fields); i++ {
		fields[i] = fmt.Sprintf(`%s.%s`, e.TableName(), fields[i])
	}

	query := fmt.Sprintf(listStmtTpl, strings.Join(fields, ","), e.TableName())
	err := database.Select(ctx, db, query, filter.StudentIDs, filter.StudentName, filter.LocationIDs, filter.Limit, filter.OffsetInteger).ScanAll(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateProfileV1 updates "name", "avatar", "phone_number" "user_group", "birthday", "gender" "updated_at" only
func (repo *UserRepo) UpdateProfileV1(ctx context.Context, db database.QueryExecer, user *entity.LegacyUser) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateProfileV1")
	defer span.End()

	fields := []string{
		// "country",
		"username",
		"name",
		"avatar",
		"user_group",
	}

	err := user.UpdatedAt.Set(time.Now())
	if err != nil {
		return err
	}

	if user.PhoneNumber.Get() != nil {
		fields = append(fields, "phone_number")
	}

	if (user.Birthday.Time != time.Time{}) {
		fields = append(fields, "birthday")
	}

	if user.Gender.Status == pgtype.Present {
		fields = append(fields, "gender")
	}

	if user.Remarks.Status == pgtype.Present {
		fields = append(fields, "remarks")
	}

	if user.LastName.Status == pgtype.Present {
		fields = append(fields, "last_name")
	}

	if user.FirstName.Status == pgtype.Present {
		fields = append(fields, "first_name")
	}

	if user.FirstNamePhonetic.Status == pgtype.Present {
		fields = append(fields, "first_name_phonetic")
	}

	if user.LastNamePhonetic.Status == pgtype.Present {
		fields = append(fields, "last_name_phonetic")
	}

	if user.FullNamePhonetic.Status == pgtype.Present {
		fields = append(fields, "full_name_phonetic")
	}

	if user.ExternalUserID.Status == pgtype.Present {
		fields = append(fields, "user_external_id")
	}

	fields = append(fields, "updated_at")

	cmdTag, err := database.UpdateFields(ctx, user, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update profile")
	}
	return nil
}

func (repo *UserRepo) GetUserGroups(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserGroupV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUserGroups")
	defer span.End()

	userGroup := &entity.UserGroupV2{}
	userGroupMem := &entity.UserGroupMember{}
	fields, _ := userGroup.FieldMap()
	// adjust fields avoid ambiguous names in query
	for index := range fields {
		fields[index] = fmt.Sprintf(`ug.%s`, fields[index])
	}

	query := fmt.Sprintf(
		`
			SELECT
				%s
			FROM
				%s ug
			LEFT JOIN
				%s ugm ON
					ug.user_group_id = ugm.user_group_id AND
					ugm.deleted_at IS NULL
			WHERE
				ugm.user_id = $1 AND
				ug.deleted_at IS NULL;
		`,
		strings.Join(fields, ","),
		userGroup.TableName(),
		userGroupMem.TableName(),
	)

	userGroups := entity.UserGroupV2s{}
	err := database.Select(ctx, db, query, &userID).ScanAll(&userGroups)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return userGroups, nil
}

func (repo *UserRepo) GetUserRoles(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (entity.Roles, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUserRoles")
	defer span.End()

	fields, _ := (&entity.Role{}).FieldMap()

	stmt := fmt.Sprintf(
		`SELECT r.%s FROM user_group_member ugm
			INNER JOIN granted_role gt ON ugm.user_group_id = gt.user_group_id
			INNER JOIN role r ON gt.role_id = r.role_id
		WHERE ugm.user_id = $1
			AND gt.deleted_at IS NULL
			AND ugm.deleted_at IS NULL
			AND r.deleted_at IS NULL`, strings.Join(fields, ", r."))

	roles := entity.Roles{}
	if err := database.Select(ctx, db, stmt, &userID).ScanAll(&roles); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return roles, nil
}

func (repo *UserRepo) GetUserGroupMembers(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserGroupMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUserGroupMembers")
	defer span.End()

	fields, _ := (&entity.UserGroupMember{}).FieldMap()
	stmt := fmt.Sprintf(`SELECT ugm.%s FROM user_group_member ugm WHERE ugm.user_id = $1 AND ugm.deleted_at IS NULL`, strings.Join(fields, ", ugm."))

	userGroupMembers := entity.UserGroupMembers{}
	if err := database.Select(ctx, db, stmt, &userID).ScanAll(&userGroupMembers); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return userGroupMembers, nil
}

func (repo *UserRepo) FindByIDUnscope(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.FindByIDUnscope")
	defer span.End()

	user := &entity.LegacyUser{}
	fields, _ := user.FieldMap()

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1",
		strings.Join(fields, ","), user.TableName())

	err := database.Select(ctx, db, sql, &id).ScanOne(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *UserRepo) SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.SoftDelete")
	defer span.End()

	sql := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE user_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &ids)
	if err != nil {
		return err
	}

	return nil
}

func (repo *UserRepo) GetUsersByUserGroupID(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUsersByUserGroupID")
	defer span.End()

	fields, _ := (&entity.LegacyUser{}).FieldMap()
	stmt := fmt.Sprintf(
		`
		SELECT u.%s
		FROM users u
		INNER JOIN user_group_member ugm
			ON ugm.user_id = u.user_id AND
			   ugm.deleted_at IS NULL
		WHERE
			ugm.user_group_id = $1 AND
			u.deleted_at IS NULL
		`,
		strings.Join(fields, ", u."),
	)

	users := entity.LegacyUsers{}
	if err := database.Select(ctx, db, stmt, &userGroupID).ScanAll(&users); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (repo *UserRepo) UpdateManyUserGroup(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, userGroup pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateManyUserGroup")
	defer span.End()

	sql := `
		UPDATE users
		SET
			user_group = $1
		WHERE
			user_id = ANY($2) AND
			deleted_at IS NULL
	`
	if _, err := db.Exec(ctx, sql, userGroup, userIDs); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

// GetBasicInfo return user entity with some special columns for rbac: user_id, user_group
// Please don't use and modify this function for business logic
func (repo *UserRepo) GetBasicInfo(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetBasicInfo")
	defer span.End()

	user := &entity.LegacyUser{}
	stmt := fmt.Sprintf(`
		SELECT 
			user_id, 
			user_group
		FROM %s
		WHERE user_id = $1
	`, user.TableName())

	var (
		userID, userGroup string
	)

	if err := db.QueryRow(ctx, stmt, &id).Scan(&userID, &userGroup); err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	user.UserID = database.Text(userID)
	user.Group = database.Text(userGroup)

	return user, nil
}
