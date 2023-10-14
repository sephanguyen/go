package conversationmgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var rbacDecider = map[string][]string{
	// TODO: use Chat Supporter role instead
	"/conversationmgmt.v1.AgoraUserMgmtService/GetAppInfo": nil,

	// Internal Command gRPC
	"/conversationmgmt.v1.ConversationModifierService/CreateConversation":        nil,
	"/conversationmgmt.v1.ConversationModifierService/AddConversationMembers":    nil,
	"/conversationmgmt.v1.ConversationModifierService/RemoveConversationMembers": nil,
	"/conversationmgmt.v1.ConversationModifierService/UpdateConversationInfo":    nil,

	// External Command gRPC
	"/conversationmgmt.v1.ConversationModifierService/DeleteMessage": nil,

	// External Query gRPC
	"/conversationmgmt.v1.ConversationReaderService/GetConversationsDetail": nil,
}

func authInterceptor(c *configurations.Config, l *zap.Logger, db database.QueryExecer) *interceptors.Auth {
	groupDecider := &interceptors.GroupDecider{
		GroupFetcher: func(ctx context.Context, userID string) ([]string, error) {
			userRepo := &repository.UserRepo{}
			return interceptors.RetrieveUserRoles(ctx, userRepo, db)
		},
		AllowedGroups: rbacDecider,
	}

	auth, err := interceptors.NewAuth(
		ignoreAuthEndpoint,
		groupDecider,
		c.Issuers,
	)
	if err != nil {
		l.Panic("err init authInterceptor", zap.Error(err))
	}

	return auth
}
