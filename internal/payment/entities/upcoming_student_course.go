package entities

import "github.com/jackc/pgtype"

type UpcomingStudentCourse struct {
	UpcomingStudentPackageID pgtype.Text
	StudentPackageID         pgtype.Text
	StudentID                pgtype.Text
	CourseID                 pgtype.Text
	LocationID               pgtype.Text
	StudentStartDate         pgtype.Timestamptz
	StudentEndDate           pgtype.Timestamptz
	CourseSlot               pgtype.Int4
	CourseSlotPerWeek        pgtype.Int4
	Weight                   pgtype.Int4
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
	DeletedAt                pgtype.Timestamptz
	PackageType              pgtype.Text
	ResourcePath             pgtype.Text
	ExecutedError            pgtype.Text
	IsExecutedByCronjob      pgtype.Bool
}

func (p *UpcomingStudentCourse) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"upcoming_student_package_id",
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
		"executed_error",
		"is_executed_by_cronjob",
	}
	values = []interface{}{
		&p.UpcomingStudentPackageID,
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
		&p.ExecutedError,
		&p.IsExecutedByCronjob,
	}
	return
}
func (p *UpcomingStudentCourse) TableName() string {
	return "upcoming_student_course"
}
