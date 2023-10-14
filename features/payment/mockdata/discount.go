package mockdata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func InsertOrgLevelDiscount(ctx context.Context, fatimaDBTrace *database.DBTrace, studentID string) (discountID string, err error) {
	// insert discount_tag
	discountTagID := idutil.ULIDNow()
	insertDiscountTagStmt := `INSERT INTO discount_tag (discount_tag_id, discount_tag_name, created_at, updated_at) VALUES ($1, $2, now(), now())`
	_, err = fatimaDBTrace.Exec(ctx, insertDiscountTagStmt, discountTagID, discountTagID)
	if err != nil {
		return "", fmt.Errorf("cannot insert discount_tag, err: %s", err)
	}

	// insert discount
	type discountParams struct {
		Name                string         `json:"name"`
		DiscountID          string         `json:"discount_id"`
		DiscountType        string         `json:"discount_type"`
		DiscountAmountType  string         `json:"discount_amount_type"`
		DiscountAmountValue pgtype.Numeric `json:"discount_amount_value"`
		AvailableFrom       time.Time      `json:"available_from"`
		AvailableUtil       time.Time      `json:"available_until"`
		Remarks             string         `json:"remarks"`
		IsArchived          bool           `json:"is_archived"`
		DiscountTagID       string         `json:"discount_tag_id"`
	}

	discountValue := pgtype.Numeric{}
	err = discountValue.Set(20)
	if err != nil {
		return "", err
	}

	discount := discountParams{
		Name:                "Sample org level discount",
		DiscountID:          idutil.ULIDNow(),
		DiscountType:        pb.DiscountType_DISCOUNT_TYPE_FAMILY.String(),
		DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
		DiscountAmountValue: discountValue,
		AvailableFrom:       time.Now().AddDate(-1, 0, 0),
		AvailableUtil:       time.Now().AddDate(1, 0, 0),
		Remarks:             "discount remarks",
		IsArchived:          false,
		DiscountTagID:       discountTagID,
	}

	insertDiscountStmt := `INSERT INTO discount (
		discount_id,
		name,
		discount_type,
		discount_amount_type,
		discount_amount_value,
		available_from,
		available_until,
		remarks,
		is_archived,
		discount_tag_id,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now());`
	_, err = fatimaDBTrace.Exec(ctx, insertDiscountStmt,
		discount.DiscountID,
		discount.Name,
		discount.DiscountType,
		discount.DiscountAmountType,
		discount.DiscountAmountValue,
		discount.AvailableFrom,
		discount.AvailableUtil,
		discount.Remarks,
		discount.IsArchived,
		discount.DiscountTagID,
	)
	if err != nil {
		return discount.DiscountID, fmt.Errorf("cannot insert discount, err: %s", err)
	}

	// insert user_discount_tag
	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 3, 0)
	insertUserDiscountTagStmt := `INSERT INTO user_discount_tag (
		user_id,
		discount_type,
		discount_tag_id,
		start_date,
	    end_date,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5, $4, now())`
	_, err = fatimaDBTrace.Exec(ctx, insertUserDiscountTagStmt,
		studentID,
		discount.DiscountAmountType,
		discount.DiscountTagID,
		startDate,
		endDate,
	)
	if err != nil {
		return discount.DiscountID, fmt.Errorf("cannot insert user_discount_tag, err: %s", err)
	}

	return discount.DiscountID, nil
}

