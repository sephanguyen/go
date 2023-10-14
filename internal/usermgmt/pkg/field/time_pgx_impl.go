package field

import (
	"database/sql/driver"
	"time"

	"github.com/jackc/pgtype"
)

func (field Time) toPGTimestamptz() pgtype.Timestamptz {
	pgTimestamptz := pgtype.Timestamptz{
		Time:   field.value,
		Status: pgtype.Status(field.status),
	}
	return pgTimestamptz
}

func (field Time) toPGTimestamptzPtr() *pgtype.Timestamptz {
	pgTimestamptz := field.toPGTimestamptz()
	return &pgTimestamptz
}

func (field *Time) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGTimestamptzPtr()
	if err := value.DecodeText(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

func (field *Time) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	value := field.toPGTimestamptzPtr()
	if err := value.DecodeBinary(ci, src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

func (field *Time) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGTimestamptzPtr().EncodeText(ci, buf)
}

func (field *Time) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return field.toPGTimestamptzPtr().EncodeBinary(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (field *Time) Scan(src interface{}) error {
	value := field.toPGTimestamptzPtr()
	if err := value.Scan(src); err != nil {
		return err
	}

	field.status = Status(value.Status)
	field.value = value.Time
	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (field *Time) Value() (driver.Value, error) {
	return field.toPGTimestamptzPtr().Value()
}

// After reports whether the time instant t is after u.
func (field *Time) After(time *Time) bool {
	return field.value.After(time.value)
}

// After reports whether the time instant t is after time.
func (field *Time) AfterTime(time time.Time) bool {
	return field.value.After(time)
}

// IsZero reports whether t represents the zero time instant,
// January 1, year 1, 00:00:00 UTC.
func (field *Time) IsZero() bool {
	return field.value.IsZero()
}

// Before reports whether the time instant t is before u.
func (field *Time) Before(time *Time) bool {
	return field.value.Before(time.value)
}

// Before reports whether the time instant t is before time.
func (field *Time) BeforeTime(time time.Time) bool {
	return field.value.Before(time)
}
