package entities

import "github.com/jackc/pgtype"

type ProductGrade struct {
	ProductID    pgtype.Text
	GradeID      pgtype.Text
	CreatedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *ProductGrade) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_id",
			"grade_id",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.ProductID,
			&e.GradeID,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductGrade) TableName() string {
	return "product_grade"
}
