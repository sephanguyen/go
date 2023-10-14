package lessonmgmt

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const PriceOrder = 500

func (s *Suite) CreateOneTaxInFatima(ctx context.Context, name string) (string, error) {
	var taxID string
	taxName := database.Text(fmt.Sprintf("Tax for create order %s", name))
	taxPercentage := database.Int4(20)
	taxCategory := database.Text("TAX_CATEGORY_INCLUSIVE")
	isArchived := database.Bool(false)

	query := `INSERT INTO tax
		(tax_id, name, tax_percentage, tax_category, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now()) RETURNING tax_id;`
	row := s.FatimaDB.QueryRow(ctx, query, idutil.ULIDNow(), taxName, taxPercentage, taxCategory, isArchived)

	err := row.Scan(&taxID)
	return taxID, err
}

func (s *Suite) CreatePackageInFatima(ctx context.Context, taxID string) ([]string, error) {
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
		var productArg AddProductParams
		var packageArg AddPackageParams

		// insert product
		productID := idutil.ULIDNow()
		productArg.ProductID = productID
		productArg.Name = fmt.Sprintf("package-%v", productID)
		productArg.ProductType = pb.ProductType_PRODUCT_TYPE_PACKAGE.String()
		productArg.AvailableFrom = time.Now()
		productArg.AvailableUtil = time.Now().AddDate(1, 0, 0)
		productArg.DisableProRatingFlag = false
		productArg.IsArchived = false
		productArg.TaxID.String = taxID
		productArg.TaxID.Valid = true

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
		row := s.FatimaDB.QueryRow(ctx, queryInsertProduct,
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
			productArg.IsArchived)
		err := row.Scan(&packageArg.PackageID)
		if err != nil {
			return nil, fmt.Errorf("cannot insert product, err: %s", err)
		}
		packageIDs = append(packageIDs, packageArg.PackageID)

		// insert package
		packageArg.PackageType = pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()
		packageArg.MaxSlot = 34
		packageArg.PackageStartDate = sql.NullTime{Time: time.Now(), Valid: true}
		packageArg.PackageEndDate = sql.NullTime{Time: time.Now().AddDate(1, 0, 0), Valid: true}

		queryInsertPackage := `INSERT INTO package(
									package_id,
                                    package_type,
                                    max_slot,
                                    package_start_date,
                                    package_end_date)
                                VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`

		_, err = s.FatimaDB.Exec(ctx, queryInsertPackage,
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

func (s *Suite) CreatePackageTypeQuantityTypeMappingInFatima(ctx context.Context) error {
	packageType := pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()
	quantityType := pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String()

	query := `INSERT INTO package_quantity_type_mapping (
			package_type,
			quantity_type,
			created_at)
		VALUES ($1, $2, now())
		ON CONFLICT DO NOTHING`

	_, err := s.FatimaDB.Exec(ctx, query, packageType, quantityType)
	if err != nil {
		return fmt.Errorf("err insert package_quantity_type_mapping: %w", err)
	}

	return nil
}

func (s *Suite) CreateProductLocationInFatima(ctx context.Context, locationID string, productIDs []string) error {
	query := `INSERT INTO product_location (location_id, product_id, created_at)
			  VALUES ($1, $2, now()) ON CONFLICT DO NOTHING`

	for _, productID := range productIDs {
		_, err := s.FatimaDB.Exec(ctx, query, locationID, productID)
		if err != nil {
			return fmt.Errorf("cannot insert product_location, err: %s", err)
		}
	}
	return nil
}

func (s *Suite) CreateProductPriceForPackageInFatima(ctx context.Context, packageIDs []string) error {
	query := `INSERT INTO product_price (product_id, price,quantity, created_at) 
			  VALUES ($1, $2, $3, now()) ON CONFLICT DO NOTHING`

	for _, packageID := range packageIDs {
		_, err := s.FatimaDB.Exec(ctx, query, packageID, PriceOrder, 3)
		if err != nil {
			return fmt.Errorf("cannot insert product_price, err: %s", err)
		}
	}
	return nil
}

func (s *Suite) CreateProductGradeInFatima(ctx context.Context, gradeID string, productIDs []string) error {
	query := `INSERT INTO product_grade (grade_id, product_id, created_at) 
			  VALUES ($1, $2, now())  ON CONFLICT DO NOTHING`

	for _, productID := range productIDs {
		_, err := s.FatimaDB.Exec(ctx, query, gradeID, productID)
		if err != nil {
			return fmt.Errorf("cannot insert product_grade, err: %s", err)
		}
	}
	return nil
}

func (s *Suite) CreatePackageCoursesInFatima(ctx context.Context, packageIDs []string, courseIDs []string) error {
	for _, packageID := range packageIDs {
		for index, courseID := range courseIDs {
			query := `INSERT INTO package_course (
				package_id,
				course_id,
				course_weight,
				max_slots_per_course,
				mandatory_flag,
				created_at
				)
			VALUES ($1, $2, $3, $4, $5, now())
			ON CONFLICT DO NOTHING`

			courseWeight := index + 1
			mandatoryFlag := (index == 0)

			_, err := s.FatimaDB.Exec(ctx, query, packageID, courseID, courseWeight, 1, mandatoryFlag)
			if err != nil {
				return fmt.Errorf("cannot insert package_course, err: %s", err)
			}
		}
	}
	return nil
}

func (s *Suite) AddGradeIDToStudents(ctx context.Context, gradeID string, studentIDs []string) error {
	query := `UPDATE students SET grade_id = $1
		WHERE student_id = ANY($2)`

	_, err := s.BobDB.Exec(ctx, query, gradeID, studentIDs)
	if err != nil {
		return fmt.Errorf("cannot update students to add grade id, err: %s", err)
	}

	return nil
}

func (s *Suite) CrateDataForOrderPackageOneTime(ctx context.Context) (taxID string, productIDs []string, err error) {
	stepState := StepStateFromContext(ctx)

	taxID, err = s.CreateOneTaxInFatima(ctx, "test-insert-package")
	if err != nil {
		return
	}

	productIDs, err = s.CreatePackageInFatima(ctx, taxID)
	if err != nil {
		return
	}

	err = multierr.Combine(
		s.CreatePackageTypeQuantityTypeMappingInFatima(ctx),
		s.CreateProductLocationInFatima(ctx, stepState.LocationIDs[0], productIDs),
		s.CreateProductPriceForPackageInFatima(ctx, productIDs),
		s.CreateProductGradeInFatima(ctx, stepState.GradeIDs[0], productIDs),
		s.CreatePackageCoursesInFatima(ctx, productIDs, stepState.CourseIDs),
		s.AddGradeIDToStudents(ctx, stepState.GradeIDs[0], stepState.StudentIds),
	)
	if err != nil {
		return
	}

	return
}

func (s *Suite) PrepareRequestForCreateOrderOneTimePackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// give a bit of time to sync location, student, and course data
	time.Sleep(5 * time.Second)

	var (
		taxID      string
		packageIDs []string
		req        pb.CreateOrderRequest
		err        error
	)
	taxID, packageIDs, err = s.CrateDataForOrderPackageOneTime(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("failed to create required data for order one time package: %s", err)
	}

	orderItems := make([]*pb.OrderItem, 0, len(packageIDs))
	billingItems := make([]*pb.BillingItem, 0, len(packageIDs))
	courseItems := make([]*pb.CourseItem, 0, len(stepState.CourseIDs))

	for index, courseID := range stepState.CourseIDs {
		courseItems = append(courseItems, &pb.CourseItem{
			CourseId:   courseID,
			CourseName: fmt.Sprintf("course-%s", courseID),
			Weight:     &wrapperspb.Int32Value{Value: int32(index + 1)},
		})
	}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:   packageIDs[0],
			CourseItems: courseItems,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[1],
			CourseItems: courseItems,
		},
		&pb.OrderItem{
			ProductId:   packageIDs[2],
			CourseItems: courseItems,
		},
	)

	billingItems = append(billingItems,
		&pb.BillingItem{
			ProductId: packageIDs[0],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			CourseItems: courseItems,
			FinalPrice:  PriceOrder,
		},
		&pb.BillingItem{
			ProductId: packageIDs[1],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		},
		&pb.BillingItem{
			ProductId: packageIDs[2],
			Price:     PriceOrder,
			Quantity:  &wrapperspb.Int32Value{Value: 3},
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			FinalPrice:  PriceOrder,
			CourseItems: courseItems,
		},
	)

	req.StudentId = stepState.StudentIds[0]
	req.LocationId = stepState.LocationIDs[0]
	req.OrderComment = "test create order one time for sync student course"
	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserCreatesAStudentCourseInOrderManagement(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, ok := stepState.Request.(*pb.CreateOrderRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *pb.CreateOrderRequest, got %T", req)
	}

	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1500)
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getStudentCourseCountInFatima(ctx context.Context) (int, error) {
	stepState := StepStateFromContext(ctx)

	count := 0
	query := `SELECT count(*) FROM student_course
		WHERE student_id = $1
		AND location_id = $2
		AND package_type = $3
		AND deleted_at is null`

	studentID := stepState.StudentIds[0]
	locationID := stepState.LocationIDs[0]
	packageType := pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()

	err := s.FatimaDB.QueryRow(ctx, query, studentID, locationID, packageType).Scan(&count)
	if err != nil {
		return count, fmt.Errorf("failed to fetch student_course count in fatima: %s", err)
	}

	return count, nil
}

