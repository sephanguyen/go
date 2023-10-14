package entities

import "github.com/jackc/pgtype"

type ProductSetting struct {
	ProductID                    pgtype.Text
	IsEnrollmentRequired         pgtype.Bool
	IsPausable                   pgtype.Bool
	IsAddedToEnrollmentByDefault pgtype.Bool
	IsOperationFee               pgtype.Bool
	CreatedAt                    pgtype.Timestamptz
	UpdatedAt                    pgtype.Timestamptz
	ResourcePath                 pgtype.Text
}

func (e *ProductSetting) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_id",
			"is_enrollment_required",
			"is_pausable",
			"is_added_to_enrollment_by_default",
			"is_operation_fee",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&e.ProductID,
			&e.IsEnrollmentRequired,
			&e.IsPausable,
			&e.IsAddedToEnrollmentByDefault,
			&e.IsOperationFee,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductSetting) TableName() string {
	return "product_setting"
}
