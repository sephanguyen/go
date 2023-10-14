package field

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

func (field TimeWithoutTz) toPGTimestamp() pgtype.Timestamp {
	pgTimestamp := pgtype.Timestamp{
		Time:   field.value.UTC(),
		Status: pgtype.Status(field.status),
	}
	return pgTimestamp
}

func (field TimeWithoutTz) toPGTimestampPtr() *pgtype.Timestamp {
	pgTimestamp := field.toPGTimestamp()
	return &pgTimestamp
}

func (field *TimeWithoutTz) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGTimestampPtr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

func (field *TimeWithoutTz) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGTimestampPtr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

func (field *TimeWithoutTz) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGTimestampPtr().EncodeText(ci, buf)
}

func (field *TimeWithoutTz) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGTimestampPtr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *TimeWithoutTz) Scan(src interface{}) error {
	value := field.toPGTimestampPtr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *TimeWithoutTz) Value() (driver.Value, error) {
	return field.toPGTimestampPtr().Value()
}
