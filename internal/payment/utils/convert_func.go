package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	MissingMandatoryData = "missing mandatory data: %v"
	ErrorParsing         = "error parsing %v: %w"
	InvalidValue         = "invalid %v: %s"
)

type SetFunc func(interface{}) error

func StringToDate(title, value string, nullable bool, setter SetFunc) error {
	var (
		timeElement time.Time
		err         error
	)
	trimmedValue := strings.TrimSpace(value)

	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}

	if len(trimmedValue) == len(LayoutISO) {
		timeElement, err = time.Parse(LayoutISO, trimmedValue)
	} else {
		timeElement, err = time.Parse(time.RFC3339, trimmedValue)
	}

	if err != nil {
		return fmt.Errorf(ErrorParsing, title, err)
	}
	return setter(timeElement)
}

func StringToFloat(title, value string, nullable bool, setter SetFunc) error {
	if value == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	floatElement, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf(ErrorParsing, title, err)
	}
	return setter(floatElement)
}

func StringToInt(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	intElement, err := strconv.Atoi(trimmedValue)
	if err != nil {
		return fmt.Errorf(ErrorParsing, title, err)
	}
	return setter(intElement)
}

func StringToBool(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	boolValue, err := strconv.ParseBool(trimmedValue)
	if err != nil {
		return fmt.Errorf(ErrorParsing, title, err)
	}
	return setter(boolValue)
}

func StringToFormatString(title, value string, nullable bool, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		if nullable {
			return setter(nil)
		}
		return fmt.Errorf(MissingMandatoryData, title)
	}
	return setter(trimmedValue)
}

