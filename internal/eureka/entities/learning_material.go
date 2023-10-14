package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
)

type LearningMaterial struct {
	BaseEntity
	ID           pgtype.Text `sql:"learning_material_id,pk"`
	TopicID      pgtype.Text `sql:"topic_id"`
	Name         pgtype.Text
	Type         pgtype.Text
	DisplayOrder pgtype.Int2
	VendorType   pgtype.Text
	IsPublished  pgtype.Bool
}

func (t *LearningMaterial) FieldMap() ([]string, []interface{}) {
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
			&t.ID,
			&t.TopicID,
			&t.Name,
			&t.Type,
			&t.DisplayOrder,
			&t.UpdatedAt,
			&t.CreatedAt,
			&t.DeletedAt,
			&t.VendorType,
			&t.IsPublished,
		}
}

func (t *LearningMaterial) TableName() string {
	return "learning_material"
}

func (t *LearningMaterial) SetDefaultVendorType() error {
	return t.VendorType.Set(sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String())
}

type LearningMaterials []*LearningMaterial

func (u *LearningMaterials) Add() database.Entity {
	e := &LearningMaterial{}
	*u = append(*u, e)

	return e
}

func (u LearningMaterials) Get() []*LearningMaterial {
	return []*LearningMaterial(u)
}
