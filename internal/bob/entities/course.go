package entities

import (
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type Course struct {
	ID                pgtype.Text `sql:"course_id,pk"`
	Name              pgtype.Text
	Country           pgtype.Text
	Subject           pgtype.Text
	Grade             pgtype.Int2
	DisplayOrder      pgtype.Int2
	SchoolID          pgtype.Int4      `sql:"school_id"`
	TeacherIDs        pgtype.TextArray `sql:"teacher_ids"`
	CourseType        pgtype.Text
	Icon              pgtype.Text
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	StartDate         pgtype.Timestamptz
	EndDate           pgtype.Timestamptz
	PresetStudyPlanID pgtype.Text `sql:"preset_study_plan_id"`
	Status            pgtype.Text
}

func (c *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"name",
		"country",
		"subject",
		"grade",
		"display_order",
		"school_id",
		"teacher_ids",
		"course_type",
		"updated_at",
		"created_at",
		"deleted_at",
		"start_date",
		"end_date",
		"preset_study_plan_id",
		"icon",
		"status",
	}
	values = []interface{}{
		&c.ID,
		&c.Name,
		&c.Country,
		&c.Subject,
		&c.Grade,
		&c.DisplayOrder,
		&c.SchoolID,
		&c.TeacherIDs,
		&c.CourseType,
		&c.UpdatedAt,
		&c.CreatedAt,
		&c.DeletedAt,
		&c.StartDate,
		&c.EndDate,
		&c.PresetStudyPlanID,
		&c.Icon,
		&c.Status,
	}
	return
}

func (*Course) TableName() string {
	return "courses"
}

type Courses []*Course

func (u *Courses) Add() database.Entity {
	e := &Course{}
	*u = append(*u, e)

	return e
}

type CourseAvailableRanges struct {
	data map[string]*CourseAvailableRange
}

func (c *CourseAvailableRanges) Add(items ...*CourseAvailableRange) {
	if c.data == nil {
		c.data = make(map[string]*CourseAvailableRange)
	}

	for _, item := range items {
		c.data[item.ID.String] = item
	}
}

func (c *CourseAvailableRanges) Get(courseID pgtype.Text) *CourseAvailableRange {
	return c.data[courseID.String]
}

func (c *CourseAvailableRanges) GetIDs() pgtype.TextArray {
	res := make([]string, 0, len(c.data))
	for k := range c.data {
		res = append(res, k)
	}

	return database.TextArray(res)
}

func (c *CourseAvailableRanges) GetArray() []*CourseAvailableRange {
	res := make([]*CourseAvailableRange, 0, len(c.data))
	for k := range c.data {
		res = append(res, c.data[k])
	}

	return res
}

type CourseAvailableRange struct {
	ID        pgtype.Text
	StartDate pgtype.Timestamptz
	EndDate   pgtype.Timestamptz
}
