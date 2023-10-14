package entities

import (
	"github.com/jackc/pgtype"
)

const (
	AcademicYearStatusActive   = "ACADEMIC_YEAR_STATUS_ACTIVE"
	AcademicYearStatusInActive = "ACADEMIC_YEAR_STATUS_INACTIVE"
)

type AcademicYear struct {
	ID            pgtype.Text
	SchoolID      pgtype.Int4
	Name          pgtype.Text
	StartYearDate pgtype.Timestamptz
	EndYearDate   pgtype.Timestamptz
	Status        pgtype.Text
	UpdatedAt     pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
}

func (rcv *AcademicYear) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"academic_year_id",
		"school_id",
		"name",
		"start_year_date",
		"end_year_date",
		"status",
		"updated_at",
		"created_at",
	}
	values = []interface{}{
		&rcv.ID,
		&rcv.SchoolID,
		&rcv.Name,
		&rcv.StartYearDate,
		&rcv.EndYearDate,
		&rcv.Status,
		&rcv.UpdatedAt,
		&rcv.CreatedAt,
	}
	return
}

func (*AcademicYear) TableName() string {
	return "academic_years"
}
