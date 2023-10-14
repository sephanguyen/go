package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type DiscountRepo struct {
}

func (r *DiscountRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, id string) (entities.Discount, error) {
	discount := &entities.Discount{}
	discountFieldNames, discountFieldValues := discount.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			discount_id = $1
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)
	row := db.QueryRow(ctx, stmt, id)
	err := row.Scan(discountFieldValues...)
	if err != nil {
		return entities.Discount{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *discount, nil
}

// Create creates Discount entity
func (r *DiscountRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Discount) error {
	ctx, span := interceptors.StartSpan(ctx, "DiscountRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.DiscountID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert Discount: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert Discount: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates Discount entity
func (r *DiscountRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Discount) error {
	ctx, span := interceptors.StartSpan(ctx, "DiscountRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "discount_id", []string{
		"name",
		"discount_type",
		"discount_amount_type",
		"discount_amount_value",
		"recurring_valid_duration",
		"available_from",
		"available_until",
		"remarks",
		"is_archived",
		"student_tag_id_validation",
		"parent_tag_id_validation",
		"discount_tag_id",
		"updated_at",
	})
	if err != nil {
		return fmt.Errorf("err update Discount: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Discount: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *DiscountRepo) GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Discount, error) {
	var discounts []entities.Discount
	discountFieldNames, _ := (&entities.Discount{}).FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			discount_id = ANY($1) AND is_archived = false
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		(&entities.Discount{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		discount := new(entities.Discount)
		_, fieldValues := discount.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		discounts = append(discounts, *discount)
	}
	return discounts, nil
}
