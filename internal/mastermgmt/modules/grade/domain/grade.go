package domain

import (
	"time"
)

type Grade struct {
	ID                string
	Name              string
	IsArchived        bool
	PartnerInternalID string
	Sequence          int
	Remarks           string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	DeletedAt         *time.Time
	ResourcePath      string
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
func (g *Grade) FieldMapWithoutRP() (fields []string, values []interface{}) {
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
	}
	return
}

func (g *Grade) TableName() string {
	return "grade"
}

func (g *Grade) ExportFieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"grade_id",
		"partner_internal_id",
		"name",
		"is_archived",
		"sequence",
		"remarks",
	}
	values = []interface{}{
		&g.ID,
		&g.PartnerInternalID,
		&g.Name,
		&g.IsArchived,
		&g.Sequence,
		&g.Remarks,
	}
	return
}
