package csp

import (
	"github.com/manabie-com/backend/features/repository/syllabus/entity"
	"github.com/manabie-com/backend/internal/eureka/entities"
)

type StepState struct {
	DefaultSchoolID int32
	OffSet          int
	Limit           int

	StudyPlanID     string
	CourseID        string
	StudyPlanItemID string
	AssignmentID    string

	StudyPlanIDs       []string
	StudyPlanItemIDs   []string
	LoStudyPlanItemIDs []string
	AssignmentIDs      []string
	ContentStructures  []*entities.ContentStructure

	StudyPlan       *entities.StudyPlan
	StudyPlanItem   *entities.StudyPlanItem
	LoStudyPlanItem *entities.LoStudyPlanItem
	CourseStudyPlan *entities.CourseStudyPlan

	StudyPlans               []*entities.StudyPlan
	StudyPlanItems           []*entities.StudyPlanItem
	LoStudyPlanItems         []*entities.LoStudyPlanItem
	CourseStudyPlans         []*entities.CourseStudyPlan
	AssignmentStudyPlanItems []*entities.AssignmentStudyPlanItem

	CourseStudyPlansListQuery       entity.GraphqlCourseStudyPlansListQuery
	CourseStudyPlansByCourseIDQuery entity.GraphqlCourseStudyPlansByCourseIDQuery
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{

		`^some valid study plans$`:                              s.someValidStudyPlans,
		`^a user inserted some course study plans to database$`: s.aUserInsertedSomeCourseStudyPlansToDatabase,
		// CourseStudyPlansList
		`^user call CourseStudyPlansList$`:                 s.userCallCourseStudyPlansList,
		`^our system return course study plans correctly$`: s.ourSystemReturnCourseStudyPlansCorrectly,
		// CourseStudyPlansByCourseId
		`^there are study plan items existed in study plan$`:                  s.thereAreStudyPlanItemsExistedInStudyPlan,
		`^there are lo study plan items existed in study plan items$`:         s.thereAreLoStudyPlanItemsExistedInStudyPlan,
		`^there are assignment study plan items existed in study plan items$`: s.thereAreAssignmentStudyPlanItemsExistedInStudyPlanItems,
		`^user call CourseStudyPlansByCourseId$`:                              s.userCallCourseStudyPlansByCourseID,
		`^our system return course study plan correctly$`:                     s.ourSystemReturnCourseStudyPlanCorrectly,
	}
	return steps
}
