package lessonmgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"

	"go.uber.org/zap"
)

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
}

var rbacDecider = map[string][]string{
	"/grpc.health.v1.Health/Watch": nil,
	// lesson reader service
	"/lessonmgmt.v1.LessonReaderService/RetrieveLessons":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonReaderService/RetrieveLessonByID":        {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonReaderService/RetrieveStudentsByLesson":  {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/lessonmgmt.v1.LessonReaderService/RetrieveLessonMedias":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead, constant.RoleStudent},
	"/lessonmgmt.v1.LessonReaderService/RetrieveLessonsV2":         {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonReaderService/RetrieveLessonsOnCalendar": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	// lesson modifier service
	"/lessonmgmt.v1.LessonModifierService/UpdateLessonSchedulingStatus":     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonModifierService/CreateLesson":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonModifierService/DeleteLesson":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonModifierService/UpdateLesson":                     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonModifierService/BulkUpdateLessonSchedulingStatus": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonModifierService/MarkStudentAsReallocate":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	// student subscription service
	"/lessonmgmt.v1.StudentSubscriptionService/RetrieveStudentSubscription":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.StudentSubscriptionService/GetStudentCourseSubscriptions":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.StudentSubscriptionService/RetrieveStudentPendingReallocate": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.StudentSubscriptionService/GetStudentCoursesAndClasses":      {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	// assigned student list
	"/lessonmgmt.v1.AssignedStudentListService/GetAssignedStudentList": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.AssignedStudentListService/GetStudentAttendance":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	// lesson report - group
	"/lessonmgmt.v1.LessonReportModifierService/SaveDraftGroupLessonReport": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead, constant.RoleReportReviewer},
	"/lessonmgmt.v1.LessonReportModifierService/SubmitGroupLessonReport":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead, constant.RoleReportReviewer},
	// lesson report - individual
	"/lessonmgmt.v1.LessonReportModifierService/SaveDraftIndividualLessonReport": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead, constant.RoleReportReviewer},
	"/lessonmgmt.v1.LessonReportModifierService/SubmitIndividualLessonReport":    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead, constant.RoleReportReviewer},

	"/lessonmgmt.v1.LessonReportReaderService/RetrievePartnerDomain": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonZoomService/GenerateZoomLink":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonZoomService/DeleteZoomLink":                {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},

	// lesson executor
	"/lessonmgmt.v1.LessonExecutorService/ExportClassrooms":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/GenerateLessonCSVTemplate": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ImportLesson":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ImportZoomAccount":         {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ExportTeacher":             {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ExportEnrolledStudent":     {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ImportClassroom":           {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ExportCourseTeachingTime":  {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonExecutorService/ImportCourseTeachingTime":  {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},

	"/lessonmgmt.v1.ZoomAccountService/ImportZoomAccount":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.ZoomAccountService/ExportZoomAccount":              {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.UserService/GetTeachersSameGrantedLocation":        {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.UserService/GetStudentsManyReferenceByNameOrEmail": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonModifierService/UpdateToRecurrence":          {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},

	// classroom
	"/lessonmgmt.v1.ClassroomReaderService/RetrieveClassroomsByLocationID": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},

	// course_location_scheduling
	"/lessonmgmt.v1.CourseLocationScheduleService/ImportCourseLocationSchedule": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.CourseLocationScheduleService/ExportCourseLocationSchedule": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},

	// lesson allocation
	"/lessonmgmt.v1.LessonAllocationReaderService/GetLessonAllocation":                    {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},
	"/lessonmgmt.v1.LessonAllocationReaderService/GetLessonScheduleByStudentSubscription": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},

	// class
	"/lessonmgmt.v1.ClassReaderService/GetByStudentSubscription": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleCentreLead, constant.RoleTeacherLead},

	// classdo
	"/lessonmgmt.v1.ClassDoAccountService/ImportClassDoAccount":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.ClassDoAccountService/ExportClassDoAccount":   {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
	"/lessonmgmt.v1.PortForwardClassDoService/PortForwardClassDo": {constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher, constant.RoleTeacherLead},
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
