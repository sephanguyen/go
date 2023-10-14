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

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	PackageOneTimeComment = "test create order package one time"
	CourseNameFormat      = "CourseName-%v"
)

type optionToPrepareDataForCreateOrderPackageOneTime struct {
	insertProductGrade               bool
	insertStudent                    bool
	insertProductPrice               bool
	insertProductLocation            bool
	insertLocation                   bool
	insertPackageQuantityTypeMapping bool
	insertPackageCourse              bool
	insertCourse                     bool
	insertPackage                    bool
	insertCourseAccessLocation       bool
	insertUserAccessLocation         bool
	insertProductSetting             bool
}

func (s *suite) prepareDataForCreateOrderOneTimePackage(ctx context.Context) (context.Context, error) {
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
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
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
		insertProductSetting:             true,
	}
	taxID,
		locationID,
		packageIDs,
		courseIDs,
		userID,
		err = s.insertAllDataForInsertOrderPackageOneTime(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(packageIDs))
	billingItems := make([]*pb.BillingItem, 0, len(packageIDs))
	courseItems := []*pb.CourseItem{
		{
			CourseId:   courseIDs[0],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[0]),
			Weight:     &wrapperspb.Int32Value{Value: 1},
		},
		{
			CourseId:   courseIDs[1],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[1]),
			Weight:     &wrapperspb.Int32Value{Value: 2},
		},
		{
			CourseId:   courseIDs[2],
			CourseName: fmt.Sprintf(CourseNameFormat, courseIDs[2]),
			Weight:     &wrapperspb.Int32Value{Value: 3},
		},
	}
	startDate := timestamppb.New(time.Now().AddDate(0, 1, 0))
	orderItems = append(
		orderItems,
		&pb.OrderItem{ProductId: packageIDs[0], CourseItems: courseItems, StartDate: startDate},
	)
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
	})

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertAllDataForInsertOrderPackageOneTime(ctx context.Context, optionPrepareData optionToPrepareDataForCreateOrderPackageOneTime) (
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

	if optionPrepareData.insertUserAccessLocation {
		err = mockdata.InsertOneUserAccessLocation(ctx, s.FatimaDBTrace, userID, locationID)
		if err != nil {
			return
		}
	}

	if optionPrepareData.insertPackage {
		productIDs, err = s.insertPackage(ctx, taxID)
		if err != nil {
			return
		}
	}

	if optionPrepareData.insertPackageQuantityTypeMapping {
		err = mockdata.InsertPackageTypeQuantityTypeMapping(ctx, s.FatimaDBTrace)
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
		if err != nil {
			return
		}
	}

	if optionPrepareData.insertCourseAccessLocation {
		for _, id := range courseIDs {
			err = mockdata.InsertOnCourseAccessLocation(ctx, s.FatimaDBTrace, id, locationID)
			if err != nil {
				return
			}
		}
	}

	if optionPrepareData.insertPackageCourse {
		err = s.insertPackageCourses(ctx, productIDs, courseIDs)
	}

	if optionPrepareData.insertProductSetting {
		for _, productID := range productIDs {
			err = s.insertProductSetting(ctx, productID)
			if err != nil {
				err = fmt.Errorf("error when insert list product setting %v", err)
				return
			}
		}
	}
	return
}

