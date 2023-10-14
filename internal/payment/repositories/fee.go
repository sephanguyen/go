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

type FeeRepo struct{}

// Create creates Product Fee entity
func (r *FeeRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Fee) error {
	ctx, span := interceptors.StartSpan(ctx, "FeeRepo.Create")
	defer span.End()

	var productID pgtype.Text

	now := time.Now()
	if err := multierr.Combine(
		e.ProductID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	err := database.InsertReturningAndExcept(ctx, &e.Product, db, []string{"resource_path"}, "product_id", &productID)
	if err != nil {
		return fmt.Errorf("err insert Product: %w", err)
	}
	e.FeeID = productID

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert Fee: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert Fee: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

// Update updates Product Fee entity
func (r *FeeRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Fee) error {
	ctx, span := interceptors.StartSpan(ctx, "FeeRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, &e.Product, db.Exec, "product_id", []string{
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

	cmdTag, err = database.UpdateFields(ctx, e, db.Exec, "fee_id", []string{
		"fee_type",
	})
	if err != nil {
		return fmt.Errorf("err update Fee: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Fee: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// GetAll get all Fee entity
func (r *FeeRepo) GetAll(ctx context.Context, db database.QueryExecer) (fees []entities.Fee, err error) {
	ctx, span := interceptors.StartSpan(ctx, "FeeRepo.GetAll")
	defer span.End()

	entity := &entities.Fee{}
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
		var fee entities.Fee
		_, fieldValues := fee.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		fees = append(fees, fee)
	}
	return
}

func (r *FeeRepo) GetFeeByID(ctx context.Context, db database.QueryExecer, feeID string) (entities.Fee, error) {
	fee := &entities.Fee{}
	packageFieldNames, packageFieldValues := fee.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			fee_id = $1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(packageFieldNames, ","),
		fee.TableName(),
	)
	row := db.QueryRow(ctx, stmt, feeID)
	err := row.Scan(packageFieldValues...)
	if err != nil {
		return entities.Fee{}, err
	}
	return *fee, nil
}
