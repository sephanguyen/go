package entities

import "github.com/jackc/pgtype"

type CompanyDetail struct {
	CompanyDetailID    pgtype.Text
	CompanyName        pgtype.Text
	CompanyAddress     pgtype.Text
	CompanyPhoneNumber pgtype.Text
	CompanyLogoURL     pgtype.Text
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
	DeletedAt          pgtype.Timestamptz
	ResourcePath       pgtype.Text
}

func (e *CompanyDetail) FieldMap() ([]string, []interface{}) {
	return []string{
			"company_detail_id",
			"company_name",
			"company_address",
			"company_phone_number",
			"company_logo_url",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.CompanyDetailID,
			&e.CompanyName,
			&e.CompanyAddress,
			&e.CompanyPhoneNumber,
			&e.CompanyLogoURL,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *CompanyDetail) TableName() string {
	return "company_detail"
}
