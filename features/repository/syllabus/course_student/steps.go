package course_student

import (
	"github.com/manabie-com/backend/features/repository/syllabus/entity"
)

type StepState struct {
	Limit  int
	Offset int

	CourseID   string
	StudentID  string
	StudentIDs []string
	CourseIDs  []string

	CourseStudentListByCourseIDQuery entity.GraphqlCourseStudentsListByCourseIdsQuery
	CourseStudentListQuery           entity.GraphqlCourseStudentsListQuery
	CourseStudentListV2Query         entity.GraphqlCourseStudentsListV2Query
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{

		`^some students assigned to many course$`:                          s.someStudentAssignedToManyCourse,
		`^user get students by call CourseStudentsListByCourseIds$`:        s.userCallCourseStudentsByCourseIds,
		`^our system must return course students by course ids correctly$`: s.ourSystemMustReturnCourseStudentsByCourseIDsCorrectly,

		`^some students assigned to course$`:                              s.someStudentAssignedToCourse,
		`^user get students by call CourseSudentsList$`:                   s.userCallCourseStudentList,
		`^our system must return course students by course id correctly$`: s.ourSystemMustReturnCourseStudentsByCourseIDCorrectly,

		`^user get students by call CourseSudentsListV(\d+)$`:                                   s.userCallCourseStudentListV2,
		`^our system must return course students by course id with limit and offset correctly$`: s.ourSystemMustReturnCourseStudentsByCourseIDWithLimitAndOffsetCorrectly,
	}
	return steps
}
