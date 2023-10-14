package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/pkg/errors"
)

type UsersRepo struct {
}

func (r *UsersRepo) FindByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (map[string]*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UsersRepo.GetUsers")
	defer span.End()

	m := &entities.User{}
	fields := database.GetFieldNames(m)

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE user_id = any($1) And deleted_at IS NULL`,
		strings.Join(fields, ","),
		m.TableName())

	rows, err := db.Query(ctx, query, database.TextArray(userIDs))
	if err != nil {
		return nil, fmt.Errorf("db.Query %w", err)
	}
	defer rows.Close()
	usersMap := map[string]*entities.User{}
	for rows.Next() {
		user := &entities.User{}
		if err := rows.Scan(database.GetScanFields(user, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}
		usersMap[user.UserID.String] = user
	}
	return usersMap, nil
}
