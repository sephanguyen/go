package mockdata

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	paymentEntities "github.com/manabie-com/backend/internal/payment/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func InsertStudentWithActiveProducts(ctx context.Context, fatimaDBTrace *database.DBTrace) (studentID string, locationID string, err error) {
	studentID, locationID, err = InsertPreconditionData(ctx, fatimaDBTrace)
	if err != nil {
		return
	}
	productIDs, err := InsertRecurringProducts(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	for _, productID := range productIDs {
		order := paymentEntities.Order{}
		err = multierr.Combine(
			order.OrderID.Set(idutil.ULIDNow()),
			order.OrderSequenceNumber.Set(1),
			order.StudentID.Set(studentID),
			order.LocationID.Set(locationID),
			order.StudentFullName.Set(idutil.ULIDNow()),
			order.OrderStatus.Set(pb.OrderStatus_ORDER_STATUS_SUBMITTED),
		)
		if err != nil {
			return
		}

		stmtOrder := `INSERT INTO public.order(
			order_id,
			order_sequence_number,
			student_id,
			location_id,
			student_full_name,
			order_status,
			updated_at,
			created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmtOrder,
			order.OrderID,
			order.OrderSequenceNumber,
			order.StudentID,
			order.LocationID,
			order.StudentFullName,
			order.OrderStatus)
		if err != nil {
			return
		}

		studentProduct := entities.StudentProduct{}
		err = multierr.Combine(
			studentProduct.StudentProductID.Set(idutil.ULIDNow()),
			studentProduct.StudentID.Set(order.StudentID),
			studentProduct.LocationID.Set(order.LocationID),
			studentProduct.ProductID.Set(productID),
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED),
			studentProduct.StartDate.Set(time.Now().AddDate(0, -1, 0)),
			studentProduct.EndDate.Set(time.Now().AddDate(0, 2, 0)),
		)
		if err != nil {
			return
		}

		stmtStudentProduct := `INSERT INTO student_product(
			student_product_id,
			student_id,
			location_id,
			product_id,
			product_status,
			start_date,
			end_date,
			updated_at,
			created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmtStudentProduct,
			studentProduct.StudentProductID,
			studentProduct.StudentID,
			studentProduct.LocationID,
			studentProduct.ProductID,
			studentProduct.ProductStatus,
			studentProduct.StartDate,
			studentProduct.EndDate)
		if err != nil {
			return
		}

		billItem := entities.BillItem{}
		err = multierr.Combine(
			billItem.OrderID.Set(order.OrderID),
			billItem.StudentProductID.Set(studentProduct.StudentProductID),
			billItem.ProductID.Set(studentProduct.ProductID),
			billItem.ProductDescription.Set(studentProduct.ProductID),
			billItem.StudentID.Set(order.StudentID),
			billItem.LocationID.Set(order.LocationID),
			billItem.BillType.Set(pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER),
			billItem.BillStatus.Set(pb.BillingStatus_BILLING_STATUS_BILLED),
			billItem.FinalPrice.Set(500),
		)
		if err != nil {
			return
		}

		stmtBillItem := `INSERT INTO bill_item(
			order_id,
			student_product_id,
			product_id,
			product_description,
			student_id,
			location_id,
			bill_type,
			billing_status,
			final_price,
			is_latest_bill_item,
			updated_at,
			created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmtBillItem,
			billItem.OrderID,
			billItem.StudentProductID,
			billItem.ProductID,
			billItem.ProductDescription,
			billItem.StudentID,
			billItem.LocationID,
			billItem.BillType,
			billItem.BillStatus,
			billItem.FinalPrice,
			true)
		if err != nil {
			return
		}
	}

	return
}

func InsertRecurringProducts(ctx context.Context, fatimaDBTrace *database.DBTrace) (productIDs []string, err error) {
	billingScheduleID, err := InsertBillingScheduleForRecurringProduct(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	taxID, err := InsertOneTax(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	productIDs, err = InsertRecurringMaterials(ctx, fatimaDBTrace, billingScheduleID, taxID)
	if err != nil {
		return
	}

	return
}

func InsertPreconditionData(ctx context.Context, fatimaDBTrace *database.DBTrace) (
	userID string,
	locationID string,
	err error,
) {
	gradeID, err := InsertOneGrade(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	locationID = constants.ManabieOrgLocation

	userID, err = InsertOneEnrolledStudent(ctx, fatimaDBTrace, gradeID, locationID)
	if err != nil {
		return
	}

	return
}

func InsertBillingScheduleForRecurringProduct(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	var billingScheduleID string
	randomStr := idutil.ULIDNow()
	name := database.Text("recurring-product-schedule-" + randomStr)
	remarks := database.Text("recurring-product-schedule-" + time.Now().Format("01-02-2006"))

	stmt := `
		INSERT INTO billing_schedule (
			billing_schedule_id,
			name,
			remarks,
			is_archived,
			created_at,
			updated_at)
		VALUES
			($1, $2, $3, $4, now(), now())
		RETURNING billing_schedule_id`
	row := fatimaDBTrace.QueryRow(ctx, stmt,
		randomStr,
		name,
		remarks,
		"false",
	)
	err := row.Scan(&billingScheduleID)
	if err != nil {
		return billingScheduleID, fmt.Errorf("cannot insert billing_schedule, err: %s", err)
	}

	return billingScheduleID, nil
}

func InsertOneLocation(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	id := idutil.ULIDNow()
	name := idutil.ULIDNow()

	stmt := `INSERT INTO locations(
		location_id,
		name,
		created_at,
		updated_at
	) VALUES ($1, $2, now(), now())`
	_, err := fatimaDBTrace.Exec(ctx, stmt, id, name)
	if err != nil {
		return id, fmt.Errorf("cannot insert location, err: %s", err)
	}

	return id, nil
}

func InsertOneGrade(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	id := idutil.ULIDNow()
	name := idutil.ULIDNow()

	stmt := `INSERT INTO grade(
		grade_id,
		name,
		is_archived,
		partner_internal_id,
		sequence,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err := fatimaDBTrace.Exec(ctx, stmt, id, name, false, id, 1)
	if err != nil {
		return "", err
	}

	return id, nil
}

func InsertOneEnrolledStudent(ctx context.Context, fatimaDBTrace *database.DBTrace, gradeID string, locationID string) (string, error) {
	id := idutil.ULIDNow()

	name := database.Text(fmt.Sprintf("Student %s", id))
	stmt := `INSERT INTO users
		(user_id, name, user_group, country, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now());`
	_, err := fatimaDBTrace.Exec(ctx, stmt, id, name, cpb.UserGroup_USER_GROUP_STUDENT.String(), "COUNTRY_VN")
	if err != nil {
		return "", err
	}

	stmt = `INSERT INTO students
		(student_id, current_grade, enrollment_status, grade_id, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now());`
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, 1, "STUDENT_ENROLLMENT_STATUS_ENROLLED", gradeID)
	if err != nil {
		return "", err
	}

	stmt = `INSERT INTO student_enrollment_status_history
		(student_id, location_id, enrollment_status, start_date, updated_at, created_at)
		VALUES ($1, $2, $3, now(), now(), now());`
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, locationID, "STUDENT_ENROLLMENT_STATUS_ENROLLED")
	if err != nil {
		return "", err
	}

	return id, nil
}

