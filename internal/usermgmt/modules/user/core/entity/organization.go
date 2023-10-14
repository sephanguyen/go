package entity

import "github.com/jackc/pgtype"

type Organization struct {
	OrganizationID      pgtype.Text
	TenantID            pgtype.Text
	Name                pgtype.Text
	ScryptSignerKey     pgtype.Text
	ScryptSaltSeparator pgtype.Text
	ScryptRounds        pgtype.Text
	ScryptMemoryCost    pgtype.Text
}

func (e *Organization) FieldMap() ([]string, []interface{}) {
	fieldNames := []string{"organization_id", "tenant_id", "name", "scrypt_signer_key", "scrypt_salt_separator", "scrypt_rounds", "scrypt_memory_cost"}
	fieldValues := []interface{}{&e.OrganizationID, &e.TenantID, &e.Name, &e.ScryptSignerKey, &e.ScryptSaltSeparator, &e.ScryptRounds, &e.ScryptMemoryCost}
	return fieldNames, fieldValues
}

// TableName returns "organizations"
func (e *Organization) TableName() string {
	return "organizations"
}

func (e *Organization) TempTableName() string {
	return "temp_organizations"
}
