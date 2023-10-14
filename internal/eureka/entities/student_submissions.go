package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// StudentSubmission record
type StudentSubmission struct {
	BaseEntity
	ID                 pgtype.Text
	StudyPlanItemID    pgtype.Text
	AssignmentID       pgtype.Text
	StudentID          pgtype.Text
	SubmissionContent  pgtype.JSONB
	CheckList          pgtype.JSONB
	SubmissionGradeID  pgtype.Text
	Note               pgtype.Text
	Status             pgtype.Text
	EditorID           pgtype.Text
	DeletedBy          pgtype.Text
	CompleteDate       pgtype.Timestamptz
	Duration           pgtype.Int4
	CorrectScore       pgtype.Float4
	TotalScore         pgtype.Float4
	UnderstandingLevel pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
}

// FieldMap return "student_submissions" columns
func (s *StudentSubmission) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_submission_id",
		"study_plan_item_id",
		"assignment_id",
		"student_id",
		"submission_content",
		"check_list",
		"note",
		"student_submission_grade_id",
		"status",
		"created_at",
		"updated_at",
		"deleted_at",
		"deleted_by",
		"editor_id",
		"complete_date",
		"duration",
		"correct_score",
		"total_score",
		"understanding_level",
		"study_plan_id",
		"learning_material_id",
	}
	values = []interface{}{
		&s.ID,
		&s.StudyPlanItemID,
		&s.AssignmentID,
		&s.StudentID,
		&s.SubmissionContent,
		&s.CheckList,
		&s.Note,
		&s.SubmissionGradeID,
		&s.Status,
		&s.BaseEntity.CreatedAt,
		&s.BaseEntity.UpdatedAt,
		&s.BaseEntity.DeletedAt,
		&s.DeletedBy,
		&s.EditorID,
		&s.CompleteDate,
		&s.Duration,
		&s.CorrectScore,
		&s.TotalScore,
		&s.UnderstandingLevel,
		&s.StudyPlanID,
		&s.LearningMaterialID,
	}
	return
}

// TableName returns "student_submissions"
func (s *StudentSubmission) TableName() string {
	return "student_submissions"
}

// StudentSubmissions to use with db helper
type StudentSubmissions []*StudentSubmission

// Add appends new StudentSubmission to StudentSubmissions slide and returns that entity
func (ss *StudentSubmissions) Add() database.Entity {
	e := &StudentSubmission{}
	*ss = append(*ss, e)

	return e
}
