package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/caching"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ErrUserNotFound use when no user found when we should expecting one
var ErrUserNotFound = errors.New("user not found")

// UserRepo stores
type UserRepo struct{}

// UserFindFilter for filtering users in DB
type UserFindFilter struct {
	IDs       pgtype.TextArray
	Email     pgtype.Text
	Phone     pgtype.Text
	UserGroup pgtype.Text
}

// Find returns nill if no filter provided
func (u *UserRepo) Find(ctx context.Context, db database.QueryExecer, filter *UserFindFilter, fields ...string) ([]*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Retrieve")
	defer span.End()

	userDTO := &entities.User{}
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

	users := entities.Users{}
	err := database.Select(ctx, db, query, args...).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

// Retrieve returns user list by ids. If not specific fields, return all fields
func (u *UserRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error) {
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
	return u.Find(ctx, db, filter, fields...)
}

// UserGroup returns empty string in case of error
func (u *UserRepo) UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UserGroup")
	defer span.End()
	return getCurrentUserGroup(ctx, db, id, u.Retrieve)
}

// ResourcePath returns the user's resource path (empty string in case of error)
func (u *UserRepo) ResourcePath(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.ResourcePath")
	defer span.End()
	return getUserResourcePath(ctx, db, id, u.Retrieve)
}

func getCurrentUserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text, retrieveFn func(context.Context, database.QueryExecer, pgtype.TextArray, ...string) ([]*entities.User, error)) (string, error) {
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

func getUserResourcePath(ctx context.Context, db database.QueryExecer, id pgtype.Text, retrieveFn func(context.Context, database.QueryExecer, pgtype.TextArray, ...string) ([]*entities.User, error)) (string, error) {
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

// StoreDeviceToken is use to update device token of user
func (u *UserRepo) StoreDeviceToken(ctx context.Context, db database.QueryExecer, userDTO *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.StoreDeviceToken")
	defer span.End()

	sql := "UPDATE users SET device_token=$1, allow_notification=$2 WHERE user_id=$3;"
	_, err := db.Exec(ctx, sql, &userDTO.DeviceToken, &userDTO.AllowNotification, &userDTO.ID)
	if err != nil {
		return errors.Wrap(err, "update token failed")
	}
	return nil
}

// UpdateProfile updates "country", "name", "avatar", "phone_number", "user_group", "updated_at" only
func (u *UserRepo) UpdateProfile(ctx context.Context, db database.QueryExecer, userDTO *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateProfile")
	defer span.End()
	fields := []string{"country", "name", "avatar", "phone_number", "user_group", "updated_at"}

	cmdTag, err := database.UpdateFields(ctx, userDTO, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update profile")
	}
	return nil
}

// UpdateProfileV1 updates "name", "avatar", "phone_number" "user_group", "updated_at" only
func (u *UserRepo) UpdateProfileV1(ctx context.Context, db database.QueryExecer, userDTO *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateProfileV1")
	defer span.End()

	fields := []string{
		// "country",
		"name",
		"avatar",
		"user_group",
		"updated_at",
	}
	if userDTO.PhoneNumber.Get() != nil {
		fields = append(fields, "phone_number")
	}

	cmdTag, err := database.UpdateFields(ctx, userDTO, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update profile")
	}
	return nil
}

// UpdateEmail updates user "email" only
func (u *UserRepo) UpdateEmail(ctx context.Context, db database.QueryExecer, userDTO *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateEmail")
	defer span.End()

	fields := []string{
		"email",
	}

	cmdTag, err := database.UpdateFields(ctx, userDTO, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update profile")
	}
	return nil
}

// UpdateLastLoginDate updates user "last_login_date" only
func (u *UserRepo) UpdateLastLoginDate(ctx context.Context, db database.QueryExecer, userDTO *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UpdateLastLoginDate")
	defer span.End()

	fields := []string{
		"last_login_date",
	}

	cmdTag, err := database.UpdateFields(ctx, userDTO, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update user last_login_date")
	}
	return nil
}

// Get similar to retrieve, get all fields of single user
func (u *UserRepo) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
	users, err := u.Retrieve(ctx, db, database.TextArray([]string{id.String}))
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, pgx.ErrNoRows
	}

	return users[0], nil
}

func (u *UserRepo) GetUsernameByUserID(ctx context.Context, db database.QueryExecer, userID string) (*entities.Username, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUsernameByUserID")
	defer span.End()

	usernameDTO := &entities.Username{}
	fields, _ := usernameDTO.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE user_id=$1",
		strings.Join(fields, ","), usernameDTO.TableName())

	err := database.Select(ctx, db, query, &userID).ScanOne(usernameDTO)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return usernameDTO, nil
}

func (u *UserRepo) GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetByEmail")
	defer span.End()

	userDTO := &entities.User{}
	fields, _ := userDTO.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (email = ANY($1))",
		strings.Join(fields, ","), userDTO.TableName())

	users := entities.Users{}
	err := database.Select(ctx, db, query, &emails).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (u *UserRepo) GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetByEmail")
	defer span.End()

	userDTO := &entities.User{}
	fields, _ := userDTO.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (phone_number = ANY($1))",
		strings.Join(fields, ","), userDTO.TableName())

	users := entities.Users{}
	err := database.Select(ctx, db, query, &phones).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

func (u *UserRepo) Create(ctx context.Context, db database.QueryExecer, userDTO *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Create")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		userDTO.UpdatedAt.Set(now),
		userDTO.CreatedAt.Set(now),
	)

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if userDTO.ResourcePath.Status == pgtype.Null {
		userDTO.ResourcePath.Set(resourcePath)
	}
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	_, err = database.Insert(ctx, userDTO, db.Exec)

	return err
}

