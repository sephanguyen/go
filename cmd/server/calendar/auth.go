package calendar

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch":                           nil,
	"/grpc.health.v1.Health/Check":                           nil,
	"/calendar.v1.DateInfoReaderService/FetchDateInfo":       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
	"/calendar.v1.DateInfoReaderService/ExportDayInfo":       {constant.RoleSchoolAdmin},
	"/calendar.v1.DateInfoModifierService/DuplicateDateInfo": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
	"/calendar.v1.DateInfoModifierService/UpsertDateInfo":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},

	"/calendar.v1.SchedulerModifierService/CreateScheduler":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
	"/calendar.v1.SchedulerModifierService/UpdateScheduler":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
	"/calendar.v1.SchedulerModifierService/CreateManySchedulers": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},

	"/calendar.v1.UserReaderService/GetStaffsByLocation":                  {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
	"/calendar.v1.UserReaderService/GetStaffsByLocationIDsAndNameOrEmail": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},

	"/calendar.v1.LessonReaderService/GetLessonDetailOnCalendar":       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
	"/calendar.v1.LessonReaderService/GetLessonIDsForBulkStatusUpdate": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreLead, constant.RoleCentreStaff, constant.RoleTeacherLead, constant.RoleTeacher},
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
