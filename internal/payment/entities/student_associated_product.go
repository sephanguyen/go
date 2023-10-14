package entities

import "github.com/jackc/pgtype"

type StudentAssociatedProduct struct {
	StudentProductID    pgtype.Text
	AssociatedProductID pgtype.Text
	CreatedAt           pgtype.Timestamptz
	DeletedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
	ResourcePath        pgtype.Text
}

func (e *StudentAssociatedProduct) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_product_id",
			"associated_product_id",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.StudentProductID,
			&e.AssociatedProductID,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *StudentAssociatedProduct) TableName() string {
	return "student_associated_product"
}
