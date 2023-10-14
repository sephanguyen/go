package entities

import "github.com/jackc/pgtype"

type Speeches struct {
	SpeechID  pgtype.Text
	Sentence  pgtype.Text
	Link      pgtype.Text
	Settings  pgtype.JSONB
	Type      pgtype.Text
	QuizID    pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
	CreatedBy pgtype.Text
	UpdatedBy pgtype.Text
}

func (s *Speeches) FieldMap() ([]string, []interface{}) {
	return []string{
			"speech_id",
			"sentence",
			"link",
			"settings",
			"type",
			"quiz_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"created_by",
			"updated_by",
		}, []interface{}{
			&s.SpeechID,
			&s.Sentence,
			&s.Link,
			&s.Settings,
			&s.Type,
			&s.QuizID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
			&s.CreatedBy,
			&s.UpdatedBy,
		}
}

func (s *Speeches) TableName() string {
	return "flashcard_speeches"
}
