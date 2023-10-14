package repo

import (
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"

	"github.com/jackc/pgtype"
)

type Grade struct {
	ID                pgtype.Text
	Name              pgtype.Text
	IsArchived        pgtype.Bool
	PartnerInternalID pgtype.Text
	Sequence          pgtype.Int4
	Remarks           pgtype.Text
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (g *Grade) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"grade_id",
		"name",
		"is_archived",
		"partner_internal_id",
		"sequence",
		"remarks",
		"updated_at",
		"created_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&g.ID,
		&g.Name,
		&g.IsArchived,
		&g.PartnerInternalID,
		&g.Sequence,
		&g.Remarks,
		&g.UpdatedAt,
		&g.CreatedAt,
		&g.DeletedAt,
		&g.ResourcePath,
	}
	return
}

func (g *Grade) TableName() string {
	return "grade"
}

func (g *Grade) ToGradeEntity() *domain.Grade {
	newGrade := &domain.Grade{
		ID:                g.ID.String,
		PartnerInternalID: g.PartnerInternalID.String,
		Name:              g.Name.String,
		IsArchived:        g.IsArchived.Bool,
		Sequence:          int(g.Sequence.Int),
		UpdatedAt:         g.UpdatedAt.Time,
		CreatedAt:         g.CreatedAt.Time,
		Remarks:           g.Remarks.String,
		ResourcePath:      g.ResourcePath.String,
	}
	if g.DeletedAt.Status == pgtype.Present {
		newGrade.DeletedAt = &g.DeletedAt.Time
	}
	return newGrade
}
