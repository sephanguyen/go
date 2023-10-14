package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ProductAccountingCategoryRepo struct {
}

func (r *ProductAccountingCategoryRepo) queueUpsert(b *pgx.Batch, productAssociatedDataAccountingCategories []*entities.ProductAccountingCategory) {
	queueFn := func(b *pgx.Batch, u *entities.ProductAccountingCategory) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range productAssociatedDataAccountingCategories {
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *ProductAccountingCategoryRepo) Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productAssociatedDataAccountingCategories []*entities.ProductAccountingCategory) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductAccountingCategoryRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM product_accounting_category WHERE product_id = $1;`, productID)
	r.queueUpsert(b, productAssociatedDataAccountingCategories)

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
