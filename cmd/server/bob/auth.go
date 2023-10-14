package bob

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	old_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	m_interceptors "github.com/manabie-com/backend/internal/mastermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	// internal access only
	"/bob.v1.InternalReaderService/VerifyAppVersion",
	"/bob.v1.InternalModifierService/SubmitQuizAnswers",

	// public access
	"/manabie.bob.UserService/CheckProfile",

	"/bob.v1.UserModifierService/ExchangeToken",
	"/bob.v1.UserModifierService/ExchangeCustomToken",

	"/grpc.health.v1.Health/Check",

	"/bob.v1.StudentReaderService/GetListSchoolIDsByStudentIDs",
	"/bob.v1.PostgresUserService/GetPostgresUserPermission",
	"/bob.v1.PostgresNamespaceService/GetPostgresNamespace",
	"/bob.v1.MediaModifierService/GenerateAudioFile",
}

var fakeJwtCtxEndpoint = []string{
	"/bob.v1.InternalModifierService/SubmitQuizAnswers",
}

var locationRestrictedCtxEndpoint = []string{
	"/bob.v1.CourseReaderService/ListCoursesByLocations",
}

var rbacDecider = map[string][]string{
	"/manabie.bob.MasterDataService/GetClientVersion":      nil,
	"/manabie.bob.MasterDataService/ImportPresetStudyPlan": {constant.RoleSchoolAdmin},

	// "/manabie.bob.Class/LeaveClass": {constant.RoleTeacher, constant.RoleStudent},
	// "/manabie.bob.Class/RemoveMember":                    {constant.RoleTeacher, constant.RoleSchoolAdmin},
	"/manabie.bob.Class/RetrieveAssignedPresetStudyPlan": {constant.RoleTeacher},
	// "/manabie.bob.Class/EditClass":                       {constant.RoleTeacher, constant.RoleSchoolAdmin},

	// TODO: remove in epic LT-24589
	"/manabie.bob.Class/JoinClass":           {constant.RoleTeacher, constant.RoleStudent},
	"/manabie.bob.Class/CreateClass":         {constant.RoleSchoolAdmin, constant.RoleTeacher},
	"/manabie.bob.Class/RetrieveClassMember": {constant.RoleTeacher, constant.RoleStudent},

	// "/manabie.bob.Class/UpdateClassCode":            {entities.UserGroupAdmin, constant.RoleTeacher, constant.RoleSchoolAdmin},
	// "/manabie.bob.Class/AddClassMember":             {entities.UserGroupAdmin, constant.RoleSchoolAdmin},
	"/manabie.bob.Class/TeacherRetrieveStreamToken": {constant.RoleTeacher},
	"/manabie.bob.Class/EndLiveLesson":              {constant.RoleTeacher},
	"/manabie.bob.Class/StudentRetrieveStreamToken": {constant.RoleStudent},
	"/manabie.bob.Class/JoinLesson":                 {constant.RoleStudent, constant.RoleTeacher},
	"/manabie.bob.Class/LeaveLesson":                {constant.RoleStudent, constant.RoleTeacher},
	"/manabie.bob.Class/UpsertMedia":                nil,
	"/manabie.bob.Class/RetrieveMedia":              nil,

	"/manabie.bob.Student/GetStudentProfile":               nil,
	"/manabie.bob.Student/CountTotalLOsFinished":           {constant.RoleStudent, constant.RoleParent},
	"/manabie.bob.Student/AssignPresetStudyPlans":          {constant.RoleStudent},
	"/manabie.bob.Student/UpdateProfile":                   {constant.RoleStudent},
	"/manabie.bob.Student/RetrieveLearningProgress":        {constant.RoleStudent, constant.RoleParent},
	"/manabie.bob.Student/RetrievePresetStudyPlanWeeklies": {constant.RoleStudent},
	"/manabie.bob.Student/RetrieveStudentStudyPlans":       {constant.RoleStudent},
	"/manabie.bob.Student/StudentPermission":               {constant.RoleStudent},
	"/manabie.bob.Student/RetrieveStat":                    {constant.RoleStudent, constant.RoleParent},
	"/manabie.bob.Student/FindStudent":                     {constant.RoleSchoolAdmin},
	"/manabie.bob.Student/CreateStudentEventLogs":          {constant.RoleStudent},
	"/manabie.bob.Student/RetrieveStudentComment":          {constant.RoleTeacher},
	"/manabie.bob.Student/RetrievePresetStudyPlans":        {constant.RoleStudent},
	"/manabie.bob.Student/UpsertStudentComment":            {constant.RoleTeacher},

	"/manabie.bob.Course/RetrieveCourses":               nil,
	"/manabie.bob.Course/RetrieveGradeMap":              nil,
	"/manabie.bob.Course/RetrieveAssignedCourses":       {constant.RoleStudent, constant.RoleTeacher},
	"/manabie.bob.Course/UpsertPresetStudyPlans":        {constant.RoleSchoolAdmin},
	"/manabie.bob.Course/UpsertPresetStudyPlanWeeklies": {constant.RoleSchoolAdmin},
	"/manabie.bob.Course/UpsertLOs":                     {constant.RoleSchoolAdmin, constant.RoleTeacher},
	"/manabie.bob.Course/RetrieveLiveLesson":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/manabie.bob.Course/RetrieveCoursesByIDs":          nil,
	"/manabie.bob.Course/CreateBrightCoveUploadUrl":     {constant.RoleSchoolAdmin, constant.RoleTeacher, constant.RoleStudent},
	"/manabie.bob.Course/FinishUploadBrightCove":        {constant.RoleSchoolAdmin, constant.RoleTeacher, constant.RoleStudent},
	"/manabie.bob.Course/RetrieveBooks":                 {constant.RoleStudent, constant.RoleTeacher},

	"/manabie.bob.UserService/ClaimsUserAuth":        nil,
	"/manabie.bob.UserService/UpdateUserDeviceToken": nil,
	"/manabie.bob.UserService/UpdateUserProfile":     nil,
	"/manabie.bob.UserService/GetCurrentUserProfile": nil,
	"/manabie.bob.UserService/GetTeacherProfiles":    {constant.RoleTeacher, constant.RoleStudent},
	"/manabie.bob.UserService/GetBasicProfile":       nil,

	"/bob.v1.UserReaderService/RetrieveBasicProfile": nil,

	"/bob.v1.UserModifierService/UpdateUserProfile":       nil,
	"/bob.v1.UserModifierService/UpdateUserLastLoginDate": nil,

	"/grpc.health.v1.Health/Watch": nil,

	"/bob.v1.CourseModifierService/RetrieveSubmissionHistory": {constant.RoleTeacher},
	"/bob.v1.CourseModifierService/UpsertAssignments":         nil,

	"/bob.v1.ClassReaderService/RetrieveClassByIDs":              constant.StaffGroupRole,
	"/bob.v1.ClassReaderService/ListStudentsByLesson":            nil,
	"/bob.v1.ClassReaderService/RetrieveClassMembers":            append(constant.StaffGroupRole, constant.RoleStudent),
	"/bob.v1.ClassReaderService/RetrieveClassMembersWithFilters": constant.StaffGroupRole,
	"/bob.v1.CourseReaderService/RetrieveBookTreeByTopicIDs":     nil,

	"/bob.v1.ClassModifierService/JoinLesson":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/bob.v1.ClassModifierService/LeaveLesson":             {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/bob.v1.ClassModifierService/ConvertMedia":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/bob.v1.ClassModifierService/EndLiveLesson":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.ClassModifierService/RetrieveWhiteboardToken": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/bob.v1.UserReaderService/SearchBasicProfile":               nil,
	"/bob.v1.CourseReaderService/ListLessonMedias":               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent, constant.RoleParent},
	"/bob.v1.CourseReaderService/ListCourses":                    nil,
	"/bob.v1.CourseReaderService/ListCoursesByLocations":         nil,
	"/bob.v1.CourseReaderService/RetrieveFlashCardStudyProgress": {constant.RoleStudent},

	"/bob.v1.LessonModifierService/PreparePublish":              {constant.RoleStudent, constant.RoleTeacher},
	"/bob.v1.LessonModifierService/Unpublish":                   {constant.RoleStudent, constant.RoleTeacher},
	"/bob.v1.LessonReaderService/GetStreamingLearners":          {constant.RoleStudent, constant.RoleTeacher},
	"/bob.v1.LessonReaderService/RetrieveLessons":               {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.LessonModifierService/CreateLiveLesson":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.LessonModifierService/UpdateLiveLesson":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.LessonModifierService/DeleteLiveLesson":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.LessonModifierService/ModifyLiveLessonState":       {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/bob.v1.LessonReaderService/GetLiveLessonState":            {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/bob.v1.LessonReaderService/RetrieveLiveLessonByLocations": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},

	"/bob.v1.LessonReportReaderService/RetrievePartnerDomain":          {constant.RoleStudent, constant.RoleTeacher, constant.RoleSchoolAdmin},
	"/bob.v1.LessonReportModifierService/CreateIndividualLessonReport": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/bob.v1.LessonReportModifierService/SubmitLessonReport":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/bob.v1.LessonReportModifierService/SaveDraftLessonReport":        {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},

	"/bob.v1.StudentSubscriptionService/RetrieveStudentSubscription":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.StudentSubscriptionService/GetStudentCourseSubscriptions": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},

	"/bob.v1.UploadService/GeneratePresignedPutObjectURL": nil,
	"/bob.v1.UploadService/GenerateResumableUploadURL":    nil,

	"/bob.v1.StudentModifierService/DeleteStudentComments":          {constant.RoleTeacher},
	"/bob.v1.NotificationReaderService/RetrieveNotificationDetail":  {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	"/bob.v1.NotificationReaderService/RetrieveNotifications":       {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	"/bob.v1.NotificationReaderService/CountUserNotification":       {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},
	"/bob.v1.NotificationReaderService/GetAnswersByFilter":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.NotificationModifierService/SetUserNotificationStatus": {constant.RoleTeacher, constant.RoleStudent, constant.RoleParent},

	"/notificationmgmt.v1.NotificationModifierService/SetStatusForUserNotifications": {constant.RoleStudent, constant.RoleParent},

	"/bob.v1.StudentReaderService/RetrieveStudentAssociatedToParentAccount": {constant.RoleParent},
	"/bob.v1.StudentReaderService/RetrieveStudentProfile":                   nil,
	"/bob.v1.StudentReaderService/RetrieveStudentSchoolHistory":             nil,

	"/bob.v1.InternalModifierService/CalculateAssignmentLearningTime": nil,

	// "/bob.v1.MediaModifierService/GenerateAudioFile": nil,

	"/bob.v1.LessonManagementService/CreateLesson":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/bob.v1.LessonManagementService/RetrieveLessons": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/bob.v1.LessonManagementService/UpdateLesson":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},
	"/bob.v1.LessonManagementService/DeleteLesson":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher},

	"/bob.v1.MasterDataImporterService/ImportLocation":             {constant.RoleSchoolAdmin},
	"/bob.v1.MasterDataImporterService/ImportLocationType":         {constant.RoleSchoolAdmin},
	"/bob.v1.MasterDataReaderService/RetrieveLocations":            nil,
	"/bob.v1.MasterDataReaderService/RetrieveLocationTypes":        nil,
	"/bob.v1.MasterDataReaderService/RetrieveLowestLevelLocations": nil,

	"/bob.v1.InternalReaderService/RetrieveTopicLOs": nil,
	"/bob.v1.InternalReaderService/RetrieveTopics":   nil,

	"/bob.v1.CourseReaderService/GetLOHighestScoresByStudyPlanItemIDs": nil,
}

func fakeSchoolAdminJwtInterceptor() *old_interceptors.FakeJwtContext {
	endpoints := map[string]struct{}{}
	for _, endpoint := range fakeJwtCtxEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	return old_interceptors.NewFakeJwtContext(endpoints, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())
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
