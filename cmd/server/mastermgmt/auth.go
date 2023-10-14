package mastermgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/configurations"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	m_interceptors "github.com/manabie-com/backend/internal/mastermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/mastermgmt.v1.VersionControlReaderService/VerifyAppVersion",
	"/mastermgmt.v1.InternalService/GetConfigurations",
	"/mastermgmt.v1.AppsmithService/GetSchemaByWorkspaceID",
	"/mastermgmt.v1.MasterInternalService/GetReserveClassesByEffectiveDate",
	"/mastermgmt.v1.MasterInternalService/DeleteReserveClassesByEffectiveDate",
}

var fakeJwtCtxEndpoint = []string{
	"/mastermgmt.v1.InternalService/GetConfigurations",
	"/mastermgmt.v1.MasterInternalService/GetReserveClassesByEffectiveDate",
	"/mastermgmt.v1.MasterInternalService/DeleteReserveClassesByEffectiveDate",
}

var locationRestrictedCtxEndpoint = []string{}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch":                                                     nil,
	"/mastermgmt.v1.MasterDataCourseService/UpsertCourses":                             {constant.RoleTeacher, constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/mastermgmt.v1.MasterDataCourseService/GetCoursesByIDs":                           nil,
	"/mastermgmt.v1.MasterDataCourseService/ExportCourses":                             {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.MasterDataCourseService/ImportCourses":                             {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.CourseAccessPathService/ImportCourseAccessPaths":                   {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.CourseAccessPathService/ExportCourseAccessPaths":                   {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.CourseTypeService/ImportCourseTypes":                               {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.LocationManagementGRPCService/ImportLocation":                      {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.LocationManagementGRPCService/ImportLocationType":                  {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.LocationManagementGRPCService/ImportLocationV2":                    {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.LocationManagementGRPCService/ImportLocationTypeV2":                {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.MasterDataReaderService/RetrieveLocations":                         nil,
	"/mastermgmt.v1.MasterDataReaderService/GetLocationTree":                           nil,
	"/mastermgmt.v1.MasterDataReaderService/RetrieveLocationTypes":                     nil,
	"/mastermgmt.v1.MasterDataReaderService/RetrieveLocationTypesV2":                   nil,
	"/mastermgmt.v1.MasterDataReaderService/RetrieveLowestLevelLocations":              nil,
	"/mastermgmt.v1.ClassService/ImportClass":                                          {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.ClassService/ExportClasses":                                        {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.ClassService/UpdateClass":                                          {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.ClassService/DeleteClass":                                          {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.ClassService/RetrieveClassesByIDs":                                 constant.StaffGroupRole,
	"/mastermgmt.v1.GradeService/ImportGrades":                                         {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.GradeService/ExportGrades":                                         {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.SubjectService/ImportSubjects":                                     {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.SubjectService/ExportSubjects":                                     {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.MasterDataReaderService/ExportLocations":                           {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.MasterDataReaderService/ExportLocationTypes":                       {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.ConfigurationService/GetConfigurations":                            nil,
	"/mastermgmt.v1.ConfigurationService/GetConfigurationByKey":                        nil,
	"/mastermgmt.v1.ExternalConfigurationService/GetExternalConfigurations":            nil,
	"/mastermgmt.v1.ExternalConfigurationService/GetExternalConfigurationByKey":        nil,
	"/mastermgmt.v1.ExternalConfigurationService/GetConfigurationByKeysAndLocations":   nil,
	"/mastermgmt.v1.ExternalConfigurationService/GetConfigurationByKeysAndLocationsV2": nil,
	"/mastermgmt.v1.AppsmithService/GetPageInfoBySlug":                                 constant.StaffGroupRole,
	// "/mastermgmt.v1.OrganizationService/CreateOrganization":               {entities.UserGroupOrganizationManager},
	"/mastermgmt.v1.ExternalConfigurationService/CreateMultiConfigurations":         {constant.RoleTeacher, constant.RoleSchoolAdmin},
	"/mastermgmt.v1.CustomEntityService/ExecuteCustomEntity":                        {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.AcademicYearService/ImportAcademicCalendar":                     {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.AcademicYearService/ExportAcademicCalendar":                     {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.AcademicYearService/RetrieveLocationsForAcademic":               {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.AcademicYearService/RetrieveLocationsByLocationTypeLevelConfig": {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.WorkingHoursService/ImportWorkingHours":                         {constant.RoleSchoolAdmin},
	"/mastermgmt.v1.TimeSlotService/ImportTimeSlots":                                {constant.RoleSchoolAdmin},

	// schedule class
	"/mastermgmt.v1.ScheduleClassService/ScheduleStudentClass":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/mastermgmt.v1.ScheduleClassService/CancelScheduledStudentClass":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/mastermgmt.v1.ScheduleClassService/RetrieveScheduledStudentClass": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/mastermgmt.v1.ScheduleClassService/BulkAssignStudentsToClass":     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
}

func fakeSchoolAdminJwtInterceptor() *gl_interceptors.FakeJwtContext {
	endpoints := map[string]struct{}{}
	for _, endpoint := range fakeJwtCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	return gl_interceptors.NewFakeJwtContext(endpoints, constant.UserGroupTeacher)
}

func locationRestrictedInterceptor(db database.Ext) *m_interceptors.LocationRestricted {
	endpoints := map[string]struct{}{}
	for _, endpoint := range locationRestrictedCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	locationRepo := &repo.LocationRepo{}
	return m_interceptors.NewLocationRestricted(endpoints, db, locationRepo)
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
