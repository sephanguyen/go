package entity

import "github.com/jackc/pgtype"

type OrganizationAuth struct {
	OrganizationID pgtype.Int4
	AuthProjectID  pgtype.Text
	AuthTenantID   pgtype.Text
}

// TableName returns "students"
func (e *OrganizationAuth) TableName() string {
	return "organization_auths"
}

func (e *OrganizationAuth) TempTableName() string {
	return "temp_organization_auths"
}

func (e *OrganizationAuth) FieldMap() ([]string, []interface{}) {
	return []string{
			"organization_id", "auth_project_id", "auth_tenant_id",
		}, []interface{}{
			&e.OrganizationID, &e.AuthProjectID, &e.AuthTenantID,
		}
}
