package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	HappyCaseWithEmptyValue       = "happy case with empty value"
	HappyCaseWithValue            = "happy case with value"
	FailCaseWithEmptyValue        = "fail case with empty value"
	FailCaseWhenConvertValue      = "fail case when convert value"
	FailCaseWhenPassingWrongValue = "Fail by pass wrong value"
	ConvertMaterial               = "Convert material"
	ConvertFee                    = "Convert fee"
	ConvertPackage                = "Convert package"
)

var quantity = int32(1)
var discountName = "discount 123"
var periodName = "BillingPeriodName"
var scheduleName = "BillingScheduleName"
var numerator = int32(1)
var denominator = int32(1)
var materialType = pb.MaterialType_MATERIAL_TYPE_RECURRING.String()
var feeType = pb.FeeType_FEE_TYPE_RECURRING.String()
var packageType = pb.PackageType_PACKAGE_TYPE_SCHEDULED.String()
var quantityType = pb.QuantityType_QUANTITY_TYPE_SLOT.String()
var courseItem = []*entities.CourseItem{
	{
		CourseID:   "1",
		CourseName: "course 1",
		Weight:     &quantity,
		Slot:       &quantity,
	},
	{
		CourseID:   "1",
		CourseName: "course 1",
		Weight:     &quantity,
		Slot:       &quantity,
	},
}

var materialBillDescriptionEntity = entities.BillingItemDescription{
	ProductType:             pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
	MaterialType:            &materialType,
	DiscountName:            &discountName,
	BillingPeriodName:       &periodName,
	BillingScheduleName:     &scheduleName,
	BillingRatioDenominator: &denominator,
	BillingRatioNumerator:   &numerator,
}
var feeBillDescriptionEntity = entities.BillingItemDescription{
	ProductType:             pb.ProductType_PRODUCT_TYPE_FEE.String(),
	FeeType:                 &feeType,
	DiscountName:            &discountName,
	BillingPeriodName:       &periodName,
	BillingScheduleName:     &scheduleName,
	BillingRatioDenominator: &denominator,
	BillingRatioNumerator:   &numerator,
}
var packageBillDescriptionEntity = entities.BillingItemDescription{
	ProductType:             pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
	PackageType:             &packageType,
	QuantityType:            &quantityType,
	CourseItems:             courseItem,
	DiscountName:            &discountName,
	BillingPeriodName:       &periodName,
	BillingScheduleName:     &scheduleName,
	BillingRatioDenominator: &denominator,
	BillingRatioNumerator:   &numerator,
}

func MockSetter(_ interface{}) error {
	return nil
}

