package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type CourseType struct {
	ID         pgtype.Text
	Name       pgtype.Text
	IsArchived pgtype.Bool
	Remarks    pgtype.Text
	UpdatedAt  pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (c *CourseType) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_type_id",
		"name",
		"is_archived",
		"remarks",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&c.ID,
		&c.Name,
		&c.IsArchived,
		&c.Remarks,
		&c.UpdatedAt,
		&c.CreatedAt,
		&c.DeletedAt,
	}
	return
}

func (*CourseType) TableName() string {
	return "course_type"
}

func (c *CourseType) ToCourseTypeEntity() *domain.CourseType {
	course := &domain.CourseType{
		CourseTypeID: c.ID.String,
		Name:         c.Name.String,
		IsArchived:   c.IsArchived.Bool,
		Remarks:      c.Remarks.String,
		CreatedAt:    c.CreatedAt.Time,
		UpdatedAt:    c.UpdatedAt.Time,
	}
	if c.DeletedAt.Status == pgtype.Present {
		course.DeletedAt = &c.DeletedAt.Time
	}
	return course
}

func NewCourseTypeFromEntity(c *domain.CourseType) (*CourseType, error) {
	dto := &CourseType{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.ID.Set(c.CourseTypeID),
		dto.Name.Set(c.Name),
		dto.IsArchived.Set(c.IsArchived),
		dto.Remarks.Set(c.Remarks),
		dto.CreatedAt.Set(c.CreatedAt),
		dto.UpdatedAt.Set(c.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not map course type entity to course dto: %w", err)
	}
	return dto, nil
}
