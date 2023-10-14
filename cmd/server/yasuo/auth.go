package yasuo

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/configurations"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/yasuo.v1.OpenSearchService/BuildConversationDocument",
	"/yasuo.v1.NotificationModifierService/SendScheduledNotification",
	"/yasuo.v1.InternalService/RetrievePushedNotificationMessages",

	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var fakeJwtCtxEndpoint = []string{
	"/yasuo.v1.NotificationModifierService/SendScheduledNotification",
}

var rbacDecider = map[string][]string{

	// "/manabie.yasuo.SchoolService/MergeSchools": {constant.UserGroupAdmin},
	// "/manabie.yasuo.SchoolService/UpdateSchool": {constant.UserGroupAdmin},

	// "/manabie.yasuo.UserService/GetUserProfile":    {constant.UserGroupAdmin},
	"/manabie.yasuo.UserService/GetBasicProfile": {constant.RoleTeacher, constant.RoleSchoolAdmin},
	// "/manabie.yasuo.UserService/UpdateUserProfile": {constant.UserGroupAdmin},
	"/manabie.yasuo.UserService/CreateUser": {constant.RoleTeacher, constant.RoleSchoolAdmin},

	// "/manabie.yasuo.SubscriptionService/CreateManualOrder":    {constant.UserGroupAdmin},
	// "/manabie.yasuo.SubscriptionService/CreatePackage":        {constant.UserGroupAdmin},
	// "/manabie.yasuo.SubscriptionService/DisableSubscription":  {constant.UserGroupAdmin},
	// "/manabie.yasuo.SubscriptionService/ToggleEnabledPackage": {constant.UserGroupAdmin},
	// "/manabie.yasuo.SubscriptionService/ExtendSubscription":   {constant.UserGroupAdmin},
	// "/manabie.yasuo.SubscriptionService/DisableOrders":        {constant.UserGroupAdmin},

	// "/manabie.yasuo.CourseService/DeleteLiveLesson":          services.SchoolPortalPermissionControl,
	// "/manabie.yasuo.CourseService/UpdateLiveLesson":          services.SchoolPortalPermissionControl,
	// "/manabie.yasuo.CourseService/CreateLiveLesson":          services.SchoolPortalPermissionControl,
	"/manabie.yasuo.CourseService/UpsertLiveCourse": services.SchoolPortalPermissionControl,
	// "/manabie.yasuo.CourseService/DeleteLiveCourse":          services.SchoolPortalPermissionControl,
	"/manabie.yasuo.CourseService/UpsertCourses":             {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/manabie.yasuo.CourseService/DeleteCourses":             {constant.RoleSchoolAdmin, constant.RoleHQStaff},
	"/manabie.yasuo.CourseService/UpsertCourseClasses":       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/manabie.yasuo.CourseService/CreateBrightCoveUploadUrl": append(services.CMSSchoolPlusPermissionControl, constant.RoleCentreLead),
	"/manabie.yasuo.CourseService/FinishUploadBrightCove":    append(services.CMSSchoolPlusPermissionControl, constant.RoleCentreLead),

	"/yasuo.v1.CourseModifierService/AttachMaterialsToCourse": {constant.RoleSchoolAdmin},

	// "/yasuo.v1.UserModifierService/UpdateUserProfile": {constant.UserGroupAdmin, constant.UserGroupSchoolAdmin},
	// "/yasuo.v1.UserModifierService/CreateUser":            {constant.UserGroupAdmin, constant.UserGroupTeacher, constant.UserGroupSchoolAdmin},
	// "/yasuo.v1.UserModifierService/AssignToParent":        {constant.UserGroupAdmin, constant.UserGroupSchoolAdmin},
	// "/yasuo.v1.UserModifierService/CreateStudent":         {constant.UserGroupAdmin, constant.UserGroupSchoolAdmin},
	"/yasuo.v1.UserModifierService/UpdateUserDeviceToken": {constant.RoleSchoolAdmin},
	// "/yasuo.v1.UserModifierService/UpdateStudent":         {constant.UserGroupAdmin, constant.UserGroupSchoolAdmin},
	// "/yasuo.v1.UserModifierService/OverrideUserPassword": {constant.UserGroupAdmin, constant.UserGroupSchoolAdmin},

	"/yasuo.v1.NotificationReaderService/RetrieveNotificationDetail": {constant.RoleSchoolAdmin},
	"/yasuo.v1.NotificationModifierService/CreateNotification":       {constant.RoleSchoolAdmin},
	"/yasuo.v1.NotificationModifierService/UpsertNotification":       {constant.RoleSchoolAdmin},
	"/yasuo.v1.NotificationModifierService/SendNotification":         {constant.RoleSchoolAdmin},
	"/yasuo.v1.NotificationModifierService/DiscardNotification":      {constant.RoleSchoolAdmin},
	"/yasuo.v1.NotificationModifierService/NotifyUnreadUser":         {constant.RoleSchoolAdmin},
	"/yasuo.v1.NotificationModifierService/SubmitQuestionnaire":      {constant.RoleStudent, constant.RoleParent},

	"/yasuo.v1.BrightcoveService/CreateBrightCoveUploadUrl":        append(services.CMSSchoolPlusPermissionControl, constant.RoleCentreLead),
	"/yasuo.v1.BrightcoveService/FinishUploadBrightCove":           append(services.CMSSchoolPlusPermissionControl, constant.RoleCentreLead),
	"/yasuo.v1.BrightcoveService/RetrieveBrightCoveProfileData":    nil,
	"/yasuo.v1.BrightcoveService/GetBrightcoveVideoInfo":           append(services.CMSSchoolPlusPermissionControl, constant.RoleCentreLead),
	"/yasuo.v1.BrightcoveService/GetVideoBrightcoveResumePosition": nil,

	"/yasuo.v1.CourseReaderService/ValidateUserSchool":      nil,
	"/yasuo.v1.UploadModifierService/UploadHtmlContent":     nil,
	"/yasuo.v1.UploadModifierService/BulkUploadHtmlContent": nil,
	"/yasuo.v1.UploadModifierService/BulkUploadFile":        nil,
	"/yasuo.v1.UploadReaderService/RetrieveUploadInfo":      nil,

	"/grpc.health.v1.Health/Check": nil,
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
