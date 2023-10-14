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
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ProductDiscountRepo struct {
}

func (r *ProductDiscountRepo) queueUpsert(b *pgx.Batch, productDiscounts []*entities.ProductDiscount) {
	queueFn := func(b *pgx.Batch, u *entities.ProductDiscount) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"
		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range productDiscounts {
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *ProductDiscountRepo) Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, e []*entities.ProductDiscount) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductDiscountRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM product_discount WHERE product_id = $1;`, productID)
	r.queueUpsert(b, e)

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (r *ProductDiscountRepo) GetByProductIDAndDiscountID(ctx context.Context, db database.QueryExecer, productID string, discountID string) (entities.ProductDiscount, error) {
	productDiscount := &entities.ProductDiscount{}
	productDiscountFieldNames, productDiscountFieldValues := productDiscount.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 AND discount_id = $2
		FOR NO KEY UPDATE
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productDiscountFieldNames, ","),
		productDiscount.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID, discountID)
	err := row.Scan(productDiscountFieldValues...)
	if err != nil {
		return entities.ProductDiscount{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *productDiscount, nil
}
