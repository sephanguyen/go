package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field *String) toPGText() pgtype.Text {
	pgText := pgtype.Text{
		String: field.String(),
		Status: pgtype.Status(field.Status()),
	}
	return pgText
}

func (field *String) toPGTextPtr() *pgtype.Text {
	pgText := field.toPGText()
	return &pgText
}

func (field *String) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGTextPtr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.String
	return nil
}

func (field *String) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGTextPtr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.String
	return nil
}

func (field *String) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGTextPtr().EncodeText(ci, buf)
}

func (field *String) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGTextPtr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *String) Scan(src interface{}) error {
	value := field.toPGTextPtr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.String
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *String) Value() (driver.Value, error) {
	return field.toPGTextPtr().Value()
}
