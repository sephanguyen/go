package managing

import (
	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/yasuo"
)

func initStepForYasuoServiceFeature(s *suite) map[string]interface{} {
	steps := map[string]interface{}{
		// create live lesson
		`^yasuo returns "([^"]*)" status code$`: s.yasuoSuite.ReturnsStatusCode,
		`^signed as "([^"]*)" account$`:         s.yasuoSuite.SignedAsAccount,
		`^a class$`:                             s.yasuoSuite.AClass,
		// `^a generate school$`:                   s.aGenerateSchool,
		`^a live course$`:     s.yasuoSuite.ALiveCourse,
		`^a teacher account$`: s.yasuoSuite.ATeacherAccount,
		`^tom must record new conversation and record new conversation_lesson$`: s.tomMustRecordNewConversationAndRecordNewConversation_lesson,

		`^tom must record message "([^"]*)" with type "([^"]*)"$`: s.tomMustRecordMessageWithType,

		// sync active student
		`^after sync active student eureka store correct student course info$`: s.afterSyncActiveStudentEurekaStoreCorrectStudentCourseInfo,
		`^school admin creates an existed course$`:                             s.schooldAdminCreatesAnExistedCourse,
		`^delete start date and end date of this student course$`:              s.deleteStartDateAndEndDateOfThisStudentCourse,
		`^eureka store student course info$`:                                   s.eurekaStoreStudentCourseInfo,
		`^run migration tool sync active student$`:                             s.runMigrationToolSyncActiveStudent,
		`^school admin creates student with course package$`:                   s.schoolAdminCreatesStudentWithCoursePackage,
	}

	return steps
}

type YasuoStepState struct {
	Random                 string
	CurrentCourseID        string
	CurrentTeacherID       string
	CurrentUserGroup       string
	CurrentLessonIDs       []string
	CurrentLessonNames     []string
	CurrentConversationIDs []string
}

func (s *suite) newYasuoSuite() {
	s.yasuoSuite = &yasuo.Suite{}
	s.yasuoSuite.Conn = s.yasuoConn
	s.yasuoSuite.BobConn = s.bobConn
	s.yasuoSuite.EurekaDB = s.eurekaDB
	s.yasuoSuite.DBTrace = s.bobDBTrace
	s.yasuoSuite.JSM = s.jsm
	s.yasuoSuite.ZapLogger = s.ZapLogger
	s.yasuoSuite.JSM = s.jsm
	s.yasuoSuite.ShamirConn = s.shamirConn
	s.yasuoSuite.ApplicantID = s.ApplicantID
	s.yasuoSuite.FirebaseClient = firebaseClient
	s.yasuoSuite.Cfg = &common.Config{
		FirebaseAPIKey: s.Cfg.FirebaseAPIKey,
	}
}
