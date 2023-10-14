package domain

import "time"

// use for exporting classes (using join)
// actually a dto, but put it in repo will cause import cycle
type ExportingClass struct {
	ClassID      string
	Name         string
	CourseID     string
	LocationID   string
	SchoolID     string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
	ResourcePath string

	CourseName   string
	LocationName string
}

func (e *ExportingClass) TableName() string {
	return "grade"
}

func (e *ExportingClass) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at", "deleted_at", "course_name", "location_name"}
	values = []interface{}{&e.ClassID, &e.Name, &e.CourseID, &e.LocationID, &e.SchoolID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.CourseName, &e.LocationName}

	return
}
