package entities

import "github.com/jackc/pgtype"

type StudentPackageLog struct {
	StudentPackageLogID  pgtype.Int4
	StudentPackageID     pgtype.Text
	UserID               pgtype.Text
	Action               pgtype.Text
	Flow                 pgtype.Text
	StudentPackageObject pgtype.JSONB
	StudentID            pgtype.Text
	CourseID             pgtype.Text
	CreatedAt            pgtype.Timestamptz
	ResourcePath         pgtype.Text
}

func (e *StudentPackageLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_package_log_id",
			"student_package_id",
			"user_id",
			"action",
			"flow",
			"student_package_object",
			"student_id",
			"course_id",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.StudentPackageLogID,
			&e.StudentPackageID,
			&e.UserID,
			&e.Action,
			&e.Flow,
			&e.StudentPackageObject,
			&e.StudentID,
			&e.CourseID,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *StudentPackageLog) GetStudentPackageObject() (*StudentPackages, error) {
	pp := &StudentPackages{}
	err := e.StudentPackageObject.AssignTo(pp)
	return pp, err
}

func (e *StudentPackageLog) TableName() string {
	return "student_package_log"
}
