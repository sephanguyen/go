package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type UserRepo struct{}

func (r *UserRepo) GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer, roleNames []string, locationID string) (userIDs []string, err error) {
	stmt := `SELECT u.user_id  
		FROM users u
		JOIN user_group_member ugm ON u.user_id = ugm.user_id 
		JOIN user_group ug ON ug.user_group_id = ugm.user_group_id  
		JOIN granted_role gr ON gr.user_group_id = ug.user_group_id 
		JOIN "role" r ON gr.role_id = r.role_id 
		JOIN granted_role_access_path grap ON grap.granted_role_id  = gr.granted_role_id  
		JOIN locations l ON l.location_id  = grap.location_id  
		WHERE l.location_id = $1
			AND r.role_name = ANY($2)
			AND u.deleted_at IS NULL 
			AND ugm.deleted_at IS NULL 
			AND ug.deleted_at IS NULL 
			AND gr.deleted_at IS NULL 
			AND grap.deleted_at IS NULL 
			AND l.deleted_at IS NULL 
			AND r.deleted_at IS null;`

	rows, err := db.Query(
		ctx,
		stmt,
		locationID,
		database.TextArray(roleNames),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs = make([]string, 0)
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("row.Scan UserRepo.GetUserIDsForLoaNotification: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func (r *UserRepo) GetByID(ctx context.Context, db database.QueryExecer, userID string) (entities.User, error) {
	user := &entities.User{}
	userFieldNames, userFieldValues := user.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			user_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(userFieldNames, ","),
		user.TableName(),
	)
	row := db.QueryRow(ctx, stmt, userID)
	err := row.Scan(userFieldValues...)
	if err != nil {
		return entities.User{}, fmt.Errorf("row.Scan UserRepo.GetByID: %w", err)
	}
	return *user, nil
}
