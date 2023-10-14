package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/payment/constant"
)

type DiscountRepo struct {
}

func (r *DiscountRepo) GetByDiscountType(
	ctx context.Context,
	db database.QueryExecer,
	discountType string,
) (
	discounts []*entities.Discount,
	err error,
) {
	discount := entities.Discount{}
	discountFieldNames, _ := discount.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			discount_type = $1
		AND
			is_archived = FALSE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)

	rows, err := db.Query(ctx, stmt, discountType)
	if err != nil {
		return
	}

	defer rows.Close()

	discounts = []*entities.Discount{}
	for rows.Next() {
		discount := new(entities.Discount)
		_, fieldValues := discount.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		discounts = append(discounts, discount)
	}
	return discounts, nil
}

func (r *DiscountRepo) GetByDiscountTagIDs(
	ctx context.Context,
	db database.QueryExecer,
	discountTagIDs []string,
) (
	discounts []*entities.Discount,
	err error,
) {
	discount := entities.Discount{}
	discountFieldNames, _ := discount.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			discount_tag_id = ANY($1)
		AND
			is_archived = FALSE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)

	rows, err := db.Query(ctx, stmt, discountTagIDs)
	if err != nil {
		return
	}

	defer rows.Close()

	discounts = []*entities.Discount{}
	for rows.Next() {
		discount := new(entities.Discount)
		_, fieldValues := discount.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		discounts = append(discounts, discount)
	}
	return discounts, nil
}

func (r *DiscountRepo) GetByID(
	ctx context.Context,
	db database.QueryExecer,
	discountID string,
) (
	entities.Discount,
	error,
) {
	discount := &entities.Discount{}
	discountFieldNames, discountFieldValues := discount.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			discount_id = $1
		AND 
			is_archived = FALSE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)
	row := db.QueryRow(ctx, stmt, discountID)
	err := row.Scan(discountFieldValues...)
	if err != nil {
		return entities.Discount{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *discount, nil
}

func (r *DiscountRepo) GetMaxDiscountByTypeAndDiscountTagIDs(
	ctx context.Context,
	db database.QueryExecer,
	discountAmountType string,
	discountTagIDs []string,
) (
	entities.Discount,
	error,
) {
	discount := &entities.Discount{}
	discountFieldNames, discountFieldValues := discount.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			discount_tag_id = ANY($1)
		AND
			discount_amount_type = $2
		AND
			is_archived = FALSE
		ORDER BY
			discount_amount_value DESC
		LIMIT 1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)
	row := db.QueryRow(ctx, stmt, discountTagIDs, discountAmountType)
	err := row.Scan(discountFieldValues...)
	if err != nil {
		return entities.Discount{}, fmt.Errorf("row.Scan: %w", err)
	}

	return *discount, nil
}

func (r *DiscountRepo) GetMaxProductDiscountByProductID(
	ctx context.Context,
	db database.QueryExecer,
	discountID string,
) (
	entities.Discount,
	error,
) {
	discount := &entities.Discount{}
	fieldNames, discountFieldValues := discount.FieldMap()
	discountFieldNames := sliceutils.Map(fieldNames, func(fieldName string) string {
		return fmt.Sprintf("d.%s", fieldName)
	})
	stmt := `
		SELECT %s
		FROM %s d
		INNER JOIN product_discount pd ON d.discount_id = pd.discount_id
		WHERE
			pd.product_id = $1
		AND 
			d.is_archived = FALSE
		AND
			d.available_from < NOW()
		AND
			d.available_until > NOW()
		ORDER BY
			d.discount_amount_type DESC,
			d.discount_amount_value DESC
		LIMIT 1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)
	row := db.QueryRow(ctx, stmt, discountID)
	err := row.Scan(discountFieldValues...)
	if err != nil {
		return entities.Discount{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *discount, nil
}
