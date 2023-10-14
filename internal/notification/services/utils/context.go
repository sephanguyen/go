package utils

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func GetNotificationInternalUserContext(ctx context.Context, orgID string, notiInternalUser *entities.NotificationInternalUser) context.Context {
	tenantWithInternalUserContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: orgID,
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       notiInternalUser.UserID.String,
		},
	})

	return tenantWithInternalUserContext
}
