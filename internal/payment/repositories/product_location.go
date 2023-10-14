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

type ProductLocationRepo struct{}

func (r *ProductLocationRepo) GetByLocationIDAndProductIDForUpdate(ctx context.Context, db database.QueryExecer, locationID string, productID string) (entities.ProductLocation, error) {
	productLocation := &entities.ProductLocation{}
	productLocationFieldNames, productLocationFieldValues := productLocation.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 and location_id = $2
		FOR NO KEY UPDATE`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productLocationFieldNames, ","),
		productLocation.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID, locationID)
	err := row.Scan(productLocationFieldValues...)
	if err != nil {
		return entities.ProductLocation{}, err
	}
	return *productLocation, nil
}

func (r *ProductLocationRepo) Replace(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productLocations []*entities.ProductLocation) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductLocationRepo.Upsert")
	defer span.End()

	batch := &pgx.Batch{}
	batch.Queue(`DELETE FROM product_location WHERE product_id = $1;`, productID)
	r.queueInsert(batch, productLocations)

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (r *ProductLocationRepo) queueInsert(batch *pgx.Batch, productLocations []*entities.ProductLocation) {
	model := entities.ProductLocation{}
	fields, _ := model.FieldMap()
	fieldsExceptResourcePath := fields[:len(fields)-1] // excepts resource_path field
	placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
	stmt := "INSERT INTO " + model.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"

	now := time.Now()
	for _, productLocation := range productLocations {
		_ = productLocation.CreatedAt.Set(now)
		_, values := productLocation.FieldMap()
		valuesExceptResourcePath := values[:len(values)-1] // excepts resource_path value
		batch.Queue(stmt, valuesExceptResourcePath...)
	}
}

func (r *ProductLocationRepo) GetLocationIDsWithProductID(ctx context.Context, db database.QueryExecer, productID string) (locationIDs []string, err error) {
	productLocation := &entities.ProductLocation{}
	productLocationFieldNames, productLocationFieldValues := productLocation.FieldMap()
	stmt := `
		SELECT %s
		FROM %s
		WHERE
		    product_id = $1`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productLocationFieldNames, ","),
		productLocation.TableName(),
	)
	rows, err := db.Query(ctx, stmt, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(productLocationFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		locationIDs = append(locationIDs, productLocation.LocationID.String)
	}
	return
}
