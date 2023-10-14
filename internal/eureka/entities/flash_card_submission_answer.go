package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type FlashCardSubmissionAnswer struct {
	BaseEntity
	StudentID          pgtype.Text
	QuizID             pgtype.Text
	SubmissionID       pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	ShuffledQuizSetID  pgtype.Text
	StudentTextAnswer  pgtype.TextArray
	CorrectTextAnswer  pgtype.TextArray
	StudentIndexAnswer pgtype.Int4Array
	CorrectIndexAnswer pgtype.Int4Array
	Point              pgtype.Int4
	IsCorrect          pgtype.BoolArray
	IsAccepted         pgtype.Bool
}

func (e *FlashCardSubmissionAnswer) FieldMap() ([]string, []interface{}) {
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
			"point",
			"is_correct",
			"is_accepted",
			"created_at",
			"updated_at",
			"deleted_at",
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
			&e.Point,
			&e.IsCorrect,
			&e.IsAccepted,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *FlashCardSubmissionAnswer) TableName() string {
	return "flash_card_submission_answer"
}

type FlashCardSubmissionAnswers []*FlashCardSubmissionAnswer

func (u *FlashCardSubmissionAnswers) Add() database.Entity {
	e := &FlashCardSubmissionAnswer{}
	*u = append(*u, e)

	return e
}

func (u FlashCardSubmissionAnswers) Get() []*FlashCardSubmissionAnswer {
	return []*FlashCardSubmissionAnswer(u)
}

type FlashCardSubmissionAnswerKey struct {
	StudentID          pgtype.Text
	SubmissionID       pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	ShuffledQuizSetID  pgtype.Text
}

func (e *FlashCardSubmissionAnswerKey) FieldMap() ([]string, []interface{}) {
	names := []string{
		"student_id",
		"submission_id",
		"study_plan_id",
		"learning_material_id",
		"shuffled_quiz_set_id",
	}
	return names, []interface{}{
		&e.StudentID,
		&e.SubmissionID,
		&e.StudyPlanID,
		&e.LearningMaterialID,
		&e.ShuffledQuizSetID,
	}
}

func (e *FlashCardSubmissionAnswerKey) TableName() string {
	return "flash_card_submission_answer"
}
