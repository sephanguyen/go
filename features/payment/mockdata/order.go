package mockdata

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

const (
	PriceOrder           = 500
	EnrolledProductPrice = 300
)

type DataForRecurringProduct struct {
	ProductIDs             []string
	BillingSchedulePeriods []*entities.BillingSchedulePeriod
	LocationID             string
	UserID                 string
	EnrolledUserID         string
	PotentialUserID        string
	DiscountIDs            []string
	TaxID                  string
	BillingScheduleID      string
	PackageCourses         []*entities.PackageCourse
	CourseIDs              []string
	LeavingReasonIDs       []string
}

type OptionToPrepareDataForCreateOrderRecurringProduct struct {
	InsertTax                            bool
	InsertDiscount                       bool
	InsertProductGrade                   bool
	InsertStudent                        bool
	InsertEnrolledStudent                bool
	InsertPotentialStudent               bool
	InsertMaterial                       bool
	InsertFee                            bool
	InsertProductPrice                   bool
	InsertEnrolledProductPrice           bool
	InsertProductPriceWithDifferentPrice bool
	InsertProductLocation                bool
	InsertLocation                       bool
	IsTaxExclusive                       bool
	InsertDiscountNotAvailable           bool
	InsertProductOutOfTime               bool
	InsertBillingSchedule                bool
	InsertBillingScheduleArchived        bool
	IsShorterPeriod                      bool
	InsertPackageCourses                 bool
	ArePackageCoursesMandatory           bool
	InsertPackageCourseScheduleBased     bool
	InsertProductDiscount                bool
	InsertMaterialUnique                 bool
	BillingScheduleStartDate             time.Time
	InsertLeavingReasons                 bool
	InsertNotificationDate               bool
	InsertProductSetting                 bool
}

type OptionToPrepareDataForCreateOrder struct {
	InsertTax                  bool
	InsertDiscount             bool
	InsertProductGrade         bool
	InsertStudent              bool
	InsertMaterial             bool
	InsertProductPrice         bool
	InsertProductLocation      bool
	InsertLocation             bool
	IsTaxExclusive             bool
	InsertFee                  bool
	InsertDiscountNotAvailable bool
	InsertProductOutOfTime     bool
	InsertProductDiscount      bool
	InsertOrgLevelDiscount     bool
	PriceOrder                 float32
	InsertNotificationDate     bool
}

