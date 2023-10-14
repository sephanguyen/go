package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	g_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/zap"
)

var RLSSimulatedEndpoint = []string{
	"/eureka.v1.CourseReaderService/ListCourseIDsByStudents", // for elasticsearch call internal
}

var ignoreAuthEndpoint = []string{
	"/grpc.health.v1.Health/Check",
	"/grpc.health.v1.Health/Watch",
	"/eureka.v1.CourseReaderService/ListCourseIDsByStudents", // for elasticsearch call internal

	"/eureka.v1.CourseReaderService/ListStudentIDsByCourseV2",
}

var teacherGroupRole = []string{constants.RoleTeacher, constants.RoleSchoolAdmin, constants.RoleHQStaff, constants.RoleCentreStaff, constants.RoleCentreLead, constants.RoleCentreManager}
var backOfficeAccessBookRole = []string{constants.RoleSchoolAdmin, constants.RoleHQStaff, constants.RoleCentreLead, constants.RoleCentreManager}
var backOfficeAccessCourseRole = []string{constants.RoleSchoolAdmin, constants.RoleHQStaff}

var rbacDecider = map[string][]string{
	"/eureka.v1.AssignmentModifierService/UpsertStudyPlanItem":      nil,
	"/eureka.v1.AssignmentModifierService/AssignStudyPlan":          nil,
	"/eureka.v1.AssignmentModifierService/DeleteAssignments":        nil,
	"/eureka.v1.AssignmentModifierService/EditAssignmentTime":       nil,
	"/eureka.v1.AssignmentModifierService/AssignAssignmentsToTopic": {constants.RoleSchoolAdmin, constants.RoleHQStaff},

	"/eureka.v1.AssignmentModifierService/UpsertAssignmentsData": nil,
	"/eureka.v1.AssignmentModifierService/UpsertAssignments":     nil,
	"/eureka.v1.AssignmentModifierService/UpsertAdHocAssignment": {constants.RoleStudent},

	"/eureka.v1.AssignmentReaderService/ListStudyPlans":                   nil,
	"/syllabus.v1.StudyPlan/ListStudentStudyPlan":                         nil,
	"/eureka.v1.AssignmentReaderService/ListCourseAssignments":            nil,
	"/eureka.v1.AssignmentReaderService/ListStudentToDoItems":             nil,
	"/eureka.v1.AssignmentReaderService/ListStudentAvailableContents":     nil,
	"/eureka.v1.AssignmentReaderService/RetrieveAssignments":              nil,
	"/eureka.v1.AssignmentReaderService/RetrieveStudyPlan":                nil,
	"/eureka.v1.AssignmentReaderService/RetrieveClassAssignmentList":      nil,
	"/eureka.v1.AssignmentReaderService/RetrieveCourseAssignmentList":     nil,
	"/eureka.v1.AssignmentReaderService/RetrieveStudentToDoList":          nil,
	"/eureka.v1.AssignmentReaderService/RetrieveStudentListInAssignment":  nil,
	"/eureka.v1.AssignmentReaderService/RetrieveStudyPlanProgress":        nil,
	"/eureka.v1.AssignmentReaderService/ListCourseTodo":                   teacherGroupRole,
	"/eureka.v1.AssignmentReaderService/GetChildStudyPlanItems":           teacherGroupRole,
	"/eureka.v1.AssignmentReaderService/RetrieveStatisticAssignmentClass": teacherGroupRole,

	"/eureka.v1.StudentAssignmentWriteService/SubmitAssignment":               append(teacherGroupRole, constants.RoleStudent),
	"/eureka.v1.StudentAssignmentWriteService/GradeStudentSubmission":         teacherGroupRole,
	"/eureka.v1.StudentAssignmentWriteService/UpdateStudentSubmissionsStatus": teacherGroupRole,

	"/eureka.v1.StudentAssignmentReaderService/ListSubmissions":          teacherGroupRole,
	"/eureka.v1.StudentAssignmentReaderService/RetrieveSubmissions":      append(teacherGroupRole, constants.RoleStudent),
	"/eureka.v1.StudentAssignmentReaderService/RetrieveSubmissionGrades": append(teacherGroupRole, constants.RoleStudent),
	"/eureka.v1.StudentAssignmentReaderService/ListSubmissionsV2":        teacherGroupRole,

	"/eureka.v1.StudentSubmissionModifierService/DeleteStudentSubmission":               teacherGroupRole,
	"/eureka.v1.StudentSubmissionReaderService/RetrieveStudentSubmissionHistoryByLoIDs": {constants.RoleStudent},

	"/eureka.v1.CourseModifierService/DuplicateBook":                          teacherGroupRole,
	"/eureka.v1.CourseModifierService/AddBooks":                               append(teacherGroupRole, constants.RoleTeacherLead),
	"/eureka.v1.CourseModifierService/UpsertLOsAndAssignments":                nil,
	"/eureka.v1.CourseModifierService/UpdateDisplayOrdersOfLOsAndAssignments": teacherGroupRole,
	"/eureka.v1.CourseModifierService/CompleteStudyPlanItem":                  {constants.RoleStudent},
	"/eureka.v1.CourseModifierService/SubmitQuizAnswers":                      {constants.RoleStudent},
	"/eureka.v1.CourseModifierService/FinishFlashCardStudyProgress":           {constants.RoleStudent},
	"/eureka.v1.CourseModifierService/UpdateFlashCardStudyProgress":           {constants.RoleStudent},

	"/eureka.v1.CourseReaderService/ListClassByCourse":      append(teacherGroupRole, constants.RoleStudent),
	"/eureka.v1.CourseReaderService/ListStudentByCourse":    nil,
	"/eureka.v1.CourseReaderService/ListStudentIDsByCourse": nil,
	"/eureka.v1.CourseReaderService/ListTopicsByStudyPlan":  append(teacherGroupRole, constants.RoleTeacherLead),
	"/eureka.v1.CourseReaderService/GetStudentsAccessPath":  nil,

	"/eureka.v1.CourseReaderService/RetrieveLOs":               nil,
	"/eureka.v1.CourseReaderService/RetrieveCourseStatistic":   teacherGroupRole,
	"/eureka.v1.CourseReaderService/RetrieveCourseStatisticV2": teacherGroupRole,
	"/eureka.v1.CourseReaderService/GetLOsByCourse":            nil,

	"/syllabus.v1.Statistics/RetrieveCourseStatistic":                teacherGroupRole,
	"/syllabus.v1.Statistics/RetrieveCourseStatisticV2":              teacherGroupRole,
	"/syllabus.v1.Statistics/RetrieveSchoolHistoryByStudentInCourse": teacherGroupRole,

	"/eureka.v1.StudyPlanReaderService/ListStudyPlanByCourse":                teacherGroupRole,
	"/eureka.v1.StudyPlanReaderService/RetrieveStudyPlanItemEventLogs":       teacherGroupRole,
	"/eureka.v1.StudyPlanReaderService/GetLOHighestScoresByStudyPlanItemIDs": nil,

	"/eureka.v1.StudyPlanItemReaderService/RetrieveMappingLmIDToStudyPlanItemID": teacherGroupRole,
	"/eureka.v1.StudyPlanReaderService/GetStudentStudyPlan":                      nil,

	"/eureka.v1.StudyPlanWriteService/ImportStudyPlan": nil,
	"/syllabus.v1.StudyPlan/ImportStudyPlan":           nil,

	"/grpc.health.v1.Health/Check": nil,
	"/grpc.health.v1.Health/Watch": nil,

	"/eureka.v1.StudyPlanModifierService/DeleteStudyPlanBelongsToACourse": nil,

	"/eureka.v1.StudyPlanModifierService/UpsertStudyPlanItemV2":          append(teacherGroupRole, constants.RoleTeacherLead),
	"/eureka.v1.StudyPlanModifierService/UpdateStudyPlanItemsSchoolDate": teacherGroupRole,
	"/eureka.v1.StudyPlanModifierService/UpdateStudyPlanItemsStatus":     teacherGroupRole,

	"/eureka.v1.StudyPlanModifierService/UpsertStudyPlan": nil,

	"/eureka.v1.StudentEventLogModifierService/CreateStudentEventLogs": {constants.RoleStudent},

	"/eureka.v1.InternalModifierService/DeleteLOStudyPlanItems":         nil,
	"/eureka.v1.InternalModifierService/UpsertAdHocIndividualStudyPlan": {constants.RoleStudent},

	"/eureka.v1.StudyPlanReaderService/GetBookIDsBelongsToStudentStudyPlan": nil,
	"/eureka.v1.StudyPlanReaderService/StudentBookStudyProgress":            append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/eureka.v1.FlashCardReaderService/RetrieveLastFlashcardStudyProgress":  append(teacherGroupRole, constants.RoleStudent),

	"/eureka.v1.BookModifierService/UpsertBooks": backOfficeAccessBookRole,
	"/eureka.v1.BookReaderService/ListBooks":     append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),

	"/eureka.v1.ChapterReaderService/ListChapters":     append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/eureka.v1.ChapterModifierService/UpsertChapters": backOfficeAccessBookRole,
	"/eureka.v1.ChapterModifierService/DeleteChapters": backOfficeAccessBookRole,

	"/eureka.v1.TopicReaderService/ListToDoItemsByTopics": append(teacherGroupRole, constants.RoleStudent),
	"/eureka.v1.TopicReaderService/RetrieveTopics":        nil,

	"/eureka.v1.TopicModifierService/Upsert":           backOfficeAccessBookRole,
	"/eureka.v1.TopicModifierService/Publish":          backOfficeAccessBookRole,
	"/eureka.v1.TopicModifierService/DeleteTopics":     backOfficeAccessBookRole,
	"/eureka.v1.TopicModifierService/AssignTopicItems": {constants.RoleSchoolAdmin, constants.RoleHQStaff},

	"/eureka.v1.FlashCardReaderService/RetrieveFlashCardStudyProgress": {constants.RoleStudent},

	"/eureka.v1.QuizReaderService/RetrieveQuizTests":             append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/eureka.v1.QuizReaderService/RetrieveTotalQuizLOs":          append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/eureka.v1.QuizReaderService/RetrieveSubmissionHistory":     teacherGroupRole,
	"/eureka.v1.QuizReaderService/ListQuizzesOfLO":               teacherGroupRole,
	"/eureka.v1.QuizModifierService/UpsertQuiz":                  teacherGroupRole,
	"/eureka.v1.QuizModifierService/UpsertSingleQuiz":            backOfficeAccessBookRole,
	"/eureka.v1.QuizModifierService/CheckQuizCorrectness":        {constants.RoleStudent},
	"/eureka.v1.QuizModifierService/UpdateDisplayOrderOfQuizSet": backOfficeAccessBookRole,

	"/eureka.v1.QuizModifierService/DeleteQuiz":           backOfficeAccessBookRole,
	"/eureka.v1.QuizModifierService/RemoveQuizFromLO":     backOfficeAccessBookRole,
	"/eureka.v1.QuizModifierService/CreateFlashCardStudy": {constants.RoleStudent},
	"/eureka.v1.QuizModifierService/CreateQuizTest":       {constants.RoleStudent},
	"/eureka.v1.QuizModifierService/CreateRetryQuizTest":  {constants.RoleStudent},

	// syllabus.v1.QuestionModifierService service
	"/syllabus.v1.QuestionService/UpsertQuestionGroup":           backOfficeAccessBookRole,
	"/syllabus.v1.QuestionService/UpdateDisplayOrderOfQuizSetV2": backOfficeAccessBookRole,
	"/syllabus.v1.QuestionService/DeleteQuestionGroup":           backOfficeAccessBookRole,

	"/eureka.v1.StudyPlanReaderService/RetrieveStat":                {constants.RoleStudent, constants.RoleParent},
	"/eureka.v1.StudyPlanReaderService/RetrieveStatV2":              {constants.RoleStudent, constants.RoleParent},
	"/eureka.v1.StudentLearningTimeReader/RetrieveLearningProgress": {constants.RoleStudent, constants.RoleParent},
	"/syllabus.v1.Statistics/RetrieveLearningProgress":              {constants.RoleStudent, constants.RoleParent},

	"/eureka.v1.LearningObjectiveModifierService/UpdateLearningObjectiveName": teacherGroupRole,
	"/eureka.v1.LearningObjectiveModifierService/UpsertLOs":                   teacherGroupRole,
	"/eureka.v1.LearningObjectiveModifierService/DeleteLos":                   teacherGroupRole,

	"/syllabus.v1.Assignment/InsertAssignment":           backOfficeAccessBookRole,
	"/syllabus.v1.Assignment/UpdateAssignment":           backOfficeAccessBookRole,
	"/eureka.v1.VisionReaderService/DetectTextFromImage": append(backOfficeAccessBookRole, constants.RoleStudent),
	"/eureka.v1.ImageToText/DetectFormula":               append(backOfficeAccessBookRole, constants.RoleStudent),

	"/syllabus.v1.LearningObjective/InsertLearningObjective": backOfficeAccessBookRole,
	"/syllabus.v1.LearningObjective/UpdateLearningObjective": backOfficeAccessBookRole,
	"/syllabus.v1.LearningObjective/ListLearningObjective":   nil,
	"/syllabus.v1.LearningObjective/UpsertLOProgression":     {constants.RoleStudent},
	"/syllabus.v1.LearningObjective/RetrieveLOProgression":   {constants.RoleStudent},

	"/syllabus.v1.Flashcard/InsertFlashcard":      backOfficeAccessBookRole,
	"/syllabus.v1.Flashcard/UpdateFlashcard":      backOfficeAccessBookRole,
	"/syllabus.v1.Flashcard/ListFlashcard":        nil,
	"/syllabus.v1.Flashcard/CreateFlashCardStudy": {constants.RoleStudent},
	"/syllabus.v1.Flashcard/FinishFlashCardStudy": {constants.RoleStudent},
	"/syllabus.v1.Flashcard/GetLastestProgress":   append(teacherGroupRole, constants.RoleStudent),

	"/syllabus.v1.LearningMaterial/DeleteLearningMaterial":     backOfficeAccessBookRole,
	"/syllabus.v1.LearningMaterial/ListLearningMaterial":       nil,
	"/syllabus.v1.LearningMaterial/UpdateLearningMaterialName": backOfficeAccessBookRole,
	"/syllabus.v1.LearningMaterial/SwapDisplayOrder":           backOfficeAccessBookRole,
	"/syllabus.v1.LearningMaterial/DuplicateBook":              teacherGroupRole,

	"/syllabus.v1.ExamLO/InsertExamLO":                      backOfficeAccessBookRole,
	"/syllabus.v1.ExamLO/UpdateExamLO":                      backOfficeAccessBookRole,
	"/syllabus.v1.ExamLO/ListExamLO":                        nil,
	"/syllabus.v1.ExamLO/ListHighestResultExamLOSubmission": append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/syllabus.v1.ExamLO/ListExamLOSubmission":              teacherGroupRole,
	"/syllabus.v1.ExamLO/ListExamLOSubmissionResult":        append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/syllabus.v1.ExamLO/ListExamLOSubmissionScore":         append(teacherGroupRole, constants.RoleStudent),
	"/syllabus.v1.ExamLO/DeleteExamLOSubmission":            {constants.RoleSchoolAdmin, constants.RoleHQStaff, constants.RoleCentreManager},
	"/syllabus.v1.ExamLO/UpsertGradeBookSetting":            {constants.RoleHQStaff, constants.RoleSchoolAdmin},
	"/syllabus.v1.ExamLO/GradeAManualGradingExamSubmission": teacherGroupRole,
	"/syllabus.v1.ExamLO/BulkApproveRejectSubmission":       {constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager},
	"/syllabus.v1.ExamLO/RetrieveMetadataTaggingResult":     nil,

	"/syllabus.v1.Assignment/ListAssignment":   nil,
	"/syllabus.v1.Assignment/SubmitAssignment": append(teacherGroupRole, constants.RoleStudent),

	"/syllabus.v1.TaskAssignment/InsertTaskAssignment":      backOfficeAccessBookRole,
	"/syllabus.v1.TaskAssignment/ListTaskAssignment":        nil,
	"/syllabus.v1.TaskAssignment/UpdateTaskAssignment":      backOfficeAccessBookRole,
	"/syllabus.v1.TaskAssignment/UpsertAdhocTaskAssignment": {constants.RoleStudent},

	"/syllabus.v1.StudyPlan/UpsertMasterInfo":                 teacherGroupRole,
	"/syllabus.v1.StudyPlan/UpsertIndividual":                 teacherGroupRole,
	"/syllabus.v1.StudyPlan/RetrieveStudyPlanIdentity":        append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/syllabus.v1.StudyPlan/UpsertSchoolDate":                 teacherGroupRole,
	"/syllabus.v1.StudyPlan/UpdateStudyPlanItemsStartEndDate": nil,
	"/syllabus.v1.StudyPlan/BulkUpdateStudyPlanItemStatus":    nil,
	"/syllabus.v1.Statistics/GetStudentProgress":              append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/syllabus.v1.StudyPlan/UpsertAllocateMarker":             {constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager, constants.RoleCentreStaff},
	"/syllabus.v1.StudyPlan/ListAllocateTeacher":              {constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager, constants.RoleCentreStaff, constants.RoleTeacher, constants.RoleTeacherLead},
	"/syllabus.v1.StudyPlan/RetrieveAllocateMarker":           {constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager, constants.RoleTeacher, constants.RoleTeacherLead, constants.RoleCentreStaff, constants.RoleCentreLead},

	"/syllabus.v1.Quiz/CreateQuizTestV2":       nil,
	"/syllabus.v1.Quiz/CreateRetryQuizTestV2":  nil,
	"/syllabus.v1.Quiz/RetrieveQuizTestsV2":    append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/syllabus.v1.Quiz/UpsertFlashcardContent": teacherGroupRole,
	"/syllabus.v1.Quiz/CheckQuizCorrectness":   {constants.RoleStudent},

	"/syllabus.v1.StudentSubmissionService/ListSubmissionsV3":         teacherGroupRole,
	"/syllabus.v1.StudentSubmissionService/ListSubmissionsV4":         teacherGroupRole,
	"/syllabus.v1.StudentSubmissionService/RetrieveSubmissionHistory": teacherGroupRole,

	"/syllabus.v1.QuestionTag/ImportQuestionTag":          {constants.RoleSchoolAdmin},
	"/syllabus.v1.QuestionTagType/ImportQuestionTagTypes": {constants.RoleSchoolAdmin},
	"/syllabus.v1.Statistics/ListGradeBook":               append(teacherGroupRole, constants.RoleTeacherLead, constants.RoleStudent, constants.RoleParent),
	"/syllabus.v1.Statistics/GetStudentStat":              {constants.RoleStudent, constants.RoleParent},
	"/syllabus.v1.Statistics/ListTagByStudentInCourse":    teacherGroupRole,

	"/syllabus.v1.StudyPlan/ListToDoItem": {constants.RoleStudent},

	"/syllabus.v1.Statistics/ListSubmissions": append(teacherGroupRole, constants.RoleStudent),

	"/syllabus.v1.StudyPlan/ListToDoItemStructuredBookTree": teacherGroupRole,

	"/syllabus.v1.LearningHistoryDataSyncService/DownloadMappingFile": nil,
	"/syllabus.v1.LearningHistoryDataSyncService/UploadMappingFile":   nil,
	// students
	"/eureka.v1.StudentService/GetStudentsByLocationAndCourse": {constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager, constants.RoleTeacher, constants.RoleTeacherLead, constants.RoleCentreStaff, constants.RoleCentreLead},

	// assessment session
	"/eureka.v1.AssessmentSessionService/GetAssessmentSessionsByCourseAndLM": nil,

	// Learnosity
	"/syllabus.v1.Assessment/GetSignedRequest":                             nil,
	"/syllabus.v1.ItemsBankService/ImportItems":                            {constants.RoleSchoolAdmin},
	"/syllabus.v1.ItemsBankService/GenerateItemBankResumableUploadURL":     {constants.RoleSchoolAdmin},
	"/syllabus.v1.ItemsBankService/GenerateListItemBankResumableUploadURL": {constants.RoleSchoolAdmin},
	"/syllabus.v1.ItemsBankService/UpsertMedia":                            {constants.RoleSchoolAdmin},
	"/syllabus.v1.ItemsBankService/DeleteMedia":                            {constants.RoleSchoolAdmin},
	"/syllabus.v1.ItemsBankService/GetItemsByLM":                           {constants.RoleHQStaff, constants.RoleSchoolAdmin, constants.RoleCentreManager, constants.RoleTeacher, constants.RoleTeacherLead, constants.RoleCentreStaff, constants.RoleCentreLead},

	// v2
	"/eureka.v2.BookService/UpsertBooks":                                 backOfficeAccessBookRole,
	"/eureka.v2.BookService/GetBookContent":                              {constants.RoleStudent},
	"/eureka.v2.BookService/GetBookHierarchyFlattenByLearningMaterialID": backOfficeAccessBookRole,

	"/eureka.v2.LearningMaterialService/UpdatePublishStatusLearningMaterials": backOfficeAccessBookRole,
	"/eureka.v2.AssessmentService/GetAssessmentSignedRequest":                 {constants.RoleStudent},
	"/eureka.v2.AssessmentService/GetLearningMaterialStatuses":                nil,
	"/eureka.v2.AssessmentService/ListAssessmentSubmissionResult":             append(teacherGroupRole, constants.RoleStudent, constants.RoleParent),
	"/eureka.v2.AssessmentService/CompleteAssessmentSession":                  {constants.RoleStudent},
	"/eureka.v2.AssessmentService/GetAssessmentSubmissionDetail":              append(teacherGroupRole, constants.RoleTeacherLead),
	"/eureka.v2.AssessmentService/AllocateMarkerSubmissions":                  append(teacherGroupRole, constants.RoleTeacherLead),
	"/eureka.v2.AssessmentService/UpdateManualGradingSubmission":              append(teacherGroupRole, constants.RoleTeacherLead),

	"/eureka.v2.CourseService/UpsertCourses":    backOfficeAccessCourseRole,
	"/eureka.v2.CourseService/ListCoursesByIds": nil,

	"/eureka.v2.ItemBankService/GetTotalItemsByLM": nil,

	"/eureka.v2.StudyPlanItemService/UpsertStudyPlanItem": {constants.RoleSchoolAdmin, constants.RoleHQStaff, constants.RoleTeacher, constants.RoleCentreManager},
	"/eureka.v2.StudyPlanService/UpsertStudyPlan":         {constants.RoleSchoolAdmin, constants.RoleHQStaff, constants.RoleTeacher, constants.RoleCentreManager},

	"/eureka.v2.StudentStudyPlanService/ListStudentStudyPlan":      {constants.RoleStudent},
	"/eureka.v2.StudentStudyPlanService/GetStudentStudyPlanStatus": {constants.RoleStudent},
	"/eureka.v2.StudentStudyPlanService/ListStudentStudyPlanItem":  {constants.RoleStudent},
	"/eureka.v2.LearningMaterialService/ListLearningMaterialInfo":  {constants.RoleStudent},

	"/eureka.v2.CerebryService/GetCerebryUserToken": append(teacherGroupRole, constants.RoleStudent),
}

func rlsSimulatedInterceptor() *g_interceptors.FakeJwtContext {
	endpoints := map[string]struct{}{}
	for _, endpoint := range RLSSimulatedEndpoint {
		endpoints[endpoint] = struct{}{}
	}
	return g_interceptors.NewFakeJwtContext(endpoints, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())
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
