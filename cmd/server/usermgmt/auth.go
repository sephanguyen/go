package usermgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/usermgmt.v2.AuthService/ExchangeCustomToken",
	"/usermgmt.v2.AuthService/GetAuthInfo",
	"/usermgmt.v2.AuthService/ResetPassword",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch": nil,

	"/usermgmt.v2.UserModifierService/CreateParentsAndAssignToStudent":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/ImportParentsAndAssignToStudent":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/CreateStudent":                                   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/ReissueUserPassword":                             nil,
	"/usermgmt.v2.UserModifierService/RemoveParentFromStudent":                         {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/UpdateParentsAndFamilyRelationship":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/UpdateStudent":                                   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/UpsertStudentCoursePackage":                      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.UserModifierService/UpdateUserLastLoginDate":                         nil,
	"/usermgmt.v2.UserModifierService/UpdateUserProfile":                               nil,
	"/usermgmt.v2.UserModifierService/GenerateImportParentsAndAssignToStudentTemplate": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},

	"/usermgmt.v2.UserReaderService/RetrieveStudentAssociatedToParentAccount": {constant.RoleParent},
	"/usermgmt.v2.UserReaderService/SearchBasicProfile":                       nil,
	"/usermgmt.v2.UserReaderService/GetBasicProfile":                          nil,

	"/usermgmt.v2.StudentService/UpsertStudent":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.StudentService/UpsertStudentComment":          {constant.RoleTeacher, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacherLead},
	"/usermgmt.v2.StudentService/DeleteStudentComments":         {constant.RoleTeacher, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacherLead},
	"/usermgmt.v2.StudentService/RetrieveStudentComment":        {constant.RoleTeacher, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacherLead},
	"/usermgmt.v2.StudentService/ImportStudent":                 {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.StudentService/ImportStudentV2":               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.StudentService/GenerateImportStudentTemplate": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/usermgmt.v2.StudentService/GetStudentProfile":             nil,

	"/usermgmt.v2.WithusStudentService/ImportWithusManagaraBaseCSV": {constant.RoleUsermgmtScheduleJob},

	"/usermgmt.v2.SchoolInfoService/ImportSchoolInfo": {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	"/usermgmt.v2.StaffService/CreateStaff":        {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/usermgmt.v2.StaffService/UpdateStaff":        {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/usermgmt.v2.StaffService/UpdateStaffSetting": {constant.RoleSchoolAdmin, constant.RoleHQStaff},

	"/usermgmt.v2.UserGroupMgmtService/CreateUserGroup":   {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/usermgmt.v2.UserGroupMgmtService/UpdateUserGroup":   {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/usermgmt.v2.UserGroupMgmtService/ValidateUserLogin": nil,

	"/usermgmt.v2.AuthService/ValidateUserIP": nil,
}

func authInterceptor(c *configurations.Config, l *zap.Logger, db database.QueryExecer) *interceptors.Auth {
	groupDecider := &interceptors.GroupDecider{
		GroupFetcher: func(ctx context.Context, userID string) ([]string, error) {
			userRepo := &repository.UserRepo{}
			return interceptors.RetrieveUserRolesV2(ctx, userRepo, db)
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
