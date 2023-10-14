package entities

import "github.com/jackc/pgtype"

type EntryExitQueue struct {
	EntryExitQueueID pgtype.Text
	StudentID        pgtype.Text
	ResourcePath     pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (e *EntryExitQueue) FieldMap() ([]string, []interface{}) {
	return []string{
			"entryexit_queue_id",
			"student_id",
			"resource_path",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.EntryExitQueueID,
			&e.StudentID,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *EntryExitQueue) TableName() string {
	return "entryexit_queue"
}
