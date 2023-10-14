package entities

import "github.com/jackc/pgtype"

type OrderItemCourse struct {
	OrderID           pgtype.Text
	PackageID         pgtype.Text
	CourseID          pgtype.Text
	CourseName        pgtype.Text
	CourseSlot        pgtype.Int4
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	OrderItemCourseID pgtype.Text
	ResourcePath      pgtype.Text
}

func (e *OrderItemCourse) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_id",
			"package_id",
			"course_id",
			"course_name",
			"course_slot",
			"created_at",
			"updated_at",
			"order_item_course_id",
			"resource_path",
		}, []interface{}{
			&e.OrderID,
			&e.PackageID,
			&e.CourseID,
			&e.CourseName,
			&e.CourseSlot,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.OrderItemCourseID,
			&e.ResourcePath,
		}
}

func (e *OrderItemCourse) TableName() string {
	return "order_item_course"
}
