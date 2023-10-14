package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type TaskAssignment struct {
	LearningMaterial
	Attachments               pgtype.TextArray
	Instruction               pgtype.Text
	RequireDuration           pgtype.Bool
	RequireCompleteDate       pgtype.Bool
	RequireUnderstandingLevel pgtype.Bool
	RequireCorrectness        pgtype.Bool
	RequireAttachment         pgtype.Bool
	RequireAssignmentNote     pgtype.Bool
}

func (t *TaskAssignment) FieldMap() (fields []string, values []interface{}) {
	fields, values = t.LearningMaterial.FieldMap()
	fields = append(fields,
		"attachments",
		"instruction",
		"require_duration",
		"require_complete_date",
		"require_understanding_level",
		"require_correctness",
		"require_attachment",
		"require_assignment_note",
	)

	values = append(values,
		&t.Attachments,
		&t.Instruction,
		&t.RequireDuration,
		&t.RequireCompleteDate,
		&t.RequireUnderstandingLevel,
		&t.RequireCorrectness,
		&t.RequireAttachment,
		&t.RequireAssignmentNote,
	)
	return
}

func (t *TaskAssignment) TableName() string {
	return "task_assignment"
}

type TaskAssignments []*TaskAssignment

func (u *TaskAssignments) Add() database.Entity {
	e := &TaskAssignment{}
	*u = append(*u, e)

	return e
}

func (u TaskAssignments) Get() []*TaskAssignment {
	return []*TaskAssignment(u)
}
