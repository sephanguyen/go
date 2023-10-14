package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type CourseLocationSchedule struct {
	CourseLocationScheduleID pgtype.Text
	CourseID                 pgtype.Text
	LocationID               pgtype.Text
	AcademiWeeks             pgtype.TextArray
	ProductTypeSchedule      pgtype.Text
	Frequency                pgtype.Int2
	TotalNoLessons           pgtype.Int2
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
}

func (c *CourseLocationSchedule) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_location_schedule_id",
		"course_id",
		"location_id",
		"academic_weeks",
		"product_type_schedule",
		"frequency",
		"total_no_lessons",
		"created_at",
		"updated_at",
	}
	values = []interface{}{
		&c.CourseLocationScheduleID,
		&c.CourseID,
		&c.LocationID,
		&c.AcademiWeeks,
		&c.ProductTypeSchedule,
		&c.Frequency,
		&c.TotalNoLessons,
		&c.CreatedAt,
		&c.UpdatedAt,
	}
	return
}

func (*CourseLocationSchedule) TableName() string {
	return "course_location_schedule"
}

type CoursesLocationSchedule []*CourseLocationSchedule

func (c *CoursesLocationSchedule) Add() database.Entity {
	e := &CourseLocationSchedule{}
	*c = append(*c, e)

	return e
}

func NewCourseLocationScheduleDTOFromCourseLocationScheduleDomain(c *domain.CourseLocationSchedule) (*CourseLocationSchedule, error) {
	dto := &CourseLocationSchedule{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.CourseLocationScheduleID.Set(c.ID),
		dto.CourseID.Set(c.CourseID),
		dto.LocationID.Set(c.LocationID),
		dto.AcademiWeeks.Set(c.AcademicWeeks),
		dto.ProductTypeSchedule.Set(c.ProductTypeSchedule),
		dto.Frequency.Set(c.Frequency),
		dto.TotalNoLessons.Set(c.TotalNoLesson),
		dto.CreatedAt.Set(c.CreatedAt),
		dto.UpdatedAt.Set(c.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from course location schedule domain to course location schedule dto: %w", err)
	}
	return dto, nil
}

func NewCourseLocationScheduleDomainFromCourseLocationScheduleDTO(c *CourseLocationSchedule) (*domain.CourseLocationSchedule, error) {
	freq := int(c.Frequency.Int)
	totalNoLesson := int(c.TotalNoLessons.Int)
	courseLocationSchedule := &domain.CourseLocationSchedule{
		ID:                  c.CourseLocationScheduleID.String,
		CourseID:            c.CourseID.String,
		LocationID:          c.LocationID.String,
		AcademicWeeks:       database.FromTextArray(c.AcademiWeeks),
		ProductTypeSchedule: domain.ProductTypeSchedule(c.ProductTypeSchedule.String),
		Frequency:           &freq,
		TotalNoLesson:       &totalNoLesson,
	}

	return courseLocationSchedule, nil
}
