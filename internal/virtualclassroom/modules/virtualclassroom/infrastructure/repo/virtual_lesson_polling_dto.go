package repo

import "github.com/jackc/pgtype"

type VirtualLessonPolling struct {
	PollID          pgtype.Text
	LessonID        pgtype.Text
	Options         pgtype.JSONB
	StudentsAnswers pgtype.JSONB
	StoppedAt       pgtype.Timestamptz
	EndedAt         pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
}

func (v *VirtualLessonPolling) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"poll_id", "lesson_id", "options", "students_answers", "stopped_at", "ended_at", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&v.PollID, &v.LessonID, &v.Options, &v.StudentsAnswers, &v.StoppedAt, &v.EndedAt, &v.UpdatedAt, &v.CreatedAt, &v.DeletedAt}
	return
}

func (*VirtualLessonPolling) TableName() string {
	return "lesson_polls"
}
