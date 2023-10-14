package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type MaterialRepo struct{}

func (r *MaterialRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, materialID string) (entities.Material, error) {
	material := &entities.Material{}
	materialFieldNames, materialFieldValues := material.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			material_id = $1
		FOR NO KEY UPDATE
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(materialFieldNames, ","),
		material.TableName(),
	)
	row := db.QueryRow(ctx, stmt, materialID)
	err := row.Scan(materialFieldValues...)
	if err != nil {
		return entities.Material{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *material, nil
}

// Create creates MaterialRepo entity
func (r *MaterialRepo) Create(ctx context.Context, tx database.QueryExecer, e *entities.Material) error {
	ctx, span := interceptors.StartSpan(ctx, "MaterialRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.ProductID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	var productID pgtype.Text
	err := database.InsertReturningAndExcept(ctx, &e.Product, tx, []string{"resource_path"}, "product_id", &productID)
	if err != nil {
		return fmt.Errorf("err insert Product: %w", err)
	}
	e.MaterialID = productID

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, tx.Exec)
	if err != nil {
		return fmt.Errorf("err insert Product Material: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert Product Material: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

// Update updates MaterialRepo entity
func (r *MaterialRepo) Update(ctx context.Context, tx database.QueryExecer, e *entities.Material) error {
	ctx, span := interceptors.StartSpan(ctx, "MaterialRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, &e.Product, tx.Exec, "product_id", []string{
		"name",
		"tax_id",
		"available_from",
		"available_until",
		"remarks",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"is_archived",
		"is_unique",
		"updated_at",
	})
	if err != nil {
		return fmt.Errorf("err update Product: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Product: %d RowsAffected", cmdTag.RowsAffected())
	}

	cmdTag, err = database.UpdateFields(ctx, e, tx.Exec, "material_id", []string{"material_type", "custom_billing_date"})
	if err != nil {
		return fmt.Errorf("err update Product Material: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Product Material: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *MaterialRepo) GetByID(ctx context.Context, db database.QueryExecer, materialID string) (entities.Material, error) {
	material := &entities.Material{}
	materialFieldNames, materialFieldValues := material.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			material_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(materialFieldNames, ","),
		material.TableName(),
	)
	row := db.QueryRow(ctx, stmt, materialID)
	err := row.Scan(materialFieldValues...)
	if err != nil {
		return entities.Material{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *material, nil
}

// GetAll get all Material entity
func (r *MaterialRepo) GetAll(ctx context.Context, db database.QueryExecer) (materials []entities.Material, err error) {
	ctx, span := interceptors.StartSpan(ctx, "MaterialRepo.GetAll")
	defer span.End()

	entity := &entities.Material{}
	fieldNames, _ := entity.FieldMap()
	stmt := fmt.Sprintf(
		constant.GetAllQuery,
		strings.Join(fieldNames, ","),
		entity.TableName(),
	)
	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var material entities.Material
		_, fieldValues := material.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		materials = append(materials, material)
	}
	return
}
