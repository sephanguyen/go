package spike

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/spike/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",

	"/spike.v1.EmailModifierService/SendEmail",
}

var fakeJwtCtxEndpoint = []string{
	"/spike.v1.EmailModifierService/SendEmail",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch": nil,
}

func fakeSchoolAdminJwtInterceptor() *gl_interceptors.FakeJwtContext {
	endpoints := map[string]struct{}{}
	for _, endpoint := range fakeJwtCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	return gl_interceptors.NewFakeJwtContext(endpoints, constant.UserGroupSchoolAdmin)
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
