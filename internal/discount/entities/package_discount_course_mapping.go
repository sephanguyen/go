package entities

import "github.com/jackc/pgtype"

type PackageDiscountCourseMapping struct {
	PackageID            pgtype.Text
	CourseCombinationIDs pgtype.Text
	DiscountTagID        pgtype.Text
	IsArchived           pgtype.Bool
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	ProductGroupID       pgtype.Text
	ResourcePath         pgtype.Text
}

func (e *PackageDiscountCourseMapping) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_id",
			"course_combination_ids",
			"discount_tag_id",
			"is_archived",
			"created_at",
			"updated_at",
			"product_group_id",
			"resource_path",
		}, []interface{}{
			&e.PackageID,
			&e.CourseCombinationIDs,
			&e.DiscountTagID,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ProductGroupID,
			&e.ResourcePath,
		}
}

func (e *PackageDiscountCourseMapping) TableName() string {
	return "package_discount_course_mapping"
}
