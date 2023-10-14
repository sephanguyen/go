package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type CourseTypeDTO struct {
	ID        pgtype.Text
	Name      pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *CourseTypeDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_type_id",
		"name",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&c.ID,
		&c.Name,
		&c.UpdatedAt,
		&c.CreatedAt,
		&c.DeletedAt,
	}
	return
}

func (*CourseTypeDTO) TableName() string {
	return "course_type"
}

func (c *CourseTypeDTO) ToCourseTypeEntity() *domain.CourseType {
	course := &domain.CourseType{
		CourseTypeID: c.ID.String,
		Name:         c.Name.String,
		CreatedAt:    c.CreatedAt.Time,
		UpdatedAt:    c.UpdatedAt.Time,
	}
	if c.DeletedAt.Status == pgtype.Present {
		course.DeletedAt = &c.DeletedAt.Time
	}
	return course
}

func NewCourseTypeFromEntity(c *domain.CourseType) (*CourseTypeDTO, error) {
	dto := &CourseTypeDTO{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.ID.Set(c.CourseTypeID),
		dto.Name.Set(c.Name),
		dto.CreatedAt.Set(c.CreatedAt),
		dto.UpdatedAt.Set(c.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from course type entity to course dto: %w", err)
	}
	return dto, nil
}
