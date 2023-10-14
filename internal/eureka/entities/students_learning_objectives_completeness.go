package entities

import "github.com/jackc/pgtype"

type StudentsLearningObjectivesCompleteness struct {
	StudentID               pgtype.Text
	LoID                    pgtype.Text
	PresetStudyPlanWeeklyID pgtype.Text
	FirstAttemptScore       pgtype.Int2
	IsFinishedQuiz          pgtype.Bool
	IsFinishedVideo         pgtype.Bool
	IsFinishedStudyGuide    pgtype.Bool
	FirstQuizCorrectness    pgtype.Float4
	FinishedQuizAt          pgtype.Timestamptz
	HighestQuizScore        pgtype.Float4
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
}

func (t *StudentsLearningObjectivesCompleteness) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id", "lo_id", "preset_study_plan_weekly_id", "first_attempt_score", "is_finished_quiz", "is_finished_video", "is_finished_study_guide", "first_quiz_correctness", "finished_quiz_at", "highest_quiz_score", "updated_at", "created_at",
		}, []interface{}{
			&t.StudentID, &t.LoID, &t.PresetStudyPlanWeeklyID, &t.FirstAttemptScore, &t.IsFinishedQuiz, &t.IsFinishedVideo, &t.IsFinishedStudyGuide, &t.FirstQuizCorrectness, &t.FinishedQuizAt, &t.HighestQuizScore, &t.UpdatedAt, &t.CreatedAt,
		}
}

func (t *StudentsLearningObjectivesCompleteness) TableName() string {
	return "students_learning_objectives_completeness"
}
