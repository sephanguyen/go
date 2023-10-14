package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type CourseAccessPath struct {
	ID         pgtype.Varchar
	LocationID pgtype.Text
	CourseID   pgtype.Text
	UpdatedAt  pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (c *CourseAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_id", "course_id", "updated_at", "created_at", "deleted_at", "id"}
	values = []interface{}{&c.LocationID, &c.CourseID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt, &c.ID}
	return
}

func (*CourseAccessPath) TableName() string {
	return "course_access_paths"
}

func NewCourseAccessPathFromEntity(l *domain.CourseAccessPath) (*CourseAccessPath, error) {
	dto := &CourseAccessPath{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.LocationID.Set(l.LocationID),
		dto.CourseID.Set(l.CourseID),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
		dto.ID.Set(l.ID),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from course_access_path entity to course_access_path dto: %w", err)
	}
	return dto, nil
}

func (c *CourseAccessPath) ToCourseAccessPathEntity() *domain.CourseAccessPath {
	cap := &domain.CourseAccessPath{
		LocationID: c.LocationID.String,
		CourseID:   c.CourseID.String,
		CreatedAt:  c.CreatedAt.Time,
		UpdatedAt:  c.UpdatedAt.Time,
		ID:         c.ID.String,
	}
	if c.DeletedAt.Status == pgtype.Present {
		cap.DeletedAt = &c.DeletedAt.Time
	}
	return cap
}
