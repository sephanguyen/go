package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
)

type ProductPriceRepo struct{}

func (r *ProductPriceRepo) GetByProductIDAndPriceType(ctx context.Context, db database.QueryExecer, productID, priceType string) ([]entities.ProductPrice, error) {
	var productPrices []entities.ProductPrice
	productPrice := &entities.ProductPrice{}
	productPriceFieldNames, productPriceFieldValues := productPrice.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 AND price_type = $2
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productPriceFieldNames, ","),
		productPrice.TableName(),
	)
	rows, err := db.Query(ctx, stmt, productID, priceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(productPriceFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		productPrices = append(productPrices, *productPrice)
	}
	return productPrices, nil
}

func (r *ProductPriceRepo) GetByProductIDAndQuantityAndPriceType(ctx context.Context, db database.QueryExecer, productID string, weight int32, priceType string) (entities.ProductPrice, error) {
	productPrice := &entities.ProductPrice{}
	productPriceFieldNames, productPriceFieldValues := productPrice.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 AND quantity = $2 AND price_type = $3
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productPriceFieldNames, ","),
		productPrice.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID, weight, priceType)
	err := row.Scan(productPriceFieldValues...)
	if err != nil {
		return entities.ProductPrice{}, err
	}
	return *productPrice, nil
}

func (r *ProductPriceRepo) GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(ctx context.Context, db database.QueryExecer, productID string, billingSchedulePeriodID string, quantity int32, priceType string) (entities.ProductPrice, error) {
	productPrice := &entities.ProductPrice{}
	productPriceFieldNames, productPriceFieldValues := productPrice.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 AND billing_schedule_period_id = $2 AND quantity = $3 AND price_type = $4
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productPriceFieldNames, ","),
		productPrice.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID, billingSchedulePeriodID, quantity, priceType)
	err := row.Scan(productPriceFieldValues...)
	if err != nil {
		return entities.ProductPrice{}, err
	}
	return *productPrice, nil
}

func (r *ProductPriceRepo) GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx context.Context, db database.QueryExecer, productID string, billingSchedulePeriodID string, priceType string) (entities.ProductPrice, error) {
	productPrice := &entities.ProductPrice{}
	productPriceFieldNames, productPriceFieldValues := productPrice.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 AND billing_schedule_period_id = $2 AND price_type = $3
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productPriceFieldNames, ","),
		productPrice.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID, billingSchedulePeriodID, priceType)
	err := row.Scan(productPriceFieldValues...)
	if err != nil {
		return entities.ProductPrice{}, err
	}
	return *productPrice, nil
}

// Create creates ProductPrice entity
func (r *ProductPriceRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.ProductPrice) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductPriceRepo.Create")
	defer span.End()

	now := time.Now()
	if err := e.CreatedAt.Set(now); err != nil {
		return fmt.Errorf("err CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"product_price_id", "resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert ProductPrice: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert ProductPrice: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *ProductPriceRepo) DeleteByProductID(ctx context.Context, db database.QueryExecer, productID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductPriceRepo.Delete")
	defer span.End()

	sql := "DELETE FROM product_price WHERE product_id = $1"
	_, err := db.Exec(ctx, sql, &productID)
	if err != nil {
		return fmt.Errorf("err delete ProductPrice: %w", err)
	}

	return nil
}
