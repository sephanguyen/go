package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type SchoolInfo struct {
	ID           pgtype.Text
	Name         pgtype.Text
	NamePhonetic pgtype.Text
	Address      pgtype.Text
	IsArchived   pgtype.Bool

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (s *SchoolInfo) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_id", "school_name", "school_name_phonetic", "address", "is_archived", "created_at", "updated_at",
		}, []interface{}{
			&s.ID, &s.Name, &s.NamePhonetic, &s.Address, &s.IsArchived, &s.CreatedAt, &s.UpdatedAt,
		}
}

// TableName returns "school_info"
func (s *SchoolInfo) TableName() string {
	return "school_info"
}

type SchoolInfos []*SchoolInfo

func (ss *SchoolInfos) Add() database.Entity {
	e := &SchoolInfo{}
	*ss = append(*ss, e)

	return e
}
