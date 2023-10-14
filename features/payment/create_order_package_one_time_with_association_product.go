package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/payment/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) insertAllDataForInsertOrderPackageOneTimeWithAssociationProduct(ctx context.Context) (
	taxID string,
	locationID string,
	packageIDs []string,
	materialIDs []string,
	courseIDs []string,
	userID string,
	gradeID string,
	err error,
) {

	gradeID, err = mockdata.InsertOneGrade(ctx, s.FatimaDBTrace)
	if err != nil {
		return
	}

	taxID, err = mockdata.InsertOneTax(ctx, s.FatimaDBTrace, "test-insert-package")
	if err != nil {
		return
	}
	locationID = constants.ManabieOrgLocation
	userID, err = mockdata.InsertOneUser(ctx, s.FatimaDBTrace, gradeID)
	if err != nil {
		return
	}
	packageIDs, err = s.insertPackage(ctx, taxID)
	if err != nil {
		return
	}
	materialIDs, err = s.insertMaterial(ctx, taxID)
	if err != nil {
		return
	}
	_ = s.insertPackageTypeQuantityTypeMapping(ctx)
	productIDs := packageIDs
	productIDs = append(productIDs, materialIDs...)
	err = mockdata.InsertProductLocation(ctx, s.FatimaDBTrace, locationID, productIDs)
	if err != nil {
		return
	}
	err = s.insertProductPriceForPackage(ctx, packageIDs)
	if err != nil {
		return
	}
	err = s.insertProductPrice(ctx, materialIDs, pb.ProductPriceType_DEFAULT_PRICE.String())
	if err != nil {
		return
	}
	err = mockdata.InsertProductGrade(ctx, s.FatimaDBTrace, gradeID, productIDs)
	if err != nil {
		return
	}
	courseIDs, err = s.insertCourses(ctx)
	if err != nil {
		return
	}
	err = s.insertPackageCourses(ctx, packageIDs, courseIDs)
	if err != nil {
		return
	}
	return
}

