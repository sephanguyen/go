package entities

import "github.com/jackc/pgtype"

type UserDiscountTag struct {
	UserID        pgtype.Text
	LocationID    pgtype.Text
	ProductID     pgtype.Text
	DiscountType  pgtype.Text
	StartDate     pgtype.Timestamptz
	EndDate       pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
	DiscountTagID pgtype.Text
}

func (p *UserDiscountTag) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_id",
		"location_id",
		"product_id",
		"discount_type",
		"start_date",
		"end_date",
		"created_at",
		"updated_at",
		"deleted_at",
		"resource_path",
		"discount_tag_id",
	}
	values = []interface{}{
		&p.UserID,
		&p.LocationID,
		&p.ProductID,
		&p.DiscountType,
		&p.StartDate,
		&p.EndDate,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
		&p.ResourcePath,
		&p.DiscountTagID,
	}
	return
}
func (p *UserDiscountTag) TableName() string {
	return "user_discount_tag"
}
