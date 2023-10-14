package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ErrUserNotFound use when no user found when we should expecting one
var ErrUserNotFound = errors.New("user not found")

type UserFindFilter struct {
	IDs       pgtype.TextArray
	Email     pgtype.Text
	Phone     pgtype.Text
	UserGroup pgtype.Text
}

type UserRepo struct{}

func (u *UserRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*User, error) {
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

// Find returns nill if no filter provided
func (u *UserRepo) Find(ctx context.Context, db database.QueryExecer, filter *UserFindFilter, fields ...string) ([]*User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.Find")
	defer span.End()

	user := &User{}
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

	users := Users{}
	err := database.Select(ctx, db, query, args...).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return users, nil
}

// UserGroup returns empty string in case of error
func (u *UserRepo) UserGroup(ctx context.Context, db database.QueryExecer, id string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.UserGroup")
	defer span.End()
	return getCurrentUserGroup(ctx, db, id, u.Retrieve)
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

func (u *UserRepo) GetUsersByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (map[string]*domain.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUsersByIDs")
	defer span.End()

	ids := database.TextArray(userIDs)
	users, err := u.Retrieve(ctx, db, ids)
	if err != nil {
		return nil, err
	}

	usersConverted := make(map[string]*domain.User)
	for _, user := range users {
		usersConverted[user.ID.String] = user.ToUserDomain()
	}

	return usersConverted, nil
}
