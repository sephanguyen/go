package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field Int16) toPGInt2() pgtype.Int2 {
	pgInt16 := pgtype.Int2{
		Int:    field.value,
		Status: pgtype.Status(field.status),
	}
	return pgInt16
}

func (field Int16) toPGInt2Ptr() *pgtype.Int2 {
	pgInt16 := field.toPGInt2()
	return &pgInt16
}

func (field *Int16) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGInt2Ptr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

func (field *Int16) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGInt2Ptr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

func (field *Int16) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGInt2Ptr().EncodeText(ci, buf)
}

func (field *Int16) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGInt2Ptr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *Int16) Scan(src interface{}) error {
	value := field.toPGInt2Ptr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Int
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *Int16) Value() (driver.Value, error) {
	return field.toPGInt2Ptr().Value()
}
