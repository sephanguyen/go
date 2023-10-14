package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field Int32) toPGInt4() pgtype.Int4 {
	pgInt32 := pgtype.Int4{
		Int:    field.value,
		Status: pgtype.Status(field.status),
	}
	return pgInt32
}

func (field Int32) toPGInt4Ptr() *pgtype.Int4 {
	pgInt32 := field.toPGInt4()
	return &pgInt32
}

func (field *Int32) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGInt4Ptr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

func (field *Int32) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGInt4Ptr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

func (field *Int32) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGInt4Ptr().EncodeText(ci, buf)
}

func (field *Int32) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGInt4Ptr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *Int32) Scan(src interface{}) error {
	value := field.toPGInt4Ptr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *Int32) Value() (driver.Value, error) {
	return field.toPGInt4Ptr().Value()
}
