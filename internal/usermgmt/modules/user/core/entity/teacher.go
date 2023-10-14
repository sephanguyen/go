package entity

import (
	"github.com/jackc/pgtype"
)

type Teacher struct {
	LegacyUser   `sql:"-"`
	ID           pgtype.Text      `sql:"teacher_id,pk"`
	SchoolIDs    pgtype.Int4Array `sql:"school_ids"`
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *Teacher) FieldMap() ([]string, []interface{}) {
	return []string{
			"teacher_id",
			"school_ids",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.ID,
			&e.SchoolIDs,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *Teacher) TableName() string {
	return "teachers"
}

// func (e *Teacher) IsInSchool(schoolID int32) bool {
// 	for _, id := range e.SchoolIDs.Elements {
// 		if id.Int == schoolID {
// 			return true
// 		}
// 	}

// 	return false
// }
