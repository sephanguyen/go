package entities

import "github.com/jackc/pgtype"

type Organization struct {
	OrganizationID pgtype.Text
	TenantID       pgtype.Text
	Name           pgtype.Text
	ResourcePath   pgtype.Text
	Country        pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *Organization) FieldMap() ([]string, []interface{}) {
	return []string{
			"organization_id",
			"tenant_id",
			"name",
			"resource_path",
			"country",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.OrganizationID,
			&e.TenantID,
			&e.Name,
			&e.ResourcePath,
			&e.Country,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *Organization) TableName() string {
	return "organizations"
}
