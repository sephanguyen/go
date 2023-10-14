package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

//nolint
const (
	StudyPlanMonitorType_STUDENT_STUDY_PLAN = "STUDENT_STUDY_PLAN"
	StudyPlanMonitorType_STUDY_PLAN_ITEM    = "STUDY_PLAN_ITEM"
)

type StudyPlanMonitor struct {
	StudyPlanMonitorID pgtype.Text
	StudentID          pgtype.Text
	CourseID           pgtype.Text
	Type               pgtype.Text
	Payload            pgtype.JSONB
	Level              pgtype.Text
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
	DeletedAt          pgtype.Timestamptz
	AutoUpsertedAt     pgtype.Timestamptz
}

type StudyPlanMonitorPayload struct {
	LoID              pgtype.Text `json:"lo_id,omitempty"`
	AssignmentID      pgtype.Text `json:"assignment_id,omitempty"`
	BookID            pgtype.Text `json:"book_id,omitempty"`
	ChapterID         pgtype.Text `json:"chapter_id,omitempty"`
	TopicID           pgtype.Text `json:"topic_id,omitempty"`
	StudyPlanID       pgtype.Text `json:"study_plan_id,omitempty"`
	MasterStudyPlanID pgtype.Text `json:"master_study_plan_id,omitempty"`
	LMDisplayOrder    pgtype.Int4 `json:"lm_display_order,omitempty"`
}

func (rcv *StudyPlanMonitor) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"study_plan_monitor_id", "student_id", "course_id", "type", "payload", "level", "created_at", "updated_at", "deleted_at", "auto_upserted_at"}
	values = []interface{}{&rcv.StudyPlanMonitorID, &rcv.StudentID, &rcv.CourseID, &rcv.Type, &rcv.Payload, &rcv.Level, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt, &rcv.AutoUpsertedAt}

	return
}

func (rcv *StudyPlanMonitor) TableName() string {
	return "study_plan_monitors"
}

type StudyPlanMonitors []*StudyPlanMonitor

func (s *StudyPlanMonitors) Add() database.Entity {
	e := &StudyPlanMonitor{}
	*s = append(*s, e)

	return e
}
