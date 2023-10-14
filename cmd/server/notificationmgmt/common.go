package notificationmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	batchCount           = 100
	scanOrganiationQuery = `
		SELECT organization_id
		FROM organizations
	`
)

func getInfoNotificationWithOffset(ctx context.Context, db database.QueryExecer, offset int, orgID, conditionStr string) ([]*Notification, error) {
	query := `
		SELECT in2.notification_id, in2.target_groups, receiver_ids, generic_receiver_ids
		FROM info_notifications in2
		WHERE in2.deleted_at IS NULL AND resource_path = $3 %s 
		ORDER BY in2.notification_id DESC
		LIMIT $1
		OFFSET $2;
	`
	rows, err := db.Query(ctx, fmt.Sprintf(query, conditionStr), batchCount, offset, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Notification{}
	for rows.Next() {
		item := &Notification{}
		err = rows.Scan(&item.NotificationID, &item.TargetGroups, &item.ReceiverIDs, &item.GenericReceiverIDs)
		if err != nil {
			return nil, fmt.Errorf("failed scan %v", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func makeTenantWithUserCtx(ctx context.Context, db *pgxpool.Pool, organizationID string) (context.Context, error) {
	var err error
	tenanCtx := auth.InjectFakeJwtToken(ctx, organizationID)

	notificationInternalUserRepo := repositories.NotificationInternalUserRepo{}
	internalUser, err := notificationInternalUserRepo.GetByOrgID(tenanCtx, db, organizationID)
	internalUserID := ""
	if err != nil {
		err = fmt.Errorf("query internal user of tenant %v has err %v", organizationID, err)
	} else {
		internalUserID = internalUser.UserID.String
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			ResourcePath: organizationID,
			UserID:       internalUserID,
		},
	}
	tenantAndUserCtx := interceptors.ContextWithJWTClaims(ctx, claim)

	return tenantAndUserCtx, err
}
