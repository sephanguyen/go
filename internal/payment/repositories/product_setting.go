package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type ProductSettingRepo struct{}

func (r *ProductSettingRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.ProductSetting) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductSettingRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert ProductSetting: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert ProductSetting: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *ProductSettingRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.ProductSetting) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductSettingRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "product_id", []string{"is_enrollment_required", "is_pausable", "is_added_to_enrollment_by_default", "is_operation_fee", "updated_at"})
	if err != nil {
		return fmt.Errorf("err update ProductSetting: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update ProductSetting: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *ProductSettingRepo) GetByID(ctx context.Context, db database.QueryExecer, productID string) (entities.ProductSetting, error) {
	productSetting := &entities.ProductSetting{}
	fieldNames, fieldValues := productSetting.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		productSetting.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.ProductSetting{}, err
	}
	return *productSetting, nil
}
