package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
)

type UserRepo struct{}

type FindUserFilter struct {
	UserIDs pgtype.TextArray
}

func NewFindUserFilter() *FindUserFilter {
	f := &FindUserFilter{}
	_ = f.UserIDs.Set(nil)
	return f
}

func (repo *UserRepo) FindUser(ctx context.Context, db database.QueryExecer, filter *FindUserFilter) ([]*entities.User, map[string]*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.FindUser")
	defer span.End()

	if filter.UserIDs.Status != pgtype.Present {
		return nil, nil, nil
	}

	user := &entities.User{}
	fieldsName := database.GetFieldNames(user)
	query := fmt.Sprintf(`
		SELECT u.%s
		FROM users u
		WHERE u.user_id = ANY($1)
		AND deleted_at IS NULL;
	`, strings.Join(fieldsName, ",u."))

	rows, err := db.Query(ctx, query, filter.UserIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	mapUserIDAndUser := make(map[string]*entities.User)
	users := make(entities.Users, 0)
	for rows.Next() {
		e := &entities.User{}
		fields := database.GetScanFields(e, database.GetFieldNames(e))
		err := rows.Scan(fields...)
		if err != nil {
			return nil, nil, err
		}
		mapUserIDAndUser[e.UserID.String] = e
		users = append(users, e)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return users, mapUserIDAndUser, nil
}
