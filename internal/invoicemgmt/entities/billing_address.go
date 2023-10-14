package entities

import "github.com/jackc/pgtype"

type BillingAddress struct {
	BillingAddressID       pgtype.Text
	UserID                 pgtype.Text
	StudentPaymentDetailID pgtype.Text
	PostalCode             pgtype.Text
	City                   pgtype.Text
	Street1                pgtype.Text
	Street2                pgtype.Text
	CreatedAt              pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
	MigratedAt             pgtype.Timestamptz
	ResourcePath           pgtype.Text
	PrefectureCode         pgtype.Text
}

func (e *BillingAddress) FieldMap() ([]string, []interface{}) {
	return []string{
			"billing_address_id",
			"user_id",
			"student_payment_detail_id",
			"postal_code",
			"city",
			"street1",
			"street2",
			"created_at",
			"updated_at",
			"deleted_at",
			"migrated_at",
			"resource_path",
			"prefecture_code",
		}, []interface{}{
			&e.BillingAddressID,
			&e.UserID,
			&e.StudentPaymentDetailID,
			&e.PostalCode,
			&e.City,
			&e.Street1,
			&e.Street2,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.MigratedAt,
			&e.ResourcePath,
			&e.PrefectureCode,
		}
}

func (e *BillingAddress) TableName() string {
	return "billing_address"
}
