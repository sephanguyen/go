package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ExamLOSubmission struct {
	BaseEntity
	SubmissionID       pgtype.Text
	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	ShuffledQuizSetID  pgtype.Text
	Status             pgtype.Text
	Result             pgtype.Text
	TeacherFeedback    pgtype.Text
	TeacherID          pgtype.Text
	MarkedAt           pgtype.Timestamptz
	RemovedAt          pgtype.Timestamptz
	TotalPoint         pgtype.Int4
	LastAction         pgtype.Text
	LastActionAt       pgtype.Timestamptz
	LastActionBy       pgtype.Text
}

func (e *ExamLOSubmission) FieldMap() ([]string, []interface{}) {
	return []string{
			"submission_id",
			"student_id",
			"study_plan_id",
			"learning_material_id",
			"shuffled_quiz_set_id",
			"status",
			"result",
			"teacher_feedback",
			"teacher_id",
			"marked_at",
			"removed_at",
			"total_point",
			"last_action",
			"last_action_at",
			"last_action_by",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&e.SubmissionID,
			&e.StudentID,
			&e.StudyPlanID,
			&e.LearningMaterialID,
			&e.ShuffledQuizSetID,
			&e.Status,
			&e.Result,
			&e.TeacherFeedback,
			&e.TeacherID,
			&e.MarkedAt,
			&e.RemovedAt,
			&e.TotalPoint,
			&e.LastAction,
			&e.LastActionAt,
			&e.LastActionBy,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
		}
}

func (e *ExamLOSubmission) TableName() string {
	return "exam_lo_submission"
}

type ExamLOSubmissions []*ExamLOSubmission

func (u *ExamLOSubmissions) Add() database.Entity {
	e := &ExamLOSubmission{}
	*u = append(*u, e)

	return e
}

func (u ExamLOSubmissions) Get() []*ExamLOSubmission {
	return []*ExamLOSubmission(u)
}
