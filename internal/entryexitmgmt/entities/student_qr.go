package entities

import "github.com/jackc/pgtype"

type StudentQR struct {
	ID        pgtype.Int4
	StudentID pgtype.Text
	QRURL     pgtype.Text
	Version   pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (e *StudentQR) FieldMap() ([]string, []interface{}) {
	return []string{
			"qr_id",
			"student_id",
			"qr_url",
			"version",
			"created_at",
			"updated_at",
		}, []interface{}{
			&e.ID,
			&e.StudentID,
			&e.QRURL,
			&e.Version,
			&e.CreatedAt,
			&e.UpdatedAt,
		}
}

func (e *StudentQR) TableName() string {
	return "student_qr"
}
