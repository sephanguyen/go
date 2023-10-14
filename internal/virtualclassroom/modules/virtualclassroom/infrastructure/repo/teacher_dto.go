package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type Teacher struct {
	ID        pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (t *Teacher) FieldMap() ([]string, []interface{}) {
	return []string{
			"teacher_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.ID,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

func (t *Teacher) TableName() string {
	return "teachers"
}

func (t *Teacher) ToTeacherDomain() *domain.Teacher {
	return &domain.Teacher{
		ID:        t.ID.String,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
		DeletedAt: t.DeletedAt.Time,
	}
}
