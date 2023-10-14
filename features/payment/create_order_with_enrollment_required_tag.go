package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareEnrollmentRequiredProductOrder(ctx context.Context, tag string, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		taxID       string
		locationID  string
		userID      string
		materialIDs []string
		req         pb.CreateOrderRequest
		err         error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:                  true,
		insertDiscount:             true,
		insertStudent:              true,
		insertMaterial:             true,
		insertProductPrice:         true,
		insertEnrolledProductPrice: true,
		insertProductLocation:      true,
		insertLocation:             false,
		insertProductGrade:         true,
		insertFee:                  false,
		insertProductDiscount:      true,
		insertMaterialUnique:       false,
		insertProductSetting:       false,
	}
	taxID,
		_,
		locationID,
		materialIDs,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch tag {
	case "true":
		productSetting := entities.ProductSetting{
			ProductID:                    pgtype.Text{String: materialIDs[0], Status: pgtype.Present},
			IsPausable:                   pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsEnrollmentRequired:         pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsOperationFee:               pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err := mockdata.InsertProductSetting(ctx, s.FatimaDBTrace, productSetting)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "false":
		productSetting := entities.ProductSetting{
			ProductID:                    pgtype.Text{String: materialIDs[0], Status: pgtype.Present},
			IsPausable:                   pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsEnrollmentRequired:         pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: false, Status: pgtype.Present},
			IsOperationFee:               pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err := mockdata.InsertProductSetting(ctx, s.FatimaDBTrace, productSetting)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	switch status {
	case "enrolled":
		err := mockdata.UpdateStudentStatus(ctx, s.FatimaDBTrace, userID, locationID, "STUDENT_ENROLLMENT_STATUS_ENROLLED", time.Now(), time.Now().AddDate(1, 0, 0))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "potential":
		err := mockdata.UpdateStudentStatus(ctx, s.FatimaDBTrace, userID, locationID, "STUDENT_ENROLLMENT_STATUS_POTENTIAL", time.Now(), time.Now().AddDate(1, 0, 0))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	orderItems := make([]*pb.OrderItem, 0, len(materialIDs))
	billingItems := make([]*pb.BillingItem, 0, len(materialIDs))

	orderItems = append(orderItems, &pb.OrderItem{ProductId: materialIDs[0]})
	billingItems = append(billingItems, &pb.BillingItem{
		ProductId: materialIDs[0],
		Price:     PriceOrder,
		Quantity:  &wrapperspb.Int32Value{Value: 1},
		TaxItem: &pb.TaxBillItem{
			TaxId:         taxID,
			TaxPercentage: 20,
			TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
			TaxAmount:     83.333336,
		},
		FinalPrice: PriceOrder,
	})

	req.StudentId = userID
	req.LocationId = locationID
	req.OrderComment = "test create order material one time"
	req.OrderItems = orderItems
	req.BillingItems = billingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) orderWithEnrollmentRequiredTagIsValidated(ctx context.Context, tag string, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if tag == "true" && status == "potential" {
		if stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "Internal") {
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("product with enrollment required tag allowed to order by student not enrolled in location")
	}

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
