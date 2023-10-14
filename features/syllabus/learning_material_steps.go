package syllabus

import (
	"github.com/manabie-com/backend/features/syllabus/assignment"
	"github.com/manabie-com/backend/features/syllabus/exam_lo"
	"github.com/manabie-com/backend/features/syllabus/flashcard"
	"github.com/manabie-com/backend/features/syllabus/learning_material"
	"github.com/manabie-com/backend/features/syllabus/learning_objective"
	"github.com/manabie-com/backend/features/syllabus/question_tag"
	"github.com/manabie-com/backend/features/syllabus/question_tag_type"
	"github.com/manabie-com/backend/features/syllabus/quiz"
	"github.com/manabie-com/backend/features/syllabus/shuffled_quiz_set"
	"github.com/manabie-com/backend/features/syllabus/student_submission"
	"github.com/manabie-com/backend/features/syllabus/task_assignment"
	"github.com/manabie-com/backend/features/syllabus/utils"
)

func initLearningMaterialStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{}

	// init entities's steps.

	loStep := learning_objective.InitStep((*learning_objective.Suite)(utils.NewEntitySuite(s.StepState.LOStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	assignmentStep := assignment.InitStep((*assignment.Suite)(utils.NewEntitySuite(s.StepState.AssignmentStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	flashcardStep := flashcard.InitStep((*flashcard.Suite)(utils.NewEntitySuite(s.StepState.FlashcardStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	lmStepState := learning_material.InitStep((*learning_material.Suite)(utils.NewEntitySuite(s.StepState.LearningMaterialStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	examLOStep := exam_lo.InitStep((*exam_lo.Suite)(utils.NewEntitySuite(s.StepState.ExamLOStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	taskAssignmentStep := task_assignment.InitStep((*task_assignment.Suite)(utils.NewEntitySuite(s.StepState.TaskAssignmentStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	shuffledQuizSetStepState := shuffled_quiz_set.InitStep((*shuffled_quiz_set.Suite)(utils.NewEntitySuite(s.StepState.ShuffledQuizSetStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	studentSubmissionState := student_submission.InitStep((*student_submission.Suite)(utils.NewEntitySuite(s.StepState.StudentSubmissionStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	questionTagStepState := question_tag.InitStep((*question_tag.Suite)(utils.NewEntitySuite(s.StepState.QuestionTagStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	questionTagTypeStepState := question_tag_type.InitStep((*question_tag_type.Suite)(utils.NewEntitySuite(s.StepState.QuestionTagTypeStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))
	quizStepState := quiz.InitStep((*quiz.Suite)(utils.NewEntitySuite(s.StepState.QuizStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))

	utils.AppendSteps(steps, loStep)
	utils.AppendSteps(steps, flashcardStep)
	utils.AppendSteps(steps, assignmentStep)
	utils.AppendSteps(steps, lmStepState)
	utils.AppendSteps(steps, examLOStep)
	utils.AppendSteps(steps, taskAssignmentStep)
	utils.AppendSteps(steps, shuffledQuizSetStepState)
	utils.AppendSteps(steps, studentSubmissionState)
	utils.AppendSteps(steps, questionTagStepState)
	utils.AppendSteps(steps, questionTagTypeStepState)
	utils.AppendSteps(steps, quizStepState)

	return steps
}
