package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// StudentSubmissionGrade record
type StudentSubmissionGrade struct {
	BaseEntity
	ID                  pgtype.Text
	StudentSubmissionID pgtype.Text
	Grade               pgtype.Numeric
	GradeContent        pgtype.JSONB
	GraderID            pgtype.Text
	GraderComment       pgtype.Text
	Status              pgtype.Text
	EditorID            pgtype.Text
}

// StudentSubmissionGradeFields cols of the table
var StudentSubmissionGradeFields = []string{
	"student_submission_grade_id",
	"student_submission_id",
	"grade",
	"grade_content",
	"grader_id",
	"grader_comment",
	"status",
	"editor_id",

	"created_at",
	"updated_at",
	"deleted_at",
}

// FieldMap return "student_submission_grades" columns
func (g *StudentSubmissionGrade) FieldMap() ([]string, []interface{}) {
	return StudentSubmissionGradeFields, []interface{}{
		&g.ID,
		&g.StudentSubmissionID,
		&g.Grade,
		&g.GradeContent,
		&g.GraderID,
		&g.GraderComment,
		&g.Status,
		&g.EditorID,

		&g.BaseEntity.CreatedAt,
		&g.BaseEntity.UpdatedAt,
		&g.BaseEntity.DeletedAt,
	}
}

// TableName returns "student_submission_grades"
func (g *StudentSubmissionGrade) TableName() string {
	return "student_submission_grades"
}

// StudentSubmissionGrades to use with db helper
type StudentSubmissionGrades []*StudentSubmissionGrade

// Add appends new StudentSubmissionGrade to StudentSubmissionGrades slide and returns that entity
func (gs *StudentSubmissionGrades) Add() database.Entity {
	e := &StudentSubmissionGrade{}
	*gs = append(*gs, e)

	return e
}
