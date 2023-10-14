package entities

import (
	"github.com/jackc/pgtype"
)

type LessonPolling struct {
	PollID          pgtype.Text
	LessonID        pgtype.Text
	Options         pgtype.JSONB
	StudentsAnswers pgtype.JSONB
	StoppedAt       pgtype.Timestamptz
	EndedAt         pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeleteAt        pgtype.Timestamptz
}

func (rcv *LessonPolling) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"poll_id", "lesson_id", "options", "students_answers", "stopped_at", "ended_at", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&rcv.PollID, &rcv.LessonID, &rcv.Options, &rcv.StudentsAnswers, &rcv.StoppedAt, &rcv.EndedAt, &rcv.UpdatedAt, &rcv.CreatedAt, &rcv.DeleteAt}
	return
}

func (*LessonPolling) TableName() string {
	return "lesson_polls"
}
