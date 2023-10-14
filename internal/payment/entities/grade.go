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
}

func (g *Grade) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"grade_id",
		"name",
		"is_archived",
		"partner_internal_id",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&g.ID,
		&g.Name,
		&g.IsArchived,
		&g.PartnerInternalID,
		&g.UpdatedAt,
		&g.CreatedAt,
		&g.DeletedAt,
	}
	return
}

func (g *Grade) TableName() string {
	return "grade"
}
