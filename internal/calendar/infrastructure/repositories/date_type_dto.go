package repositories

import (
	"github.com/manabie-com/backend/internal/calendar/domain/dto"

	"github.com/jackc/pgtype"
)

type DateType struct {
	DateTypeID   pgtype.Text
	DisplayName  pgtype.Text
	IsArchived   pgtype.Bool
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (dt *DateType) FieldMap() ([]string, []interface{}) {
	return []string{
			"day_type_id",
			"display_name",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&dt.DateTypeID,
			&dt.DisplayName,
			&dt.IsArchived,
			&dt.CreatedAt,
			&dt.UpdatedAt,
			&dt.DeletedAt,
			&dt.ResourcePath,
		}
}

func (dt *DateType) TableName() string {
	return "day_type"
}

func (dt *DateType) ConvertToDTO() *dto.DateType {
	return &dto.DateType{
		DateTypeID:  dt.DateTypeID.String,
		DisplayName: dt.DisplayName.String,
		IsArchived:  dt.IsArchived.Bool,
		CreatedAt:   dt.CreatedAt.Time,
		UpdatedAt:   dt.UpdatedAt.Time,
		DeletedAt:   dt.DeletedAt.Time,
	}
}
