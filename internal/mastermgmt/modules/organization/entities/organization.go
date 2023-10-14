package entities

import "github.com/jackc/pgtype"

type Organization struct {
	ID           pgtype.Text `sql:"organization_id"`
	TenantID     pgtype.Text
	Name         pgtype.Text
	ResourcePath pgtype.Text
	DomainName   pgtype.Text
	LogoURL      pgtype.Text
	Country      pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *Organization) FieldMap() ([]string, []interface{}) {
	return []string{
			"organization_id",
			"tenant_id",
			"name",
			"resource_path",
			"domain_name",
			"logo_url",
			"country",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.ID,
			&e.TenantID,
			&e.Name,
			&e.ResourcePath,
			&e.DomainName,
			&e.LogoURL,
			&e.Country,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *Organization) TableName() string {
	return "organizations"
}
