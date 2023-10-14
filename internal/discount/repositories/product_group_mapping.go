package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ProductGroupMappingRepo struct {
}

func (r *ProductGroupMappingRepo) GetByProductID(ctx context.Context, db database.QueryExecer, productID string) (productGroupMappings []*entities.ProductGroupMapping, err error) {
	productGroupMappingEntity := entities.ProductGroupMapping{}
	productGroupMappingFieldNames, _ := productGroupMappingEntity.FieldMap()
	stmt := `
		SELECT
			%s
		FROM 
			%s
		WHERE
			product_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productGroupMappingFieldNames, ","),
		productGroupMappingEntity.TableName(),
	)

	rows, err := db.Query(ctx, stmt, productID)
	if err != nil {
		return
	}
	defer rows.Close()

	productGroupMappings = []*entities.ProductGroupMapping{}
	for rows.Next() {
		productGroupMapping := new(entities.ProductGroupMapping)
		_, fieldValues := productGroupMapping.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		productGroupMappings = append(productGroupMappings, productGroupMapping)
	}
	return productGroupMappings, nil
}

func (r *ProductGroupMappingRepo) queueUpsert(b *pgx.Batch, productGroupMappings []*entities.ProductGroupMapping) {
	queueFn := func(b *pgx.Batch, u *entities.ProductGroupMapping) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"
		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range productGroupMappings {
		_ = u.CreatedAt.Set(now)
		_ = u.UpdatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *ProductGroupMappingRepo) Upsert(ctx context.Context, db database.QueryExecer, productGroupID pgtype.Text, e []*entities.ProductGroupMapping) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductGroupMappingRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM product_group_mapping WHERE product_group_id = $1;`, productGroupID)
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
