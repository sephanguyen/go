package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type ProductGroupRepo struct {
}

func (r *ProductGroupRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.ProductGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductGroupRepo.Create")
	defer span.End()

	now := time.Now()
	id := idutil.ULIDNow()

	if err := multierr.Combine(
		e.ProductGroupID.Set(id),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine ProductGroupID.Set,UpdatedAt.Set and CreatedAt.Set: %w", err)
	}

	if err := e.CreatedAt.Set(now); err != nil {
		return fmt.Errorf("err CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert ProductGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert ProductGroup: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *ProductGroupRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.ProductGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductGroupRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "product_group_id", []string{"group_name", "group_tag", "discount_type", "is_archived", "updated_at"})
	if err != nil {
		return fmt.Errorf("err update ProductGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update ProductGroup: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *ProductGroupRepo) GetByID(ctx context.Context, db database.QueryExecer, productGroupID string) (entities.ProductGroup, error) {
	productGroup := &entities.ProductGroup{}
	productGroupFieldNames, productGroupFieldValues := productGroup.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_group_id = $1
		AND
			is_archived = FALSE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productGroupFieldNames, ","),
		productGroup.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productGroupID)
	err := row.Scan(productGroupFieldValues...)
	if err != nil {
		return entities.ProductGroup{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *productGroup, nil
}
