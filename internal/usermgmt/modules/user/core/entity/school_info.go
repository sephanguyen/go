package entity

import "github.com/jackc/pgtype"

type SchoolInfo struct {
	ID           pgtype.Text
	Name         pgtype.Text
	NamePhonetic pgtype.Text
	PartnerID    pgtype.Text
	LevelID      pgtype.Text
	Address      pgtype.Text
	IsArchived   pgtype.Bool
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (s *SchoolInfo) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"school_id",
		"school_name",
		"school_name_phonetic",
		"school_partner_id",
		"school_level_id",
		"address",
		"is_archived",
		"updated_at",
		"created_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&s.ID,
		&s.Name,
		&s.NamePhonetic,
		&s.PartnerID,
		&s.LevelID,
		&s.Address,
		&s.IsArchived,
		&s.UpdatedAt,
		&s.CreatedAt,
		&s.DeletedAt,
		&s.ResourcePath,
	}
	return
}

func (*SchoolInfo) TableName() string {
	return "school_info"
}
