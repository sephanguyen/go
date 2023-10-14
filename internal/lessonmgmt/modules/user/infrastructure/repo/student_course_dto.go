package repo

import (
	"github.com/jackc/pgtype"
)

type StudentCourse struct {
	StudentID         pgtype.Text
	CourseID          pgtype.Text
	LocationID        pgtype.Text
	CourseSlot        pgtype.Int4
	CourseSlotPerWeek pgtype.Int4
}

func (u *StudentCourse) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"course_id",
			"location_id",
			"course_slot",
			"course_slot_per_week",
		}, []interface{}{
			&u.StudentID,
			&u.CourseID,
			&u.LocationID,
			&u.CourseSlot,
			&u.CourseSlotPerWeek,
		}
}

func (u *StudentCourse) TableName() string {
	return "student_course"
}
