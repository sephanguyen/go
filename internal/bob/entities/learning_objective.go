package entities

import (
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type LearningObjective struct {
	ID            pgtype.Text `sql:"lo_id,pk"`
	Name          pgtype.Text
	Country       pgtype.Text
	Grade         pgtype.Int2
	Subject       pgtype.Text
	TopicID       pgtype.Text `sql:"topic_id"`
	MasterLoID    pgtype.Text `sql:"master_lo_id"`
	DisplayOrder  pgtype.Int2
	VideoScript   pgtype.Text
	Prerequisites pgtype.TextArray
	Video         pgtype.Text
	StudyGuide    pgtype.Text
	SchoolID      pgtype.Int4 `sql:"school_id"`
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	Type          pgtype.Text
}

func (t *LearningObjective) FieldMap() ([]string, []interface{}) {
	return []string{
			"lo_id", "name", "country", "grade", "subject", "topic_id", "master_lo_id", "display_order", "video_script", "prerequisites", "video", "study_guide", "school_id", "updated_at", "created_at", "deleted_at", "type",
		}, []interface{}{
			&t.ID, &t.Name, &t.Country, &t.Grade, &t.Subject, &t.TopicID, &t.MasterLoID, &t.DisplayOrder, &t.VideoScript, &t.Prerequisites, &t.Video, &t.StudyGuide, &t.SchoolID, &t.UpdatedAt, &t.CreatedAt, &t.DeletedAt, &t.Type,
		}
}

func (t *LearningObjective) TableName() string {
	return "learning_objectives"
}

type LearningObjectives []*LearningObjective

func (u *LearningObjectives) Add() database.Entity {
	e := &LearningObjective{}
	*u = append(*u, e)

	return e
}

type CopiedLearningObjective struct {
	CopiedLoID pgtype.Text
	LoID       pgtype.Text
}

type BookLearningObjective struct {
	LearningObjective
	BookID pgtype.Text
}
