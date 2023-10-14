package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Assignment struct {
	BaseEntity
	ID              pgtype.Text
	Content         pgtype.JSONB
	Attachment      pgtype.TextArray
	Settings        pgtype.JSONB
	CheckList       pgtype.JSONB
	Type            pgtype.Text
	Status          pgtype.Text
	MaxGrade        pgtype.Int4
	Instruction     pgtype.Text
	Name            pgtype.Text
	IsRequiredGrade pgtype.Bool
	DisplayOrder    pgtype.Int4
	OriginalTopic   pgtype.Text
	TopicID         pgtype.Text
}

func (rcv *Assignment) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"assignment_id",
		"content",
		"attachment",
		"settings",
		"check_list",
		"created_at",
		"updated_at",
		"deleted_at",
		"max_grade",
		"status",
		"instruction",
		"type",
		"name",
		"is_required_grade",
		"display_order",
		"original_topic",
		"topic_id",
	}
	values = []interface{}{
		&rcv.ID,
		&rcv.Content,
		&rcv.Attachment,
		&rcv.Settings,
		&rcv.CheckList,
		&rcv.CreatedAt,
		&rcv.UpdatedAt,
		&rcv.DeletedAt,
		&rcv.MaxGrade,
		&rcv.Status,
		&rcv.Instruction,
		&rcv.Type,
		&rcv.Name,
		&rcv.IsRequiredGrade,
		&rcv.DisplayOrder,
		&rcv.OriginalTopic,
		&rcv.TopicID,
	}
	return
}

func (rcv *Assignment) TableName() string {
	return "assignments"
}

type AssignmentSetting struct {
	AllowLateSubmission       bool `json:"allow_late_submission"`
	AllowResubmission         bool `json:"allow_resubmission"`
	RequireAttachment         bool `json:"require_attachment"`
	RequireAssignmentNote     bool `json:"require_assignment_note"`
	RequireVideoSubmission    bool `json:"require_video_submission"`
	RequireCompleteDate       bool `json:"require_complete_date"`
	RequireDuration           bool `json:"require_duration"`
	RequireCorrectness        bool `json:"require_correctness"`
	RequireUnderstandingLevel bool `json:"require_understanding_level"`
}

type AssignmentCheckList struct {
	CheckList map[string]bool
}

type AssignmentContent struct {
	TopicID string   `json:"topic_id"`
	LoIDs   []string `json:"lo_id"`
}

type Assignments []*Assignment

func (u *Assignments) Add() database.Entity {
	e := &Assignment{}
	*u = append(*u, e)

	return e
}

type BookAssignment struct {
	Assignment
	BookID    pgtype.Text
	ChapterID pgtype.Text
	TopicID   pgtype.Text
}