func StringToProductType(title, value string, setter SetFunc) error {
	var productType pb.ProductType
	switch value {
	case "1":
		productType = pb.ProductType_PRODUCT_TYPE_PACKAGE
	case "2":
		productType = pb.ProductType_PRODUCT_TYPE_MATERIAL
	case "3":
		productType = pb.ProductType_PRODUCT_TYPE_FEE
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(productType.String())
}

func StringToPackageType(title, value string, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	var packageType pb.PackageType
	switch trimmedValue {
	case "1":
		packageType = pb.PackageType_PACKAGE_TYPE_ONE_TIME
	case "2":
		packageType = pb.PackageType_PACKAGE_TYPE_SLOT_BASED
	case "3":
		packageType = pb.PackageType_PACKAGE_TYPE_FREQUENCY
	case "4":
		packageType = pb.PackageType_PACKAGE_TYPE_SCHEDULED
	case "":
		return fmt.Errorf(MissingMandatoryData, title)
	default:
		return fmt.Errorf(InvalidValue, title, trimmedValue)
	}
	return setter(packageType.String())
}

func StringToMaterialType(title, value string, setter SetFunc) error {
	var materialType pb.MaterialType
	switch value {
	case "1":
		materialType = pb.MaterialType_MATERIAL_TYPE_ONE_TIME
	case "2":
		materialType = pb.MaterialType_MATERIAL_TYPE_RECURRING
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(materialType.String())
}

func StringToFeeType(title, value string, setter SetFunc) error {
	var feeType pb.FeeType
	switch value {
	case "1":
		feeType = pb.FeeType_FEE_TYPE_ONE_TIME
	case "2":
		feeType = pb.FeeType_FEE_TYPE_RECURRING
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(feeType.String())
}

func StringToLeavingReasonType(title, value string, setter SetFunc) error {
	var leavingReasonType pb.LeavingReasonType
	switch value {
	case "1":
		leavingReasonType = pb.LeavingReasonType_LEAVING_REASON_TYPE_WITHDRAWAL
	case "2":
		leavingReasonType = pb.LeavingReasonType_LEAVING_REASON_TYPE_GRADUATE
	case "3":
		leavingReasonType = pb.LeavingReasonType_LEAVING_REASON_TYPE_LOA
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(leavingReasonType.String())
}

func StringToDiscountType(title, value string, setter SetFunc) error {
	var discountType pb.DiscountType
	switch value {
	case "1":
		discountType = pb.DiscountType_DISCOUNT_TYPE_REGULAR
	case "2":
		discountType = pb.DiscountType_DISCOUNT_TYPE_COMBO
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(discountType.String())
}

func StringToDiscountAmountType(title, value string, setter SetFunc) error {
	var discountAmountType pb.DiscountAmountType
	switch value {
	case "1":
		discountAmountType = pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE
	case "2":
		discountAmountType = pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_FIXED_AMOUNT
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(discountAmountType.String())
}

func StringToTaxCategory(title, value string, setter SetFunc) error {
	var taxCategory pb.TaxCategory
	switch value {
	case "1":
		taxCategory = pb.TaxCategory_TAX_CATEGORY_INCLUSIVE
	case "2":
		taxCategory = pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE
	default:
		return fmt.Errorf(InvalidValue, title, value)
	}
	return setter(taxCategory.String())
}

func StringToQuantityType(title, value string, setter SetFunc) error {
	trimmedValue := strings.TrimSpace(value)
	var quantityType pb.QuantityType
	switch trimmedValue {
	case "1":
		quantityType = pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT
	case "2":
		quantityType = pb.QuantityType_QUANTITY_TYPE_SLOT
	case "3":
		quantityType = pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK
	case "":
		return fmt.Errorf(MissingMandatoryData, title)
	default:
		return fmt.Errorf(InvalidValue, title, trimmedValue)
	}
	return setter(quantityType.String())
}

func GetKeyMapFromOrderItem(orderItem *pb.OrderItem) (key string) {
	if orderItem.PackageAssociatedId != nil {
		key = fmt.Sprintf("%s_%s", orderItem.ProductId, orderItem.PackageAssociatedId.Value)
	} else {
		key = orderItem.ProductId
	}
	return
}

func GetKeyMapFromOrderItemV2(orderItem *pb.OrderItem, orderType pb.OrderType) (key string, err error) {
	if orderType == pb.OrderType_ORDER_TYPE_WITHDRAWAL ||
		orderType == pb.OrderType_ORDER_TYPE_GRADUATE {
		if orderItem.StudentProductId == nil {
			err = status.Error(codes.FailedPrecondition, fmt.Sprintf("missing student product in withdraw order item with product id %v", orderItem.ProductId))
			return
		}
		key = fmt.Sprintf("%s_%s", orderItem.ProductId, orderItem.StudentProductId.Value)
		return
	}

	if orderItem.PackageAssociatedId != nil {
		key = fmt.Sprintf("%s_%s", orderItem.ProductId, orderItem.PackageAssociatedId.Value)
	} else {
		key = orderItem.ProductId
	}
	return
}

func GetKeyMapFromBillItem(billingItem *pb.BillingItem) (key string) {
	if billingItem.PackageAssociatedId != nil {
		key = fmt.Sprintf("%s_%s", billingItem.ProductId, billingItem.PackageAssociatedId.Value)
	} else {
		key = billingItem.ProductId
	}
	return
}

func GetKeyMapFromBillItemV2(billingItem *pb.BillingItem, orderType pb.OrderType) (key string, err error) {
	if orderType == pb.OrderType_ORDER_TYPE_WITHDRAWAL ||
		orderType == pb.OrderType_ORDER_TYPE_GRADUATE {
		if billingItem.StudentProductId == nil {
			err = status.Error(codes.FailedPrecondition, fmt.Sprintf("missing student product in withdraw bill item with product id %v", billingItem.ProductId))
			return
		}
		key = fmt.Sprintf("%s_%s", billingItem.ProductId, billingItem.StudentProductId.Value)
		return
	}

	if billingItem.PackageAssociatedId != nil {
		key = fmt.Sprintf("%s_%s", billingItem.ProductId, billingItem.PackageAssociatedId.Value)
	} else {
		key = billingItem.ProductId
	}
	return
}

func GetKeyFromStudentAssociatedProduct(association *pb.ProductAssociation) (associatedKey string, packageKey string) {
	associatedKey = fmt.Sprintf("%s_%s", association.ProductId, association.PackageId)
	packageKey = fmt.Sprintf("%s", association.PackageId)
	return
}

func PagingToFromAndLimit(paging *cpb.Paging) (from, limit int64, err error) {
	if paging == nil {
		err = status.Error(codes.Internal, "not found paging")
		return
	}
	from = int64(0)
	limit = int64(paging.Limit)
	switch u := paging.Offset.(type) {
	case *cpb.Paging_OffsetInteger:
		from = u.OffsetInteger
	case *cpb.Paging_OffsetCombined:
		from = u.OffsetCombined.OffsetInteger
	default:
	}
	return
}

func ConvertCommonPaging(totalItems int, fromIdx int64, limit int64) (prevPage *cpb.Paging, nextPage *cpb.Paging, err error) {
	if fromIdx > int64(totalItems) {
		err = status.Error(codes.Internal, "Error offset")
		return
	}

	prevOffset := fromIdx - limit
	if prevOffset >= 0 {
		prevPage = &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: prevOffset,
			},
		}
	}

	nextOffset := fromIdx + limit
	if uint32(nextOffset) < uint32(totalItems) {
		nextPage = &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: nextOffset,
			},
		}
	}

	return
}

func ConvertBillItemEntityToBillItemDescription(billItem entities.BillItem) (billItemDescription *pb.BillItemDescription, err error) {
	billItemDescription = &pb.BillItemDescription{}
	description, err := billItem.GetBillingItemDescription()
	if err != nil {
		return
	}
	billItemDescription.ProductName = description.ProductName
	if billItem.ProductID.Status != pgtype.Present {
		return
	}
	billItemDescription.ProductId = billItem.ProductID.String
	billItemDescription.ProductType = pb.ProductType(pb.ProductType_value[description.ProductType])
	switch description.ProductType {
	case pb.ProductType_PRODUCT_TYPE_MATERIAL.String():
		billItemDescription.MaterialType = pb.MaterialType(pb.MaterialType_value[*description.MaterialType])
	case pb.ProductType_PRODUCT_TYPE_FEE.String():
		billItemDescription.FeeType = pb.FeeType(pb.FeeType_value[*description.FeeType])
	case pb.ProductType_PRODUCT_TYPE_PACKAGE.String():
		billItemDescription.PackageType = pb.PackageType(pb.PackageType_value[*description.PackageType])
		billItemDescription.QuantityType = pb.QuantityType(pb.QuantityType_value[*description.QuantityType])
		billItemDescription.CourseItems = ConvertEntityCourseItemsToProtoCourseItems(description.CourseItems)
	}

	if description.BillingRatioDenominator != nil {
		billItemDescription.BillingRatioDenominator = wrapperspb.Int32(*description.BillingRatioDenominator)
	}

	if description.BillingRatioNumerator != nil {
		billItemDescription.BillingRatioNumerator = wrapperspb.Int32(*description.BillingRatioNumerator)
	}

	if description.BillingPeriodName != nil {
		billItemDescription.BillingPeriodName = wrapperspb.String(*description.BillingPeriodName)
	}
	return
}

func ConvertEntityCourseItemsToProtoCourseItems(entityCourseItems []*entities.CourseItem) (protoCourseItems []*pb.CourseItem) {
	protoCourseItems = make([]*pb.CourseItem, 0, len(entityCourseItems))
	for _, item := range entityCourseItems {
		protoCourseItem := &pb.CourseItem{
			CourseId:   item.CourseID,
			CourseName: item.CourseName,
		}

		if item.Weight != nil {
			protoCourseItem.Weight = wrapperspb.Int32(*item.Weight)
		}

		if item.Slot != nil {
			protoCourseItem.Slot = wrapperspb.Int32(*item.Slot)
		}
		protoCourseItems = append(protoCourseItems, protoCourseItem)
	}
	return
}

func ConvertEntityBillItemToProtoProductInfoAndLocationInfo(entityBillItems []*entities.BillItem) (
	productInfos []*pb.ProductInfo,
	locationInfo *pb.LocationInfo,
	err error,
) {
	productInfos = make([]*pb.ProductInfo, 0, len(entityBillItems))
	for _, item := range entityBillItems {
		var (
			billItemDescription *entities.BillingItemDescription
			productInfo         *pb.ProductInfo
		)
		productInfo = &pb.ProductInfo{}
		billItemDescription, err = item.GetBillingItemDescription()
		if err != nil {
			err = status.Errorf(codes.Internal, "getting bill item description have error : %s", err.Error())
			return
		}
		if locationInfo == nil {
			locationInfo = &pb.LocationInfo{
				LocationId:   item.LocationID.String,
				LocationName: item.LocationName.String,
			}
		}
		productInfo.ProductName = billItemDescription.ProductName
		if item.ProductID.Status != pgtype.Present {
			productInfos = append(productInfos, productInfo)
			continue
		}
		productInfo.ProductId = item.ProductID.String
		productInfo.ProductType = pb.ProductType(pb.ProductType_value[billItemDescription.ProductType])
		switch productInfo.ProductType {
		case pb.ProductType_PRODUCT_TYPE_MATERIAL:
			productInfo.MaterialType = pb.MaterialType(pb.MaterialType_value[*billItemDescription.MaterialType])
		case pb.ProductType_PRODUCT_TYPE_FEE:
			productInfo.FeeType = pb.FeeType(pb.FeeType_value[*billItemDescription.FeeType])
		case pb.ProductType_PRODUCT_TYPE_PACKAGE:
			productInfo.PackageType = pb.PackageType(pb.PackageType_value[*billItemDescription.PackageType])
			productInfo.QuantityType = pb.QuantityType(pb.QuantityType_value[*billItemDescription.QuantityType])
		}
		productInfos = append(productInfos, productInfo)
	}
	return
}

func ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(
	billItem entities.BillItem,
	studentProduct entities.StudentProduct,
	orderType string,
	orderItem entities.OrderItem,
) (orderProduct *pb.RetrieveListOfOrderDetailProductsResponse_OrderProduct, err error) {
	var (
		billItemDescription *entities.BillingItemDescription
		productTypeValue    pb.ProductType
	)

	billItemDescription, err = billItem.GetBillingItemDescription()
	if err != nil {
		err = status.Errorf(codes.Internal, "getting description from bill item have error: %s", err.Error())
		return
	}
	productTypeValue = pb.ProductType(pb.ProductType_value[billItemDescription.ProductType])
	orderProduct = &pb.RetrieveListOfOrderDetailProductsResponse_OrderProduct{
		ProductId:        billItemDescription.ProductID,
		ProductName:      billItemDescription.ProductName,
		ProductType:      productTypeValue,
		StudentProductId: billItem.StudentProductID.String,
	}

	if billItemDescription.DiscountName != nil {
		orderProduct.DiscountInfo = &pb.RetrieveListOfOrderDetailProductsResponse_OrderProduct_DiscountInfo{
			DiscountName: *billItemDescription.DiscountName,
			DiscountId:   billItem.DiscountID.String,
		}
	}

	if billItemDescription.BillingRatioNumerator != nil {
		orderProduct.BillingRatioNumerator = wrapperspb.Int32(*billItemDescription.BillingRatioNumerator)
	}

	if billItemDescription.BillingRatioDenominator != nil {
		orderProduct.BillingRatioDenominator = wrapperspb.Int32(*billItemDescription.BillingRatioDenominator)
	}

	if billItemDescription.BillingPeriodName != nil {
		orderProduct.BillingPeriodName = wrapperspb.String(*billItemDescription.BillingPeriodName)
	}

	if billItemDescription.BillingScheduleName != nil {
		orderProduct.BillingScheduleName = wrapperspb.String(*billItemDescription.BillingScheduleName)
	}
	switch orderType {
	case pb.OrderType_ORDER_TYPE_LOA.String():
		if orderItem.StartDate.Status == pgtype.Present {
			orderProduct.StartDate = timestamppb.New(orderItem.StartDate.Time)
		}
		if orderItem.EndDate.Status == pgtype.Present {
			orderProduct.EndDate = timestamppb.New(orderItem.EndDate.Time)
		}
		if studentProduct.ProductStatus.Status == pgtype.Present {
			orderProduct.ProductStatus = pb.StudentProductStatus(pb.StudentProductStatus_value[studentProduct.ProductStatus.String])
		}
		if orderItem.StartDate.Time.After(orderItem.CreatedAt.Time) {
			orderProduct.StudentProductLabel = pb.StudentProductLabel_PAUSE_SCHEDULED
		} else {
			orderProduct.StudentProductLabel = pb.StudentProductLabel_PAUSED
		}
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(), pb.OrderType_ORDER_TYPE_GRADUATE.String():
		if orderItem.EffectiveDate.Status == pgtype.Present {
			orderProduct.StartDate = timestamppb.New(orderItem.EffectiveDate.Time)
		}
		if studentProduct.ProductStatus.Status == pgtype.Present {
			orderProduct.ProductStatus = pb.StudentProductStatus(pb.StudentProductStatus_value[studentProduct.ProductStatus.String])
		}
		if studentProduct.StudentProductLabel.Status == pgtype.Present {
			orderProduct.StudentProductLabel = pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String])
		}
	case pb.OrderType_ORDER_TYPE_UPDATE.String():
		if studentProduct.ProductStatus.Status == pgtype.Present {
			orderProduct.ProductStatus = pb.StudentProductStatus(pb.StudentProductStatus_value[studentProduct.ProductStatus.String])
		}
		if studentProduct.StudentProductLabel.Status == pgtype.Present {
			orderProduct.StudentProductLabel = pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String])
		}
		if orderItem.EffectiveDate.Status == pgtype.Present {
			orderProduct.StartDate = timestamppb.New(studentProduct.StartDate.Time)
		}
	default:
		if studentProduct.StartDate.Status == pgtype.Present {
			orderProduct.StartDate = timestamppb.New(studentProduct.StartDate.Time)
		}
		if studentProduct.ProductStatus.Status == pgtype.Present && orderProduct.ProductStatus == pb.StudentProductStatus_CANCELLED {
			orderProduct.ProductStatus = pb.StudentProductStatus_ORDERED
		} else {
			orderProduct.ProductStatus = pb.StudentProductStatus(pb.StudentProductStatus_value[studentProduct.ProductStatus.String])
		}
		if studentProduct.StudentProductLabel.Status == pgtype.Present {
			orderProduct.StudentProductLabel = pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String])
		}
	}

	if studentProduct.UpdatedToStudentProductID.Status == pgtype.Present {
		orderProduct.UpdatedToStudentProductId = wrapperspb.String(studentProduct.UpdatedToStudentProductID.String)
	}

	if studentProduct.UpdatedFromStudentProductID.Status == pgtype.Present {
		orderProduct.UpdatedFromStudentProductId = wrapperspb.String(studentProduct.UpdatedFromStudentProductID.String)
	}

	switch productTypeValue {
	case pb.ProductType_PRODUCT_TYPE_MATERIAL:
		orderProduct.MaterialType = pb.MaterialType(pb.MaterialType_value[*billItemDescription.MaterialType])
	case pb.ProductType_PRODUCT_TYPE_FEE:
		orderProduct.FeeType = pb.FeeType(pb.FeeType_value[*billItemDescription.FeeType])
	case pb.ProductType_PRODUCT_TYPE_PACKAGE:
		orderProduct.PackageType = pb.PackageType(pb.PackageType_value[*billItemDescription.PackageType])
		orderProduct.QuantityType = pb.QuantityType(pb.QuantityType_value[*billItemDescription.QuantityType])
		orderProduct.CourseItems = ConvertEntityCourseItemsToProtoCourseItems(billItemDescription.CourseItems)
		if orderProduct.PackageType == pb.PackageType_PACKAGE_TYPE_ONE_TIME ||
			orderProduct.PackageType == pb.PackageType_PACKAGE_TYPE_SLOT_BASED {
			orderProduct.StartDate = nil
		}
	}
	return
}

