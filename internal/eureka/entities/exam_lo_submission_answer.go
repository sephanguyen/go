package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ExamLOSubmissionAnswer struct {
	BaseEntity
	StudentID           pgtype.Text
	QuizID              pgtype.Text
	SubmissionID        pgtype.Text
	StudyPlanID         pgtype.Text
	LearningMaterialID  pgtype.Text
	ShuffledQuizSetID   pgtype.Text
	StudentTextAnswer   pgtype.TextArray
	CorrectTextAnswer   pgtype.TextArray
	StudentIndexAnswer  pgtype.Int4Array
	CorrectIndexAnswer  pgtype.Int4Array
	IsCorrect           pgtype.BoolArray
	IsAccepted          pgtype.Bool
	Point               pgtype.Int4
	SubmittedKeysAnswer pgtype.TextArray
	CorrectKeysAnswer   pgtype.TextArray
}

func (e *ExamLOSubmissionAnswer) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"quiz_id",
			"submission_id",
			"study_plan_id",
			"learning_material_id",
			"shuffled_quiz_set_id",
			"student_text_answer",
			"correct_text_answer",
			"student_index_answer",
			"correct_index_answer",
			"is_correct",
			"is_accepted",
			"point",
			"updated_at",
			"created_at",
			"deleted_at",
			"submitted_keys_answer",
			"correct_keys_answer",
		}, []interface{}{
			&e.StudentID,
			&e.QuizID,
			&e.SubmissionID,
			&e.StudyPlanID,
			&e.LearningMaterialID,
			&e.ShuffledQuizSetID,
			&e.StudentTextAnswer,
			&e.CorrectTextAnswer,
			&e.StudentIndexAnswer,
			&e.CorrectIndexAnswer,
			&e.IsCorrect,
			&e.IsAccepted,
			&e.Point,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.SubmittedKeysAnswer,
			&e.CorrectKeysAnswer,
		}
}

func (e *ExamLOSubmissionAnswer) TableName() string {
	return "exam_lo_submission_answer"
}

type ExamLOSubmissionAnswers []*ExamLOSubmissionAnswer

func (u *ExamLOSubmissionAnswers) Add() database.Entity {
	e := &ExamLOSubmissionAnswer{}
	*u = append(*u, e)

	return e
}

func (u ExamLOSubmissionAnswers) Get() []*ExamLOSubmissionAnswer {
	return []*ExamLOSubmissionAnswer(u)
}
