package entities

import (
	"github.com/jackc/pgtype"
)

type PackageCourse struct {
	PackageID         pgtype.Text
	CourseID          pgtype.Text
	MandatoryFlag     pgtype.Bool
	CourseWeight      pgtype.Int4
	MaxSlotsPerCourse pgtype.Int4
	CreatedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (pc *PackageCourse) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"package_id",
		"course_id",
		"mandatory_flag",
		"course_weight",
		"max_slots_per_course",
		"created_at",
		"resource_path",
	}
	values = []interface{}{
		&pc.PackageID,
		&pc.CourseID,
		&pc.MandatoryFlag,
		&pc.CourseWeight,
		&pc.MaxSlotsPerCourse,
		&pc.CreatedAt,
		&pc.ResourcePath,
	}
	return
}

func (pc *PackageCourse) TableName() string {
	return "package_course"
}