func (s *suite) prepareDataForCreateOrderOneTimePackageWithAssociationProduct(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		locationID  string
		userID      string
		packageIDs  []string
		materialIDs []string
		courseIDs   []string
		req         pb.CreateOrderRequest
		err         error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	taxID,
		locationID,
		packageIDs,
		materialIDs,
		courseIDs,
		userID,
		_,
		err = s.insertAllDataForInsertOrderPackageOneTimeWithAssociationProduct(ctx)
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

	orderItems = append(
		orderItems,
		&pb.OrderItem{
			ProductId: packageIDs[0], CourseItems: courseItems,
			ProductAssociations: []*pb.ProductAssociation{
				{
					PackageId:   packageIDs[0],
					CourseId:    courseIDs[0],
					ProductId:   materialIDs[0],
					ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				},
			},
		},
		&pb.OrderItem{
			ProductId:           materialIDs[0],
			PackageAssociatedId: wrapperspb.String(packageIDs[0]),
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
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice:          PriceOrder,
		PackageAssociatedId: wrapperspb.String(packageIDs[0]),
	})

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreateOrderOneTimePackageWithAssociationRecurringProduct(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		locationID  string
		userID      string
		packageIDs  []string
		materialIDs []string
		courseIDs   []string
		gradeID     string
		req         pb.CreateOrderRequest
		err         error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	taxID,
		locationID,
		packageIDs,
		materialIDs,
		courseIDs,
		userID,
		gradeID,
		err = s.insertAllDataForInsertOrderPackageOneTimeWithAssociationProduct(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"

	orderItems := make([]*pb.OrderItem, 0, len(packageIDs))
	billingItems := make([]*pb.BillingItem, 0, len(packageIDs))
	upcomingBillingItems := make([]*pb.BillingItem, 0, len(packageIDs))
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

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            true,
		InsertProductLocation:         false,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertMaterial:                true,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		InsertProductDiscount:         true,
		BillingScheduleStartDate:      time.Now().AddDate(0, -2, 0),
	}

	data, err := mockdata.InsertDataForRecurringProduct(ctx, s.FatimaDBTrace, defaultOptionPrepareData, gradeID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = mockdata.InsertProductLocation(ctx, s.FatimaDBTrace, locationID, []string{data.ProductIDs[0]})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	orderItems = append(
		orderItems,
		&pb.OrderItem{
			ProductId: packageIDs[0], CourseItems: courseItems,
			ProductAssociations: []*pb.ProductAssociation{
				{
					PackageId:   packageIDs[0],
					CourseId:    courseIDs[0],
					ProductId:   materialIDs[0],
					ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				},
				{
					PackageId:   packageIDs[0],
					CourseId:    courseIDs[0],
					ProductId:   data.ProductIDs[0],
					ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				},
			},
		},
		&pb.OrderItem{
			ProductId:           data.ProductIDs[0],
			DiscountId:          &wrapperspb.StringValue{Value: data.DiscountIDs[1]},
			StartDate:           &timestamppb.Timestamp{Seconds: defaultOptionPrepareData.BillingScheduleStartDate.AddDate(0, 0, 1).Unix()},
			PackageAssociatedId: wrapperspb.String(packageIDs[0]),
		},
		&pb.OrderItem{
			ProductId:           materialIDs[0],
			PackageAssociatedId: wrapperspb.String(packageIDs[0]),
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
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice:          PriceOrder,
		PackageAssociatedId: wrapperspb.String(packageIDs[0]),
	}, &pb.BillingItem{
		ProductId:               data.ProductIDs[0],
		BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[0].BillingSchedulePeriodID.String},
		Price:                   PriceOrder,
		Quantity:                &wrapperspb.Int32Value{Value: 1},
		TaxItem: &pb.TaxBillItem{
			TaxId:         data.TaxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     getInclusivePercentTax(PriceOrder-10, 20),
		},
		FinalPrice: PriceOrder - 10,
		DiscountItem: &pb.DiscountBillItem{
			DiscountId:          data.DiscountIDs[1],
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
			DiscountAmountValue: 10,
			DiscountAmount:      10,
		},
		PackageAssociatedId: wrapperspb.String(packageIDs[0]),
	}, &pb.BillingItem{
		ProductId:               data.ProductIDs[0],
		BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[1].BillingSchedulePeriodID.String},
		Price:                   PriceOrder,
		Quantity:                &wrapperspb.Int32Value{Value: 1},
		TaxItem: &pb.TaxBillItem{
			TaxId:         data.TaxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     getInclusivePercentTax(PriceOrder-10, 20),
		},
		FinalPrice: PriceOrder - 10,
		DiscountItem: &pb.DiscountBillItem{
			DiscountId:          data.DiscountIDs[1],
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
			DiscountAmountValue: 10,
			DiscountAmount:      10,
		},
		PackageAssociatedId: wrapperspb.String(packageIDs[0]),
	}, &pb.BillingItem{
		ProductId:               data.ProductIDs[0],
		BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[2].BillingSchedulePeriodID.String},
		Price:                   PriceOrder,
		Quantity:                &wrapperspb.Int32Value{Value: 1},
		TaxItem: &pb.TaxBillItem{
			TaxId:         data.TaxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     getInclusivePercentTax(PriceOrder-10, 20),
		},
		FinalPrice: PriceOrder - 10,
		DiscountItem: &pb.DiscountBillItem{
			DiscountId:          data.DiscountIDs[1],
			DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
			DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
			DiscountAmountValue: 10,
			DiscountAmount:      10,
		},
		PackageAssociatedId: wrapperspb.String(packageIDs[0]),
	})
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               data.ProductIDs[0],
			BillingSchedulePeriodId: &wrapperspb.StringValue{Value: data.BillingSchedulePeriods[3].BillingSchedulePeriodID.String},
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         data.TaxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     getInclusivePercentTax(PriceOrder-10, 20),
			},
			FinalPrice: PriceOrder - 10,
			DiscountItem: &pb.DiscountBillItem{
				DiscountId:          data.DiscountIDs[1],
				DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
				DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT,
				DiscountAmountValue: 10,
				DiscountAmount:      10,
			},
			PackageAssociatedId: wrapperspb.String(packageIDs[0]),
		},
	)
	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreateOrderOneTimePackageWithAssociationProductWithDuplicatedProduct(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		locationID  string
		userID      string
		packageIDs  []string
		materialIDs []string
		courseIDs   []string

		req pb.CreateOrderRequest
		err error
	)
	taxID,
		locationID,
		packageIDs,
		materialIDs,
		courseIDs,
		userID,
		_,
		err = s.insertAllDataForInsertOrderPackageOneTimeWithAssociationProduct(ctx)
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

	orderItems = append(
		orderItems,
		&pb.OrderItem{
			ProductId:   packageIDs[0],
			CourseItems: courseItems,
			ProductAssociations: []*pb.ProductAssociation{
				{
					PackageId:   packageIDs[0],
					CourseId:    courseIDs[0],
					ProductId:   materialIDs[0],
					ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				},
			},
		},
		&pb.OrderItem{
			ProductId:           materialIDs[0],
			PackageAssociatedId: wrapperspb.String(packageIDs[0]),
		},
		&pb.OrderItem{ProductId: materialIDs[0]})
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
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice:          PriceOrder,
		PackageAssociatedId: wrapperspb.String(packageIDs[0]),
	}, &pb.BillingItem{
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice: PriceOrder,
	})

	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getOrderProductAssociatedOfPackageList(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	reqCreateOrders := stepState.Request.(*pb.CreateOrderRequest)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentProducts, err := s.getListStudentProductBaseOnStudentID(ctx, reqCreateOrders.StudentId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
		StudentProductId: studentProducts.StudentProductID.String,
		Paging: &cpb.Paging{
			Limit: 2,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveListOfOrderAssociatedProductOfPackages(contextWithToken(ctx), &req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResponseOrderProductAssociated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateOrderRequest)
	resp := stepState.Response.(*pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse)

	expectedResponse := &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse{
		Items: []*pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct{},
		NextPage: &cpb.Paging{
			Limit: 2,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		PreviousPage: &cpb.Paging{},
		TotalItems:   2,
	}

	for _, billItems := range req.BillingItems {
		itemOfRequest := &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct{
			LocationInfo: &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct_LocationInfo{
				LocationId: req.LocationId,
			},
			ProductId: billItems.ProductId,
		}
		expectedResponse.Items = append(expectedResponse.Items, itemOfRequest)
	}

	for _, item := range resp.Items {
		if item.LocationInfo == nil {
			return ctx, fmt.Errorf("wrong location info")
		}

		if item.Duration == nil {
			return ctx, fmt.Errorf("wrong duration info")
		}

		if item.Status != pb.StudentProductStatus_ORDERED {
			return ctx, fmt.Errorf("wrong BillingStatus: %v", item.Status)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getListStudentProductBaseOnStudentID(ctx context.Context, studentID string) (studentProduct entities.StudentProduct, err error) {
	studentProductFieldNames, studentProductFieldValues := (&studentProduct).FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1 AND
			is_associated = false
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentProductFieldNames, ","),
		studentProduct.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, studentID)
	err = row.Scan(studentProductFieldValues...)
	if err != nil {
		return entities.StudentProduct{}, err
	}
	return studentProduct, nil
}
