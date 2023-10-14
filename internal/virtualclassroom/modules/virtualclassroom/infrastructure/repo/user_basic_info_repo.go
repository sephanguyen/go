package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type UserBasicInfoRepo struct{}

func (u *UserBasicInfoRepo) GetUserInfosByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) ([]domain.UserBasicInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserBasicInfoRepo.GetUserInfosByIDs")
	defer span.End()

	dto := &UserBasicInfo{}
	fields, values := dto.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE user_id = ANY($1)
		AND deleted_at IS NULL `,
		strings.Join(fields, ", "),
		dto.TableName(),
	)

	rows, err := db.Query(ctx, query, &userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user infos, db.Query: %w", err)
	}
	defer rows.Close()

	userInfos := []domain.UserBasicInfo{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		userInfos = append(userInfos, dto.ToUserBasicInfoDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return userInfos, nil
}