func (s *Suite) getStudentCourseCountInBob(ctx context.Context) (int, error) {
	stepState := StepStateFromContext(ctx)

	count := 0
	query := `SELECT count(*) FROM student_course
		WHERE student_id = $1
		AND location_id = $2
		AND package_type = $3
		AND deleted_at is null`

	studentID := stepState.StudentIds[0]
	locationID := stepState.LocationIDs[0]
	packageType := pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()

	err := s.BobDBTrace.QueryRow(ctx, query, studentID, locationID, packageType).Scan(&count)
	if err != nil {
		return count, fmt.Errorf("failed to fetch student_course count in bob: %s", err)
	}

	return count, nil
}

func (s *Suite) StudentCourseDataInBobDBSyncSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// wait for sync process done
	time.Sleep(5 * time.Second)

	retryCount := 0
	var (
		fatimaCount int
		bobCount    int
		err         error
	)
	for fatimaCount != bobCount && retryCount < 5 {
		fatimaCount, err = s.getStudentCourseCountInFatima(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		bobCount, err = s.getStudentCourseCountInBob(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// wait for another time for sync
		time.Sleep(5 * time.Second)
		retryCount++
	}

	if fatimaCount != bobCount && retryCount >= 5 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("counts dosen't match, in fatima: %v, in bob: %v", fatimaCount, bobCount)
	}

	return StepStateToContext(ctx, stepState), nil
}
