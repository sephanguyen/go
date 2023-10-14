package entity

import "github.com/jackc/pgtype"

type UserAddress struct {
	UserAddressID pgtype.Text
	UserID        pgtype.Text
	AddressType   pgtype.Text
	PostalCode    pgtype.Text
	PrefectureID  pgtype.Text
	City          pgtype.Text
	FirstStreet   pgtype.Text
	SecondStreet  pgtype.Text

	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (s *UserAddress) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_address_id",
		"user_id",
		"address_type",
		"postal_code",
		"prefecture_id",
		"city",
		"first_street",
		"second_street",
		"created_at",
		"updated_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&s.UserAddressID,
		&s.UserID,
		&s.AddressType,
		&s.PostalCode,
		&s.PrefectureID,
		&s.City,
		&s.FirstStreet,
		&s.SecondStreet,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.DeletedAt,
		&s.ResourcePath,
	}
	return
}

func (*UserAddress) TableName() string {
	return "user_address"
}
