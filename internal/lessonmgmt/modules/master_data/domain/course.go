package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Course struct {
	CourseID       pgtype.Text
	Name           pgtype.Text
	TeachingMethod pgtype.Text
}

func (c *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "name", "teaching_method"}
	values = []interface{}{&c.CourseID, &c.Name, &c.TeachingMethod}
	return
}

func (c *Course) TableName() string {
	return "courses"
}

type CourseRepository interface {
	GetByIDs(ctx context.Context, db database.Ext, id []string) ([]*Course, error)
}
