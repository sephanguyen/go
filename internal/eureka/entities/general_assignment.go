package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type GeneralAssignment struct {
	LearningMaterial
	Attachments            pgtype.TextArray
	MaxGrade               pgtype.Int4
	Instruction            pgtype.Text
	IsRequiredGrade        pgtype.Bool
	AllowResubmission      pgtype.Bool
	RequireAttachment      pgtype.Bool
	AllowLateSubmission    pgtype.Bool
	RequireAssignmentNote  pgtype.Bool
	RequireVideoSubmission pgtype.Bool
}

func (e *GeneralAssignment) FieldMap() (fields []string, values []interface{}) {
	fields, values = e.LearningMaterial.FieldMap()
	fields = append(fields,
		"attachments",
		"max_grade",
		"instruction",
		"is_required_grade",
		"allow_resubmission",
		"require_attachment",
		"allow_late_submission",
		"require_assignment_note",
		"require_video_submission",
	)

	values = append(values,
		&e.Attachments,
		&e.MaxGrade,
		&e.Instruction,
		&e.IsRequiredGrade,
		&e.AllowResubmission,
		&e.RequireAttachment,
		&e.AllowLateSubmission,
		&e.RequireAssignmentNote,
		&e.RequireVideoSubmission,
	)
	return
}

func (e *GeneralAssignment) TableName() string {
	return "assignment"
}

type GeneralAssignments []*GeneralAssignment

func (e *GeneralAssignments) Add() database.Entity {
	u := &GeneralAssignment{}
	*e = append(*e, u)

	return u
}

func (e GeneralAssignments) Get() []*GeneralAssignment {
	return []*GeneralAssignment(e)
}
