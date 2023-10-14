package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"

	"github.com/jackc/pgtype"
)

type LearningMaterialDto struct {
	ID           pgtype.Text `sql:"learning_material_id,pk"`
	TopicID      pgtype.Text `sql:"topic_id"`
	Name         pgtype.Text
	Type         pgtype.Text
	DisplayOrder pgtype.Int2
	VendorType   pgtype.Text
	IsPublished  pgtype.Bool
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

func (lm *LearningMaterialDto) FieldMap() ([]string, []interface{}) {
	return []string{
			"learning_material_id",
			"topic_id",
			"name",
			"type",
			"display_order",
			"updated_at",
			"created_at",
			"deleted_at",
			"vendor_type",
			"is_published",
		}, []interface{}{
			&lm.ID,
			&lm.TopicID,
			&lm.Name,
			&lm.Type,
			&lm.DisplayOrder,
			&lm.UpdatedAt,
			&lm.CreatedAt,
			&lm.DeletedAt,
			&lm.VendorType,
			&lm.IsPublished,
		}
}

func (lm *LearningMaterialDto) TableName() string {
	return "learning_material"
}

func (lm *LearningMaterialDto) ToEntity() domain.LearningMaterial {
	return domain.LearningMaterial{
		ID:           lm.ID.String,
		TopicID:      lm.TopicID.String,
		Name:         lm.Name.String,
		Type:         constants.LearningMaterialType(lm.Type.String),
		DisplayOrder: int(lm.DisplayOrder.Int),
		Published:    lm.IsPublished.Bool,
		CreatedAt:    lm.CreatedAt.Time,
		UpdatedAt:    lm.UpdatedAt.Time,
		DeletedAt:    nil,
	}
}