func (u *UserRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.User) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fields, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, values...)
	}

	b := &pgx.Batch{}
	now := time.Now()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, u := range users {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		if u.ResourcePath.Status == pgtype.Null {
			u.ResourcePath.Set(resourcePath)
		}

		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(users); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("user not inserted")
		}
	}

	return nil
}

func (u *UserRepo) FindUsersMatchConditions(ctx context.Context, db database.Ext, conditions entities.TargetConditions) ([]*entities.User, error) {
	var users entities.Users

	user := &entities.User{}
	fieldName := database.GetFieldNames(user)
	var userGroup, country, platform, subscription pgtype.TextArray

	multierr.Combine(
		userGroup.Set(nil),
		country.Set(nil),
		platform.Set(nil),
		subscription.Set(nil),
	)
	if len(conditions.UserGroup) > 0 {
		userGroup.Set(conditions.UserGroup)
	}

	if len(conditions.Country) > 0 {
		country.Set(conditions.Country)
	}

	if len(conditions.Platform) > 0 {
		platform.Set(conditions.Platform)
	}

	joinQuery := ""
	appendQuery := ""
	if conditions.IsTester {
		appendQuery += "AND is_tester = true"
	}

	stmt := `SELECT %s FROM %s %s WHERE 
		($1::text[] IS NULL OR user_group = ANY($1)) 
		AND ($2::text[] IS NULL OR country = ANY($2))
		AND ($3::text[] IS NULL OR platform = ANY($3))
		%s`

	if len(conditions.Subscription) > 0 {
		subscription.Set(conditions.Subscription)
		joinQuery = "JOIN student_subscription AS ss ON users.user_id = student_subscription.student_id"
		appendQuery += " AND ($4 IS NULL OR ss.plan_id = ANY($4)"
		query := fmt.Sprintf(stmt, strings.Join(fieldName, " ,"), user.TableName(), joinQuery, appendQuery)
		err := database.Select(ctx, db, query, userGroup, country, platform, subscription).ScanAll(&users)
		if err != nil {
			return nil, err
		}
	} else {
		query := fmt.Sprintf(stmt, strings.Join(fieldName, " ,"), user.TableName(), joinQuery, appendQuery)
		err := database.Select(ctx, db, query, userGroup, country, platform).ScanAll(&users)
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

// UserRepository interface implemented by both cache and DB
type UserRepository interface {
	Find(ctx context.Context, db database.QueryExecer, filter *UserFindFilter, fields ...string) ([]*entities.User, error)
	Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
	UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error)
	ResourcePath(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error)
	StoreDeviceToken(ctx context.Context, db database.QueryExecer, u *entities.User) error
	UpdateProfile(ctx context.Context, db database.QueryExecer, u *entities.User) error
	UpdateProfileV1(ctx context.Context, db database.QueryExecer, u *entities.User) error
	UpdateEmail(ctx context.Context, db database.QueryExecer, u *entities.User) error
	UpdateLastLoginDate(ctx context.Context, db database.QueryExecer, u *entities.User) error
	Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error)
	GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entities.User, error)
	GetByPhone(ctx context.Context, db database.QueryExecer, phones pgtype.TextArray) ([]*entities.User, error)
	Create(ctx context.Context, db database.QueryExecer, u *entities.User) error
	CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entities.User) error
}

// UserRepoWrapper decides to use cache or not
type UserRepoWrapper struct {
	LocalCacher caching.LocalCacher
	UserRepository
}

// UserGroup use cache wrapper
func (c *UserRepoWrapper) UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error) {
	if caching.IsNoCache(ctx) {
		return c.UserRepository.UserGroup(ctx, db, id)
	}

	v, found := c.LocalCacher.Get(ctx, "profile", id.String)
	if found {
		return v.(string), nil
	}

	group, err := c.UserRepository.UserGroup(ctx, db, id)
	if err != nil {
		return "", err
	}
	c.LocalCacher.Set(ctx, "profile", id.String, group, 30*time.Minute)
	return group, nil
}

func (u *UserRepo) SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	sql := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE user_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &ids)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (u *UserRepo) FindByIDUnscope(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error) {
	userDTO := &entities.User{}
	fields, _ := userDTO.FieldMap()

	sql := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE user_id = $1",
		strings.Join(fields, ","), userDTO.TableName())

	err := database.Select(ctx, db, sql, &id).ScanOne(userDTO)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return userDTO, nil
}

func (u *UserRepo) Update(ctx context.Context, db database.QueryExecer, s *entities.User) error {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Update")
	defer span.End()

	now := time.Now()
	s.UpdatedAt.Set(now)

	cmdTag, err := database.Update(ctx, s, db.Exec, "user_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update user")
	}

	return nil
}

const listStmtTpl = `SELECT %s 
FROM %s WHERE deleted_at is NULL AND user_id = ANY($1::_TEXT)  
	AND (($2::TEXT IS NULL) OR (name LIKE $2 OR given_name LIKE $2))
ORDER BY name LIMIT $3 OFFSET $4`

type SearchProfileFilter struct {
	Limit         uint
	OffsetInteger uint

	StudentIDs  pgtype.TextArray
	StudentName pgtype.Text
}

// SearchProfile return array basic profile according StudentIds
func (u *UserRepo) SearchProfile(ctx context.Context, db database.QueryExecer, filter *SearchProfileFilter) ([]*entities.User, error) {
	e := &entities.User{}
	result := make(entities.Users, 0, len(filter.StudentIDs.Elements))
	fields, _ := e.FieldMap()
	err := database.Select(ctx, db, fmt.Sprintf(listStmtTpl, strings.Join(fields, ","), e.TableName()), filter.StudentIDs, filter.StudentName, filter.Limit, filter.OffsetInteger).ScanAll(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
