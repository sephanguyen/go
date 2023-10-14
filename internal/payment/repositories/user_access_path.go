package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type UserAccessPathRepo struct{}

func (r *UserAccessPathRepo) GetUserAccessPathByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (mapUserAccess map[string]interface{}, err error) {
	var (
		valueMap interface{}
	)
	mapUserAccess = make(map[string]interface{})
	stmt :=
		`
		SELECT user_id,location_id
		FROM 
			%s
		WHERE 
			user_id = ANY($1) AND deleted_at IS NULL
		`
	stmt = fmt.Sprintf(
		stmt,
		(&entities.UserAccessPaths{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID, locationID string
		err = rows.Scan(&userID, &locationID)
		if err != nil {
			err = fmt.Errorf("row.Scan: %w", err)
			return
		}
		key := fmt.Sprintf("%v_%v", locationID, userID)
		mapUserAccess[key] = valueMap
	}
	return
}
