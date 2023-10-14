package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field Boolean) toPGBoolean() pgtype.Bool {
	pgBool := pgtype.Bool{
		Bool:   field.value,
		Status: pgtype.Status(field.status),
	}
	return pgBool
}

func (field Boolean) toPGBooleanPtr() *pgtype.Bool {
	pgBool := field.toPGBoolean()
	return &pgBool
}

func (field *Boolean) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGBooleanPtr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Bool
	return nil
}

func (field *Boolean) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGBooleanPtr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Bool
	return nil
}

func (field *Boolean) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGBooleanPtr().EncodeText(ci, buf)
}

func (field *Boolean) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGBooleanPtr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *Boolean) Scan(src interface{}) error {
	value := field.toPGBooleanPtr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Bool
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *Boolean) Value() (driver.Value, error) {
	return field.toPGBooleanPtr().Value()
}
