package entities

import "github.com/jackc/pgtype"

type Class struct {
	ClassID      pgtype.Text
	CourseID     pgtype.Text
	LocationID   pgtype.Text
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (g *Class) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"class_id",
		"course_id",
		"location_id",
		"updated_at",
		"created_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&g.ClassID,
		&g.CourseID,
		&g.LocationID,
		&g.UpdatedAt,
		&g.CreatedAt,
		&g.DeletedAt,
		&g.ResourcePath,
	}
	return
}

func (*Class) TableName() string {
	return "class"
}