func ConvertEntityStudentProductAndBillItemToOrderProductInStudentBilling(
	billItem *entities.BillItem,
	studentProduct *entities.StudentProduct,
) (
	orderProduct *pb.RetrieveListOfOrderProductsResponse_OrderProduct,
	err error,
) {
	var (
		billingItemDescription *entities.BillingItemDescription
		productTypeValue       pb.ProductType
	)
	billingItemDescription, err = billItem.GetBillingItemDescription()
	if err != nil {
		return nil, err
	}
	productTypeValue = pb.ProductType(pb.ProductType_value[billingItemDescription.ProductType])
	orderProduct = &pb.RetrieveListOfOrderProductsResponse_OrderProduct{
		LocationInfo: &pb.RetrieveListOfOrderProductsResponse_OrderProduct_LocationInfo{
			LocationName: billItem.LocationName.String,
			LocationId:   billItem.LocationID.String,
		},
		Duration: &pb.RetrieveListOfOrderProductsResponse_OrderProduct_Duration{
			From: timestamppb.New(studentProduct.StartDate.Time),
			To:   timestamppb.New(studentProduct.EndDate.Time),
		},
		ProductId:                   billingItemDescription.ProductID,
		ProductName:                 billingItemDescription.ProductName,
		ProductType:                 productTypeValue,
		Status:                      pb.StudentProductStatus(pb.StudentProductStatus_value[studentProduct.ProductStatus.String]),
		StudentProductId:            studentProduct.StudentProductID.String,
		StudentProductLabel:         pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String]),
		UpdatedFromStudentProductId: wrapperspb.String(studentProduct.UpdatedFromStudentProductID.String),
		UpdatedToStudentProductId:   wrapperspb.String(studentProduct.UpdatedToStudentProductID.String),
	}

	if billItem.DiscountID.Status == pgtype.Present {
		orderProduct.DiscountInfo = &pb.RetrieveListOfOrderProductsResponse_OrderProduct_DiscountInfo{
			DiscountName: *billingItemDescription.DiscountName,
			DiscountId:   billItem.DiscountID.String,
		}
	}
	if billingItemDescription.BillingRatioNumerator != nil {
		orderProduct.BillingRatioNumerator = wrapperspb.Int32(*billingItemDescription.BillingRatioNumerator)
	}

	if billingItemDescription.BillingRatioDenominator != nil {
		orderProduct.BillingRatioDenominator = wrapperspb.Int32(*billingItemDescription.BillingRatioDenominator)
	}

	if billingItemDescription.BillingPeriodName != nil {
		orderProduct.BillingPeriodName = wrapperspb.String(*billingItemDescription.BillingPeriodName)
	}

	if billingItemDescription.BillingScheduleName != nil {
		orderProduct.BillingScheduleName = wrapperspb.String(*billingItemDescription.BillingScheduleName)
	}

	if studentProduct.UpdatedToStudentProductID.Status == pgtype.Present {
		orderProduct.UpdatedToStudentProductId = wrapperspb.String(studentProduct.UpdatedToStudentProductID.String)
	}

	if studentProduct.UpdatedFromStudentProductID.Status == pgtype.Present {
		orderProduct.UpdatedFromStudentProductId = wrapperspb.String(studentProduct.UpdatedFromStudentProductID.String)
	}

	if studentProduct.StudentProductLabel.Status == pgtype.Present {
		orderProduct.StudentProductLabel = pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String])
	}

	switch productTypeValue {
	case pb.ProductType_PRODUCT_TYPE_MATERIAL:
		orderProduct.MaterialType = pb.MaterialType(pb.MaterialType_value[*billingItemDescription.MaterialType])
	case pb.ProductType_PRODUCT_TYPE_FEE:
		orderProduct.FeeType = pb.FeeType(pb.FeeType_value[*billingItemDescription.FeeType])
	case pb.ProductType_PRODUCT_TYPE_PACKAGE:
		orderProduct.PackageType = pb.PackageType(pb.PackageType_value[*billingItemDescription.PackageType])
		orderProduct.QuantityType = pb.QuantityType(pb.QuantityType_value[*billingItemDescription.QuantityType])
		orderProduct.CourseItems = ConvertEntityCourseItemsToProtoCourseItems(billingItemDescription.CourseItems)
	}
	return
}

