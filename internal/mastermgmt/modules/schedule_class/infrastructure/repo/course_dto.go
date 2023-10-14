package repo

import "github.com/jackc/pgtype"

type Course struct {
	CourseID pgtype.Text
	Name     pgtype.Text
}

func (c *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "name"}
	values = []interface{}{&c.CourseID, &c.Name}
	return
}

func (c *Course) TableName() string {
	return "courses"
}
