package fatima

import (
	"context"

	"github.com/manabie-com/backend/internal/fatima/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var rbacDecider = map[string][]string{
	"/fatima.v1.AccessibilityReadService/RetrieveAccessibility":        nil,
	"/fatima.v1.AccessibilityReadService/RetrieveStudentAccessibility": {constant.RoleSchoolAdmin},

	// "/fatima.v1.SubscriptionModifierService/CreatePackage": nil,
	// "/fatima.v1.SubscriptionModifierService/ToggleActivePackage":        nil,
	"/fatima.v1.SubscriptionModifierService/AddStudentPackage":                  nil,
	"/fatima.v1.SubscriptionModifierService/AddStudentPackageCourse":            nil,
	"/fatima.v1.SubscriptionModifierService/EditTimeStudentPackage":             nil,
	"/fatima.v1.SubscriptionModifierService/ToggleActiveStudentPackage":         nil,
	"/fatima.v1.SubscriptionModifierService/ListStudentPackage":                 nil,
	"/fatima.v1.SubscriptionModifierService/ListStudentPackageV2":               nil,
	"/fatima.v1.SubscriptionModifierService/RegisterStudentClass":               nil,
	"/fatima.v1.SubscriptionModifierService/RetrieveStudentPackagesUnderCourse": nil,

	"/fatima.v1.CourseReaderService/ListStudentByCourse": nil,

	"/grpc.health.v1.Health/Check": nil,
	"/grpc.health.v1.Health/Watch": nil,
}
var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
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
