package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type LOProgression struct {
	BaseEntity
	ProgressionID      pgtype.Text
	ShuffledQuizSetID  pgtype.Text
	LastIndex          pgtype.Int4
	QuizExternalIDs    pgtype.TextArray
	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	SessionID          pgtype.Text
}

func (e *LOProgression) FieldMap() ([]string, []interface{}) {
	names := []string{
		"progression_id",
		"shuffled_quiz_set_id",
		"last_index",
		"student_id",
		"study_plan_id",
		"learning_material_id",
		"quiz_external_ids",
		"session_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	return names, []interface{}{
		&e.ProgressionID,
		&e.ShuffledQuizSetID,
		&e.LastIndex,
		&e.StudentID,
		&e.StudyPlanID,
		&e.LearningMaterialID,
		&e.QuizExternalIDs,
		&e.SessionID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
}

func (e *LOProgression) TableName() string {
	return "lo_progression"
}

type LOProgressions []*LOProgression

func (u *LOProgressions) Add() database.Entity {
	e := &LOProgression{}
	*u = append(*u, e)

	return e
}

func (u LOProgressions) Get() []*LOProgression {
	return []*LOProgression(u)
}
