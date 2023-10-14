package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Check": nil,
	"/grpc.health.v1.Health/Watch": nil,

	// endpoints
	"/virtualclassroom.v1.VirtualClassroomReaderService/RetrieveWhiteboardToken": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomReaderService/GetLiveLessonState":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomReaderService/GetUserInformation":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.VirtualClassroomModifierService/JoinLiveLesson":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomModifierService/LeaveLiveLesson":             {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomModifierService/EndLiveLesson":               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/virtualclassroom.v1.VirtualClassroomModifierService/ModifyVirtualClassroomState": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomModifierService/PreparePublish":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomModifierService/Unpublish":                   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.VirtualLessonReaderService/GetLiveLessonsByLocations": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonID":     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonIDs":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualLessonReaderService/GetLessons":                {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualLessonReaderService/GetClassDoURL":             {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.VirtualClassroomChatService/GetConversationID":         {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.VirtualClassroomChatService/GetPrivateConversationIDs": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.LessonRecordingService/StartRecording":               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/virtualclassroom.v1.LessonRecordingService/GetRecordingByLessonID":       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/virtualclassroom.v1.LessonRecordingService/GetRecordingDownloadLinkByID": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/virtualclassroom.v1.LessonRecordingService/StopRecording":                {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.LiveRoomModifierService/JoinLiveRoom":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.LiveRoomModifierService/LeaveLiveRoom":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.LiveRoomModifierService/EndLiveRoom":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/virtualclassroom.v1.LiveRoomModifierService/ModifyLiveRoomState":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.LiveRoomModifierService/PreparePublishLiveRoom": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.LiveRoomModifierService/UnpublishLiveRoom":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.LiveRoomReaderService/GetLiveRoomState":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.LiveRoomReaderService/GetWhiteboardToken": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/virtualclassroom.v1.ZegoCloudService/GetAuthenticationToken":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.ZegoCloudService/GetAuthenticationTokenV2": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/virtualclassroom.v1.ZegoCloudService/GetChatConfig":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
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
