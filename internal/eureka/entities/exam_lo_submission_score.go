package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ExamLOSubmissionScore struct {
	BaseEntity
	SubmissionID      pgtype.Text
	QuizID            pgtype.Text
	TeacherID         pgtype.Text
	ShuffledQuizSetID pgtype.Text
	TeacherComment    pgtype.Text
	IsCorrect         pgtype.BoolArray
	IsAccepted        pgtype.Bool
	Point             pgtype.Int4
}

func (e *ExamLOSubmissionScore) FieldMap() ([]string, []interface{}) {
	return []string{
			"submission_id",
			"quiz_id",
			"teacher_id",
			"shuffled_quiz_set_id",
			"teacher_comment",
			"is_correct",
			"is_accepted",
			"point",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&e.SubmissionID,
			&e.QuizID,
			&e.TeacherID,
			&e.ShuffledQuizSetID,
			&e.TeacherComment,
			&e.IsCorrect,
			&e.IsAccepted,
			&e.Point,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
		}
}

func (e *ExamLOSubmissionScore) TableName() string {
	return "exam_lo_submission_score"
}

type ExamLOSubmissionScores []*ExamLOSubmissionScore

func (u *ExamLOSubmissionScores) Add() database.Entity {
	e := &ExamLOSubmissionScore{}
	*u = append(*u, e)

	return e
}

func (u ExamLOSubmissionScores) Get() []*ExamLOSubmissionScore {
	return []*ExamLOSubmissionScore(u)
}
