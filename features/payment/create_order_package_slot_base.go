package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForCreateOrderSlotBasePackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID      string
		locationID string
		userID     string
		packageIDs []string
		courseIDs  []string
		req        pb.CreateOrderRequest
		err        error
	)
	defaultOptionPrepareData := optionToPrepareDataForCreateOrderPackageOneTime{
		insertStudent:                    true,
		insertPackage:                    true,
		insertPackageCourse:              true,
		insertCourse:                     true,
		insertProductPrice:               true,
		insertProductLocation:            true,
		insertLocation:                   false,
		insertProductGrade:               true,
		insertPackageQuantityTypeMapping: true,
	}
	taxID,
		locationID,
		packageIDs,
		courseIDs,
		userID,
		err = s.insertAllDataForInsertOrderPackageSlotBase(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order package based slot"

	orderItems := make([]*pb.OrderItem, 0, len(packageIDs))
	billingItems := make([]*pb.BillingItem, 0, len(packageIDs))
	courseItems := []*pb.CourseItem{
		{
			CourseId:   courseIDs[0],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[0]),
			Slot:       &wrapperspb.Int32Value{Value: 1},
		},
		{
			CourseId:   courseIDs[1],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[1]),
			Slot:       &wrapperspb.Int32Value{Value: 2},
		},
		{
			CourseId:   courseIDs[2],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[2]),
			Slot:       &wrapperspb.Int32Value{Value: 3},
		},
	}
	startDate := timestamppb.New(time.Now().AddDate(1, 1, 0))
	orderItems = append(
		orderItems,
		&pb.OrderItem{ProductId: packageIDs[0], CourseItems: courseItems, StartDate: startDate},
		&pb.OrderItem{
			ProductId:   packageIDs[1],
			CourseItems: courseItems,
			StartDate:   startDate,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[2],
			CourseItems: courseItems,
			StartDate:   startDate,
		})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: packageIDs[0],
		Price:     PriceOrder,
		Quantity:  &wrapperspb.Int32Value{Value: 6},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		CourseItems: courseItems,
		FinalPrice:  PriceOrder,
	}, &pb.BillingItem{
		ProductId: packageIDs[1],
		Price:     PriceOrder,
		Quantity:  &wrapperspb.Int32Value{Value: 6},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice:  PriceOrder,
		CourseItems: courseItems,
	}, &pb.BillingItem{
		ProductId: packageIDs[2],
		Price:     PriceOrder,
		Quantity:  &wrapperspb.Int32Value{Value: 6},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice:  PriceOrder,
		CourseItems: courseItems,
	})

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertAllDataForInsertOrderPackageSlotBase(ctx context.Context, optionPrepareData optionToPrepareDataForCreateOrderPackageOneTime) (
	taxID string,
	locationID string,
	productIDs []string,
	courseIDs []string,
	userID string,
	err error,
) {
	gradeID, err := mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
	if err != nil {
		return
	}

	taxID, err = mockdata.InsertOneTax(ctx, s.FatimaDBTrace, "test-insert-package")
	if err != nil {
		return
	}

	if optionPrepareData.insertLocation {
		locationID, err = mockdata.InsertOneLocation(ctx, s.FatimaDBTrace)
		if err != nil {
			return
		}
	} else {
		locationID = constants.ManabieOrgLocation
	}

	if optionPrepareData.insertStudent {
		userID, err = mockdata.InsertOneUser(ctx, s.FatimaDBTrace, gradeID)
	}

	if optionPrepareData.insertPackage {
		productIDs, err = s.insertPackageBaseSlot(ctx, taxID)
	}

	if optionPrepareData.insertPackageQuantityTypeMapping {
		err = s.insertPackageTypeQuantityTypeMapping(ctx)
		if err != nil {
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"package_quantity_type_mapping_pk\"") {
				return
			}
		}
	}

	if optionPrepareData.insertProductLocation {
		err = mockdata.InsertProductLocation(ctx, s.FatimaDBTrace, locationID, productIDs)
	}

	if optionPrepareData.insertProductPrice {
		err = s.insertProductPriceForPackage(ctx, productIDs)
	}

	if optionPrepareData.insertProductGrade {
		err = mockdata.InsertProductGrade(ctx, s.FatimaDBTrace, gradeID, productIDs)
	}

	if optionPrepareData.insertCourse {
		courseIDs, err = s.insertCourses(ctx)
	}

	if optionPrepareData.insertPackageCourse {
		err = s.insertPackageCourseSlotBase(ctx, productIDs, courseIDs)
	}
	return
}

