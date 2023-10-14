package tom

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	glInterceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	master_interceptors "github.com/manabie-com/backend/internal/mastermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/tom/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var fakeJwtCtxEndpoint = []string{
	"/tom.v1.ConversationReaderService/ListConversationByLessons", // for call to sync data
}
var ignoreAuthEndpoint = []string{
	"/tom.v1.ConversationReaderService/ListConversationByLessons", // for call to sync data
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var locationRestrictedCtxEndpoint = []string{
	"/tom.v1.ChatReaderService/ListConversationsInSchoolV2",
}

var rbacDecider = map[string][]string{
	"/manabie.tom.ChatService/SendMessage":                        nil,
	"/manabie.tom.ChatService/SeenMessage":                        nil,
	"/manabie.tom.ChatService/ConversationList":                   nil,
	"/manabie.tom.ChatService/RetrievePushedNotificationMessages": nil,
	"/manabie.tom.ChatService/ConversationDetail":                 nil,
	"/manabie.tom.ChatService/ConversationByStudentQuestion":      nil,
	"/manabie.tom.ChatService/GetConversation":                    nil,
	"/manabie.tom.ChatService/ConversationByClass":                nil,
	"/manabie.tom.ChatService/ConversationByLesson":               nil,
	"/manabie.tom.ChatService/Subscribe":                          nil,
	"/manabie.tom.ChatService/StudentRaiseHand":                   nil,
	"/manabie.tom.ChatService/TeacherAllowAllStudentToChat":       nil,
	"/manabie.tom.ChatService/TeacherProhibitAllStudentToChat":    nil,
	"/manabie.tom.ChatService/TeacherAllowStudentToSpeak":         nil,
	"/manabie.tom.ChatService/TeacherProhibitStudentToSpeak":      nil,
	"/manabie.tom.ChatService/DeleteMessage":                      nil,
	"/manabie.tom.ChatService/StreamingEvent":                     nil,
	"/manabie.tom.ChatService/StudentPutHandDown":                 nil,
	"/manabie.tom.ChatService/RetrieveConversationEvents":         nil,
	"/manabie.tom.ChatService/StudentAcceptToSpeak":               nil,
	"/manabie.tom.ChatService/StudentDeclineToSpeak":              nil,
	"/manabie.tom.ChatService/SubscribeV2":                        nil,
	"/manabie.tom.ChatService/PingSubscribeV2":                    nil,

	"/grpc.health.v1.Health/Check": nil,
	"/grpc.health.v1.Health/Watch": nil,

	"/tom.v1.ChatModifierService/TeacherAllowAllStudentToSpeak":               nil,
	"/tom.v1.ChatModifierService/TeacherProhibitAllStudentToSpeak":            nil,
	"/tom.v1.ChatModifierService/TeacherAllowStudentToShowCamera":             nil,
	"/tom.v1.ChatModifierService/TeacherProhibitStudentToShowCamera":          nil,
	"/tom.v1.ChatModifierService/TeacherHandOffAllStudent":                    nil,
	"/tom.v1.ChatModifierService/TeacherHandOffStudent":                       nil,
	"/tom.v1.ChatModifierService/JoinConversations":                           nil,
	"/tom.v1.ChatModifierService/JoinAllConversations":                        nil,
	"/tom.v1.ChatModifierService/JoinAllConversationsWithLocations":           nil,
	"/tom.v1.ChatModifierService/LeaveConversations":                          nil,
	"/tom.v1.ChatModifierService/DeleteMessage":                               nil,
	"/tom.v1.ChatReaderService/RetrieveConversationMemberLatestEvent":         nil,
	"/tom.v1.ChatReaderService/GetConversationV2":                             nil,
	"/tom.v1.ChatReaderService/ListConversationsInSchool":                     nil,
	"/tom.v1.ChatReaderService/ListConversationsInSchoolV2":                   nil,
	"/tom.v1.ChatReaderService/ListConversationsInSchoolWithLocations":        nil,
	"/tom.v1.ChatReaderService/RetrieveTotalUnreadMessage":                    nil,
	"/tom.v1.ChatReaderService/RetrieveTotalUnreadConversationsWithLocations": nil,
	"/tom.v1.LessonChatReaderService/LiveLessonConversationDetail":            nil,
	"/tom.v1.LessonChatReaderService/LiveLessonConversationMessages":          nil,
	"/tom.v1.LessonChatReaderService/RefreshLiveLessonSession":                nil,
	"/tom.v1.LessonChatModifierService/CreateLiveLessonPrivateConversation":   nil,
	"/tom.v1.LessonChatReaderService/LiveLessonPrivateConversationMessages":   nil,

	"/tom.v1.ConversationReaderService/ListConversationIDs":       {constant.RoleSchoolAdmin}, // for call to sync data
	"/tom.v1.ConversationReaderService/ListConversationByLessons": {constant.RoleSchoolAdmin}, // for call to sync data
	"/tom.v1.ConversationReaderService/ListConversationByUsers":   {constant.RoleSchoolAdmin}, // for call to sync data
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

func locationRestrictedInterceptor(db database.Ext) *master_interceptors.LocationRestricted {
	endpoints := map[string]struct{}{}
	for _, endpoint := range locationRestrictedCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	locationRepo := &repo.LocationRepo{}
	return master_interceptors.NewLocationRestricted(endpoints, db, locationRepo)
}

func fakeSchoolAdminJwtInterceptor() *glInterceptors.FakeJwtContext {
	endpoints := map[string]struct{}{}
	for _, endpoint := range fakeJwtCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	return glInterceptors.NewFakeJwtContext(endpoints, constant.UserGroupSchoolAdmin)
}
