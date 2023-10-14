package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type Course struct {
	ID        pgtype.Text
	Name      pgtype.Text
	Status    pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *Course) FieldMap() ([]string, []interface{}) {
	return []string{
			"course_id",
			"name",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&c.ID,
			&c.Name,
			&c.Status,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		}
}

func (c *Course) TableName() string {
	return "courses"
}

func (c *Course) ToCourseDomain() *domain.Course {
	return &domain.Course{
		ID:        c.ID.String,
		Name:      c.Name.String,
		Status:    c.Status.String,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
		DeletedAt: c.DeletedAt.Time,
	}
}
