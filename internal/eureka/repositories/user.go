package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// this repo get from replicating table
type UserRepo struct{}

func (r *UserRepo) GetCountryByUserID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (string, error) {
	var country pgtype.Text
	query := "SELECT country FROM users WHERE user_id = $1"
	if err := db.QueryRow(ctx, query, &studentID).Scan(&country); err != nil {
		return "", fmt.Errorf("db.QueryRowEx: %w", err)
	}
	return country.String, nil
}

func (r *UserRepo) GetUsersByIDsAndName(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, name string, limit, offset uint32) ([]*entities.User, error) {
	user := &entities.User{}
	fields, _ := user.FieldMap()
	users := &entities.Users{}

	stmt := `SELECT %s FROM %s 
	WHERE user_id = ANY($1::_TEXT) 
	AND deleted_at IS NULL
	AND ($2::text IS NULL OR %s.name ILIKE '%%' || $2 || '%%')
	ORDER BY last_name ASC, first_name ASC
	LIMIT $3 OFFSET $4`
	stmt = fmt.Sprintf(stmt, strings.Join(fields, ", "), user.TableName(), user.TableName())

	if err := database.Select(ctx, db, stmt, ids, name, limit, offset).ScanAll(users); err != nil {
		return nil, err
	}

	return *users, nil
}

func (r *UserRepo) CountUsersByIDsAndName(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, name string) (int32, error) {
	user := &entities.User{}
	var totalUser int32

	stmt := `SELECT count(*) FROM %s WHERE user_id = ANY($1::_TEXT) AND deleted_at IS NULL
	AND ($2::text IS NULL OR %s.name ILIKE '%%' || $2 || '%%')`
	stmt = fmt.Sprintf(stmt, user.TableName(), user.TableName())

	if err := db.QueryRow(ctx, stmt, ids, name).Scan(&totalUser); err != nil {
		return 0, err
	}

	return totalUser, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities.User, error) {
	user := &entities.User{}
	fields, _ := user.FieldMap()

	stmt := `SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL`
	stmt = fmt.Sprintf(stmt, strings.Join(fields, ", "), user.TableName())

	if err := database.Select(ctx, db, stmt, userID).ScanOne(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) GetUsersByIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) ([]*entities.User, error) {
	user := &entities.User{}
	fields, _ := user.FieldMap()
	users := &entities.Users{}

	stmt := `SELECT %s FROM %s 
	WHERE user_id = ANY($1::_TEXT) 
	ORDER BY last_name ASC, first_name ASC`
	stmt = fmt.Sprintf(stmt, strings.Join(fields, ", "), user.TableName())

	if err := database.Select(ctx, db, stmt, userIDs).ScanAll(users); err != nil {
		return nil, err
	}

	return *users, nil
}
