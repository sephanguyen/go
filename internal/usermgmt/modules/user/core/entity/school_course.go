package entity

import "github.com/jackc/pgtype"

type SchoolCourse struct {
	ID           pgtype.Text
	Name         pgtype.Text
	NamePhonetic pgtype.Text
	PartnerID    pgtype.Text
	SchoolID     pgtype.Text
	IsArchived   pgtype.Bool
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (s *SchoolCourse) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"school_course_id",
		"school_course_name",
		"school_course_name_phonetic",
		"school_course_partner_id",
		"school_id",
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
		&s.SchoolID,
		&s.IsArchived,
		&s.UpdatedAt,
		&s.CreatedAt,
		&s.DeletedAt,
		&s.ResourcePath,
	}
	return
}

func (*SchoolCourse) TableName() string {
	return "school_course"
}
