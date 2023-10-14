package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type ProductRepo struct{}

type ProductListFilter struct {
	ProductTypes  []*pb.ProductSpecificType
	StudentGrades []string
	ProductName   string
	ProductStatus string
	Limit         *int64
	Offset        *int64
}

const (
	productActive   = " AND (p.available_from <= NOW() AND p.available_until >= NOW())"
	productInactive = " AND (p.available_from > NOW() OR p.available_until < NOW())"
)

func (r *ProductRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, productID string) (entities.Product, error) {
	product := &entities.Product{}
	productFieldNames, productFieldValues := product.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1
		FOR NO KEY UPDATE`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		product.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID)
	err := row.Scan(productFieldValues...)
	if err != nil {
		return entities.Product{}, err
	}
	return *product, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Product, error) {
	product := &entities.Product{}
	productFieldNames, productFieldValues := product.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 AND is_archived = false
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		product.TableName(),
	)
	row := db.QueryRow(ctx, stmt, entitiesID)
	err := row.Scan(productFieldValues...)
	if err != nil {
		return entities.Product{}, err
	}
	return *product, nil
}

func (r *ProductRepo) GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error) {
	var products []entities.Product
	product := &entities.Product{}
	productFieldNames, productFieldValues := product.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM
			%s
		WHERE
			product_id = ANY($1) AND is_archived = false
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		product.TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(productFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		products = append(products, *product)
	}
	return products, nil
}

func (r *ProductRepo) GetByIDsForExport(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error) {
	var products []entities.Product
	product := &entities.Product{}
	productFieldNames, productFieldValues := product.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM
			%s
		WHERE
			product_id = ANY($1)
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productFieldNames, ","),
		product.TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(productFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		products = append(products, *product)
	}
	return products, nil
}

func (r *ProductRepo) GetProductStatsByFilter(ctx context.Context, db database.QueryExecer, filter ProductListFilter) (productStats entities.ProductStats, err error) {
	getListOfProductsQuery, args := r.buildGetListOfProductsWithFilterQuery(filter)
	_, fieldValues := (&productStats).FieldProductStatsMap()
	stmt := fmt.Sprintf(`
			SELECT
				COUNT(filtered_product.product_id) AS total_items,
				SUM(
				    CASE 
				        WHEN filtered_product.available_from <= NOW() AND
				             filtered_product.available_until >= NOW() THEN 1 ELSE 0
				    END
				) AS total_active,
			    SUM(
			        CASE 
				        WHEN filtered_product.available_from > NOW() OR
				             filtered_product.available_until < NOW() THEN 1 ELSE 0
				    END
				) AS total_inactive
			FROM (%s) AS filtered_product;
			`,
		getListOfProductsQuery,
	)
	row := db.QueryRow(ctx, stmt, args...)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf("row.Scan: %w", err)
		return
	}
	return
}

func (r *ProductRepo) GetProductsByFilter(ctx context.Context, db database.QueryExecer, filter ProductListFilter) (products []entities.Product, err error) {
	getListOfProductsQuery, args := r.buildGetListOfProductsWithFilterQuery(filter)

	rows, err := db.Query(ctx, getListOfProductsQuery, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		product := new(entities.Product)
		_, fieldValues := product.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		products = append(products, *product)
	}
	return
}

func (r *ProductRepo) GetProductIDsByProductTypeAndOrderID(ctx context.Context, db database.QueryExecer, productType, orderID string) (productIDs []string, err error) {
	query := `
			select p.product_id  
			from product p join order_item oi on p.product_id = oi.product_id 
            where p.product_type = $1 and oi.order_id = $2;`
	rows, err := db.Query(ctx, query, productType, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		productID := new(string)
		err = rows.Scan(productID)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		productIDs = append(productIDs, *productID)
	}
	return
}

func (r *ProductRepo) buildGetListOfProductsWithFilterQuery(filter ProductListFilter) (query string, args []interface{}) {
	argsIndex := 1
	product := entities.Product{}
	productGrade := entities.ProductGrade{}
	fieldNames, _ := product.FieldMap()

	query = fmt.Sprintf(`
				SELECT %s FROM "%s" p
				WHERE
					is_archived = false AND p.name ~*'.*%s.*'`, strings.Join(fieldNames, ","), product.TableName(), filter.ProductName)

	if len(filter.StudentGrades) > 0 {
		query += fmt.Sprintf(` 
		AND 
			p.product_id IN (
        		SELECT product_id FROM public."%s" pg WHERE pg.grade_id = ANY($%d)
			)`, productGrade.TableName(), argsIndex)
		args = append(args, filter.StudentGrades)
		argsIndex++
	}
	if len(filter.ProductTypes) > 0 {
		var (
			packageTypes  []string
			materialTypes []string
			feeTypes      []string
			typeStrings   []string
		)
		for _, productType := range filter.ProductTypes {
			switch productType.ProductType {
			case pb.ProductType_PRODUCT_TYPE_PACKAGE:
				packageTypes = append(packageTypes, productType.PackageType.String())
			case pb.ProductType_PRODUCT_TYPE_MATERIAL:
				materialTypes = append(materialTypes, productType.MaterialType.String())
			case pb.ProductType_PRODUCT_TYPE_FEE:
				feeTypes = append(feeTypes, productType.FeeType.String())
			}
		}

		if len(packageTypes) > 0 {
			packageE := entities.Package{}
			packageString := fmt.Sprintf(`
			p.product_id IN (SELECT package_id FROM public."%s" pk WHERE pk.package_type = ANY($%d))`, packageE.TableName(), argsIndex)
			args = append(args, packageTypes)
			typeStrings = append(typeStrings, packageString)
			argsIndex++
		}
		if len(materialTypes) > 0 {
			material := entities.Material{}
			materialString := fmt.Sprintf(`
			p.product_id IN (SELECT material_id FROM public."%s" mt WHERE mt.material_type = ANY($%d))`, material.TableName(), argsIndex)
			args = append(args, materialTypes)
			typeStrings = append(typeStrings, materialString)
			argsIndex++
		}
		if len(feeTypes) > 0 {
			fee := entities.Fee{}
			feeString := fmt.Sprintf(`
			p.product_id IN (SELECT fee_id FROM public."%s" f WHERE f.fee_type = ANY($%d))`, fee.TableName(), argsIndex)
			args = append(args, materialTypes)
			typeStrings = append(typeStrings, feeString)
			argsIndex++
		}
		query += fmt.Sprintf(` AND ( %s )`, strings.Join(typeStrings, " OR "))
	}
	switch filter.ProductStatus {
	case pb.ProductStatus_PRODUCT_STATUS_ACTIVE.String():
		query += productActive
	case pb.ProductStatus_PRODUCT_STATUS_INACTIVE.String():
		query += productInactive
	}
	if filter.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argsIndex)
		args = append(args, *filter.Limit)
		argsIndex++
	}
	if filter.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argsIndex)
		args = append(args, *filter.Offset)
		argsIndex++
	}

	return
}
