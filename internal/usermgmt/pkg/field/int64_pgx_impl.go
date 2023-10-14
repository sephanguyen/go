package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field Int64) toPGInt8() pgtype.Int8 {
	pgInt64 := pgtype.Int8{
		Int:    field.value,
		Status: pgtype.Status(field.status),
	}
	return pgInt64
}

func (field Int64) toPGInt8Ptr() *pgtype.Int8 {
	pgInt64 := field.toPGInt8()
	return &pgInt64
}

func (field *Int64) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGInt8Ptr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

func (field *Int64) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGInt8Ptr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

func (field *Int64) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGInt8Ptr().EncodeText(ci, buf)
}

func (field *Int64) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGInt8Ptr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *Int64) Scan(src interface{}) error {
	value := field.toPGInt8Ptr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *Int64) Value() (driver.Value, error) {
	return field.toPGInt8Ptr().Value()
}
