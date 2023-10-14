package syllabus

import (
	csp "github.com/manabie-com/backend/features/repository/syllabus/course_study_plan"
	studentstudyplan "github.com/manabie-com/backend/features/repository/syllabus/student_study_plan"
	"github.com/manabie-com/backend/features/repository/syllabus/study_plan"
	"github.com/manabie-com/backend/features/repository/syllabus/utils"
)

func initStudyPlanStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{}

	// init entities's steps.
	studyPlanStep := study_plan.InitStep((*study_plan.Suite)(utils.NewEntitySuite(s.StepState.StudyPlanStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))

	utils.AppendSteps(steps, studyPlanStep)
	courseStudyPlanSteps := csp.InitStep((*csp.Suite)(utils.NewEntitySuite(s.StepState.CourseStudyPlanStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	utils.AppendSteps(steps, courseStudyPlanSteps)
	// init entities's steps.
	studentStudyPlanStep := studentstudyplan.InitStep((*studentstudyplan.Suite)(utils.NewEntitySuite(s.StepState.StudentStudyPlanStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))

	utils.AppendSteps(steps, studentStudyPlanStep)
	return steps
}

type StudyPlanStepState struct {
}
