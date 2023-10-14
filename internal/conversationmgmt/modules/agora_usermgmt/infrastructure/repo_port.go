package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type AgoraUserRepo interface {
	Create(ctx context.Context, db database.QueryExecer, agoraUser *models.AgoraUser) error
	CreateAgoraUserFailure(ctx context.Context, db database.QueryExecer, agoraUserFailure *models.AgoraUserFailure) error
}

type UserBasicInfoRepo interface {
	GetUsers(ctx context.Context, db database.QueryExecer, userIDs []string) ([]*models.UserBasicInfo, error)
}