func ConvertEntityStudentProductAndBillItemToOrderProductAssociatedInStudentBilling(
	billItem *entities.BillItem,
	studentProduct *entities.StudentProduct,
) (
	orderProduct *pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct,
	err error,
) {
	var (
		billingItemDescription *entities.BillingItemDescription
		productTypeValue       pb.ProductType
	)
	billingItemDescription, err = billItem.GetBillingItemDescription()
	if err != nil {
		return nil, err
	}
	productTypeValue = pb.ProductType(pb.ProductType_value[billingItemDescription.ProductType])
	orderProduct = &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct{
		LocationInfo: &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct_LocationInfo{
			LocationName: billItem.LocationName.String,
			LocationId:   billItem.LocationID.String,
		},
		Duration: &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct_Duration{
			From: timestamppb.New(studentProduct.StartDate.Time),
			To:   timestamppb.New(studentProduct.EndDate.Time),
		},
		ProductId:                   billingItemDescription.ProductID,
		ProductName:                 billingItemDescription.ProductName,
		ProductType:                 productTypeValue,
		Status:                      pb.StudentProductStatus(pb.StudentProductStatus_value[studentProduct.ProductStatus.String]),
		StudentProductId:            studentProduct.StudentProductID.String,
		StudentProductLabel:         pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String]),
		UpdatedFromStudentProductId: wrapperspb.String(studentProduct.UpdatedFromStudentProductID.String),
		UpdatedToStudentProductId:   wrapperspb.String(studentProduct.UpdatedToStudentProductID.String),
	}

	if billItem.DiscountID.Status == pgtype.Present {
		orderProduct.DiscountInfo = &pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse_OrderProduct_DiscountInfo{
			DiscountName: *billingItemDescription.DiscountName,
			DiscountId:   billItem.DiscountID.String,
		}
	}
	if billingItemDescription.BillingRatioNumerator != nil {
		orderProduct.BillingRatioNumerator = wrapperspb.Int32(*billingItemDescription.BillingRatioNumerator)
	}

	if billingItemDescription.BillingRatioDenominator != nil {
		orderProduct.BillingRatioDenominator = wrapperspb.Int32(*billingItemDescription.BillingRatioDenominator)
	}

	if billingItemDescription.BillingPeriodName != nil {
		orderProduct.BillingPeriodName = wrapperspb.String(*billingItemDescription.BillingPeriodName)
	}

	if billingItemDescription.BillingScheduleName != nil {
		orderProduct.BillingScheduleName = wrapperspb.String(*billingItemDescription.BillingScheduleName)
	}

	if studentProduct.UpdatedToStudentProductID.Status == pgtype.Present {
		orderProduct.UpdatedToStudentProductId = wrapperspb.String(studentProduct.UpdatedToStudentProductID.String)
	}

	if studentProduct.UpdatedFromStudentProductID.Status == pgtype.Present {
		orderProduct.UpdatedFromStudentProductId = wrapperspb.String(studentProduct.UpdatedFromStudentProductID.String)
	}

	if studentProduct.StudentProductLabel.Status == pgtype.Present {
		orderProduct.StudentProductLabel = pb.StudentProductLabel(pb.StudentProductLabel_value[studentProduct.StudentProductLabel.String])
	}

	switch productTypeValue {
	case pb.ProductType_PRODUCT_TYPE_MATERIAL:
		orderProduct.MaterialType = pb.MaterialType(pb.MaterialType_value[*billingItemDescription.MaterialType])
	case pb.ProductType_PRODUCT_TYPE_FEE:
		orderProduct.FeeType = pb.FeeType(pb.FeeType_value[*billingItemDescription.FeeType])
	case pb.ProductType_PRODUCT_TYPE_PACKAGE:
		err = status.Errorf(codes.Internal, "Error when convert entity have product type package")
	}
	return
}

func ConvertOrderItemType(orderType pb.OrderType, billItem *pb.BillingItem) (typeOfOrder OrderType) {
	switch orderType {
	case pb.OrderType_ORDER_TYPE_NEW:
		typeOfOrder = OrderCreate
	case pb.OrderType_ORDER_TYPE_UPDATE:
		if billItem.IsCancelBillItem != nil && billItem.IsCancelBillItem.Value {
			typeOfOrder = OrderCancel
		} else {
			typeOfOrder = OrderUpdate
		}
	case pb.OrderType_ORDER_TYPE_ENROLLMENT:
		typeOfOrder = OrderEnrollment
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL:
		typeOfOrder = OrderWithdraw
	case pb.OrderType_ORDER_TYPE_GRADUATE:
		typeOfOrder = OrderGraduate
	case pb.OrderType_ORDER_TYPE_LOA:
		typeOfOrder = OrderLOA
	case pb.OrderType_ORDER_TYPE_RESUME:
		typeOfOrder = OrderResume
	}
	return
}
