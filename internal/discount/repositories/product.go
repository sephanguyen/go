package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ProductRepo struct{}

func (r *ProductRepo) GetByID(ctx context.Context, db database.QueryExecer, productID string) (entities.Product, error) {
	product := &entities.Product{}
	productFieldNames, productFieldValues := product.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		product.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID)
	err := row.Scan(productFieldValues...)
	if err != nil {
		return entities.Product{}, fmt.Errorf("row.Scan ProductRepo.GetByID: %w", err)
	}
	return *product, nil
}
