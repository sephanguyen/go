package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type BillItemRepo struct {
}

func (r *BillItemRepo) GetLastBillItemOfStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
) (
	entities.BillItem,
	error,
) {
	billItem := &entities.BillItem{}
	billItemFieldNames, billItemFieldValues := billItem.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			student_product_id = $1 AND is_latest_bill_item = true
		ORDER BY
			billing_date DESC
		LIMIT 1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentProductID)
	err := row.Scan(billItemFieldValues...)
	if err != nil {
		return entities.BillItem{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *billItem, nil
}
