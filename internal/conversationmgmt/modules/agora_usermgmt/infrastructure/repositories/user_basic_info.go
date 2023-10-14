package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type UserBasicInfoRepo struct{}

func (u *UserBasicInfoRepo) GetUsers(ctx context.Context, db database.QueryExecer, userIDs []string) ([]*models.UserBasicInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserBasicInfoRepo.GetUsers")
	defer span.End()

	fields := database.GetFieldNames(&models.UserBasicInfo{})
	query := fmt.Sprintf(`
		SELECT %s 
		FROM users 
		WHERE user_id = ANY($1) 	
			AND deleted_at IS NULL 
	`, strings.Join(fields, ", "))

	rows, err := db.Query(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userBasicInfo := make([]*models.UserBasicInfo, 0)
	for rows.Next() {
		u := new(models.UserBasicInfo)
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
