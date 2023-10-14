package entities

import "github.com/jackc/pgtype"

type StudentEntryExitRecords struct {
	ID        pgtype.Int4
	StudentID pgtype.Text
	EntryAt   pgtype.Timestamptz
	ExitAt    pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (e *StudentEntryExitRecords) FieldMap() ([]string, []interface{}) {
	return []string{
			"entryexit_id",
			"student_id",
			"entry_at",
			"exit_at",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.ID,
			&e.StudentID,
			&e.EntryAt,
			&e.ExitAt,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *StudentEntryExitRecords) TableName() string {
	return "student_entryexit_records"
}
