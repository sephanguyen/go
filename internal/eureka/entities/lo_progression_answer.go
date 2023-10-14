package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type LOProgressionAnswer struct {
	BaseEntity
	ProgressionAnswerID pgtype.Text
	ShuffledQuizSetID   pgtype.Text
	QuizExternalID      pgtype.Text
	ProgressionID       pgtype.Text

	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text

	StudentTextAnswers  pgtype.TextArray
	StudentIndexAnswers pgtype.Int4Array
	SubmittedKeysAnswer pgtype.TextArray
}

func (e *LOProgressionAnswer) FieldMap() ([]string, []interface{}) {
	names := []string{
		"progression_answer_id",
		"shuffled_quiz_set_id",
		"quiz_external_id",
		"progression_id",

		"student_id",
		"study_plan_id",
		"learning_material_id",

		"student_text_answer",
		"student_index_answer",
		"submitted_keys_answer",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	return names, []interface{}{
		&e.ProgressionAnswerID,
		&e.ShuffledQuizSetID,
		&e.QuizExternalID,
		&e.ProgressionID,
		&e.StudentID,
		&e.StudyPlanID,
		&e.LearningMaterialID,
		&e.StudentTextAnswers,
		&e.StudentIndexAnswers,
		&e.SubmittedKeysAnswer,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
}

func (e *LOProgressionAnswer) TableName() string {
	return "lo_progression_answer"
}

type LOProgressionAnswers []*LOProgressionAnswer

func (u *LOProgressionAnswers) Add() database.Entity {
	e := &LOProgressionAnswer{}
	*u = append(*u, e)

	return e
}

func (u LOProgressionAnswers) Get() []*LOProgressionAnswer {
	return []*LOProgressionAnswer(u)
}
