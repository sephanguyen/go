package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
)

type NotificationInternalUserRepo struct{}

func (r *NotificationInternalUserRepo) GetByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (*entities.NotificationInternalUser, error) {
	fields := strings.Join(database.GetFieldNames(&entities.NotificationInternalUser{}), ",")
	ent := &entities.NotificationInternalUser{}

	err := database.Select(ctx, db, fmt.Sprintf(`
		SELECT %s
		FROM notification_internal_user niu
		WHERE niu.resource_path = $1
			AND niu.is_system = true
			AND niu.deleted_at IS NULL
		LIMIT 1
	`, fields), database.Text(orgID)).ScanOne(ent)
	if err != nil {
		return nil, err
	}

	return ent, nil
}
