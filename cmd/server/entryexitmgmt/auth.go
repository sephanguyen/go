package entryexitmgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/entryexitmgmt/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch":                                nil,
	"/entryexitmgmt.v1.EntryExitService/CreateEntryExit":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/entryexitmgmt.v1.EntryExitService/UpdateEntryExit":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/entryexitmgmt.v1.EntryExitService/DeleteEntryExit":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/entryexitmgmt.v1.EntryExitService/Scan":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/entryexitmgmt.v1.EntryExitService/RetrieveEntryExitRecords": {constant.RoleParent, constant.RoleStudent, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/entryexitmgmt.v1.EntryExitService/RetrieveStudentQRCode":    {constant.RoleParent, constant.RoleStudent, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/entryexitmgmt.v1.EntryExitService/GenerateBatchQRCodes":     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
}

func authInterceptor(c *configurations.Config, l *zap.Logger, db database.QueryExecer) *interceptors.Auth {
	groupDecider := &interceptors.GroupDecider{
		GroupFetcher: func(ctx context.Context, userID string) ([]string, error) {
			userRepo := &repository.UserRepo{}
			return interceptors.RetrieveUserRoles(ctx, userRepo, db)
		},
		AllowedGroups: rbacDecider,
	}

	a, err := interceptors.NewAuth(
		ignoreAuthEndpoint,
		groupDecider,
		c.Issuers,
	)
	if err != nil {
		l.Panic("err init authInterceptor", zap.Error(err))
	}

	return a
}