func InsertAllDataForInsertOrder(ctx context.Context, fatimaDBTrace *database.DBTrace, optionPrepareData OptionToPrepareDataForCreateOrder, name string) (
	taxID string,
	discountIDs []string,
	locationID string,
	productIDs []string,
	userID string,
	err error,
) {
	gradeID, err := InsertOneGrade(ctx, fatimaDBTrace)
	if err != nil {
		return
	}

	if optionPrepareData.InsertTax {
		if optionPrepareData.IsTaxExclusive {
			taxID, err = InsertOneTaxExclusive(ctx, fatimaDBTrace, name)
			if err != nil {
				return
			}
		} else {
			taxID, err = InsertOneTax(ctx, fatimaDBTrace, name)
			if err != nil {
				return
			}
		}
	}

	if optionPrepareData.InsertDiscount {
		discountIDs, err = InsertOneDiscountAmount(ctx, fatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertDiscountNotAvailable {
		discountIDs, err = InsertOneDiscountAmountNotAvailable(ctx, fatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertLocation {
		locationID, err = InsertOneLocation(ctx, fatimaDBTrace)
		if err != nil {
			return
		}
	} else {
		locationID = constants.ManabieOrgLocation
	}

	if optionPrepareData.InsertStudent {
		userID, err = InsertOneUser(ctx, fatimaDBTrace, gradeID)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertMaterial {
		productIDs, err = InsertMaterial(ctx, fatimaDBTrace, taxID)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertFee {
		productIDs, err = InsertFee(ctx, fatimaDBTrace, taxID)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertProductOutOfTime {
		productIDs, err = InsertMaterialOutOfTime(ctx, fatimaDBTrace, taxID)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertProductLocation {
		err = InsertProductLocation(ctx, fatimaDBTrace, locationID, productIDs)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertProductPrice {
		err = InsertProductPrice(ctx, fatimaDBTrace, productIDs, optionPrepareData.PriceOrder)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertProductGrade {
		err = InsertProductGrade(ctx, fatimaDBTrace, gradeID, productIDs)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertProductDiscount {
		err = InsertProductDiscount(ctx, fatimaDBTrace, productIDs, discountIDs)
		if err != nil {
			return
		}
	}

	if optionPrepareData.InsertOrgLevelDiscount {
		var orgLevelDiscountID string
		orgLevelDiscountID, err = InsertOrgLevelDiscount(ctx, fatimaDBTrace, userID)
		if err != nil {
			return
		}

		discountIDs = append(discountIDs, orgLevelDiscountID)
	}
	return
}

func InsertOneTax(ctx context.Context, fatimaDBTrace *database.DBTrace, name string) (string, error) {
	taxName := database.Text(fmt.Sprintf("Tax for create order %s", name))
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

func InsertOneTaxExclusive(ctx context.Context, fatimaDBTrace *database.DBTrace, name string) (string, error) {
	taxName := database.Text(fmt.Sprintf("Tax for create order %s", name))
	taxPercentage := database.Int4(20)
	taxCategory := database.Text("TAX_CATEGORY_EXCLUSIVE")
	isArchived := database.Bool(false)
	stmt := `INSERT INTO tax
		(tax_id, name, tax_percentage, tax_category, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now()) RETURNING tax_id;`
	row := fatimaDBTrace.QueryRow(ctx, stmt, idutil.ULIDNow(), taxName, taxPercentage, taxCategory, isArchived)
	var taxID string
	err := row.Scan(&taxID)
	return taxID, err
}

func InsertOneDiscountAmount(ctx context.Context, fatimaDBTrace *database.DBTrace, name string) ([]string, error) {
	type addDiscountParams struct {
		Name                string         `json:"name"`
		DiscountType        string         `json:"discount_type"`
		DiscountAmountType  string         `json:"discount_amount_type"`
		DiscountAmountValue pgtype.Numeric `json:"discount_amount_value"`
		AvailableFrom       time.Time      `json:"available_from"`
		AvailableUtil       time.Time      `json:"available_until"`
		Remarks             string         `json:"remarks"`
		IsArchived          bool           `json:"is_archived"`
	}
	discountValue := pgtype.Numeric{}
	discountValue.Set(20)
	var discountIDs []string
	var discountID string
	discountList := []addDiscountParams{
		{
			Name:                fmt.Sprintf("Discount for create %s with type fixed amount", name),
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
			DiscountAmountValue: discountValue,
			AvailableFrom:       time.Now(),
			AvailableUtil:       time.Now().AddDate(1, 0, 0),
			Remarks:             "discount remarks",
			IsArchived:          false,
		},
		{
			Name:                fmt.Sprintf("Discount for create %s with type percentage", name),
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
			DiscountAmountValue: discountValue,
			AvailableFrom:       time.Now(),
			AvailableUtil:       time.Now().AddDate(1, 0, 0),
			Remarks:             "discount remarks",
			IsArchived:          false,
		},
	}

	for i := 0; i < 2; i++ {
		stmt := `INSERT INTO discount
		(discount_id, name, discount_type, discount_amount_type, discount_amount_value, available_from, available_until, remarks,  is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now()) RETURNING discount_id;`
		row := fatimaDBTrace.QueryRow(ctx, stmt, idutil.ULIDNow(), discountList[i].Name, discountList[i].DiscountType, discountList[i].DiscountAmountType, discountList[i].DiscountAmountValue, discountList[i].AvailableFrom, discountList[i].AvailableUtil, discountList[i].Remarks, discountList[i].IsArchived)

		err := row.Scan(&discountID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert discount, err: %s", err)
		}

		discountIDs = append(discountIDs, discountID)
	}

	return discountIDs, nil
}

func InsertOneDiscountAmountNotAvailable(ctx context.Context, fatimaDBTrace *database.DBTrace, name string) ([]string, error) {
	type addDiscountParams struct {
		Name                string         `json:"name"`
		DiscountType        string         `json:"discount_type"`
		DiscountAmountType  string         `json:"discount_amount_type"`
		DiscountAmountValue pgtype.Numeric `json:"discount_amount_value"`
		AvailableFrom       time.Time      `json:"available_from"`
		AvailableUtil       time.Time      `json:"available_until"`
		Remarks             string         `json:"remarks"`
		IsArchived          bool           `json:"is_archived"`
	}
	discountValue := pgtype.Numeric{}
	discountValue.Set(20)
	var discountIDs []string
	var discountID string
	discountList := []addDiscountParams{
		{
			Name:                fmt.Sprintf("Discount for create %s with type fixed amount", name),
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
			DiscountAmountValue: discountValue,
			AvailableFrom:       time.Now().AddDate(1, 0, 0),
			AvailableUtil:       time.Now().AddDate(2, 0, 0),
			Remarks:             "discount remarks",
			IsArchived:          false,
		},
		{
			Name:                fmt.Sprintf("Discount for create %s with type percentage", name),
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
			DiscountAmountValue: discountValue,
			AvailableFrom:       time.Now().AddDate(1, 0, 0),
			AvailableUtil:       time.Now().AddDate(2, 0, 0),
			Remarks:             "discount remarks",
			IsArchived:          false,
		},
		{
			Name:                fmt.Sprintf("Discount for create %s with type percentage", name),
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
			DiscountAmountValue: discountValue,
			AvailableFrom:       time.Now().AddDate(-2, 0, 0),
			AvailableUtil:       time.Now().AddDate(-1, 0, 0),
			Remarks:             "discount remarks",
			IsArchived:          false,
		},
	}

	for i := 0; i < len(discountList); i++ {
		stmt := `INSERT INTO discount
		(discount_id, name, discount_type, discount_amount_type, discount_amount_value, available_from, available_until, remarks,  is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now()) RETURNING discount_id;`
		row := fatimaDBTrace.QueryRow(ctx, stmt, idutil.ULIDNow(), discountList[i].Name, discountList[i].DiscountType, discountList[i].DiscountAmountType, discountList[i].DiscountAmountValue, discountList[i].AvailableFrom, discountList[i].AvailableUtil, discountList[i].Remarks, discountList[i].IsArchived)

		err := row.Scan(&discountID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert discount, err: %s", err)
		}

		discountIDs = append(discountIDs, discountID)
	}

	return discountIDs, nil
}

func InsertOneLocation(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	id := idutil.ULIDNow()
	name := database.Text("Location for create order material one time " + id)
	stmt := `INSERT INTO locations
		(location_id, name, created_at, updated_at)
		VALUES ($1, $2, now(), now())`
	_, err := fatimaDBTrace.Exec(ctx, stmt, id, name)
	if err != nil {
		return id, fmt.Errorf("cannot insert location, err: %s", err)
	}
	return id, nil
}

func InsertOneUser(ctx context.Context, fatimaDBTrace *database.DBTrace, gradeID string) (string, error) {
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
		VALUES ($1, $2, $3,$4, now(), now());`
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, 1, "STUDENT_ENROLLMENT_STATUS_POTENTIAL", gradeID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func InsertOneGrade(ctx context.Context, fatimaDBTrace *database.DBTrace) (string, error) {
	gradeID := idutil.ULIDNow()
	stmt := `INSERT INTO grade
	(grade_id, name, is_archived, partner_internal_id, sequence, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err := fatimaDBTrace.Exec(ctx, stmt, gradeID, "grade_name", false, gradeID, 1)
	if err != nil {
		return "", err
	}

	return gradeID, nil
}

func InsertOneEnrolledUser(ctx context.Context, fatimaDBTrace *database.DBTrace, gradeID string, locationID string) (string, error) {
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

func InsertOnePotentialUser(ctx context.Context, fatimaDBTrace *database.DBTrace, gradeID string) (string, error) {
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
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, 1, "STUDENT_ENROLLMENT_STATUS_POTENTIAL", gradeID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func InsertMaterialOutOfTime(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
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

	type AddMaterialParams struct {
		MaterialID        string       `json:"material_id"`
		MaterialType      string       `json:"material_type"`
		CustomBillingDate sql.NullTime `json:"custom_billing_date"`
	}
	var materialIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var materialArg AddMaterialParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("material-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		productArg.AvailableFrom = time.Now().AddDate(1, 0, 0)
		productArg.AvailableUtil = time.Now().AddDate(2, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.TaxID = sql.NullString{String: taxID, Valid: true}
		stmt := `INSERT INTO product (
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
		row := fatimaDBTrace.QueryRow(ctx, stmt,
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
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		materialIDs = append(materialIDs, materialArg.MaterialID)
		queryInsertPackage := `INSERT INTO material (material_id, material_type, custom_billing_date) VALUES ($1, $2, $3)
		`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
		materialArg.CustomBillingDate = sql.NullTime{Time: time.Now(), Valid: true}
		_, err = fatimaDBTrace.Exec(ctx, queryInsertPackage,
			materialArg.MaterialID,
			materialArg.MaterialType,
			materialArg.CustomBillingDate,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert material, err: %s", err)
		}
	}
	return materialIDs, nil
}

func InsertProductLocation(ctx context.Context, fatimaDBTrace *database.DBTrace, locationID string, materialIDs []string) error {
	queryInsertProductLocation := `INSERT INTO product_location (location_id, product_id, created_at) VALUES ($1, $2, now())
		`
	for _, materialID := range materialIDs {
		_, err := fatimaDBTrace.Exec(ctx, queryInsertProductLocation,
			locationID,
			materialID,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_location, err: %s", err)
		}
	}
	return nil
}

func InsertProductGrade(ctx context.Context, fatimaDBTrace *database.DBTrace, gradeID string, materialIDs []string) error {
	queryInsertProductGrade := `INSERT INTO product_grade (grade_id, product_id, created_at) VALUES ($1, $2, now())
		`
	for _, materialID := range materialIDs {
		_, err := fatimaDBTrace.Exec(ctx, queryInsertProductGrade,
			gradeID,
			materialID,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_grade, err: %s", err)
		}
	}
	return nil
}

func InsertProductDiscount(ctx context.Context, fatimaDBTrace *database.DBTrace, materialIDs []string, discountIDs []string) error {
	queryInsertProductDiscount := `INSERT INTO product_discount (product_id, discount_id, created_at) VALUES ($1, $2, now())
		`
	for _, productID := range materialIDs {
		for _, discountID := range discountIDs {
			_, err := fatimaDBTrace.Exec(ctx, queryInsertProductDiscount, productID, discountID)
			if err != nil {
				return fmt.Errorf("canot insert product_discount, err: %s", err)
			}
		}
	}
	return nil
}

func InsertProductPrice(ctx context.Context, fatimaDBTrace *database.DBTrace, materialIDs []string, priceOrder float32) error {
	queryInsertProductPrice := `INSERT INTO product_price (product_id, price, created_at) VALUES ($1, $2, now())
		`
	for _, materialID := range materialIDs {
		_, err := fatimaDBTrace.Exec(ctx, queryInsertProductPrice,
			materialID,
			priceOrder,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_price, err: %s", err)
		}
	}
	return nil
}

func InsertProductPriceForRecurringProducts(ctx context.Context, fatimaDBTrace *database.DBTrace, materialIDs []string, billingPeriods []*entities.BillingSchedulePeriod, priceOrder float32, priceType string) error {
	queryInsertProductPrice := `INSERT INTO product_price (product_id, price, billing_schedule_period_id, created_at, price_type) VALUES ($1, $2, $3, now(), $4)
		`
	for _, materialID := range materialIDs {
		for _, period := range billingPeriods {
			_, err := fatimaDBTrace.Exec(ctx, queryInsertProductPrice,
				materialID,
				priceOrder,
				period.BillingSchedulePeriodID,
				priceType,
			)
			if err != nil {
				return fmt.Errorf("cannot insert product_price, err: %s", err)
			}
		}
	}
	return nil
}

func InsertEnrolledProductPriceForRecurringProducts(ctx context.Context, fatimaDBTrace *database.DBTrace, materialIDs []string, billingPeriods []*entities.BillingSchedulePeriod, priceOrder float32) error {
	queryInsertProductPrice := `INSERT INTO product_price (product_id, price, billing_schedule_period_id, created_at) VALUES ($1, $2, $3, now())
		`
	for _, materialID := range materialIDs {
		for _, period := range billingPeriods {
			_, err := fatimaDBTrace.Exec(ctx, queryInsertProductPrice,
				materialID,
				priceOrder,
				period.BillingSchedulePeriodID,
			)
			if err != nil {
				return fmt.Errorf("cannot insert product_price, err: %s", err)
			}
		}
	}
	return nil
}
func InsertDifferentProductPriceForRecurringProducts(ctx context.Context, fatimaDBTrace *database.DBTrace, materialIDs []string, billingPeriods []*entities.BillingSchedulePeriod, priceOrder float32) error {
	queryInsertProductPrice := `INSERT INTO product_price (product_id, price, billing_schedule_period_id, price_type, created_at) VALUES ($1, $2, $3, $4, now())
		`
	for _, materialID := range materialIDs {
		for i, period := range billingPeriods {
			_, err := fatimaDBTrace.Exec(ctx, queryInsertProductPrice,
				materialID,
				priceOrder+float32(50*i),
				period.BillingSchedulePeriodID,
				"DEFAULT_PRICE",
			)
			if err != nil {
				return fmt.Errorf("cannot insert product_price, err: %s", err)
			}
			_, err = fatimaDBTrace.Exec(ctx, queryInsertProductPrice,
				materialID,
				(priceOrder-100)+float32(50*i),
				period.BillingSchedulePeriodID,
				"ENROLLED_PRICE",
			)
			if err != nil {
				return fmt.Errorf("cannot insert product_price, err: %s", err)
			}
		}
	}
	return nil
}

func InsertMaterial(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
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

	type AddMaterialParams struct {
		MaterialID        string       `json:"material_id"`
		MaterialType      string       `json:"material_type"`
		CustomBillingDate sql.NullTime `json:"custom_billing_date"`
	}
	var materialIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var materialArg AddMaterialParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("material-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.TaxID = sql.NullString{String: taxID, Valid: true}
		stmt := `INSERT INTO product (
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
		row := fatimaDBTrace.QueryRow(ctx, stmt,
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
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		materialIDs = append(materialIDs, materialArg.MaterialID)
		queryInsertPackage := `INSERT INTO material (material_id, material_type, custom_billing_date) VALUES ($1, $2, $3)
		`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
		materialArg.CustomBillingDate = sql.NullTime{Time: time.Now(), Valid: true}
		_, err = fatimaDBTrace.Exec(ctx, queryInsertPackage,
			materialArg.MaterialID,
			materialArg.MaterialType,
			materialArg.CustomBillingDate,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert material, err: %s", err)
		}
	}
	return materialIDs, nil
}

func InsertFee(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string) ([]string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
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

	type AddFeeParams struct {
		FeeID   string `json:"fee_id"`
		FeeType string `json:"fee_type"`
	}
	var feeIDs []string
	for i := 0; i < 3; i++ {
		var productArg AddProductParams
		var feeArg AddFeeParams
		randomStr := idutil.ULIDNow()
		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("fee-for-create-order-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_FEE.String()
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.TaxID = sql.NullString{String: taxID, Valid: true}
		stmt := `INSERT INTO product (
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
		row := fatimaDBTrace.QueryRow(ctx, stmt,
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
		err := row.Scan(&feeArg.FeeID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		feeIDs = append(feeIDs, feeArg.FeeID)
		queryInsertPackage := `INSERT INTO fee (fee_id, fee_type) VALUES ($1, $2)
		`
		feeArg.FeeType = pb.FeeType_FEE_TYPE_ONE_TIME.String()
		_, err = fatimaDBTrace.Exec(ctx, queryInsertPackage,
			feeArg.FeeID,
			feeArg.FeeType,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert fee, err: %s", err)
		}
	}
	return feeIDs, nil
}

func InsertBillingScheduleForRecurringProduct(ctx context.Context, fatimaDBTrace *database.DBTrace, isArchived bool) (string, error) {
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
		isArchived,
	)
	err := row.Scan(&billingScheduleID)
	if err != nil {
		return billingScheduleID, fmt.Errorf("cannot insert billingScheduleID, err: %s", err)
	}

	return billingScheduleID, nil
}

func InsertBillingSchedulePeriodForRecurringProduct(ctx context.Context, fatimaDBTrace *database.DBTrace, billingSchedule entities.BillingSchedule, scheduleStartDate time.Time) error {
	for i := 0; i < 4; i++ {
		randomStr := idutil.ULIDNow()
		periodStartDate := scheduleStartDate.AddDate(0, 0, (i * 29))
		periodEndDate := periodStartDate.AddDate(0, 0, 28)
		periodBillingDate := periodStartDate.AddDate(0, -1, 15)

		billingPeriod := entities.BillingSchedulePeriod{
			Name: pgtype.Text{
				String: fmt.Sprintf("billing-period-%s", randomStr),
				Status: pgtype.Present,
			},
			BillingScheduleID: pgtype.Text{
				String: "1",
				Status: pgtype.Present,
			},
			StartDate: pgtype.Timestamptz{
				Time:   periodStartDate,
				Status: pgtype.Present,
			},
			EndDate: pgtype.Timestamptz{
				Time:   periodEndDate,
				Status: pgtype.Present,
			},
			BillingDate: pgtype.Timestamptz{
				Time:   periodBillingDate,
				Status: pgtype.Present,
			},
			IsArchived: pgtype.Bool{
				Bool:   false,
				Status: pgtype.Present,
			},
		}

		stmt := `
		INSERT INTO billing_schedule_period (
			billing_schedule_period_id,
			name,
			billing_schedule_id,
			start_date,
			end_date,
			billing_date,
			is_archived,
			created_at,
			updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, now(), now())`
		_, err := fatimaDBTrace.Exec(
			ctx,
			stmt,
			randomStr,
			billingPeriod.Name,
			billingSchedule.BillingScheduleID,
			billingPeriod.StartDate,
			billingPeriod.EndDate,
			billingPeriod.BillingDate,
			billingPeriod.IsArchived,
		)
		if err != nil {
			return fmt.Errorf("cannot insert billing schedule period, err: %s", err)
		}
	}

	return nil
}

func InsertBillingSchedulePeriodForRecurringProduct_ShorterPeriod(ctx context.Context, fatimaDBTrace *database.DBTrace, billingSchedule entities.BillingSchedule, scheduleStartDate time.Time) error {
	for i := 0; i < 4; i++ {
		randomStr := idutil.ULIDNow()
		periodStartDate := scheduleStartDate.AddDate(0, 0, (i * 19))
		periodEndDate := periodStartDate.AddDate(0, 0, 18)
		periodBillingDate := periodStartDate.AddDate(0, 0, -15)

		billingPeriod := entities.BillingSchedulePeriod{
			Name: pgtype.Text{
				String: fmt.Sprintf("billing-period-%s", randomStr),
				Status: pgtype.Present,
			},
			BillingScheduleID: pgtype.Text{
				String: "1",
				Status: pgtype.Present,
			},
			StartDate: pgtype.Timestamptz{
				Time:   periodStartDate,
				Status: pgtype.Present,
			},
			EndDate: pgtype.Timestamptz{
				Time:   periodEndDate,
				Status: pgtype.Present,
			},
			BillingDate: pgtype.Timestamptz{
				Time:   periodBillingDate,
				Status: pgtype.Present,
			},
			IsArchived: pgtype.Bool{
				Bool:   false,
				Status: pgtype.Present,
			},
		}

		stmt := `
		INSERT INTO billing_schedule_period (
			billing_schedule_period_id,
			name,
			billing_schedule_id,
			start_date,
			end_date,
			billing_date,
			is_archived,
			created_at,
			updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, now(), now())`
		_, err := fatimaDBTrace.Exec(
			ctx,
			stmt,
			randomStr,
			billingPeriod.Name,
			billingSchedule.BillingScheduleID,
			billingPeriod.StartDate,
			billingPeriod.EndDate,
			billingPeriod.BillingDate,
			billingPeriod.IsArchived,
		)
		if err != nil {
			return fmt.Errorf("cannot insert billing schedule period, err: %s", err)
		}
	}

	return nil
}

func InsertBillingRatioForRecurringProduct(ctx context.Context, fatimaDBTrace *database.DBTrace, billingSchedule entities.BillingSchedule) error {
	billingPeriods, err := GetBillingPeriodBySchedule(ctx, fatimaDBTrace, billingSchedule)
	if err != nil {
		return errors.Wrap(err, "error retrieving billing periods")
	}

	for _, bp := range billingPeriods {
		ratioStartDate := bp.StartDate.Time.UTC()
		var ratioEndDate time.Time

		for j := 4; j > 0; j-- {
			ratioEndDate = ratioStartDate.AddDate(0, 0, 7)

			if j == 1 {
				ratioEndDate = bp.EndDate.Time.UTC()
			}

			billingRatio := entities.BillingRatio{
				BillingSchedulePeriodID: pgtype.Text{
					String: bp.BillingSchedulePeriodID.String,
					Status: pgtype.Present,
				},
				StartDate: pgtype.Timestamptz{
					Time:   ratioStartDate,
					Status: pgtype.Present,
				},
				EndDate: pgtype.Timestamptz{
					Time:   ratioEndDate,
					Status: pgtype.Present,
				},
				BillingRatioNumerator: pgtype.Int4{
					Int:    int32(j),
					Status: pgtype.Present,
				},
				BillingRatioDenominator: pgtype.Int4{
					Int:    4,
					Status: pgtype.Present,
				},
				IsArchived: pgtype.Bool{
					Bool:   false,
					Status: pgtype.Present,
				},
			}

			stmt := `
				INSERT INTO billing_ratio (
					billing_ratio_id,
					start_date,
					end_date,
					billing_schedule_period_id,
					billing_ratio_numerator,
					billing_ratio_denominator,
					is_archived,
					created_at,
					updated_at)
				VALUES
					($1, $2, $3, $4, $5, $6, $7, now(), now())`
			_, err = fatimaDBTrace.Exec(
				ctx,
				stmt,
				idutil.ULIDNow(),
				billingRatio.StartDate,
				billingRatio.EndDate,
				billingRatio.BillingSchedulePeriodID,
				billingRatio.BillingRatioNumerator,
				billingRatio.BillingRatioDenominator,
				billingRatio.IsArchived)
			if err != nil {
				return fmt.Errorf("cannot insert billing ratio, err: %s", err)
			}

			ratioStartDate = ratioStartDate.AddDate(0, 0, 8)
		}
	}
	return nil
}

func InsertRecurringFees(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string, billingScheduleID string) ([]string, error) {
	type ProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                string         `json:"tax_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    string         `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type FeeParams struct {
		FeeID        string         `json:"fee_id"`
		MaterialType string         `json:"fee_type"`
		ResourcePath sql.NullString `json:"resource_path"`
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
			materialArg       FeeParams
			productSettingArg ProductSettingParams
		)

		randomStr := idutil.ULIDNow()
		currentTime := time.Now()

		productArg.ProductID = randomStr
		productArg.Name = fmt.Sprintf("material-recurring-%v", randomStr)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_FEE.String()
		productArg.AvailableFrom = currentTime.AddDate(-1, 0, 0)
		productArg.AvailableUtil = currentTime.AddDate(1, 0, 0)
		productArg.IsArchived = false
		productArg.TaxID = taxID
		productArg.BillingScheduleID = billingScheduleID
		productArg.DisableProRatingFlag = false
		if i%2 != 0 {
			productArg.DisableProRatingFlag = true
		}

		productSettingArg.ProductID = randomStr
		productSettingArg.IsEnrollmentRequired = false
		productSettingArg.IsPausable = true
		productSettingArg.IsAddedToEnrollmentByDefault = false
		productSettingArg.IsOperationFee = false

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
		err := row.Scan(&materialArg.FeeID)
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

		stmtInsertFee := `
			INSERT INTO fee (
				fee_id,
				fee_type)
			VALUES ($1, $2)`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_RECURRING.String()
		_, err = fatimaDBTrace.Exec(
			ctx,
			stmtInsertFee,
			materialArg.FeeID,
			materialArg.MaterialType,
		)
		if err != nil {
			return productIDs, fmt.Errorf("cannot insert fee, err: %s", err)
		}

		productIDs = append(productIDs, materialArg.FeeID)
	}
	return productIDs, nil
}

func InsertDiscountForRecurringProduct(ctx context.Context, fatimaDBTrace *database.DBTrace) ([]string, error) {
	type discountParams struct {
		Name                   string         `json:"name"`
		DiscountType           string         `json:"discount_type"`
		DiscountAmountType     string         `json:"discount_amount_type"`
		DiscountAmountValue    pgtype.Numeric `json:"discount_amount_value"`
		RecurringValidDuration pgtype.Int4    `json:"recurring_valid_duration"`
		AvailableFrom          time.Time      `json:"available_from"`
		AvailableUtil          time.Time      `json:"available_until"`
		Remarks                string         `json:"remarks"`
		IsArchived             bool           `json:"is_archived"`
	}
	discountValue := pgtype.Numeric{}
	err := discountValue.Set(10)
	if err != nil {
		return nil, err
	}

	recurringValidDurationInfinite := pgtype.Int4{Status: pgtype.Null}
	recurringValidDurationFinite := pgtype.Int4{Int: 2, Status: pgtype.Present}

	var discountID string
	discountList := []discountParams{
		{
			Name:                   "Discount for create recurring product fixed amount finite use",
			DiscountType:           pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:     pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
			DiscountAmountValue:    discountValue,
			RecurringValidDuration: recurringValidDurationFinite,
			AvailableFrom:          time.Now().AddDate(-1, 0, 0),
			AvailableUtil:          time.Now().AddDate(1, 0, 0),
			Remarks:                "discount for recurring product test",
			IsArchived:             false,
		},
		{
			Name:                   "Discount for create recurring product fixed amount infinite use",
			DiscountType:           pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:     pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT.String(),
			DiscountAmountValue:    discountValue,
			RecurringValidDuration: recurringValidDurationInfinite,
			AvailableFrom:          time.Now().AddDate(-1, 0, 0),
			AvailableUtil:          time.Now().AddDate(1, 0, 0),
			Remarks:                "discount for recurring product test",
			IsArchived:             false,
		},
		{
			Name:                   "Discount for create recurring product percent type finite use",
			DiscountType:           pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:     pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
			DiscountAmountValue:    discountValue,
			RecurringValidDuration: recurringValidDurationFinite,
			AvailableFrom:          time.Now().AddDate(-1, 0, 0),
			AvailableUtil:          time.Now().AddDate(1, 0, 0),
			Remarks:                "discount for recurring product test",
			IsArchived:             false,
		},
		{
			Name:                   "Discount for create recurring product percent type infinite use",
			DiscountType:           pb.DiscountType_DISCOUNT_TYPE_REGULAR.String(),
			DiscountAmountType:     pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(),
			DiscountAmountValue:    discountValue,
			RecurringValidDuration: recurringValidDurationInfinite,
			AvailableFrom:          time.Now().AddDate(-1, 0, 0),
			AvailableUtil:          time.Now().AddDate(1, 0, 0),
			Remarks:                "discount for recurring product test",
			IsArchived:             false,
		},
	}
	discountIDs := make([]string, len(discountList))

	for i, discount := range discountList {
		stmt := `
			INSERT INTO discount (
				discount_id,
				name,
				discount_type,
				discount_amount_type,
				discount_amount_value,
				recurring_valid_duration,
				available_from,
				available_until,
				remarks, 
				is_archived,
				created_at,
				updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
			RETURNING discount_id;`
		row := fatimaDBTrace.QueryRow(
			ctx,
			stmt,
			idutil.ULIDNow(),
			discount.Name,
			discount.DiscountType,
			discount.DiscountAmountType,
			discount.DiscountAmountValue,
			discount.RecurringValidDuration,
			discount.AvailableFrom,
			discount.AvailableUtil,
			discount.Remarks,
			discount.IsArchived)

		err := row.Scan(&discountID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert discount, err: %s", err)
		}

		discountIDs[i] = discountID
	}

	return discountIDs, nil
}

func GetBillingSchedule(ctx context.Context, fatimaDBTrace *database.DBTrace, id string) (entities.BillingSchedule, error) {
	schedule := entities.BillingSchedule{}

	stmt := `
		SELECT
			billing_schedule_id,
			name,
		    remarks,
			is_archived
		FROM
			billing_schedule
		WHERE
			billing_schedule_id = $1
		`
	rows, err := fatimaDBTrace.Query(
		ctx,
		stmt,
		id,
	)
	if err != nil {
		return schedule, errors.Wrap(err, "query billing_schedule")
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&schedule.BillingScheduleID,
			&schedule.Name,
			&schedule.Remarks,
			&schedule.IsArchived,
		)
		if err != nil {
			return schedule, errors.WithMessage(err, "rows.Scan billing schedule")
		}
	}
	return schedule, nil
}

func GetBillingPeriodBySchedule(ctx context.Context, fatimaDBTrace *database.DBTrace, billingSchedule entities.BillingSchedule) ([]*entities.BillingSchedulePeriod, error) {
	billingPeriods := []*entities.BillingSchedulePeriod{}
	stmt :=
		`
		SELECT
			billing_schedule_period_id,
			name,
			billing_schedule_id,
			start_date,
			end_date,
			billing_date,
			remarks,
			is_archived
		FROM
			billing_schedule_period
		WHERE
			billing_schedule_id = $1
		ORDER BY
			billing_date
		`
	rows, err := fatimaDBTrace.Query(
		ctx,
		stmt,
		billingSchedule.BillingScheduleID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query billing_schedule_period")
	}
	defer rows.Close()

	for rows.Next() {
		period := &entities.BillingSchedulePeriod{}
		err := rows.Scan(
			&period.BillingSchedulePeriodID,
			&period.Name,
			&period.BillingScheduleID,
			&period.StartDate,
			&period.EndDate,
			&period.BillingDate,
			&period.Remarks,
			&period.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan billing schedule period")
		}
		billingPeriods = append(billingPeriods, period)
	}
	return billingPeriods, nil
}

func InsertRecurringUniqueMaterials(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string, billingScheduleID string) ([]string, error) {
	type ProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                string         `json:"tax_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    string         `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		IsUnique             bool           `json:"is_unique"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type MaterialParams struct {
		MaterialID        string         `json:"material_id"`
		MaterialType      string         `json:"material_type"`
		CustomBillingDate sql.NullTime   `json:"custom_billing_date"`
		ResourcePath      sql.NullString `json:"resource_path"`
	}

	productIDs := []string{}

	for i := 0; i < 3; i++ {
		var productArg ProductParams
		var materialArg MaterialParams
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
		productArg.IsUnique = true
		productArg.DisableProRatingFlag = false
		if i%2 != 0 {
			productArg.DisableProRatingFlag = true
		}

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
				is_unique,
                updated_at,
                created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now(), now())
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
			productArg.IsUnique,
		)
		err := row.Scan(&materialArg.MaterialID)
		if err != nil {
			return productIDs, fmt.Errorf("cannot insert product, err: %s", err)
		}
		stmtInsertMaterial := `
			INSERT INTO material (
				material_id,
				material_type)
			VALUES ($1, $2)`
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_RECURRING.String()
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

func InsertRecurringMaterials(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string, billingScheduleID string) ([]string, error) {
	type ProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                string         `json:"tax_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    string         `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type MaterialParams struct {
		MaterialID        string         `json:"material_id"`
		MaterialType      string         `json:"material_type"`
		CustomBillingDate sql.NullTime   `json:"custom_billing_date"`
		ResourcePath      sql.NullString `json:"resource_path"`
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
		if i%2 != 0 {
			productArg.DisableProRatingFlag = true
		}

		productSettingArg.ProductID = randomStr
		productSettingArg.IsEnrollmentRequired = false
		productSettingArg.IsPausable = true
		productSettingArg.IsAddedToEnrollmentByDefault = false
		productSettingArg.IsOperationFee = false

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
		materialArg.MaterialType = pb.MaterialType_MATERIAL_TYPE_RECURRING.String()
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

func InsertPackageCourses(ctx context.Context, fatimaDBTrace *database.DBTrace, taxID string, billingScheduleID string, isPackageCourseScheduleBased bool, arePackageCoursesMandatory bool) ([]string, []string, error) {
	type AddProductParams struct {
		ProductID            string         `json:"product_id"`
		Name                 string         `json:"name"`
		ProductType          string         `json:"product_type"`
		TaxID                sql.NullString `json:"tax_id"`
		AvailableFrom        time.Time      `json:"available_from"`
		AvailableUtil        time.Time      `json:"available_until"`
		Remarks              sql.NullString `json:"remarks"`
		CustomBillingPeriod  sql.NullTime   `json:"custom_billing_period"`
		BillingScheduleID    string         `json:"billing_schedule_id"`
		DisableProRatingFlag bool           `json:"disable_pro_rating_flag"`
		IsArchived           bool           `json:"is_archived"`
		UpdatedAt            time.Time      `json:"updated_at"`
		CreatedAt            time.Time      `json:"created_at"`
	}

	type AddPackageParams struct {
		PackageID        string    `json:"package_id"`
		PackageType      string    `json:"package_type"`
		MaxSlot          int32     `json:"max_slot"`
		PackageStartDate time.Time `json:"package_start_date"`
		PackageEndDate   time.Time `json:"package_end_date"`
	}

	type AddCourseParams struct {
		CourseID  string    `json:"course_id"`
		Name      string    `json:"course_name"`
		Grade     int32     `json:"grade"`
		UpdatedAt time.Time `json:"updated_at"`
		CreatedAt time.Time `json:"created_at"`
	}

	type AddPackageCourseParams struct {
		PackageID         string    `json:"package_id"`
		CourseID          string    `json:"course_id"`
		MandatoryFlag     bool      `json:"mandatory_flag"`
		CourseWeight      int32     `json:"course_weight"`
		MaxSlotsPerCourse int32     `json:"max_slots_per_course"`
		CreatedAt         time.Time `json:"created_at"`
	}

	type ProductSettingParams struct {
		ProductID                    string
		IsEnrollmentRequired         bool
		IsPausable                   bool
		IsAddedToEnrollmentByDefault bool
		IsOperationFee               bool
	}

	var (
		packageIDs        []string
		courseIDs         []string
		productArgs       AddProductParams
		packageArgs       AddPackageParams
		courseArgs        AddCourseParams
		packageCourseArgs AddPackageCourseParams
		productSettingArg ProductSettingParams
	)

	randomStr := idutil.ULIDNow()

	productArgs.ProductID = randomStr
	productArgs.Name = fmt.Sprintf("package-%v", randomStr)
	productArgs.ProductType = pb.ProductType_PRODUCT_TYPE_PACKAGE.String()
	productArgs.AvailableFrom = time.Now().AddDate(-1, 0, 0)
	productArgs.AvailableUtil = time.Now().AddDate(1, 0, 0)
	productArgs.DisableProRatingFlag = false
	productArgs.IsArchived = false
	productArgs.TaxID.String = taxID
	productArgs.BillingScheduleID = billingScheduleID
	productArgs.TaxID.Valid = true

	productSettingArg.ProductID = randomStr
	productSettingArg.IsEnrollmentRequired = false
	productSettingArg.IsPausable = true
	productSettingArg.IsAddedToEnrollmentByDefault = false
	productSettingArg.IsOperationFee = false

	queryInsertProduct := `INSERT INTO product(
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
	row := fatimaDBTrace.QueryRow(ctx, queryInsertProduct,
		productArgs.ProductID,
		productArgs.Name,
		productArgs.ProductType,
		productArgs.TaxID,
		productArgs.AvailableFrom,
		productArgs.AvailableUtil,
		productArgs.Remarks,
		productArgs.CustomBillingPeriod,
		productArgs.BillingScheduleID,
		productArgs.DisableProRatingFlag,
		productArgs.IsArchived)
	err := row.Scan(&packageArgs.PackageID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot insert product, err: %s", err)
	}

	if isPackageCourseScheduleBased {
		packageArgs.PackageType = pb.PackageType_PACKAGE_TYPE_SCHEDULED.String()
	} else {
		packageArgs.PackageType = pb.PackageType_PACKAGE_TYPE_FREQUENCY.String()
	}

	packageArgs.MaxSlot = 5
	packageArgs.PackageStartDate = time.Now().AddDate(-1, 0, 0)
	packageArgs.PackageEndDate = time.Now().AddDate(1, 0, 0)
	queryInsertPackage := `INSERT INTO package (
					package_id,
                	package_type,
                	max_slot,
                	package_start_date,
                	package_end_date)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING package_id`
	row = fatimaDBTrace.QueryRow(ctx, queryInsertPackage,
		packageArgs.PackageID,
		packageArgs.PackageType,
		packageArgs.MaxSlot,
		packageArgs.PackageStartDate,
		packageArgs.PackageEndDate,
	)
	err = row.Scan(&packageArgs.PackageID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot insert package, err: %s", err)
	}

	packageIDs = append(packageIDs, packageArgs.PackageID)

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
		return packageIDs, courseIDs, fmt.Errorf("cannot insert product setting, err: %s", err)
	}

	for j := 0; j < 3; j++ {
		randomCourseID := idutil.ULIDNow()
		courseArgs.CourseID = randomCourseID
		courseArgs.Name = fmt.Sprintf("course-%s", randomCourseID)
		courseArgs.Grade = 1
		queryInsertCourse := `INSERT INTO courses (
					course_id,
					name,
					grade,
					updated_at,
					created_at)
				VALUES ($1, $2, $3, NOW(), NOW())
				RETURNING course_id`
		row := fatimaDBTrace.QueryRow(ctx, queryInsertCourse,
			courseArgs.CourseID,
			courseArgs.Name,
			courseArgs.Grade,
		)
		err := row.Scan(&courseArgs.CourseID)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot insert course, err: %s", err)
		}

		courseIDs = append(courseIDs, courseArgs.CourseID)
		packageCourseArgs.PackageID = packageArgs.PackageID
		packageCourseArgs.CourseID = courseArgs.CourseID
		packageCourseArgs.CourseWeight = 1
		packageCourseArgs.MaxSlotsPerCourse = 2
		packageCourseArgs.MandatoryFlag = arePackageCoursesMandatory
		queryInsertPackageCourse := `INSERT INTO package_course (
				package_id,
				course_id,
				mandatory_flag,
				course_weight,
				max_slots_per_course,
				created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())`
		_, err = fatimaDBTrace.Exec(ctx, queryInsertPackageCourse,
			packageCourseArgs.PackageID,
			packageCourseArgs.CourseID,
			packageCourseArgs.MandatoryFlag,
			packageCourseArgs.CourseWeight,
			packageCourseArgs.MaxSlotsPerCourse)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot insert course, err: %s", err)
		}
	}
	return packageIDs, courseIDs, nil
}

func GetPackageCourseByPackageIDs(ctx context.Context, fatimaDBTrace *database.DBTrace, packageIDs []string) ([]*entities.PackageCourse, error) {
	packageCourses := []*entities.PackageCourse{}
	stmt := `
		SELECT
			package_id,
			course_id,
			mandatory_flag,
			course_weight,
			max_slots_per_course,
			created_at
		FROM
			package_course
		WHERE
			package_id = $1`
	for _, packageID := range packageIDs {
		rows, err := fatimaDBTrace.Query(
			ctx,
			stmt,
			packageID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "query package course")
		}
		defer rows.Close()
		for rows.Next() {
			packageCourse := &entities.PackageCourse{}
			err := rows.Scan(
				&packageCourse.PackageID,
				&packageCourse.CourseID,
				&packageCourse.MandatoryFlag,
				&packageCourse.CourseWeight,
				&packageCourse.MaxSlotsPerCourse,
				&packageCourse.CreatedAt,
			)
			if err != nil {
				return nil, errors.WithMessage(err, "rows.Scan package course")
			}
			packageCourses = append(packageCourses, packageCourse)
		}
	}
	return packageCourses, nil
}

func InsertProductPriceForQtyPackage(ctx context.Context, fatimaDBTrace *database.DBTrace, packageIDs []string, billingPeriods []*entities.BillingSchedulePeriod) error {
	queryInsertProductPrice := `INSERT INTO product_price (
			product_id,
			billing_schedule_period_id,
			quantity,
			price,
			created_at)
		VALUES ($1, $2, $3, $4, now())`
	basePrice := 100

	for _, packageID := range packageIDs {
		for _, period := range billingPeriods {
			for j := 1; j <= 5; j++ {
				_, err := fatimaDBTrace.Exec(ctx, queryInsertProductPrice,
					packageID,
					period.BillingSchedulePeriodID,
					j,
					(j)*basePrice,
				)
				if err != nil {
					return fmt.Errorf("cannot insert product_price, err: %s", err)
				}
			}
		}
	}
	return nil
}

func InsertPackageTypeQuantityTypeMapping(ctx context.Context, fatimaDBTrace *database.DBTrace) error {
	packageTypeAndQuantityTypeMapping := []*entities.PackageQuantityTypeMapping{
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_ONE_TIME.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String(), Status: pgtype.Present},
		},
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_SLOT_BASED.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(), Status: pgtype.Present},
		},
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String(), Status: pgtype.Present},
		},
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_SCHEDULED.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String(), Status: pgtype.Present},
		},
	}

	queryInsertPackageQuantityTypeMapping := `INSERT INTO package_quantity_type_mapping (
			package_type,
			quantity_type,
			created_at)
		VALUES ($1, $2, now())
		ON CONFLICT DO NOTHING`
	for _, item := range packageTypeAndQuantityTypeMapping {
		_, err := fatimaDBTrace.Exec(ctx, queryInsertPackageQuantityTypeMapping, item.PackageType, item.QuantityType)
		if err != nil {
			return fmt.Errorf("err insert packageTypeAndQuantityTypeMapping: %w", err)
		}
	}
	return nil
}

func UpdateDisabledProratingFlagByProductID(ctx context.Context, fatimaDBTrace *database.DBTrace, productID string, disabledProrating bool) (err error) {
	stmt := `UPDATE product SET disable_pro_rating_flag = $1, updated_at = now()
		WHERE product_id = $2`
	_, err = fatimaDBTrace.Exec(ctx, stmt, disabledProrating, productID)
	if err != nil {
		return fmt.Errorf("err updating disabled prorating flag of product ID: %s", productID)
	}
	return nil
}

func InsertDataForRecurringProduct(ctx context.Context, fatimaDBTrace *database.DBTrace, options OptionToPrepareDataForCreateOrderRecurringProduct, gradeID string) (data DataForRecurringProduct, err error) {
	name := "recurring product test"

	if len(gradeID) == 0 {
		gradeID, err = InsertOneGrade(ctx, fatimaDBTrace)
		if err != nil {
			return data, err
		}
	}

	if options.InsertBillingSchedule {
		data.BillingScheduleID, err = InsertBillingScheduleForRecurringProduct(ctx, fatimaDBTrace, options.InsertBillingScheduleArchived)
		if err != nil {
			return data, err
		}

		billingSchedule, err := GetBillingSchedule(ctx, fatimaDBTrace, data.BillingScheduleID)
		if err != nil {
			return data, err
		}

		if options.IsShorterPeriod {
			err = InsertBillingSchedulePeriodForRecurringProduct_ShorterPeriod(ctx, fatimaDBTrace, billingSchedule, options.BillingScheduleStartDate)
			if err != nil {
				return data, err
			}
		} else {
			err = InsertBillingSchedulePeriodForRecurringProduct(ctx, fatimaDBTrace, billingSchedule, options.BillingScheduleStartDate)
			if err != nil {
				return data, err
			}
		}

		err = InsertBillingRatioForRecurringProduct(ctx, fatimaDBTrace, billingSchedule)
		if err != nil {
			return data, err
		}

		data.BillingSchedulePeriods, err = GetBillingPeriodBySchedule(ctx, fatimaDBTrace, billingSchedule)
		if err != nil {
			return data, err
		}
	}

	data.TaxID, data.DiscountIDs, data.LocationID, data.UserID, data.EnrolledUserID, data.PotentialUserID, err = InsertPreconditionData(ctx, fatimaDBTrace, options, data, gradeID, name)
	if err != nil {
		return
	}

	if options.InsertMaterial {
		data.ProductIDs, err = InsertRecurringMaterials(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID)
		if err != nil {
			return
		}
	}

	if options.InsertMaterialUnique {
		data.ProductIDs, err = InsertRecurringUniqueMaterials(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID)
		if err != nil {
			return
		}
	}

	if options.InsertFee {
		data.ProductIDs, err = InsertRecurringFees(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID)
		if err != nil {
			return
		}
	}

	if options.InsertPackageCourses {
		err = InsertPackageTypeQuantityTypeMapping(ctx, fatimaDBTrace)
		if err != nil {
			return
		}

		data.ProductIDs, data.CourseIDs, err = InsertPackageCourses(ctx, fatimaDBTrace, data.TaxID, data.BillingScheduleID, options.InsertPackageCourseScheduleBased, options.ArePackageCoursesMandatory)
		if err != nil {
			return
		}
		data.PackageCourses, err = GetPackageCourseByPackageIDs(ctx, fatimaDBTrace, data.ProductIDs)
		if err != nil {
			return
		}
		err = InsertProductPriceForQtyPackage(ctx, fatimaDBTrace, data.ProductIDs, data.BillingSchedulePeriods)
		if err != nil {
			return
		}
	}

	if options.InsertLeavingReasons {
		data.LeavingReasonIDs, err = InsertLeavingReasonsAndReturnID(ctx, fatimaDBTrace)
		if err != nil {
			return
		}
	}

	err = InsertReferenceData(ctx, fatimaDBTrace, options, data, gradeID)
	if err != nil {
		return
	}
	return
}

func InsertLeavingReasonsAndReturnID(ctx context.Context, fatimaDBTrace *database.DBTrace) (leavingReasonIDs []string, err error) {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := fmt.Sprintf("Cat " + randomStr)
		leavingReasonType := database.Text("1")
		remarks := fmt.Sprintf("Remark " + randomStr)
		isArchived := true
		stmt := `INSERT INTO leaving_reason
		(leaving_reason_id, name, leaving_reason_type,  remark, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())`

		_, err := fatimaDBTrace.Exec(ctx, stmt, randomStr, name, leavingReasonType, remarks, isArchived)
		if err != nil {
			return nil, fmt.Errorf("cannot insert leaving_reason, err: %s", err)
		}
		leavingReasonIDs = append(leavingReasonIDs, randomStr)
	}
	return
}

func InsertReferenceData(ctx context.Context, fatimaDBTrace *database.DBTrace, options OptionToPrepareDataForCreateOrderRecurringProduct, data DataForRecurringProduct, gradeID string) (
	err error,
) {
	if options.InsertProductGrade {
		err = InsertProductGrade(ctx, fatimaDBTrace, gradeID, data.ProductIDs)
		if err != nil {
			return
		}
	}

	if options.InsertProductPrice {
		err = InsertProductPriceForRecurringProducts(ctx, fatimaDBTrace, data.ProductIDs, data.BillingSchedulePeriods, PriceOrder, pb.ProductPriceType_DEFAULT_PRICE.String())
		if err != nil {
			return
		}
	}
	if options.InsertEnrolledProductPrice {
		err = InsertProductPriceForRecurringProducts(ctx, fatimaDBTrace, data.ProductIDs, data.BillingSchedulePeriods, EnrolledProductPrice, pb.ProductPriceType_ENROLLED_PRICE.String())
		if err != nil {
			return
		}
	}

	if options.InsertProductPriceWithDifferentPrice {
		err = InsertDifferentProductPriceForRecurringProducts(ctx, fatimaDBTrace, data.ProductIDs, data.BillingSchedulePeriods, PriceOrder)
		if err != nil {
			return
		}
	}

	if options.InsertProductLocation {
		err = InsertProductLocation(ctx, fatimaDBTrace, data.LocationID, data.ProductIDs)
		if err != nil {
			return
		}
	}

	if options.InsertProductDiscount {
		err = InsertProductDiscount(ctx, fatimaDBTrace, data.ProductIDs, data.DiscountIDs)
		if err != nil {
			return
		}
	}
	if options.InsertNotificationDate {
		_, err = InsertSomeNotificationDates(ctx, fatimaDBTrace)
		if err != nil {
			return
		}
	}
	if options.InsertProductSetting {
		for _, productID := range data.ProductIDs {
			productSetting := entities.ProductSetting{
				ProductID: pgtype.Text{
					String: productID,
					Status: pgtype.Present,
				},
				IsEnrollmentRequired: pgtype.Bool{
					Bool:   false,
					Status: pgtype.Present,
				},
				IsPausable: pgtype.Bool{
					Bool:   true,
					Status: pgtype.Present,
				},
				IsAddedToEnrollmentByDefault: pgtype.Bool{
					Bool:   false,
					Status: pgtype.Present,
				},
				IsOperationFee: pgtype.Bool{
					Bool:   false,
					Status: pgtype.Present,
				},
			}
			err = InsertProductSetting(ctx, fatimaDBTrace, productSetting)
			if err != nil {
				err = fmt.Errorf("error when insert list product setting %v", err)
				return
			}
		}
	}
	return
}

func InsertPreconditionData(ctx context.Context, fatimaDBTrace *database.DBTrace, options OptionToPrepareDataForCreateOrderRecurringProduct, data DataForRecurringProduct, gradeID string, name string) (
	taxID string,
	discountIDs []string,
	locationID string,
	userID string,
	enrolledUserID string,
	potentialUserID string,
	err error,
) {
	if options.InsertTax {
		if options.IsTaxExclusive {
			taxID, err = InsertOneTaxExclusive(ctx, fatimaDBTrace, name)
			if err != nil {
				return
			}
		} else {
			taxID, err = InsertOneTax(ctx, fatimaDBTrace, name)
			if err != nil {
				return
			}
		}
	}

	if options.InsertDiscount {
		discountIDs, err = InsertDiscountForRecurringProduct(ctx, fatimaDBTrace)
		if err != nil {
			return
		}
	}

	if options.InsertDiscountNotAvailable {
		discountIDs, err = InsertOneDiscountAmountNotAvailable(ctx, fatimaDBTrace, name)
		if err != nil {
			return
		}
	}

	if options.InsertLocation {
		locationID, err = InsertOneLocation(ctx, fatimaDBTrace)
		if err != nil {
			return
		}
	} else {
		locationID = constants.ManabieOrgLocation
	}

	if options.InsertStudent {
		userID, err = InsertOneUser(ctx, fatimaDBTrace, gradeID)
		if err != nil {
			return
		}
	}

	if options.InsertEnrolledStudent {
		enrolledUserID, err = InsertOneEnrolledUser(ctx, fatimaDBTrace, gradeID, locationID)
		if err != nil {
			return
		}
	}

	if options.InsertPotentialStudent {
		potentialUserID, err = InsertOnePotentialUser(ctx, fatimaDBTrace, gradeID)
		if err != nil {
			return
		}
	}
	return
}

func InsertOneUserAccessLocation(ctx context.Context, fatimaDBTrace *database.DBTrace, userID string, locationID string) error {
	stmt := `INSERT INTO user_access_paths
		(user_id, location_id, updated_at, created_at)
		VALUES ($1, $2, now(), now());`
	_, err := fatimaDBTrace.Exec(ctx, stmt, userID, locationID)

	return err
}

func InsertOnCourseAccessLocation(ctx context.Context, fatimaDBTrace *database.DBTrace, courseID string, locationID string) error {
	stmt := `INSERT INTO course_access_paths
		(course_id, location_id, updated_at, created_at)
		VALUES ($1, $2, now(), now());`
	_, err := fatimaDBTrace.Exec(ctx, stmt, courseID, locationID)

	return err
}

func InsertOneClass(ctx context.Context, fatimaDBTrace *database.DBTrace, courseID, locationID string) (string, error) {
	classID := idutil.ULIDNow()
	stmt := `INSERT INTO class
	(class_id, course_id, location_id, updated_at, created_at)
	VALUES ($1, $2, $3, now(), now())`
	_, err := fatimaDBTrace.Exec(ctx, stmt, classID, courseID, locationID)
	if err != nil {
		return "", err
	}

	return classID, nil
}

func UpdateStudentStatus(ctx context.Context, fatimaDBTrace *database.DBTrace, userID string, locationID string, status string, startDate time.Time, endDate time.Time) error {
	stmt := `INSERT INTO student_enrollment_status_history
	(student_id, location_id, enrollment_status, start_date, end_date, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err := fatimaDBTrace.Exec(ctx, stmt, userID, locationID, status, startDate, endDate)
	if err != nil {
		return err
	}

	return nil
}

func InsertStudentWithEnrollmentStatus(ctx context.Context, fatimaDBTrace *database.DBTrace, enrollmentStatus string, startDate time.Time) (data DataForRecurringProduct, err error) {
	data.UserID, err = InsertOneUser(ctx, fatimaDBTrace, "")
	if err != nil {
		return
	}

	data.LocationID = constants.ManabieOrgLocation

	switch enrollmentStatus {
	case "STUDENT_ENROLLMENT_STATUS_POTENTIAL":
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_POTENTIAL", startDate, startDate.AddDate(1, 0, 0))
	case "STUDENT_ENROLLMENT_STATUS_ENROLLED":
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_ENROLLED", startDate, startDate.AddDate(1, 0, 0))
	case "STUDENT_ENROLLMENT_STATUS_WITHDRAWN":
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_ENROLLED", startDate.AddDate(0, -1, 0), startDate.AddDate(0, 0, -1))
		if err != nil {
			return
		}
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_WITHDRAWN", startDate, startDate.AddDate(1, 0, 0))
	case "STUDENT_ENROLLMENT_STATUS_GRADUATE":
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_ENROLLED", startDate.AddDate(0, -1, 0), startDate.AddDate(0, 0, -1))
		if err != nil {
			return
		}
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_GRADUATE", startDate, startDate.AddDate(1, 0, 0))
	case "STUDENT_ENROLLMENT_STATUS_LOA":
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_ENROLLED", startDate.AddDate(0, -1, 0), startDate.AddDate(0, 0, -1))
		if err != nil {
			return
		}
		err = UpdateStudentStatus(ctx, fatimaDBTrace, data.UserID, data.LocationID, "STUDENT_ENROLLMENT_STATUS_LOA", startDate, startDate.AddDate(1, 0, 0))
	}

	return
}

func InsertProductSetting(ctx context.Context, fatimaDBTrace *database.DBTrace, productSetting entities.ProductSetting) error {
	deleteStmt := `DELETE FROM product_setting WHERE product_id = $1`
	_, err := fatimaDBTrace.Exec(ctx, deleteStmt, productSetting.ProductID)
	if err != nil {
		err = fmt.Errorf("cannot delete product_setting, err: %s", err)
		return err
	}

	stmt := `INSERT INTO product_setting
	(product_id, is_enrollment_required, is_pausable, is_added_to_enrollment_by_default, is_operation_fee, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err = fatimaDBTrace.Exec(ctx, stmt, productSetting.ProductID, productSetting.IsEnrollmentRequired, productSetting.IsPausable, productSetting.IsAddedToEnrollmentByDefault, productSetting.IsOperationFee)
	if err != nil {
		err = fmt.Errorf("cannot insert product_setting, err: %s", err)
		return err
	}

	return nil
}

func InsertNotificationDate(ctx context.Context, fatimaDBTrace *database.DBTrace, orderType string, notificationDate int) (id string, err error) {
	id = idutil.ULIDNow()
	stmt :=
		`
		INSERT INTO notification_date(
			notification_date_id,
			order_type,
			notification_date,
			is_archived,
			created_at,
			updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
        ON CONFLICT DO NOTHING
		`
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, orderType, notificationDate, false)
	if err != nil {
		return
	}
	return
}

func InsertNotificationDateAndReturnID(ctx context.Context, fatimaDBTrace *database.DBTrace, orderType string, notificationDate int) (id string, err error) {
	id = idutil.ULIDNow()
	stmt :=
		`
		INSERT INTO notification_date(
			notification_date_id,
			order_type,
			notification_date,
			is_archived,
			created_at,
			updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
        ON CONFLICT DO NOTHING
		`
	_, err = fatimaDBTrace.Exec(ctx, stmt, id, orderType, notificationDate, false)
	if err != nil {
		return
	}
	return
}

func InsertSomeNotificationDates(ctx context.Context, fatimaDBTrace *database.DBTrace) (ids []string, err error) {
	ids = make([]string, 0)
	orderTypes := []string{
		pb.OrderType_ORDER_TYPE_LOA.String(),
		pb.OrderType_ORDER_TYPE_RESUME.String(),
		pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
		pb.OrderType_ORDER_TYPE_GRADUATE.String(),
		pb.OrderType_ORDER_TYPE_NEW.String(),
		pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
	}
	notificationDate := 10
	for _, orderType := range orderTypes {
		var id string
		id, err = InsertNotificationDateAndReturnID(ctx, fatimaDBTrace, orderType, notificationDate)
		if err != nil {
			return
		}
		ids = append(ids, id)
	}
	return
}
