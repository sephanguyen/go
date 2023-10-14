package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field Date) toPGDate() pgtype.Date {
	pgDate := pgtype.Date{
		Time:   field.value,
		Status: pgtype.Status(field.status),
	}
	return pgDate
}

func (field Date) toPGDatePtr() *pgtype.Date {
	pgDate := field.toPGDate()
	return &pgDate
}

func (field *Date) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGDatePtr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

func (field *Date) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGDatePtr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

func (field *Date) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGDatePtr().EncodeText(ci, buf)
}

func (field *Date) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGDatePtr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *Date) Scan(src interface{}) error {
	value := field.toPGDatePtr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *Date) Value() (driver.Value, error) {
	return field.toPGDatePtr().Value()
}