func InsertOrgDiscount(ctx context.Context, fatimaDBTrace *database.DBTrace, discountType string) (discountID string, discountTagID string, err error) {
	discountTag := entities.DiscountTag{}
	discountTagID = idutil.ULIDNow()

	selectable := true
	if discountType == pb.DiscountType_DISCOUNT_TYPE_COMBO.String() || discountType == pb.DiscountType_DISCOUNT_TYPE_SIBLING.String() {
		selectable = false
	}

	err = multierr.Combine(
		discountTag.DiscountTagID.Set(discountTagID),
		discountTag.DiscountTagName.Set(discountType),
		discountTag.Selectable.Set(selectable),
		discountTag.IsArchived.Set(false),
	)
	if err != nil {
		return
	}

	insertDiscountTagStmt := `INSERT INTO discount_tag (
		discount_tag_id,
		discount_tag_name,
		selectable,
		is_archived,
		updated_at,
		created_at)
	VALUES ($1, $2, $3, $4, now(), now())`
	_, err = fatimaDBTrace.Exec(ctx, insertDiscountTagStmt,
		discountTag.DiscountTagID,
		discountTag.DiscountTagName,
		discountTag.Selectable,
		discountTag.IsArchived,
	)
	if err != nil {
		return
	}

	discount := entities.Discount{}
	discountID = idutil.ULIDNow()

	err = multierr.Combine(
		discount.DiscountID.Set(discountID),
		discount.Name.Set(discountType),
		discount.DiscountType.Set(discountType),
		discount.DiscountAmountType.Set(pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE),
		discount.DiscountAmountValue.Set(50),
		discount.AvailableFrom.Set(time.Now().AddDate(-1, 0, 0)),
		discount.AvailableUntil.Set(time.Now().AddDate(1, 0, 0)),
		discount.IsArchived.Set(false),
		discount.DiscountTagID.Set(discountTagID),
	)
	if err != nil {
		return
	}

	insertDiscountStmt := `INSERT INTO discount (
			discount_id,
			name,
			discount_type,
			discount_amount_type,
			discount_amount_value,
			available_from,
			available_until,
			is_archived,
			discount_tag_id,
			updated_at,
			created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())`
	_, err = fatimaDBTrace.Exec(ctx, insertDiscountStmt,
		discount.DiscountID,
		discount.Name,
		discount.DiscountType,
		discount.DiscountAmountType,
		discount.DiscountAmountValue,
		discount.AvailableFrom,
		discount.AvailableUntil,
		discount.IsArchived,
		discount.DiscountTagID,
	)

	return
}

func InsertUserDiscountTag(ctx context.Context, fatimaDBTrace *database.DBTrace, userDiscountTag *entities.UserDiscountTag) (err error) {
	var stmt string
	switch userDiscountTag.DiscountType.String {
	case pb.DiscountType_DISCOUNT_TYPE_COMBO.String(), pb.DiscountType_DISCOUNT_TYPE_SIBLING.String():
		stmt = `
			INSERT INTO user_discount_tag (
				user_id,
				discount_type,
				discount_tag_id,
				product_id,
				product_group_id,
				updated_at,
				created_at)
			VALUES ($1, $2, $3,$4,$5, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmt,
			userDiscountTag.UserID,
			userDiscountTag.DiscountType,
			userDiscountTag.DiscountTagID,
			userDiscountTag.ProductID,
			userDiscountTag.ProductGroupID,
		)
	default:
		stmt = `
			INSERT INTO user_discount_tag (
				user_id,
				discount_type,
				discount_tag_id,
				updated_at,
				created_at)
			VALUES ($1, $2, $3, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmt,
			userDiscountTag.UserID,
			userDiscountTag.DiscountType,
			userDiscountTag.DiscountTagID,
		)
	}
	if err != nil {
		return err
	}

	return nil
}

func GetDiscountByID(ctx context.Context, fatimaDBTrace *database.DBTrace, discountID string) (discount *entities.Discount, err error) {
	discount = &entities.Discount{}
	discountFieldNames, discountValues := discount.FieldMap()
	stmt := `SELECT %s FROM %s WHERE discount_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountFieldNames, ","),
		discount.TableName(),
	)
	row := fatimaDBTrace.QueryRow(ctx, stmt, discountID)
	err = row.Scan(discountValues...)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return
}

func GetStudentProductsByStudentID(ctx context.Context, fatimaDBTrace *database.DBTrace, studentID string) (studentProducts []entities.StudentProduct, err error) {
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()

	stmt := `SELECT %s FROM %s WHERE student_id = $1`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProductEntity.TableName(),
	)
	rows, err := fatimaDBTrace.Query(ctx, stmt, studentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		studentProduct := new(entities.StudentProduct)
		_, fieldValues := studentProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		studentProducts = append(studentProducts, *studentProduct)
	}

	return studentProducts, nil
}