func (s *suite) insertPackage(ctx context.Context, taxID string) ([]string, error) {
	packageIDs := make([]string, 0, 5)
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
		row := s.FatimaDBTrace.QueryRow(ctx, stmt,
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

		packageArg.PackageType = pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()
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

func (s *suite) insertCourses(ctx context.Context) ([]string, error) {
	courseIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		e := entities.Course{}
		courseID := idutil.ULIDNow()
		courseIDs = append(courseIDs, courseID)
		now := time.Now()
		if err := multierr.Combine(
			e.UpdatedAt.Set(now),
			e.CreatedAt.Set(now),
			e.CourseID.Set(courseID),
			e.TeachingMethod.Set(nil),
			e.Grade.Set(1),
			e.Name.Set(fmt.Sprintf(CourseNameFormat, courseID)),
		); err != nil {
			return nil, fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, &e, []string{"resource_path"}, s.FatimaDBTrace.Exec)
		if err != nil {
			return nil, fmt.Errorf("err insert course: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return nil, fmt.Errorf("err insert course: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return courseIDs, nil
}

func (s *suite) insertPackageCourses(ctx context.Context, packageIDs []string, courseIDs []string) error {
	for _, packageID := range packageIDs {
		for index, courseID := range courseIDs {
			packageCourse := entities.PackageCourse{}
			err := multierr.Combine(
				packageCourse.PackageID.Set(packageID),
				packageCourse.CourseID.Set(courseID),
				packageCourse.CourseWeight.Set(index+1),
				packageCourse.CreatedAt.Set(time.Now()),
				packageCourse.MaxSlotsPerCourse.Set(1),
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

func (s *suite) insertPackageTypeQuantityTypeMapping(ctx context.Context) error {
	packageTypeAndQuantityTypeMapping := []*entities.PackageQuantityTypeMapping{
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_ONE_TIME.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String(), Status: pgtype.Present},
			CreatedAt:    pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
		},
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_SLOT_BASED.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_SLOT.String(), Status: pgtype.Present},
			CreatedAt:    pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
		},
		{
			PackageType:  pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), Status: pgtype.Present},
			QuantityType: pgtype.Text{String: pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String(), Status: pgtype.Present},
			CreatedAt:    pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
		},
	}
	for _, item := range packageTypeAndQuantityTypeMapping {
		cmdTag, err := database.InsertExcept(ctx, item, []string{"resource_path"}, s.FatimaDBTrace.Exec)
		if err != nil {
			return fmt.Errorf("err insert packageTypeAndQuantityTypeMapping: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert packageTypeAndQuantityTypeMapping: %d RowsAffected", cmdTag.RowsAffected())
		}
	}
	return nil
}

func (s *suite) insertProductPriceForPackage(ctx context.Context, packageIDs []string) error {
	insertProductPriceStmt := `INSERT INTO product_price (product_id, price,quantity, created_at) VALUES ($1, $2, $3, now())
		`
	for _, packageID := range packageIDs {
		_, err := s.FatimaDBTrace.Exec(ctx, insertProductPriceStmt,
			packageID,
			PriceOrder,
			6,
		)
		if err != nil {
			return fmt.Errorf("cannot insert product_price, err: %s", err)
		}
	}
	return nil
}

func (s *suite) createOrderOneTimePackageSuccess(ctx context.Context) (context.Context, error) {
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

func (s *suite) subscribeStudentCourseEventSync(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	//stepState.FoundChanForJetStream = make(chan interface{}, 1)
	//studentCourseEventSyncOptions := nats.Option{
	//	JetStreamOptions: []nats.JSSubOption{
	//		nats.ManualAck(), nats.AckWait(30 * time.Second),
	//		nats.MaxDeliver(10),
	//		nats.Bind(constants.StreamStudentCourseEventSync, constants.DurableStudentCourseEventSync),
	//		nats.DeliverSubject(constants.DeliverStudentCourseEventSync),
	//	},
	//}
	//
	//handlerStudentCourseEventSync := func(ctx context.Context, data []byte) (bool, error) {
	//	var syncStudentCourses []*pb.EventSyncStudentPackageCourse
	//	err := json.Unmarshal(data, &syncStudentCourses)
	//	if err != nil {
	//		return false, err
	//	}
	//	stepState.FoundChanForJetStream <- syncStudentCourses
	//	return false, nil
	//}
	//
	//sub, err := s.JSM.QueueSubscribe(constants.SubjectStudentCourseEventSync, constants.QueueStudentCourseEventSync, studentCourseEventSyncOptions, handlerStudentCourseEventSync)
	//if err != nil {
	//	return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe StudentCourseEventSync: %w", err)
	//}
	//stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eventPublishedSignalStudentCourseEventSync(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()

	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			studentCourseSync := data.([]*pb.EventSyncStudentPackageCourse)
			if len(studentCourseSync) > 0 {
				return StepStateToContext(ctx, stepState), nil
			}
		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}
