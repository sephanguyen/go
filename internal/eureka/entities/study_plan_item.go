package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ContentStructure struct {
	CourseID     string `json:"course_id"`
	BookID       string `json:"book_id"`
	ChapterID    string `json:"chapter_id"`
	TopicID      string `json:"topic_id"`
	LoID         string `json:"lo_id"`
	AssignmentID string `json:"assignment_id"`
}

type StudyPlanItem struct {
	BaseEntity
	ID                      pgtype.Text
	StudyPlanID             pgtype.Text
	AvailableFrom           pgtype.Timestamptz
	AvailableTo             pgtype.Timestamptz
	StartDate               pgtype.Timestamptz
	EndDate                 pgtype.Timestamptz
	CompletedAt             pgtype.Timestamptz
	ContentStructure        pgtype.JSONB // unmarshal to ContentStructure struct above
	ContentStructureFlatten pgtype.Text
	DisplayOrder            pgtype.Int4
	CopyStudyPlanItemID     pgtype.Text
	Status                  pgtype.Text
	SchoolDate              pgtype.Timestamptz
}

func (e *StudyPlanItem) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"study_plan_item_id",
		"study_plan_id",
		"available_from",
		"available_to",
		"start_date",
		"end_date",
		"updated_at",
		"created_at",
		"deleted_at",
		"completed_at",
		"content_structure",
		"display_order",
		"copy_study_plan_item_id",
		"content_structure_flatten",
		"status",
		"school_date",
	}
	values = []interface{}{
		&e.ID,
		&e.StudyPlanID,
		&e.AvailableFrom,
		&e.AvailableTo,
		&e.StartDate,
		&e.EndDate,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
		&e.CompletedAt,
		&e.ContentStructure,
		&e.DisplayOrder,
		&e.CopyStudyPlanItemID,
		&e.ContentStructureFlatten,
		&e.Status,
		&e.SchoolDate,
	}
	return
}

func (e *StudyPlanItem) TableName() string {
	return "study_plan_items"
}

type StudyPlanItems []*StudyPlanItem

func (rcv *StudyPlanItems) Add() database.Entity {
	e := &StudyPlanItem{}
	*rcv = append(*rcv, e)

	return e
}

// addition field student id
type StudentStudyPlanItem struct {
	StudentID pgtype.Text
	StudyPlanItem
}
