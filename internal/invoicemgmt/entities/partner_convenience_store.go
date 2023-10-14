package entities

import "github.com/jackc/pgtype"

type PartnerConvenienceStore struct {
	PartnerConvenienceStoreID pgtype.Text
	ManufacturerCode          pgtype.Int4
	CompanyCode               pgtype.Int4
	ShopCode                  pgtype.Text
	CompanyName               pgtype.Text
	CompanyTelNumber          pgtype.Text
	PostalCode                pgtype.Text
	Address1                  pgtype.Text
	Address2                  pgtype.Text
	Message1                  pgtype.Text
	Message2                  pgtype.Text
	Message3                  pgtype.Text
	Message4                  pgtype.Text
	Message5                  pgtype.Text
	Message6                  pgtype.Text
	Message7                  pgtype.Text
	Message8                  pgtype.Text
	Message9                  pgtype.Text
	Message10                 pgtype.Text
	Message11                 pgtype.Text
	Message12                 pgtype.Text
	Message13                 pgtype.Text
	Message14                 pgtype.Text
	Message15                 pgtype.Text
	Message16                 pgtype.Text
	Message17                 pgtype.Text
	Message18                 pgtype.Text
	Message19                 pgtype.Text
	Message20                 pgtype.Text
	Message21                 pgtype.Text
	Message22                 pgtype.Text
	Message23                 pgtype.Text
	Message24                 pgtype.Text
	Remarks                   pgtype.Text
	IsArchived                pgtype.Bool
	CreatedAt                 pgtype.Timestamptz
	UpdatedAt                 pgtype.Timestamptz
	DeletedAt                 pgtype.Timestamptz
	ResourcePath              pgtype.Text
}

func (e *PartnerConvenienceStore) FieldMap() ([]string, []interface{}) {
	return []string{
			"partner_convenience_store_id",
			"manufacturer_code",
			"company_code",
			"shop_code",
			"company_name",
			"company_tel_number",
			"postal_code",
			"address1",
			"address2",
			"message1",
			"message2",
			"message3",
			"message4",
			"message5",
			"message6",
			"message7",
			"message8",
			"message9",
			"message10",
			"message11",
			"message12",
			"message13",
			"message14",
			"message15",
			"message16",
			"message17",
			"message18",
			"message19",
			"message20",
			"message21",
			"message22",
			"message23",
			"message24",
			"remarks",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.PartnerConvenienceStoreID,
			&e.ManufacturerCode,
			&e.CompanyCode,
			&e.ShopCode,
			&e.CompanyName,
			&e.CompanyTelNumber,
			&e.PostalCode,
			&e.Address1,
			&e.Address2,
			&e.Message1,
			&e.Message2,
			&e.Message3,
			&e.Message4,
			&e.Message5,
			&e.Message6,
			&e.Message7,
			&e.Message8,
			&e.Message9,
			&e.Message10,
			&e.Message11,
			&e.Message12,
			&e.Message13,
			&e.Message14,
			&e.Message15,
			&e.Message16,
			&e.Message17,
			&e.Message18,
			&e.Message19,
			&e.Message20,
			&e.Message21,
			&e.Message22,
			&e.Message23,
			&e.Message24,
			&e.Remarks,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (*PartnerConvenienceStore) TableName() string {
	return "partner_convenience_store"
}
