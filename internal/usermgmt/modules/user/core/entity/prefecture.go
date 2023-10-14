package entity

import "github.com/jackc/pgtype"

type Prefecture struct {
	ID             pgtype.Text
	PrefectureCode pgtype.Text
	Country        pgtype.Text
	Name           pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (s *Prefecture) FieldMap() ([]string, []interface{}) {
	fields := []string{
		"prefecture_id",
		"prefecture_code",
		"country",
		"name",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	values := []interface{}{
		&s.ID,
		&s.PrefectureCode,
		&s.Country,
		&s.Name,
		&s.UpdatedAt,
		&s.CreatedAt,
		&s.DeletedAt,
	}
	return fields, values
}

func (*Prefecture) TableName() string {
	return "prefecture"
}
