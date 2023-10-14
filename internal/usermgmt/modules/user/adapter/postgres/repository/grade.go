package repository

import (
	"github.com/jackc/pgtype"
)

type GradeEntity struct {
	ID                pgtype.Text
	Name              pgtype.Text
	IsArchived        pgtype.Bool
	PartnerInternalID pgtype.Text
	Sequence          pgtype.Int4
	OrganizationID    pgtype.Text
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (g *GradeEntity) FieldMap() ([]string, []interface{}) {
	return []string{
			"grade_id",
			"name",
			"is_archived",
			"partner_internal_id",
			"sequence",
			"resource_path",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&g.ID,
			&g.Name,
			&g.IsArchived,
			&g.PartnerInternalID,
			&g.Sequence,
			&g.OrganizationID,
			&g.UpdatedAt,
			&g.CreatedAt,
			&g.DeletedAt,
		}
}

func (g *GradeEntity) TableName() string {
	return "grade"
}
