package entities

import "github.com/jackc/pgtype"

type StudentDiscountTracker struct {
	DiscountTrackerID            pgtype.Text
	StudentID                    pgtype.Text
	LocationID                   pgtype.Text
	StudentProductID             pgtype.Text
	ProductID                    pgtype.Text
	ProductGroupID               pgtype.Text
	DiscountType                 pgtype.Text
	DiscountStatus               pgtype.Text
	DiscountStartDate            pgtype.Timestamptz
	DiscountEndDate              pgtype.Timestamptz
	StudentProductStartDate      pgtype.Timestamptz
	StudentProductEndDate        pgtype.Timestamptz
	StudentProductStatus         pgtype.Text
	UpdatedFromDiscountTrackerID pgtype.Text
	UpdatedToDiscountTrackerID   pgtype.Text
	CreatedAt                    pgtype.Timestamptz
	UpdatedAt                    pgtype.Timestamptz
	DeletedAt                    pgtype.Timestamptz
	ResourcePath                 pgtype.Text
}

func (p *StudentDiscountTracker) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"discount_tracker_id",
		"student_id",
		"location_id",
		"student_product_id",
		"product_id",
		"product_group_id",
		"discount_type",
		"discount_status",
		"discount_start_date",
		"discount_end_date",
		"student_product_start_date",
		"student_product_end_date",
		"student_product_status",
		"updated_from_discount_tracker_id",
		"updated_to_discount_tracker_id",
		"created_at",
		"updated_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&p.DiscountTrackerID,
		&p.StudentID,
		&p.LocationID,
		&p.StudentProductID,
		&p.ProductID,
		&p.ProductGroupID,
		&p.DiscountType,
		&p.DiscountStatus,
		&p.DiscountStartDate,
		&p.DiscountEndDate,
		&p.StudentProductStartDate,
		&p.StudentProductEndDate,
		&p.StudentProductStatus,
		&p.UpdatedFromDiscountTrackerID,
		&p.UpdatedToDiscountTrackerID,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
		&p.ResourcePath,
	}
	return
}
func (p *StudentDiscountTracker) TableName() string {
	return "student_discount_tracker"
}
