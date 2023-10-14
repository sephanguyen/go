package syllabus

import (
	"github.com/manabie-com/backend/features/repository/syllabus/assignment"
	"github.com/manabie-com/backend/features/repository/syllabus/book"
	"github.com/manabie-com/backend/features/repository/syllabus/course_student"
	"github.com/manabie-com/backend/features/repository/syllabus/exam_lo_submission"
	"github.com/manabie-com/backend/features/repository/syllabus/learning_objectives"
	"github.com/manabie-com/backend/features/repository/syllabus/shuffled_quiz_set"
	"github.com/manabie-com/backend/features/repository/syllabus/student_event_log"
	"github.com/manabie-com/backend/features/repository/syllabus/student_latest_submissions"
	"github.com/manabie-com/backend/features/repository/syllabus/student_submissions"
	"github.com/manabie-com/backend/features/repository/syllabus/user"
	"github.com/manabie-com/backend/features/repository/syllabus/utils"
)

func initLearningMaterialStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{}

	// init entities's steps.
	bookStep := book.InitStep((*book.Suite)(utils.NewEntitySuite(s.StepState.BookStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	courseStudentStep := course_student.InitStep((*course_student.Suite)(utils.NewEntitySuite(s.StepState.CourseStudentStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	assignmentStep := assignment.InitStep((*assignment.Suite)(utils.NewEntitySuite(s.StepState.AssignmentStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	userStep := user.InitStep((*user.Suite)(utils.NewEntitySuite(s.StepState.UserStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	shuffledQuizSetStep := shuffled_quiz_set.InitStep((*shuffled_quiz_set.Suite)(utils.NewEntitySuite(s.StepState.ShuffledQuizSetStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	studentSubmissionsStep := student_submissions.InitStep((*student_submissions.Suite)(utils.NewEntitySuite(s.StepState.StudentSubmissionStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	studentLatestSubmissionsStep := student_latest_submissions.InitStep((*student_latest_submissions.Suite)(utils.NewEntitySuite(s.StepState.StudentLatestSubmissionStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	studentEventLogStep := student_event_log.InitStep((*student_event_log.Suite)(utils.NewEntitySuite(s.StepState.StudentEventLogStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	examLOSubmissionStep := exam_lo_submission.InitStep((*exam_lo_submission.Suite)(utils.NewEntitySuite(s.StepState.ExamLOSubmissionStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))
	learningObjectiveStep := learning_objectives.InitStep((*learning_objectives.Suite)(utils.NewEntitySuite(s.StepState.LearningObjectivesStepState, s.DB, s.BobDBTrace, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))

	utils.AppendSteps(steps, courseStudentStep)
	utils.AppendSteps(steps, bookStep)
	utils.AppendSteps(steps, assignmentStep)
	utils.AppendSteps(steps, userStep)
	utils.AppendSteps(steps, shuffledQuizSetStep)
	utils.AppendSteps(steps, studentSubmissionsStep)
	utils.AppendSteps(steps, studentLatestSubmissionsStep)
	utils.AppendSteps(steps, studentEventLogStep)
	utils.AppendSteps(steps, examLOSubmissionStep)
	utils.AppendSteps(steps, learningObjectiveStep)

	return steps
}
