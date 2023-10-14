package timesheet

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/configuration"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
}

var staffRoles = []string{
	constant.RoleSchoolAdmin,
	constant.RoleHQStaff,
	constant.RoleCentreManager,
	constant.RoleCentreLead,
	constant.RoleCentreStaff,
	constant.RoleTeacherLead,
	constant.RoleTeacher,
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch":                                                     nil,
	"/timesheet.v1.TimesheetService/CreateTimesheet":                                   staffRoles,
	"/timesheet.v1.TimesheetService/UpdateTimesheet":                                   staffRoles,
	"/timesheet.v1.TimesheetService/CountTimesheets":                                   staffRoles,
	"/timesheet.v1.TimesheetService/CountTimesheetsV2":                                 staffRoles,
	"/timesheet.v1.TimesheetService/CountSubmittedTimesheets":                          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager},
	"/timesheet.v1.TimesheetStateMachineService/DeleteTimesheet":                       staffRoles,
	"/timesheet.v1.TimesheetStateMachineService/SubmitTimesheet":                       staffRoles,
	"/timesheet.v1.ImportMasterDataService/ImportTimesheetConfig":                      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/timesheet.v1.TimesheetStateMachineService/ApproveTimesheet":                      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager},
	"/timesheet.v1.TimesheetStateMachineService/CancelApproveTimesheet":                {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager},
	"/timesheet.v1.TimesheetStateMachineService/ConfirmTimesheet":                      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/timesheet.v1.AutoCreateTimesheetService/UpdateAutoCreateTimesheetFlag":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager},
	"/timesheet.v1.TimesheetStateMachineService/CancelSubmissionTimesheet":             staffRoles,
	"/timesheet.v1.StaffTransportationExpenseService/UpsertStaffTransportationExpense": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager},
	"/timesheet.v1.TimesheetConfirmationService/GetConfirmationPeriodByDate":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager},
	"/timesheet.v1.TimesheetConfirmationService/ConfirmTimesheet":                      {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/timesheet.v1.TimesheetConfirmationService/GetTimesheetLocationList":              {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/timesheet.v1.TimesheetConfirmationService/GetNonConfirmedLocationCount":          {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/timesheet.v1.LocationService/GetGrantedLocationsOfStaff":                         staffRoles,
}

func authInterceptor(c *configuration.Config, l *zap.Logger, db database.QueryExecer) *interceptors.Auth {
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
