package entities

import (
	"github.com/jackc/pgtype"
)

type Grade struct {
	ID                pgtype.Text
	Name              pgtype.Text
	IsArchived        pgtype.Bool
	PartnerInternalID pgtype.Varchar
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
	Sequence          pgtype.Int4
}

func (e *Grade) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"grade_id",
		"name",
		"is_archived",
		"partner_internal_id",
		"updated_at",
		"created_at",
		"deleted_at",
		"resource_path",
		"sequence",
	}
	values = []interface{}{
		&e.ID,
		&e.Name,
		&e.IsArchived,
		&e.PartnerInternalID,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
		&e.ResourcePath,
		&e.Sequence,
	}
	return
}

func (e *Grade) TableName() string {
	return "grade"
}
