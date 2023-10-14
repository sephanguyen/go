package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type LearningObjectiveV2 struct {
	LearningMaterial

	ManualGrading pgtype.Bool
	Video         pgtype.Text
	StudyGuide    pgtype.Text
	VideoScript   pgtype.Text
}

type LearningObjectiveBaseV2 struct {
	LearningObjectiveV2
	TotalQuestion pgtype.Int4
}

func (e *LearningObjectiveV2) FieldMap() ([]string, []interface{}) {
	fields, values := e.LearningMaterial.FieldMap()
	fields = append(fields,
		"video",
		"study_guide",
		"video_script",
		"manual_grading",
	)

	values = append(values, &e.Video, &e.StudyGuide, &e.VideoScript, &e.ManualGrading)
	return fields, values
}

func (e *LearningObjectiveV2) TableName() string {
	return "learning_objective"
}

type LearningObjectiveV2s []*LearningObjectiveV2

func (u *LearningObjectiveV2s) Add() database.Entity {
	e := &LearningObjectiveV2{}
	*u = append(*u, e)

	return e
}

func (u LearningObjectiveV2s) Get() []*LearningObjectiveV2 {
	return []*LearningObjectiveV2(u)
}
