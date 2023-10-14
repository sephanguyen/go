package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/status"
)

const CustomBillingItemsName = "custom billing items"

func (s *suite) customBillingIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	err := s.checkCreatedOrderDetailsAndActionLogs(ctx, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*pb.CreateCustomBillingRequest)
	res := stepState.Response.(*pb.CreateCustomBillingResponse)

	orderItems, err := s.getOrderItemsCustom(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	foundOrderItems := countOrderItemCustom(orderItems, req.CustomBillingItems)
	if foundOrderItems < len(req.CustomBillingItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error missing create custom order items")
	}

	billItems, err := s.getBillItemsCustom(ctx, res.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	foundBillItem := countBillItemCustom(billItems, req.CustomBillingItems, req.LocationId)
	if foundBillItem < len(req.CustomBillingItems) {
		fmt.Println(foundBillItem, len(req.CustomBillingItems))
		return StepStateToContext(ctx, stepState), fmt.Errorf("error missing create custom billing items")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreatingCustomBilling(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		taxID      string
		locationID string
		userID     string
		req        pb.CreateCustomBillingRequest
		err        error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        false,
		insertStudent:         true,
		insertFee:             false,
		insertProductPrice:    false,
		insertProductLocation: false,
		insertLocation:        false,
		insertProductGrade:    false,
		insertMaterial:        false,
		insertProductDiscount: true,
	}
	taxID, _, locationID, _, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, CustomBillingItemsName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create custom billing"

	req.CustomBillingItems = []*pb.CustomBillingItem{
		{
			Name:  "Default custom billing item",
			Price: 500,
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
		},
	}
	req.OrderType = pb.OrderType_ORDER_TYPE_CUSTOM_BILLING
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreatingCustomBillingWithAccountCategory(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		taxID              string
		locationID         string
		userID             string
		accountCategoryIDs []string
		req                pb.CreateCustomBillingRequest
		err                error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        false,
		insertStudent:         true,
		insertFee:             false,
		insertProductPrice:    false,
		insertProductLocation: false,
		insertLocation:        false,
		insertProductGrade:    false,
		insertMaterial:        false,
		insertProductDiscount: true,
	}
	taxID, _, locationID, _, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, CustomBillingItemsName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	accountCategoryIDs, err = s.insertSomeAccountingCategoriesForCustomBillItem(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create custom billing"

	req.CustomBillingItems = []*pb.CustomBillingItem{
		{
			Name:  "Default custom billing item",
			Price: 500,
			TaxItem: &pb.TaxBillItem{
				TaxId:         taxID,
				TaxPercentage: 20,
				TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
				TaxAmount:     83.333336,
			},
			AccountCategoryIds: accountCategoryIDs,
		},
	}
	req.OrderType = pb.OrderType_ORDER_TYPE_CUSTOM_BILLING
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) submitCustomBillingRequest(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
		CreateCustomBilling(contextWithToken(ctx), stepState.Request.(*pb.CreateCustomBillingRequest))
	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
			CreateCustomBilling(contextWithToken(ctx), stepState.Request.(*pb.CreateCustomBillingRequest))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesErrorMessageForCreateCustomBillingWith(ctx context.Context, expectedErrMessage, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	req := stepState.Request.(*pb.CreateCustomBillingRequest)
	switch testcase {
	case "invalid tax category":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.CustomBillingItems[0].Name, pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String(), pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String())
	case "invalid tax percentage":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.CustomBillingItems[0].Name, req.CustomBillingItems[0].TaxItem.TaxPercentage, 20)
	case "invalid tax amount":
		expectedErrMessage = fmt.Sprintf(expectedErrMessage, req.CustomBillingItems[0].TaxItem.TaxAmount, 83.333336)
	default:
	}
	if !strings.Contains(stt.Message(), expectedErrMessage) {
		return ctx, fmt.Errorf("expecting %s, got %s error message ", expectedErrMessage, stt.Message())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestForCreateCustomBillingWith(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID      string
		locationID string
		userID     string
		err        error
	)

	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        false,
		insertStudent:         true,
		insertFee:             false,
		insertProductPrice:    false,
		insertProductLocation: false,
		insertLocation:        false,
		insertProductGrade:    false,
		insertMaterial:        false,
		insertProductDiscount: true,
	}
	taxID, _, locationID, _, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, CustomBillingItemsName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := pb.CreateCustomBillingRequest{}
	req.OrderType = pb.OrderType_ORDER_TYPE_CUSTOM_BILLING
	req.OrderComment = "test create custom billing"

	switch testcase {
	case "missing location":
		req.StudentId = userID
		req.LocationId = ""
		req.CustomBillingItems = []*pb.CustomBillingItem{
			{
				Name:  "Default custom billing item",
				Price: 500,
				TaxItem: &pb.TaxBillItem{
					TaxId:         taxID,
					TaxPercentage: 20,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     83.333336,
				},
			},
		}
	case "missing name":
		req.StudentId = userID
		req.LocationId = locationID
		req.CustomBillingItems = []*pb.CustomBillingItem{
			{
				Name:  "Default custom billing item",
				Price: 500,
				TaxItem: &pb.TaxBillItem{
					TaxId:         taxID,
					TaxPercentage: 20,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     83.333336,
				},
			},
			{
				Price: 500,
				TaxItem: &pb.TaxBillItem{
					TaxId:         taxID,
					TaxPercentage: 20,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     83.333336,
				},
			},
		}
	case "not-exist student":
		req.StudentId = "invalid user"
		req.LocationId = locationID
	case "not-exist location":
		defaultPrepareDataSettings.insertTax = false
		defaultPrepareDataSettings.insertLocation = false
		_, _, _, _, userID, err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, CustomBillingItemsName)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.StudentId = userID
		req.LocationId = "invalid location"
	case "invalid tax category":
		req.StudentId = userID
		req.LocationId = locationID
		req.CustomBillingItems = []*pb.CustomBillingItem{
			{
				Name:  "Default custom billing item",
				Price: 500,
				TaxItem: &pb.TaxBillItem{
					TaxId:         taxID,
					TaxPercentage: 20,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
					TaxAmount:     83.333336,
				},
			},
		}
	case "invalid tax percentage":
		req.StudentId = userID
		req.LocationId = locationID
		req.CustomBillingItems = []*pb.CustomBillingItem{
			{
				Name:  "Default custom billing item",
				Price: 500,
				TaxItem: &pb.TaxBillItem{
					TaxId:         taxID,
					TaxPercentage: 10,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     83.333336,
				},
			},
		}
	case "invalid tax amount":
		req.StudentId = userID
		req.LocationId = locationID
		req.CustomBillingItems = []*pb.CustomBillingItem{
			{
				Name:  "Default custom billing item",
				Price: 500,
				TaxItem: &pb.TaxBillItem{
					TaxId:         taxID,
					TaxPercentage: 20,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     80,
				},
			},
		}
	default:
		fmt.Println("None testcase")
	}

	stepState.Request = &req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getOrderItemsCustom(ctx context.Context, orderID string) ([]entities.OrderItem, error) {
	var orderItems []entities.OrderItem
	orderItem := &entities.OrderItem{}
	orderItemFieldNames, orderItemFieldValues := orderItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(orderItemFieldNames, ","),
		orderItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(orderItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		orderItems = append(orderItems, *orderItem)
	}
	return orderItems, nil
}

func countOrderItemCustom(dbOrderItems []entities.OrderItem, orderItems []*pb.CustomBillingItem) int {
	foundOrderItem := 0
	for _, item := range orderItems {
		for _, dbItem := range dbOrderItems {
			if item.Name == dbItem.ProductName.String {
				foundOrderItem++
			}
		}
	}
	return foundOrderItem
}

func (s *suite) getBillItemsCustom(ctx context.Context, orderID string) ([]entities.BillItem, error) {
	var billItems []entities.BillItem
	billItem := &entities.BillItem{}
	billItemFieldNames, billItemFieldValues := billItem.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			order_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(billItemFieldNames, ","),
		billItem.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(billItemFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		billItems = append(billItems, *billItem)
	}
	return billItems, nil
}

func countBillItemCustom(dbBillItems []entities.BillItem, billingItems []*pb.CustomBillingItem, locationID string) int {
	foundBillItem := 0
	for _, item := range billingItems {
		for _, dbItem := range dbBillItems {
			if item.Name == dbItem.ProductDescription.String &&
				dbItem.TaxCategory.String == item.TaxItem.TaxCategory.String() &&
				float32(dbItem.TaxPercentage.Int) == item.TaxItem.TaxPercentage &&
				dbItem.TaxID.String == item.TaxItem.TaxId &&
				IsEqualNumericAndFloat32(dbItem.TaxAmount, item.TaxItem.TaxAmount) &&
				dbItem.BillStatus.String == pb.BillingStatus_BILLING_STATUS_BILLED.String() &&
				dbItem.BillType.String == pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String() &&
				dbItem.BillFrom.Status == pgtype.Null &&
				dbItem.BillTo.Status == pgtype.Null &&
				IsEqualNumericAndFloat32(dbItem.FinalPrice, item.Price) &&
				dbItem.LocationID.String == locationID {
				foundBillItem++
			}
		}
	}
	return foundBillItem
}

func (s *suite) insertSomeAccountingCategoriesForCustomBillItem(ctx context.Context) (ids []string, err error) {
	for i := 0; i < 5; i++ {
		randomStr := idutil.ULIDNow()
		name := database.Text("Cat " + randomStr)
		remarks := database.Text("Remarks " + randomStr)
		isArchived := database.Bool(false)
		stmt := `INSERT INTO accounting_category
		(accounting_category_id, name, remarks, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())`
		_, err = s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, remarks, isArchived)
		if err != nil {
			return nil, fmt.Errorf("cannot insert accounting category, err: %s", err)
		}
		ids = append(ids, randomStr)
	}
	return
}