func InsertProductGroupMappingForSpecialDiscount(ctx context.Context, fatimaDBTrace *database.DBTrace, discountType string) (productGroupIDS []string, productIDs []string, err error) {
	randomID := idutil.ULIDNow()
	var groupTag string

	switch discountType {
	case pb.DiscountType_DISCOUNT_TYPE_COMBO.String():
		groupTag = fmt.Sprintf("COMBO-%s", randomID)
	case pb.DiscountType_DISCOUNT_TYPE_SIBLING.String():
		groupTag = fmt.Sprintf("SIBLING-%s", randomID)
	default:
		return nil, nil, fmt.Errorf("invalid discount type for product group: %v", discountType)
	}

	productGroupA, err := InsertProductGroup(ctx, fatimaDBTrace, groupTag, discountType)
	if err != nil {
		return
	}
	productGroupIDS = append(productGroupIDS, productGroupA.ProductGroupID.String)

	productGroupB, err := InsertProductGroup(ctx, fatimaDBTrace, groupTag, discountType)
	if err != nil {
		return
	}
	productGroupIDS = append(productGroupIDS, productGroupB.ProductGroupID.String)

	productIDs, err = InsertRecurringProducts(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	stmt := `INSERT INTO product_group_mapping(
		product_group_id,
		product_id,
		created_at,
		updated_at
	) VALUES ($1, $2, now(), now())`

	for i, productID := range productIDs {
		if i == 0 {
			_, err = fatimaDBTrace.Exec(ctx, stmt, productGroupA.ProductGroupID, productID)
			if err != nil {
				return
			}
		} else {
			_, err = fatimaDBTrace.Exec(ctx, stmt, productGroupB.ProductGroupID, productID)
			if err != nil {
				return
			}
		}
	}

	return
}

func InsertProductGroup(ctx context.Context, fatimaDBTrace *database.DBTrace, groupTag, discountType string) (productGroup entities.ProductGroup, err error) {
	productGroup = entities.ProductGroup{
		ProductGroupID: pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
		GroupName:      pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
		DiscountType:   pgtype.Text{String: discountType, Status: pgtype.Present},
		GroupTag:       pgtype.Text{String: groupTag, Status: pgtype.Present},
	}

	stmt := `INSERT INTO product_group(
		product_group_id,
		group_name,
		group_tag,
		discount_type,
		created_at,
		updated_at
	) VALUES ($1, $2, $3,$4, now(), now())`
	_, err = fatimaDBTrace.Exec(ctx, stmt, productGroup.ProductGroupID, productGroup.GroupName, productGroup.GroupTag, productGroup.DiscountType)
	if err != nil {
		return entities.ProductGroup{}, err
	}

	return productGroup, nil
}

func InsertRecurringMaterials(ctx context.Context, fatimaDBTrace *database.DBTrace, billingScheduleID string, taxID string) ([]string, error) {
	type ProductParams struct {
		ProductID            string
		Name                 string
		ProductType          string
		TaxID                string
		AvailableFrom        time.Time
		AvailableUtil        time.Time
		Remarks              sql.NullString
		CustomBillingPeriod  sql.NullTime
		BillingScheduleID    string
		DisableProRatingFlag bool
		IsArchived           bool
	}

	type MaterialParams struct {
		MaterialID        string
		MaterialType      string
		CustomBillingDate sql.NullTime
		ResourcePath      sql.NullString
	}

	type ProductSettingParams struct {
		ProductID                    string
		IsEnrollmentRequired         bool
		IsPausable                   bool
		IsAddedToEnrollmentByDefault bool
		IsOperationFee               bool
	}

	productIDs := []string{}

	for i := 0; i < 3; i++ {
		var (
			productArg        ProductParams
			materialArg       MaterialParams
			productSettingArg ProductSettingParams
		)

		randomStr := idutil.ULIDNow()
		currentTime := time.Now()

		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("material-recurring-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		productArg.AvailableFrom = currentTime.AddDate(-1, 0, 0)
		productArg.AvailableUtil = currentTime.AddDate(1, 0, 0)
		productArg.IsArchived = false
		productArg.TaxID = taxID
		productArg.BillingScheduleID = billingScheduleID
		productArg.DisableProRatingFlag = false

		productSettingArg.ProductID = randomStr
		productSettingArg.IsEnrollmentRequired = false
		productSettingArg.IsPausable = true
		productSettingArg.IsAddedToEnrollmentByDefault = false
		productSettingArg.IsOperationFee = false

		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_RECURRING.String()

		stmtInsertProduct := `
			INSERT INTO product (
				product_id,
        		name,
            	product_type,
                tax_id,
                available_from,
                available_until,
                remarks,
                custom_billing_period,
                billing_schedule_id,
                disable_pro_rating_flag,
                is_archived,
                updated_at,
                created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
			RETURNING product_id`
		row := fatimaDBTrace.QueryRow(ctx, stmtInsertProduct,
			productArg.ProductID,
			productArg.Name,
			productArg.ProductType,
			productArg.TaxID,
			productArg.AvailableFrom,
			productArg.AvailableUtil,
			productArg.Remarks,
			productArg.CustomBillingPeriod,
			productArg.BillingScheduleID,
			productArg.DisableProRatingFlag,
			productArg.IsArchived,
		)
		err := row.Scan(&materialArg.MaterialID)
		if err != nil {
			return productIDs, fmt.Errorf("cannot insert product, err: %s", err)
		}

		stmtInsertProductSetting := `
			INSERT INTO product_setting (
				product_id,
				is_enrollment_required,
				is_pausable,
				is_added_to_enrollment_by_default,
				is_operation_fee,
				created_at,
				updated_at)
			VALUES ($1, $2, $3, $4, $5, now(), now())`
		_, err = fatimaDBTrace.Exec(
			ctx,
			stmtInsertProductSetting,
			productSettingArg.ProductID,
			productSettingArg.IsEnrollmentRequired,
			productSettingArg.IsPausable,
			productSettingArg.IsAddedToEnrollmentByDefault,
			productSettingArg.IsOperationFee,
		)
		if err != nil {
			return productIDs, fmt.Errorf("cannot insert product setting, err: %s", err)
		}

		stmtInsertMaterial := `
			INSERT INTO material (
				material_id,
				material_type)
			VALUES ($1, $2)`
		_, err = fatimaDBTrace.Exec(
			ctx,
			stmtInsertMaterial,
			materialArg.MaterialID,
			materialArg.MaterialType,
		)
		if err != nil {
			return productIDs, fmt.Errorf("cannot insert material, err: %s", err)
		}

		productIDs = append(productIDs, materialArg.MaterialID)
	}
	return productIDs, nil
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
			VALUES ($1, $2, $3, $4, $5, now(), now())`
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

func InsertOneTax(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	taxName := database.Text("Tax for discount automation mock data")
	taxPercentage := database.Int4(20)
	taxCategory := database.Text("TAX_CATEGORY_INCLUSIVE")
	isArchived := database.Bool(false)
	stmt := `INSERT INTO tax
		(tax_id, name, tax_percentage, tax_category, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now()) RETURNING tax_id;`
	row := fatimaDBTrace.QueryRow(ctx, stmt, idutil.ULIDNow(), taxName, taxPercentage, taxCategory, isArchived)
	var taxID string
	err := row.Scan(&taxID)
	return taxID, err
}

func InsertNSiblingsAndReturnIDs(ctx context.Context, fatimaDBTrace *database.DBTrace, siblingCount int) (siblingIDs []string, err error) {
	siblingIDs = []string{}

	parentID, err := InsertOneParent(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	for i := 0; i < siblingCount; i++ {
		var studentID string
		studentID, err = InsertOneStudent(ctx, fatimaDBTrace, "1")
		if err != nil {
			return
		}

		err = InsertParentChildRelationship(ctx, fatimaDBTrace, parentID, studentID)
		if err != nil {
			return
		}

		siblingIDs = append(siblingIDs, studentID)
	}

	return
}

func InsertOneParent(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	id := idutil.ULIDNow()

	name := database.Text(fmt.Sprintf("Student %s", id))
	stmt := `INSERT INTO users
		(user_id, name, user_group, country, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now());`
	_, err := fatimaDBTrace.Exec(ctx, stmt, id, name, cpb.UserGroup_USER_GROUP_PARENT.String(), "COUNTRY_VN")
	if err != nil {
		return "", err
	}

	return id, nil
}

func InsertOneStudent(ctx context.Context, fatimaDBTrace *database.DBTrace, gradeID string) (string, error) {
	id := idutil.ULIDNow()

	name := database.Text(fmt.Sprintf("Student %s", id))
	stmt := `INSERT INTO users
		(user_id, name, user_group, country, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now());`
	_, err := fatimaDBTrace.Exec(ctx, stmt, id, name, cpb.UserGroup_USER_GROUP_STUDENT.String(), "COUNTRY_VN")
	if err != nil {
		return "", err
	}

	stmt = `INSERT INTO students
		(student_id, current_grade, enrollment_status, grade_id, updated_at, created_at)
		VALUES ($1, $2, $3, $4, now(), now());`
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, 1, "STUDENT_ENROLLMENT_STATUS_ENROLLED", gradeID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func InsertParentChildRelationship(ctx context.Context, fatimaDBTrace *database.DBTrace, parentID string, studentID string) error {
	stmt := `INSERT INTO student_parents
		(parent_id, student_id, relationship, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now());`
	_, err := fatimaDBTrace.Exec(ctx, stmt, parentID, studentID, "")
	if err != nil {
		return err
	}

	return nil
}

func InsertOrderForStudentWithProducts(
	ctx context.Context,
	fatimaDBTrace *database.DBTrace,
	studentID string,
	productIDs []string,
) (
	orderID string,
	locationID string,
	studentProductIDs []string,
	err error,
) {
	locationID = constants.ManabieOrgLocation

	order := paymentEntities.Order{}
	err = multierr.Combine(
		order.OrderID.Set(idutil.ULIDNow()),
		order.OrderSequenceNumber.Set(1),
		order.StudentID.Set(studentID),
		order.LocationID.Set(locationID),
		order.StudentFullName.Set(idutil.ULIDNow()),
		order.OrderStatus.Set(pb.OrderStatus_ORDER_STATUS_SUBMITTED),
	)
	if err != nil {
		return
	}

	orderID = order.OrderID.String

	stmtOrder := `INSERT INTO public.order(
		order_id,
		order_sequence_number,
		student_id,
		location_id,
		student_full_name,
		order_status,
		updated_at,
		created_at)
	VALUES ($1, $2, $3, $4, $5, $6, now(), now())`
	_, err = fatimaDBTrace.Exec(ctx, stmtOrder,
		order.OrderID,
		order.OrderSequenceNumber,
		order.StudentID,
		order.LocationID,
		order.StudentFullName,
		order.OrderStatus)
	if err != nil {
		return
	}

	studentProductIDs = []string{}

	for _, productID := range productIDs {
		studentProduct := entities.StudentProduct{}
		err = multierr.Combine(
			studentProduct.StudentProductID.Set(idutil.ULIDNow()),
			studentProduct.StudentID.Set(order.StudentID),
			studentProduct.LocationID.Set(order.LocationID),
			studentProduct.ProductID.Set(productID),
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED),
			studentProduct.StartDate.Set(time.Now().AddDate(0, -1, 0)),
			studentProduct.EndDate.Set(time.Now().AddDate(0, 2, 0)),
		)
		if err != nil {
			return
		}

		stmtStudentProduct := `INSERT INTO student_product(
			student_product_id,
			student_id,
			location_id,
			product_id,
			product_status,
			start_date,
			end_date,
			updated_at,
			created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmtStudentProduct,
			studentProduct.StudentProductID,
			studentProduct.StudentID,
			studentProduct.LocationID,
			studentProduct.ProductID,
			studentProduct.ProductStatus,
			studentProduct.StartDate,
			studentProduct.EndDate)
		if err != nil {
			return
		}

		billItem := entities.BillItem{}
		err = multierr.Combine(
			billItem.OrderID.Set(order.OrderID),
			billItem.StudentProductID.Set(studentProduct.StudentProductID),
			billItem.ProductID.Set(studentProduct.ProductID),
			billItem.ProductDescription.Set(studentProduct.ProductID),
			billItem.StudentID.Set(order.StudentID),
			billItem.LocationID.Set(order.LocationID),
			billItem.BillType.Set(pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER),
			billItem.BillStatus.Set(pb.BillingStatus_BILLING_STATUS_BILLED),
			billItem.FinalPrice.Set(500),
		)
		if err != nil {
			return
		}

		stmtBillItem := `INSERT INTO bill_item(
			order_id,
			student_product_id,
			product_id,
			product_description,
			student_id,
			location_id,
			bill_type,
			billing_status,
			final_price,
			is_latest_bill_item,
			updated_at,
			created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())`
		_, err = fatimaDBTrace.Exec(ctx, stmtBillItem,
			billItem.OrderID,
			billItem.StudentProductID,
			billItem.ProductID,
			billItem.ProductDescription,
			billItem.StudentID,
			billItem.LocationID,
			billItem.BillType,
			billItem.BillStatus,
			billItem.FinalPrice,
			true)
		if err != nil {
			return
		}
		studentProductIDs = append(studentProductIDs, studentProduct.StudentProductID.String)
	}

	return orderID, locationID, studentProductIDs, nil
}

func InsertPackage(ctx context.Context, taxID string, fatimaDBTrace *database.DBTrace) (string, error) {
	var packageID string
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
		ProductTag           sql.NullString `json:"product_tag"`
		ProductPartnerID     sql.NullString `json:"product_partner_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    sql.NullString `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type AddPackageParams struct {
		PackageID        string       `json:"package_id"`
		PackageType      string       `json:"package_type"`
		MaxSlot          int32        `json:"max_slot"`
		PackageStartDate sql.NullTime `json:"package_start_date"`
		PackageEndDate   sql.NullTime `json:"package_end_date"`
	}
	var arg AddProductParams
	var packageArg AddPackageParams
	randomStr := idutil.ULIDNow()
	arg.ProductID = randomStr
	arg.Name = fmt.Sprintf("package-%v", randomStr)
	arg.ProductType = pb.ProductType_PRODUCT_TYPE_PACKAGE.String()
	arg.AvailableFrom = time.Now()
	arg.AvailableUtil = time.Now().AddDate(1, 0, 0)
	arg.DisableProRatingFlag = false
	arg.IsArchived = false
	arg.TaxID.String = taxID
	arg.TaxID.Valid = true

	stmt := `INSERT INTO product(
				    product_id,
                    name,
                    product_type,
                    tax_id,
					product_tag,
                    product_partner_id,
                    available_from,
                    available_until,
                    remarks,
                    custom_billing_period,
                    billing_schedule_id,
                    disable_pro_rating_flag,
                    is_archived,
                    updated_at,
                    created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now(), now())
                RETURNING product_id`
	row := fatimaDBTrace.QueryRow(ctx, stmt,
		arg.ProductID,
		arg.Name,
		arg.ProductType,
		arg.TaxID,
		arg.ProductTag,
		arg.ProductPartnerID,
		arg.AvailableFrom,
		arg.AvailableUtil,
		arg.Remarks,
		arg.CustomBillingPeriod,
		arg.BillingScheduleID,
		arg.DisableProRatingFlag,
		arg.IsArchived)
	err := row.Scan(&packageArg.PackageID)
	if err != nil {
		return "", fmt.Errorf("cannot insert product, err: %s", err)
	}
	packageID = packageArg.PackageID
	queryInsertPackage := `INSERT INTO package(
									package_id,
                                    package_type,
                                    max_slot,
                                    package_start_date,
                                    package_end_date)
                                VALUES ($1, $2, $3, $4, $5)`

	packageArg.PackageType = pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()
	packageArg.MaxSlot = 34
	packageArg.PackageStartDate = sql.NullTime{Time: time.Now().Truncate(24 * time.Hour), Valid: true}
	packageArg.PackageEndDate = sql.NullTime{Time: time.Now().AddDate(1, 0, 0).Truncate(24 * time.Hour), Valid: true}
	_, err = fatimaDBTrace.Exec(ctx, queryInsertPackage,
		packageArg.PackageID,
		packageArg.PackageType,
		packageArg.MaxSlot,
		packageArg.PackageStartDate,
		packageArg.PackageEndDate,
	)
	if err != nil {
		return "", fmt.Errorf("cannot insert package, err: %s", err)
	}

	return packageID, nil
}

func InsertCourse(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	e := paymentEntities.Course{}
	courseID := idutil.ULIDNow()
	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
		e.CourseID.Set(courseID),
		e.TeachingMethod.Set(nil),
		e.Grade.Set(1),
		e.Name.Set(fmt.Sprintf("test-course-%s", courseID)),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, &e, []string{"resource_path"}, fatimaDBTrace.Exec)
	if err != nil {
		return "", fmt.Errorf("err insert course: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return "", fmt.Errorf("err insert course: %d RowsAffected", cmdTag.RowsAffected())
	}

	return courseID, nil
}