func TestStringToDate(t *testing.T) {
	t.Run(HappyCaseWithEmptyValue, func(t *testing.T) {
		err := StringToDate("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWithEmptyValue, func(t *testing.T) {
		err := StringToDate("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(MissingMandatoryData, "test"), err)
	})

	t.Run("happy case with value (RFC3339)", func(t *testing.T) {
		err := StringToDate("test", "2021-12-07T00:00:00-07:00", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run("happy case with value (ISO layout)", func(t *testing.T) {
		err := StringToDate("test", "2021-12-07", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run(FailCaseWhenConvertValue, func(t *testing.T) {
		err := StringToDate("test", "121", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, `error parsing test: parsing time "121" as "2006-01-02T15:04:05Z07:00": cannot parse "121" as "2006"`, err.Error())
	})
}

func TestStringToInt(t *testing.T) {
	t.Run(HappyCaseWithEmptyValue, func(t *testing.T) {
		err := StringToInt("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWithEmptyValue, func(t *testing.T) {
		err := StringToInt("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(MissingMandatoryData, "test"), err)
	})

	t.Run(HappyCaseWithValue, func(t *testing.T) {
		err := StringToInt("test", "2", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run(FailCaseWhenConvertValue, func(t *testing.T) {
		err := StringToInt("test", "asa", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, `error parsing test: strconv.Atoi: parsing "asa": invalid syntax`, err.Error())
	})
}

func TestStringToFloat(t *testing.T) {
	t.Run(HappyCaseWithEmptyValue, func(t *testing.T) {
		err := StringToFloat("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWithEmptyValue, func(t *testing.T) {
		err := StringToFloat("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(MissingMandatoryData, "test"), err)
	})

	t.Run(HappyCaseWithValue, func(t *testing.T) {
		err := StringToFloat("test", "2.0", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run(FailCaseWhenConvertValue, func(t *testing.T) {
		err := StringToFloat("test", "asa", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, `error parsing test: strconv.ParseFloat: parsing "asa": invalid syntax`, err.Error())
	})
}

func TestStringToBool(t *testing.T) {
	t.Run(HappyCaseWithEmptyValue, func(t *testing.T) {
		err := StringToBool("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWithEmptyValue, func(t *testing.T) {
		err := StringToBool("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(MissingMandatoryData, "test"), err)
	})

	t.Run(HappyCaseWithValue, func(t *testing.T) {
		err := StringToBool("test", "1", false, MockSetter)
		require.Nil(t, err)
	})

	t.Run(FailCaseWhenConvertValue, func(t *testing.T) {
		err := StringToBool("test", "asa", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, `error parsing test: strconv.ParseBool: parsing "asa": invalid syntax`, err.Error())
	})
}

func TestStringToFormatString(t *testing.T) {
	t.Run(HappyCaseWithEmptyValue, func(t *testing.T) {
		err := StringToFormatString("test", "", true, MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWithEmptyValue, func(t *testing.T) {
		err := StringToFormatString("test", "", false, MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, fmt.Errorf(MissingMandatoryData, "test"), err)
	})
}

func TestStringToProductType(t *testing.T) {
	t.Run("Happy case with type Package", func(t *testing.T) {
		err := StringToProductType("test", "1", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type Material", func(t *testing.T) {
		err := StringToProductType("test", "3", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type Fee", func(t *testing.T) {
		err := StringToProductType("test", "3", MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWhenPassingWrongValue, func(t *testing.T) {
		err := StringToProductType("test", "4", MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, "invalid test: 4", err.Error())
	})
}

func TestStringToPackageType(t *testing.T) {
	t.Run("Happy case with type ONE_TIME package", func(t *testing.T) {
		err := StringToPackageType("test", "1", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type SLOT_BASED", func(t *testing.T) {
		err := StringToPackageType("test", "2", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type FREQUENCY", func(t *testing.T) {
		err := StringToPackageType("test", "3", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type SCHEDULED", func(t *testing.T) {
		err := StringToPackageType("test", "4", MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWhenPassingWrongValue, func(t *testing.T) {
		err := StringToPackageType("test", "5", MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, "invalid test: 5", err.Error())
	})
}

func TestStringToFeeType(t *testing.T) {
	t.Run("Happy case with type ONE_TIME fee", func(t *testing.T) {
		err := StringToFeeType("test", "1", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type RECURRING", func(t *testing.T) {
		err := StringToFeeType("test", "2", MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWhenPassingWrongValue, func(t *testing.T) {
		err := StringToFeeType("test", "3", MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, "invalid test: 3", err.Error())
	})
}

func TestStringToMaterialType(t *testing.T) {
	t.Run("Happy case with type ONE_TIME material", func(t *testing.T) {
		err := StringToMaterialType("test", "1", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type RECURRING", func(t *testing.T) {
		err := StringToMaterialType("test", "2", MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWhenPassingWrongValue, func(t *testing.T) {
		err := StringToMaterialType("test", "3", MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, "invalid test: 3", err.Error())
	})
}

func TestStringToLeavingReasonType(t *testing.T) {
	t.Run("Happy case with type WITHDRAWAL", func(t *testing.T) {
		err := StringToLeavingReasonType("test", "1", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type GRADUATE", func(t *testing.T) {
		err := StringToLeavingReasonType("test", "2", MockSetter)
		require.Nil(t, err)
	})
	t.Run("Happy case with type LOA", func(t *testing.T) {
		err := StringToLeavingReasonType("test", "3", MockSetter)
		require.Nil(t, err)
	})
	t.Run(FailCaseWhenPassingWrongValue, func(t *testing.T) {
		err := StringToLeavingReasonType("test", "4", MockSetter)
		require.NotNil(t, err)
		assert.Equal(t, "invalid test: 4", err.Error())
	})
}

func TestGetKeyMapFromBillItem(t *testing.T) {
	t.Run("key just have product id", func(t *testing.T) {
		key := GetKeyMapFromBillItem(&pb.BillingItem{
			ProductId: "1",
		})
		require.Equal(t, "1", key)
	})
	t.Run("key have product id and associated package", func(t *testing.T) {
		key := GetKeyMapFromBillItem(&pb.BillingItem{
			ProductId:           "1",
			PackageAssociatedId: wrapperspb.String("2"),
		})
		require.Equal(t, "1_2", key)
	})
}

func TestGetKeyMapFromOrderItem(t *testing.T) {
	t.Run("key just have product id", func(t *testing.T) {
		key := GetKeyMapFromOrderItem(&pb.OrderItem{
			ProductId: "1",
		})
		require.Equal(t, "1", key)
	})
	t.Run("key have product id and associated package", func(t *testing.T) {
		key := GetKeyMapFromOrderItem(&pb.OrderItem{
			ProductId:           "1",
			PackageAssociatedId: wrapperspb.String("2"),
		})
		require.Equal(t, "1_2", key)
	})
}

func TestConvertBillItemEntityToBillItemDescription(t *testing.T) {
	t.Run(ConvertMaterial, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(materialBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		description, err := ConvertBillItemEntityToBillItemDescription(billItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertFee, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(feeBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		description, err := ConvertBillItemEntityToBillItemDescription(billItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertPackage, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(packageBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		description, err := ConvertBillItemEntityToBillItemDescription(billItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
}

func TestConvertEntityBillItemToProtoProductInfoAndLocationInfo(t *testing.T) {
	materialBillItem := entities.BillItem{}
	_ = materialBillItem.BillingItemDescription.Set(materialBillDescriptionEntity)
	_ = materialBillItem.LocationID.Set("1")
	_ = materialBillItem.LocationName.Set("1")

	feeBillItem := entities.BillItem{}
	_ = feeBillItem.BillingItemDescription.Set(feeBillDescriptionEntity)
	_ = feeBillItem.LocationID.Set("1")
	_ = feeBillItem.LocationName.Set("1")

	packageBillItem := entities.BillItem{}
	_ = packageBillItem.BillingItemDescription.Set(packageBillDescriptionEntity)
	_ = packageBillItem.LocationID.Set("1")
	_ = packageBillItem.LocationName.Set("1")

	billItems := []*entities.BillItem{&materialBillItem, &feeBillItem, &packageBillItem}
	productInfos, locationInfo, err := ConvertEntityBillItemToProtoProductInfoAndLocationInfo(billItems)
	assert.Nil(t, err)
	assert.NotNil(t, productInfos)
	assert.NotNil(t, locationInfo)
}

func TestConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(t *testing.T) {
	t.Run(ConvertMaterial, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(materialBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		orderType := pb.OrderType_ORDER_TYPE_WITHDRAWAL.String()
		orderItem := entities.OrderItem{}
		_ = orderItem.EffectiveDate.Set(time.Now())
		_ = orderItem.OrderItemID.Set("1")
		_ = orderItem.OrderID.Set("1")
		_ = orderItem.ProductID.Set("1")
		description, err := ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(billItem, studentProduct, orderType, orderItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertFee, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(feeBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		orderType := pb.OrderType_ORDER_TYPE_WITHDRAWAL.String()
		orderItem := entities.OrderItem{}
		_ = orderItem.EffectiveDate.Set(time.Now())
		_ = orderItem.OrderItemID.Set("1")
		_ = orderItem.OrderID.Set("1")
		_ = orderItem.ProductID.Set("1")
		description, err := ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(billItem, studentProduct, orderType, orderItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertPackage, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(packageBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		orderType := pb.OrderType_ORDER_TYPE_WITHDRAWAL.String()
		orderItem := entities.OrderItem{}
		_ = orderItem.EffectiveDate.Set(time.Now())
		_ = orderItem.OrderItemID.Set("1")
		_ = orderItem.OrderID.Set("1")
		_ = orderItem.ProductID.Set("1")
		description, err := ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(billItem, studentProduct, orderType, orderItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertPackage, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(packageBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		orderType := pb.OrderType_ORDER_TYPE_GRADUATE.String()
		orderItem := entities.OrderItem{}
		_ = orderItem.EffectiveDate.Set(time.Now())
		_ = orderItem.OrderItemID.Set("1")
		_ = orderItem.OrderID.Set("1")
		_ = orderItem.ProductID.Set("1")
		description, err := ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(billItem, studentProduct, orderType, orderItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertPackage, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(packageBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		orderType := pb.OrderType_ORDER_TYPE_LOA.String()
		orderItem := entities.OrderItem{}
		_ = orderItem.StartDate.Set(time.Now())
		_ = orderItem.EndDate.Set(time.Now().AddDate(0, 1, 0))
		_ = orderItem.OrderItemID.Set("1")
		_ = orderItem.OrderID.Set("1")
		_ = orderItem.ProductID.Set("1")
		description, err := ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(billItem, studentProduct, orderType, orderItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertPackage, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(packageBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		orderType := pb.OrderType_ORDER_TYPE_NEW.String()
		orderItem := entities.OrderItem{}
		_ = orderItem.StartDate.Set(time.Now())
		_ = orderItem.OrderItemID.Set("1")
		_ = orderItem.OrderID.Set("1")
		_ = orderItem.ProductID.Set("1")
		description, err := ConvertEntityBillItemAndStudentProductToOrderProductInOrderDetail(billItem, studentProduct, orderType, orderItem)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
}

func TestConvertEntityStudentProductAndBillItemToOrderProductInStudentBilling(t *testing.T) {
	t.Run(ConvertMaterial, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(materialBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		description, err := ConvertEntityStudentProductAndBillItemToOrderProductInStudentBilling(&billItem, &studentProduct)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertFee, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(feeBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		description, err := ConvertEntityStudentProductAndBillItemToOrderProductInStudentBilling(&billItem, &studentProduct)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
	t.Run(ConvertPackage, func(t *testing.T) {
		billItem := entities.BillItem{}
		_ = billItem.BillingItemDescription.Set(packageBillDescriptionEntity)
		_ = billItem.LocationID.Set("1")
		_ = billItem.LocationName.Set("1")
		studentProduct := entities.StudentProduct{}
		_ = studentProduct.StartDate.Set(time.Now())
		_ = studentProduct.UpdatedToStudentProductID.Set("1")
		_ = studentProduct.UpdatedFromStudentProductID.Set("1")
		_ = studentProduct.StudentProductLabel.Set("1")
		description, err := ConvertEntityStudentProductAndBillItemToOrderProductInStudentBilling(&billItem, &studentProduct)
		assert.Nil(t, err)
		assert.NotNil(t, description)
	})
}
