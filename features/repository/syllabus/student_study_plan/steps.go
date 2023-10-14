package studentstudyplan

import (
	"github.com/manabie-com/backend/features/repository/syllabus/entity"
	"github.com/manabie-com/backend/internal/eureka/entities"
)

type StepState struct {
	DefaultSchoolID int32
	CourseID        string
	CourseIDs       []string

	StudentID  string
	StudentIDs []string

	NumberOfBooks int
	BookIDs       []string
	Books         []*entities.Book

	ChapterIDs []string
	Chapters   []*entities.Chapter

	TopicIDs []string
	Topics   []*entities.Topic

	StudyPlan                      *entities.StudyPlan
	NumberOfStudentStudyPlansAdded int
	StudyPlanID                    string
	StudyPlans                     []*entities.StudyPlan
	StudyPlanIDs                   []string

	StudyPlanItems   []*entities.StudyPlanItem
	StudyPlanItemIDs []string

	Los   []*entities.LearningObjective
	LoIDs []string

	LoStudyPlanItems []*entities.LoStudyPlanItem

	Assignments   []*entities.Assignment
	AssignmentIDs []string

	AssignmentStudyPlanItems []*entities.AssignmentStudyPlanItem

	StudentStudyPlans []*entities.StudentStudyPlan

	StudentStudyPlanQuery        entity.GraphqlStudentStudyPlansByCourseIDQuery
	StudentStudyPlansManyV2Query entity.StudentStudyPlansManyV2

	ContentStructures []*entities.ContentStructure

	Error error

	studyPlanIDsResponse []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^<student_study_plan>some valid study plans$`:           s.someValidStudyPlans,
		`^some students register to the course$`:                 s.someStudentsRegisterToTheCourse,
		`^a user inserted some student study plans to database$`: s.aUserInsertedSomeStudentStudyPlansToDatabase,
		// StudentStudyPlansManyV2
		`^user call StudentStudyPlansManyV2$`:               s.userCallStudentStudyPlansManyV2,
		`^our system return student study plans correctly$`: s.ourSystemReturnStudentStudyPlansCorrectly,
		// StudentStudyPlansByCourseId
		`^<student_study_plan>there are study plan items existed in study plan$`:                  s.thereAreStudyPlanItemsExistedInStudyPlan,
		`^<student_study_plan>there are lo study plan items existed in study plan items$`:         s.thereAreLoStudyPlanItemsExistedInStudyPlan,
		`^<student_study_plan>there are assignment study plan items existed in study plan items$`: s.thereAreAssignmentStudyPlanItemsExistedInStudyPlanItems,
		`^user call StudentStudyPlansByCourseId$`:                                                 s.userCallStudentStudyPlansByCourseId,
		`^our system return student study plan correctly$`:                                        s.ourSystemReturnStudentStudyPlanCorrectly,

		`^return a student study plan$`:                 s.returnAStudentStudyPlan,
		`^user call FindStudentStudyPlanWithCourseIDs$`: s.userCallFindStudentStudyPlanWithCourseIDs,
	}
	return steps
}
