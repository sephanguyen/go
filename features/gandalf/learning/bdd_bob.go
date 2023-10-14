package learning

import (
	"github.com/manabie-com/backend/features/bob"
)

func initStepForBobServiceFeature(s *suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^returns "([^"]*)" status code$`: s.bobSuite.ReturnsStatusCode,

		// create live lesson
		`^some live courses with school id$`:     s.bobSuite.CreateLiveCourse,
		`^some medias$`:                          s.bobSuite.CreateMedias,
		`^some student accounts with school id$`: s.bobSuite.CreateStudentAccounts,
		`^some teacher accounts with school id$`: s.bobSuite.CreateTeacherAccounts,
		`^there is a live lesson with "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)", teachers, courses and learners be created$`: s.bobSuite.ThereIsLiveLessonInDB,
		`^user creates live lesson with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", teachers, courses and learners$`:             s.bobSuite.UserCreateLiveLesson,
		`^user signed as admin$`:                       s.bobSuite.SignedInAdmin,
		`^user creates lesson with missing "([^"]*)"$`: s.bobSuite.UserCreateLiveLessonWithMissing,
		`^user cannot create any lesson$`:              s.bobSuite.UserCannotCreateAnyLesson,
		`^user signed as teacher$`:                     s.bobSuite.SignedInTeacher,
		`^teacher get a live lesson with "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)", teachers, courses and learners be created$`: s.bobSuite.TeacherGetALiveLesson,
		`^user retrieve list lessons by above courses$`:                         s.bobSuite.UserRetrieveListLessonsByAboveCourses,
		`^teacher get a conversation in a room for this lesson with "([^"]*)"$`: s.teacherGetNewConversationForThisLesson,
	}
	return steps
}

type BobStepState struct {
	CurrentSchoolID  int32
	CurrentStudentId string
}

func (s *suite) newBobSuite() {
	s.bobSuite = &bob.Suite{}
	s.bobSuite.DB = s.bobDB
	s.bobSuite.Conn = s.bobConn
	s.bobSuite.ZapLogger = s.ZapLogger
	s.bobSuite.JSM = s.jsm
	s.bobSuite.ShamirConn = s.shamirConn
	s.bobSuite.ApplicantID = s.ApplicantID
}
