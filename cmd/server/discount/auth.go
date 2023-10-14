package discount

import (
	"context"

	"github.com/manabie-com/backend/internal/discount/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var fakeJwtCtxEndpoint = []string{
	"/discount.v1.InternalService/AutoSelectHighestDiscount",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch":                                            nil,
	"/discount.v1.DiscountService/RetrieveActiveStudentDiscountTag":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/discount.v1.DiscountService/UpsertStudentDiscountTag":                   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/discount.v1.ImportMasterDataService/ImportDiscountTag":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/discount.v1.ImportMasterDataService/ImportProductGroup":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/discount.v1.ImportMasterDataService/ImportProductGroupMapping":          {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/discount.v1.InternalService/AutoSelectHighestDiscount":                  {constant.RolePaymentScheduleJob, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead},
	"/discount.v1.ImportMasterDataService/ImportPackageDiscountSetting":       {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/discount.v1.ImportMasterDataService/ImportPackageDiscountCourseMapping": {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/discount.v1.ExportService/ExportMasterData":                             {constant.RoleSchoolAdmin, constant.RoleHQStaff},
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

	authInterceptor, err := interceptors.NewAuth(
		ignoreAuthEndpoint,
		groupDecider,
		c.Issuers,
	)
	if err != nil {
		l.Panic("err init authInterceptor", zap.Error(err))
	}

	return authInterceptor
}