func (s *suite) insertPackageBaseSlot(ctx context.Context, taxID string) ([]string, error) {
	packageIDs := make([]string, 0, 5)
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

	type AddPackageParams struct {
		PackageID        string       `json:"package_id"`
		PackageType      string       `json:"package_type"`
		MaxSlot          int32        `json:"max_slot"`
		PackageStartDate sql.NullTime `json:"package_start_date"`
		PackageEndDate   sql.NullTime `json:"package_end_date"`
	}
	for i := 0; i < 5; i++ {
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
		row := s.FatimaDBTrace.QueryRow(ctx, stmt,
			arg.ProductID,
			arg.Name,
			arg.ProductType,
			arg.TaxID,
			arg.AvailableFrom,
			arg.AvailableUtil,
			arg.Remarks,
			arg.CustomBillingPeriod,
			arg.BillingScheduleID,
			arg.DisableProRatingFlag,
			arg.IsArchived)
		err := row.Scan(&packageArg.PackageID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		packageIDs = append(packageIDs, packageArg.PackageID)
		queryInsertPackage := `INSERT INTO package(
									package_id,
                                    package_type,
                                    max_slot,
                                    package_start_date,
                                    package_end_date)
                                VALUES ($1, $2, $3, $4, $5)`

		packageArg.PackageType = pb.PackageType_PACKAGE_TYPE_SLOT_BASED.String()
		packageArg.MaxSlot = 34
		packageArg.PackageStartDate = sql.NullTime{Time: time.Now().Truncate(24 * time.Hour), Valid: true}
		packageArg.PackageEndDate = sql.NullTime{Time: time.Now().AddDate(1, 0, 0).Truncate(24 * time.Hour), Valid: true}
		_, err = s.FatimaDBTrace.Exec(ctx, queryInsertPackage,
			packageArg.PackageID,
			packageArg.PackageType,
			packageArg.MaxSlot,
			packageArg.PackageStartDate,
			packageArg.PackageEndDate,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot insert package, err: %s", err)
		}
	}
	return packageIDs, nil
}

func (s *suite) insertPackageCourseSlotBase(ctx context.Context, packageIDs []string, courseIDs []string) error {
	for _, packageID := range packageIDs {
		for index, courseID := range courseIDs {
			packageCourse := entities.PackageCourse{}
			err := multierr.Combine(
				packageCourse.PackageID.Set(packageID),
				packageCourse.CourseID.Set(courseID),
				packageCourse.CourseWeight.Set(index+1),
				packageCourse.CreatedAt.Set(time.Now()),
				packageCourse.MaxSlotsPerCourse.Set(4),
				packageCourse.MandatoryFlag.Set(index == 0),
			)
			if err != nil {
				return err
			}
			cmdTag, err := database.InsertExcept(ctx, &packageCourse, []string{"resource_path"}, s.FatimaDBTrace.Exec)
			if err != nil {
				return fmt.Errorf("err insert package course: %w", err)
			}

			if cmdTag.RowsAffected() != 1 {
				return fmt.Errorf("err insert package course: %d RowsAffected", cmdTag.RowsAffected())
			}
		}
	}
	return nil
}

func (s *suite) createOrderSlotBasePackageSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	err := s.checkCreatedOrderDetailsAndActionLogs(ctx, pb.OrderType_ORDER_TYPE_NEW)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.validateCreatedOrderItemsAndBillItemsForOneTimeProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
