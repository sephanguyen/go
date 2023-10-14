package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type OldClass struct {
	ID        pgtype.Int4 `sql:"class_id,pk"`
	Name      pgtype.Text
	Status    pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (o *OldClass) FieldMap() ([]string, []interface{}) {
	return []string{
			"class_id",
			"name",
			"status",
			"created_at",
			"updated_at",
		}, []interface{}{
			&o.ID,
			&o.Name,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
		}
}

func (o *OldClass) TableName() string {
	return "classes"
}

func (o *OldClass) ToOldClassDomain() *domain.OldClass {
	return &domain.OldClass{
		ID:        o.ID.Int,
		Name:      o.Name.String,
		Status:    o.Status.String,
		CreatedAt: o.CreatedAt.Time,
		UpdatedAt: o.UpdatedAt.Time,
	}
}
