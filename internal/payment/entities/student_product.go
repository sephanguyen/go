package entities

import "github.com/jackc/pgtype"

type StudentProduct struct {
	StudentProductID            pgtype.Text
	StudentID                   pgtype.Text
	ProductID                   pgtype.Text
	UpcomingBillingDate         pgtype.Timestamptz
	StartDate                   pgtype.Timestamptz
	EndDate                     pgtype.Timestamptz
	ProductStatus               pgtype.Text
	ApprovalStatus              pgtype.Text
	UpdatedAt                   pgtype.Timestamptz
	CreatedAt                   pgtype.Timestamptz
	DeletedAt                   pgtype.Timestamptz
	ResourcePath                pgtype.Text
	LocationID                  pgtype.Text
	UpdatedFromStudentProductID pgtype.Text
	UpdatedToStudentProductID   pgtype.Text
	StudentProductLabel         pgtype.Text
	IsUnique                    pgtype.Bool
	IsAssociated                pgtype.Bool
	RootStudentProductID        pgtype.Text
	VersionNumber               pgtype.Int4
}

func (p *StudentProduct) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_product_id",
		"student_id",
		"product_id",
		"upcoming_billing_date",
		"start_date",
		"end_date",
		"product_status",
		"approval_status",
		"updated_at",
		"created_at",
		"deleted_at",
		"location_id",
		"student_product_label",
		"updated_from_student_product_id",
		"updated_to_student_product_id",
		"is_unique",
		"root_student_product_id",
		"is_associated",
		"version_number",
		"resource_path",
	}
	values = []interface{}{
		&p.StudentProductID,
		&p.StudentID,
		&p.ProductID,
		&p.UpcomingBillingDate,
		&p.StartDate,
		&p.EndDate,
		&p.ProductStatus,
		&p.ApprovalStatus,
		&p.UpdatedAt,
		&p.CreatedAt,
		&p.DeletedAt,
		&p.LocationID,
		&p.StudentProductLabel,
		&p.UpdatedFromStudentProductID,
		&p.UpdatedToStudentProductID,
		&p.IsUnique,
		&p.RootStudentProductID,
		&p.IsAssociated,
		&p.VersionNumber,
		&p.ResourcePath,
	}
	return
}
func (p *StudentProduct) TableName() string {
	return "student_product"
}
