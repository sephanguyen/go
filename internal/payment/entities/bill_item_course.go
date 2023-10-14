package entities

import "github.com/jackc/pgtype"

type BillItemCourse struct {
	BillItemSequenceNumber pgtype.Int4
	CourseID               pgtype.Text
	CourseName             pgtype.Text
	CourseWeight           pgtype.Int4
	CourseSlot             pgtype.Int4
	CreatedAt              pgtype.Timestamptz
	ResourcePath           pgtype.Text
}

func (e *BillItemCourse) FieldMap() ([]string, []interface{}) {
	return []string{
			"bill_item_sequence_number",
			"course_id",
			"course_name",
			"course_weight",
			"course_slot",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.BillItemSequenceNumber,
			&e.CourseID,
			&e.CourseName,
			&e.CourseWeight,
			&e.CourseSlot,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *BillItemCourse) TableName() string {
	return "bill_item_course"
}
