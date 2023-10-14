package entities

import "github.com/jackc/pgtype"

type PackageDiscountSetting struct {
	PackageID      pgtype.Text
	MinSlotTrigger pgtype.Int4
	MaxSlotTrigger pgtype.Int4
	DiscountTagID  pgtype.Text
	IsArchived     pgtype.Bool
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	ProductGroupID pgtype.Text
	ResourcePath   pgtype.Text
}

func (e *PackageDiscountSetting) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_id",
			"min_slot_trigger",
			"max_slot_trigger",
			"discount_tag_id",
			"is_archived",
			"created_at",
			"updated_at",
			"product_group_id",
			"resource_path",
		}, []interface{}{
			&e.PackageID,
			&e.MinSlotTrigger,
			&e.MaxSlotTrigger,
			&e.DiscountTagID,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ProductGroupID,
			&e.ResourcePath,
		}
}

func (e *PackageDiscountSetting) TableName() string {
	return "package_discount_setting"
}
