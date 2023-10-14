package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type TaggedUser struct {
	UserID pgtype.Text
	TagID  pgtype.Text

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (s *TaggedUser) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id", "tag_id", "created_at", "updated_at",
		}, []interface{}{
			&s.UserID, &s.TagID, &s.CreatedAt, &s.UpdatedAt,
		}
}

// TableName returns "school_info"
func (s *TaggedUser) TableName() string {
	return "tagged_user"
}

type TaggedUsers []*TaggedUser

func (ss *TaggedUsers) Add() database.Entity {
	e := &TaggedUser{}
	*ss = append(*ss, e)

	return e
}
