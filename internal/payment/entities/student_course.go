package entities

import "github.com/jackc/pgtype"

type StudentCourse struct {
	StudentPackageID  pgtype.Text
	StudentID         pgtype.Text
	CourseID          pgtype.Text
	LocationID        pgtype.Text
	StudentStartDate  pgtype.Timestamptz
	StudentEndDate    pgtype.Timestamptz
	CourseSlot        pgtype.Int4
	CourseSlotPerWeek pgtype.Int4
	Weight            pgtype.Int4
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	PackageType       pgtype.Text
	ResourcePath      pgtype.Text
}

func (p *StudentCourse) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_package_id",
		"student_id",
		"course_id",
		"location_id",
		"student_start_date",
		"student_end_date",
		"course_slot",
		"course_slot_per_week",
		"weight",
		"created_at",
		"updated_at",
		"deleted_at",
		"package_type",
		"resource_path",
	}
	values = []interface{}{
		&p.StudentPackageID,
		&p.StudentID,
		&p.CourseID,
		&p.LocationID,
		&p.StudentStartDate,
		&p.StudentEndDate,
		&p.CourseSlot,
		&p.CourseSlotPerWeek,
		&p.Weight,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
		&p.PackageType,
		&p.ResourcePath,
	}
	return
}
func (p *StudentCourse) TableName() string {
	return "student_course"
}
