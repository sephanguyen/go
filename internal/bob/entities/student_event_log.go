package entities

import "github.com/jackc/pgtype"

type StudentEventLog struct {
	ID        pgtype.Int4
	StudentID pgtype.Text
	EventID   pgtype.Varchar
	EventType pgtype.Varchar
	Payload   pgtype.JSONB
	CreatedAt pgtype.Timestamptz
}

func (s *StudentEventLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_event_log_id", "student_id", "event_id", "event_type", "payload", "created_at",
		}, []interface{}{
			&s.ID, &s.StudentID, &s.EventID, &s.EventType, &s.Payload, &s.CreatedAt,
		}
}

// TableName returns "student_event_logs"
func (s *StudentEventLog) TableName() string {
	return "student_event_logs"
}
