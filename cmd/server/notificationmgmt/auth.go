package notificationmgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	gl_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/config"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/notificationmgmt.v1.NotificationModifierService/SendScheduledNotification",

	"/notificationmgmt.v1.InternalService/RetrievePushedNotificationMessages",

	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var fakeJwtCtxEndpoint = []string{
	"/notificationmgmt.v1.NotificationModifierService/SendScheduledNotification",
}

var rbacDecider = map[string][]string{
	// Deprecated
	"/bob.v1.NotificationReaderService/RetrieveNotificationDetail": {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	// Deprecated
	"/bob.v1.NotificationReaderService/RetrieveNotifications": {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	// Deprecated
	"/bob.v1.NotificationReaderService/CountUserNotification": {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	// Deprecated
	"/bob.v1.NotificationReaderService/GetAnswersByFilter": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	// Deprecated
	"/bob.v1.NotificationModifierService/SetUserNotificationStatus": {constant.RoleStudent, constant.RoleParent},

	// Deprecated
	"/yasuo.v1.NotificationModifierService/UpsertNotification": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	// Deprecated
	"/yasuo.v1.NotificationModifierService/SendNotification": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	// Deprecated
	"/yasuo.v1.NotificationModifierService/DiscardNotification": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	// Deprecated
	"/yasuo.v1.NotificationModifierService/NotifyUnreadUser": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	// Deprecated
	"/yasuo.v1.NotificationModifierService/SubmitQuestionnaire": {constant.RoleStudent, constant.RoleParent},

	"/notificationmgmt.v1.NotificationReaderService/RetrieveNotificationDetail": {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	"/notificationmgmt.v1.NotificationReaderService/RetrieveNotifications":      {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	"/notificationmgmt.v1.NotificationReaderService/CountUserNotification":      {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	"/notificationmgmt.v1.NotificationReaderService/GetAnswersByFilter":         {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationReaderService/GetNotificationsByFilter":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationReaderService/RetrieveGroupAudience":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationReaderService/GetQuestionnaireAnswersCSV": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationReaderService/RetrieveDraftAudience":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},

	"/notificationmgmt.v2.NotificationReaderService/RetrieveNotificationDetail": {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},

	"/notificationmgmt.v1.NotificationModifierService/UpsertNotification":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationModifierService/SendNotification":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationModifierService/DiscardNotification":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationModifierService/NotifyUnreadUser":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.NotificationModifierService/SubmitQuestionnaire":           {constant.RoleStudent, constant.RoleParent},
	"/notificationmgmt.v1.NotificationModifierService/SetStatusForUserNotifications": {constant.RoleStudent, constant.RoleParent},
	"/notificationmgmt.v1.NotificationModifierService/UpdateUserDeviceToken":         {constant.RoleStudent, constant.RoleParent, constant.RoleTeacher, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},
	"/notificationmgmt.v1.NotificationModifierService/UpsertQuestionnaireTemplate":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleCentreLead, constant.RoleTeacher, constant.RoleTeacherLead},
	"/notificationmgmt.v1.NotificationModifierService/DeleteNotification":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},

	"/notificationmgmt.v1.MediaModifierService/UpsertMedia": nil,

	// Deprecated
	"/notificationmgmt.v1.TagMgmtModifierService/DeleteTag":  {constant.RoleSchoolAdmin},
	"/notificationmgmt.v1.TagMgmtModifierService/UpsertTag":  {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/notificationmgmt.v1.TagMgmtModifierService/ImportTags": {constant.RoleSchoolAdmin},
	// Deprecated
	"/notificationmgmt.v1.TagMgmtReaderService/GetTagsByFilter":   {constant.RoleSchoolAdmin},
	"/notificationmgmt.v1.TagMgmtReaderService/ExportTags":        {constant.RoleSchoolAdmin},
	"/notificationmgmt.v1.TagMgmtReaderService/CheckExistTagName": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},

	"/notificationmgmt.v1.SystemNotificationReaderService/RetrieveSystemNotifications":   nil,
	"/notificationmgmt.v1.SystemNotificationModifierService/SetSystemNotificationStatus": nil,

	// Deprecated
	"/manabie.bob.UserService/UpdateUserDeviceToken": {constant.RoleStudent, constant.RoleParent, constant.RoleTeacher, constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff},

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

func authInterceptor(c *config.Config, l *zap.Logger, db database.QueryExecer) *interceptors.Auth {
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
