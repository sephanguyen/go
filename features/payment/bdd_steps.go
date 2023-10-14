package payment

import (
	"regexp"
	"sync"

	"github.com/manabie-com/backend/features/helper"

	"github.com/cucumber/godog"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		// Echo
		`^an echo message$`:                s.anEchoMessage,
		`^"([^"]*)" echo a message$`:       s.echoAMessage,
		`^the message is echoed$`:          s.theMessageIsEchoed,
		`^receives "([^"]*)" status code$`: s.receivesStatusCode,

		// Import Accounting Category
		`^an accounting category valid request payload with "([^"]*)"$`:                     s.anAccountingCategoryValidRequestPayloadWith,
		`^"([^"]*)" importing accounting category$`:                                         s.importingAccountingCategory,
		`^the invalid accounting category lines are returned with error$`:                   s.theInvalidAccountingCategoryLinesAreReturnedWithError,
		`^an accounting category valid request payload with incorrect data with "([^"]*)"$`: s.anAccountingCategoryValidRequestPayloadWithIncorrectData,
		`^the import accounting category transaction is rolled back$`:                       s.theImportAccountingCategoryTransactionIsRolledBack,
		`^the valid accounting category lines are imported successfully$`:                   s.theValidAccountingCategoryLinesAreImportedSuccessfully,
		`^an accounting category invalid "([^"]*)" request payload$`:                        s.anAccountingCategoryInvalidRequestPayload,

		// Import Tax
		`^a tax valid request payload with "([^"]*)"$`:                     s.aTaxValidRequestPayloadWith,
		`^"([^"]*)" importing tax$`:                                        s.importingTax,
		`^the valid tax lines are imported successfully$`:                  s.theValidTaxLinesAreImportedSuccessfully,
		`^a tax valid request payload with incorrect data with "([^"]*)"$`: s.aTaxValidRequestPayloadWithIncorrectData,
		`^the import tax transaction is rolled back$`:                      s.theImportTaxTransactionIsRolledBack,
		`^the invalid tax lines are returned with error$`:                  s.theInvalidTaxLinesAreReturnedWithError,
		`^a tax invalid "([^"]*)" request payload$`:                        s.aTaxInvalidRequestPayload,

		// Import Billing Schedule
		`^an billing schedule valid request payload with "([^"]*)"$`:                     s.anBillingScheduleValidRequestPayloadWith,
		`^"([^"]*)" importing billing schedule$`:                                         s.importingBillingSchedule,
		`^the invalid billing schedule lines are returned with error$`:                   s.theInvalidBillingScheduleLinesAreReturnedWithError,
		`^an billing schedule valid request payload with incorrect data with "([^"]*)"$`: s.anBillingScheduleValidRequestPayloadWithIncorrectData,
		`^the import billing schedule transaction is rolled back$`:                       s.theImportBillingScheduleTransactionIsRolledBack,
		`^the valid billing schedule lines are imported successfully$`:                   s.theValidBillingScheduleLinesAreImportedSuccessfully,
		`^an billing schedule invalid "([^"]*)" request payload$`:                        s.anBillingScheduleInvalidRequestPayload,

		// Import Billing Schedule Period
		`^an billing schedule period valid request payload with "([^"]*)"$`:                     s.anBillingSchedulePeriodValidRequestPayloadWith,
		`^"([^"]*)" importing billing schedule period$`:                                         s.importingBillingSchedulePeriod,
		`^the invalid billing schedule period lines are returned with error$`:                   s.theInvalidBillingSchedulePeriodLinesAreReturnedWithError,
		`^the valid billing schedule period lines are imported successfully$`:                   s.theValidBillingSchedulePeriodLinesAreImportedSuccessfully,
		`^an billing schedule period valid request payload with incorrect data with "([^"]*)"$`: s.anBillingSchedulePeriodValidRequestPayloadWithIncorrectData,
		`^the import billing schedule period transaction is rolled back$`:                       s.theImportBillingSchedulePeriodTransactionIsRolledBack,
		`^an billing schedule period invalid "([^"]*)" request payload$`:                        s.anBillingSchedulePeriodInvalidRequestPayload,

		// Import Discount
		`^an discount invalid "([^"]*)" request payload$`:                        s.anDiscountInvalidRequestPayload,
		`^an discount valid request payload with "([^"]*)"$`:                     s.anDiscountValidRequestPayloadWith,
		`^"([^"]*)" importing discount$`:                                         s.importingDiscount,
		`^an discount valid request payload with incorrect data with "([^"]*)"$`: s.anDiscountValidRequestPayloadWithIncorrectData,
		`^the import discount transaction is rolled back$`:                       s.theImportDiscountTransactionIsRolledBack,
		`^the invalid discount lines are returned with error$`:                   s.theInvalidDiscountLinesAreReturnedWithError,
		`^the valid discount lines are imported successfully$`:                   s.theValidDiscountLinesAreImportedSuccessfully,

		// Import Product Associated Data - Accounting Category
		`^a product accounting category valid request payload with correct data with "([^"]*)"$`:   s.aProductAccountingCategoryValidRequestPayloadWithCorrectDataWith,
		`^a product accounting category valid request payload with incorrect data with "([^"]*)"$`: s.aProductAccountingCategoryValidRequestPayloadWithIncorrectDataWith,
		`^a product accounting category invalid request payload with "([^"]*)"$`:                   s.aProductAccountingCategoryInvalidRequestPayload,
		`^"([^"]*)" importing product accounting category$`:                                        s.importingProductAccountingCategory,
		`^the valid product accounting category lines are imported successfully$`:                  s.theValidProductAccountingCategoryLinesAreImportedSuccessfully,
		`^the import product accounting category transaction is rolled back$`:                      s.theImportProductAccountingCategoryTransactionIsRolledBack,
		`^the invalid product accounting category lines are returned with error$`:                  s.theInvalidProductAccountingCategoryLinesAreReturnedWithError,

		// Import Product - Grade
		`^a product grade valid request payload with correct data with "([^"]*)"$`:   s.aProductGradeValidRequestPayloadWithCorrectDataWith,
		`^a product grade valid request payload with incorrect data with "([^"]*)"$`: s.aProductGradeValidRequestPayloadWithIncorrectDataWith,
		`^a product grade invalid request payload with "([^"]*)"$`:                   s.aProductGradeInvalidRequestPayload,
		`^"([^"]*)" importing product grade$`:                                        s.importingProductGrade,
		`^the valid product grade lines are imported successfully$`:                  s.theValidProductGradeLinesAreImportedSuccessfully,
		`^the import product grade transaction is rolled back$`:                      s.theImportProductGradeTransactionIsRolledBack,
		`^the invalid product grade lines are returned with error$`:                  s.theInvalidProductGradeLinesAreReturnedWithError,

		// Sync Grade
		`^an grade valid request payload$`:                s.anGradeValidRequestPayload,
		`^"([^"]*)" importing grade$`:                     s.importingGrade,
		`^payment save consistent record grade with bob$`: s.paymentMustSaveConsistentGradeWithBobData,

		// Sync Location
		`^prepare location data$`:         s.prepareLocationDataForInsert,
		`^insert location data from bob$`: s.insertLocationFromBobDB,
		`^payment must record location$`:  s.checkSyncLocationSuccess,

		// Import Package
		`^an package valid request payload with "([^"]*)"$`:                     s.anPackageValidRequestPayloadWith,
		`^"([^"]*)" importing package$`:                                         s.importingPackage,
		`^the valid package lines are imported successfully$`:                   s.theValidPackageLinesAreImportedSuccessfully,
		`^an package valid request payload with incorrect data with "([^"]*)"$`: s.anPackageValidRequestPayloadWithIncorrectData,
		`^the import package transaction is rolled back$`:                       s.theImportPackageTransactionIsRolledBack,
		`^the invalid package lines are returned with error$`:                   s.theInvalidPackageLinesAreReturnedWithError,
		`^an package invalid "([^"]*)" request payload$`:                        s.anPackageInvalidRequestPayload,

		// Import Product Price
		`^an product price valid request payload with correct data with "([^"]*)"$`:   s.anProductPriceValidRequestPayloadWithCorrectDataWith,
		`^an product price valid request payload with incorrect data with "([^"]*)"$`: s.anProductPriceValidRequestPayloadWithIncorrectDataWith,
		`^an product price invalid request payload with "([^"]*)"$`:                   s.anProductPriceInvalidRequestPayloadWith,
		`^"([^"]*)" importing product price$`:                                         s.importingProductPrice,
		`^the valid product price lines are imported successfully$`:                   s.theValidProductPriceLinesAreImportedSuccessfully,
		`^the import product price transaction is rolled back$`:                       s.theImportProductPriceTransactionIsRolledBack,
		`^the invalid product price lines are returned with error$`:                   s.theInvalidProductPriceLinesAreReturnedWithError,

		// Import Product Material
		`^an material valid request payload with "([^"]*)"`:                      s.anMaterialValidRequestPayloadWith,
		`^"([^"]*)" importing material`:                                          s.importingMaterial,
		`^the valid material lines are imported successfully$`:                   s.theValidMaterialLinesAreImportedSuccessfully,
		`^an material valid request payload with incorrect data with "([^"]*)"$`: s.anMaterialValidRequestPayloadWithIncorrectData,
		`^the import material transaction is rolled back$`:                       s.theImportMaterialTransactionIsRolledBack,
		`^the invalid material lines are returned with error$`:                   s.theInvalidMaterialLinesAreReturnedWithError,
		`^an material invalid "([^"]*)" request payload$`:                        s.anMaterialInvalidRequestPayload,

		// Import Product Fee
		`^an fee valid request payload with "([^"]*)"`:                      s.anFeeValidRequestPayloadWith,
		`^"([^"]*)" importing fee$`:                                         s.importingFee,
		`^the valid fee lines are imported successfully$`:                   s.theValidFeeLinesAreImportedSuccessfully,
		`^an fee valid request payload with incorrect data with "([^"]*)"$`: s.anFeeValidRequestPayloadWithIncorrectData,
		`^the import fee transaction is rolled back$`:                       s.theImportFeeTransactionIsRolledBack,
		`^the invalid fee lines are returned with error$`:                   s.theInvalidFeeLinesAreReturnedWithError,
		`^an fee invalid "([^"]*)" request payload$`:                        s.anFeeInvalidRequestPayload,

		// Import Package Course
		`^an package course valid request payload with "([^"]*)"`:    s.anPackageCoursesValidRequestPayloadWith,
		`^"([^"]*)" importing package course$`:                       s.importingPackageCourses,
		`^the valid package course lines are imported successfully$`: s.theValidPackageCoursesLinesAreImportedSuccessfully,
		`^the invalid package course lines are returned with error$`: s.theInvalidPackageCoursesLinesAreReturnedWithError,
		`^an package course invalid "([^"]*)" request payload$`:      s.anPackageCoursesInvalidRequestPayload,

		// Import Product Location
		`^an product location valid request payload with correct data with "([^"]*)"`:  s.anProductLocationsValidRequestPayloadWithCorrectDataWith,
		`^an product location valid request payload with incorrect data with"([^"]*)"`: s.anProductLocationsValidRequestPayloadWithIncorrectDataWith,
		`^an product location invalid "([^"]*)" request payload$`:                      s.anProductLocationsInvalidRequestPayload,
		`^"([^"]*)" importing product location$`:                                       s.importingProductLocations,
		`^the valid product location lines are imported successfully$`:                 s.theValidProductLocationsLinesAreImportedSuccessfully,
		`^the import product location transaction is rolled back$`:                     s.theImportProductLocationTransactionIsRolledBack,
		`^the invalid product location lines are returned with error$`:                 s.theInvalidProductLocationsLinesAreReturnedWithError,

		// Import all-in-one csv file for test
		`^a valid request payload for importing all-in-one csv file for test$`: s.aValidRequestPayloadForImportingAllInOneCsvFileForTest,
		`^"([^"]*)" import all-in-one csv file for test$`:                      s.importAllInOneCsvFileForTest,
		`^the valid all-in-one csv file for test is imported successfully$`:    s.theValidAllInOneCsvFileForTestIsImportedSuccessfully,

		// Import Leaving reason
		`^an leaving reason invalid "([^"]*)" request payload$`:                        s.anLeavingReasonInvalidRequestPayload,
		`^an leaving reason valid request payload with "([^"]*)"$`:                     s.anLeavingReasonValidRequestPayloadWith,
		`^"([^"]*)" importing leaving reason$`:                                         s.importingLeavingReason,
		`^an leaving reason valid request payload with incorrect data with "([^"]*)"$`: s.anLeavingReasonValidRequestPayloadWithIncorrectData,
		`^the import leaving reason transaction is rolled back$`:                       s.theImportLeavingReasonTransactionIsRolledBack,
		`^the invalid leaving reason lines are returned with error$`:                   s.theInvalidLeavingReasonLinesAreReturnedWithError,
		`^the valid leaving reason lines are imported successfully$`:                   s.theValidLeavingReasonLinesAreImportedSuccessfully,

		// Import Billing ratio
		`^a billing ratio valid request payload with "([^"]*)"$`:                     s.aBillingRatioValidRequestPayloadWith,
		`^"([^"]*)" importing billing ratio$`:                                        s.importingBillingRatio,
		`^the valid billing ratio lines are imported successfully$`:                  s.theValidBillingRatioLinesAreImportedSuccessfully,
		`^a billing ratio valid request payload with incorrect data with "([^"]*)"$`: s.aBillingRatioValidRequestPayloadWithIncorrectData,
		`^the import billing ratio transaction is rolled back$`:                      s.theImportBillingRatioTransactionIsRolledBack,
		`^the invalid billing ratio lines are returned with error$`:                  s.theInvalidBillingRatioLinesAreReturnedWithError,
		`^a billing ratio invalid "([^"]*)" request payload$`:                        s.aBillingRatioInvalidRequestPayload,

		// Import Package Quantity Type Mapping
		`^a package quantity type mapping valid request payload with correct data with "([^"]*)"$`:   s.aPackageQuantityTypeMappingValidRequestPayloadWithCorrectDataWith,
		`^a package quantity type mapping valid request payload with incorrect data with "([^"]*)"$`: s.aPackageQuantityTypeMappingValidRequestPayloadWithIncorrectDataWith,
		`^a package quantity type mapping invalid request payload with "([^"]*)"$`:                   s.aPackageQuantityTypeMappingInvalidRequestPayload,
		`^"([^"]*)" importing package quantity type mapping$`:                                        s.importingPackageQuantityTypeMapping,
		`^the valid package quantity type mapping lines are imported successfully$`:                  s.theValidPackageQuantityTypeMappingLinesAreImportedSuccessfully,
		`^the import package quantity type mapping transaction is rolled back$`:                      s.theImportPackageQuantityTypeMappingTransactionIsRolledBack,
		`^the invalid package quantity type mapping lines are returned with error$`:                  s.theInvalidPackageQuantityTypeMappingLinesAreReturnedWithError,

		// Import Product Setting
		`^a product setting valid request payload with correct data with "([^"]*)"$`:   s.aProductSettingValidRequestPayloadWithCorrectData,
		`^a product setting valid request payload with incorrect data with "([^"]*)"$`: s.aProductSettingValidRequestPayloadWithIncorrectData,
		`^the valid product setting lines are imported successfully$`:                  s.theValidProductSettingLinesAreImportedSuccessfully,
		`^"([^"]*)" importing product setting$`:                                        s.importingProductSetting,
		`^the import product setting transaction is rolled back$`:                      s.theImportProductSettingTransactionIsRolledBack,
		`^the invalid product setting lines are returned with error$`:                  s.theInvalidProductSettingLinesAreReturnedWithError,
		`^a product setting invalid request payload with "([^"]*)"$`:                   s.aProductSettingInvalidRequestPayload,

		// Crete Order One Time Material
		`^prepare data for create order one time material with "([^"]*)"$`: s.prepareDataForCreateOrderOneTimeMaterial,
		`^"([^"]*)" submit order$`:                                                             s.userSubmitOrder,
		`^order one time material is created successfully$`:                                    s.createOrderOneTimeMaterialSuccess,
		`^request for create order one time material with "([^"]*)"$`:                          s.prepareDataForCreateOrderOneTimeMaterialWithCase,
		`^receives "([^"]*)" error message for create order one time material with "([^"]*)"$`: s.receivesErrMessageForCreateOneTimeMaterial,

		// Create Order One Time Fee
		`^prepare data for creating order one time fee$`:                                  s.prepareDataForCreateOrderOneTimeFee,
		`^order one time fee is created successfully$`:                                    s.createOrderOneTimeFeeSuccess,
		`^request for create order one time fee with "([^"]*)"$`:                          s.prepareDataForCreateOrderOneTimeFeeWithCase,
		`^receives "([^"]*)" error message for create order one time fee with "([^"]*)"$`: s.receivesErrMessageForCreateOneTimeFee,

		// Create Order One Time Package
		`^prepare data for create order one time package$`:           s.prepareDataForCreateOrderOneTimePackage,
		`^order one time package is created successfully$`:           s.createOrderOneTimePackageSuccess,
		`^an event must be published to signal student course sync$`: s.eventPublishedSignalStudentCourseEventSync,

		// Create Order Slot base Package
		`^prepare data for create order slot base package$`: s.prepareDataForCreateOrderSlotBasePackage,
		`^order slot base package is created successfully$`: s.createOrderSlotBasePackageSuccess,

		// Kafka sync Order data to Elasticsearch
		`^prepare order data for elastic sync$`:                      s.prepareOrderRecordUpdate,
		`^a record is "([^"]*)" in order table$`:                     s.orderRecordUpdatedInDB,
		`^the record "([^"]*)" must be reflected in ES order index$`: s.orderRecordReflectedInES,

		// Kafka sync user bounded context
		`^admin inserts a user record to bob$`: s.adminInsertsAUserRecordToBob,
		`^payment user table will be updated$`: s.paymentUserTableWillBeUpdated,

		// Kafka sync User Access Paths
		`^a record is inserted in user access paths in bob$`:  s.insertUserAccessPathsRecordToBob,
		`^the user access paths must be recorded in payment$`: s.userAccessPathsRecordedInPayment,

		// Get order list
		`^prepare data for getting order list$`:                           s.prepareDataForGetOrderList,
		`^"([^"]*)" create orders data for getting order list$`:           s.createOrders,
		`^"([^"]*)" get order list after creating orders with "([^"]*)"$`: s.getOrderListWithFilter,

		// Create Order Recurring Fee
		`^prepare data for create order recurring fee with "([^"]*)"$`:                     s.prepareDataForCreateOrderRecurringFeeWithValidRequest,
		`^order recurring fee is created successfully$`:                                    s.createOrderRecurringFeeSuccess,
		`^request for create order recurring fee with "([^"]*)"$`:                          s.prepareDataForCreateOrderRecurringFeeWithInvalidRequest,
		`^receives "([^"]*)" error message for create order recurring fee with "([^"]*)"$`: s.receivesErrMessageForCreateRecurringFee,

		// Create Order Recurring Material
		`^prepare data for create order recurring material with "([^"]*)"$`:                     s.prepareDataForCreateOrderRecurringMaterialWithValidRequest,
		`^order recurring material is created successfully$`:                                    s.createOrderRecurringMaterialSuccess,
		`^request for create order recurring material with "([^"]*)"$`:                          s.prepareDataForCreateOrderRecurringMaterialWithInvalidRequest,
		`^receives "([^"]*)" error message for create order recurring material with "([^"]*)"$`: s.receivesErrMessageForCreateRecurringMaterial,

		// Import Associated Products by Material
		`^associated products by material valid request payload with "([^"]*)"$`:                     s.associatedProductsByMaterialValidRequestPayloadWith,
		`^"([^"]*)" importing associated products by material$`:                                      s.importingAssociatedProductsByMaterial,
		`^the invalid associated products by material lines are returned with error$`:                s.theInvalidAssociatedProductsByMaterialLinesAreReturnedWithError,
		`^the valid associated products by material lines are imported successfully$`:                s.theValidAssociatedProductsByMaterialLinesAreImportedSuccessfully,
		`^associated products by material invalid "([^"]*)" request payload$`:                        s.associatedProductsByMaterialInvalidRequestPayload,
		`^associated products by material valid request payload with incorrect data with "([^"]*)"$`: s.associatedProductsByMaterialValidRequestPayloadWithIncorrectDataWith,
		`^the import associated products by material transaction is rolled back$`:                    s.theImportAssociatedProductsByMaterialTransactionIsRolledBack,

		// Import Associated Products by Fee
		`^associated products by fee valid request payload with "([^"]*)"$`:                     s.associatedProductsByFeeValidRequestPayloadWith,
		`^"([^"]*)" importing associated products by fee$`:                                      s.importingAssociatedProductsByFee,
		`^the invalid associated products by fee lines are returned with error$`:                s.theInvalidAssociatedProductsByFeeLinesAreReturnedWithError,
		`^the valid associated products by fee lines are imported successfully$`:                s.theValidAssociatedProductsByFeeLinesAreImportedSuccessfully,
		`^associated products by fee invalid "([^"]*)" request payload$`:                        s.associatedProductsByFeeInvalidRequestPayload,
		`^associated products by fee valid request payload with incorrect data with "([^"]*)"$`: s.associatedProductsByFeeValidRequestPayloadWithIncorrectDataWith,
		`^the import associated products by fee transaction is rolled back$`:                    s.theImportAssociatedProductsByFeeTransactionIsRolledBack,

		// Get orders items list
		`^prepare data for get list order items create "([^"]*)" "([^"]*)"$`: s.prepareDataForGetOderItemsList,
		`^"([^"]*)" create "([^"]*)" orders data for get list order items$`:  s.createOrdersForOrderItems,
		`^"([^"]*)" get list order items with "([^"]*)"$`:                    s.getOrderItemsListWithFilter,

		// Create Order Enrollment
		`^order enrollment is created successfully$`:                                    s.orderEnrollmentIsCreatedSuccessfully,
		`^prepare data for create order enrollment with "([^"]*)"$`:                     s.prepareDataForCreateOrderEnrollmentWith,
		`^receives "([^"]*)" error message for create order enrollment with "([^"]*)"$`: s.receivesErrorMessageForCreateOrderEnrollmentWith,
		`^request for create order enrollment with "([^"]*)"$`:                          s.requestForCreateOrderEnrollmentWith,

		// Create Order Package One Time With Product Association
		`^prepare data for create order one time package with association product$`:                        s.prepareDataForCreateOrderOneTimePackageWithAssociationProduct,
		`^prepare data for create order one time package with association product and duplicated product$`: s.prepareDataForCreateOrderOneTimePackageWithAssociationProductWithDuplicatedProduct,
		`^prepare data for create order one time package with association recurring product$`:              s.prepareDataForCreateOrderOneTimePackageWithAssociationRecurringProduct,
		`^"([^"]*)" get list order package with association product$`:                                      s.getOrderProductAssociatedOfPackageList,
		`^check response data of successfully$`:                                                            s.checkResponseOrderProductAssociated,

		// Get bill items list
		`^prepare data for get list bill items create "([^"]*)" "([^"]*)"$`: s.prepareDataForGetBillItemsList,
		`^"([^"]*)" create "([^"]*)" orders data for get list bill items$`:  s.createOrdersForBillItems,
		`^"([^"]*)" get list bill items with "([^"]*)"$`:                    s.getBillItemsListWithFilter,

		// Update Bill Item Status used by Invoice Service
		`^there is an existing bill items from order records with "([^"]*)"$`: s.thereIsAnExistingBillItemsFromOrderRecords,
		`^"([^"]*)" request payload to update bill items status$`:             s.requestPayloadToUpdateBillItemsStatus,
		`^"([^"]*)" submitted the request using "([^"]*)"$`:                   s.submittedTheRequest,
		`^response has no errors$`:                                            s.responseHasNoErrors,
		`^bill items status updated "([^"]*)"$`:                               s.billItemsStatusAreUpdated,
		`^invoiced bill items will have invoiced order$`:                      s.invoicedBillItemsWillHaveInvoicedOrder,

		// Update Order Status
		`^order status updated "([^"]*)"$`:                   s.orderStatusUpdated,
		`^"([^"]*)" request payload to update order status$`: s.requestPayloadToUpdateOrderStatus,
		`^there is an existing order from order records$`:    s.thereIsAnExistingOrderFromOrderRecords,
		`^"([^"]*)" submitted the update order request$`:     s.submittedTheUpdateOrderRequest,
		`^update order status response has no errors$`:       s.updateOrderStatusResponseHasNoErrors,

		// Create Custom Billing
		`^custom billing is created successfully$`:                                    s.customBillingIsCreatedSuccessfully,
		`^prepare data for creating custom billing$`:                                  s.prepareDataForCreatingCustomBilling,
		`^prepare data for creating custom billing with account category$`:            s.prepareDataForCreatingCustomBillingWithAccountCategory,
		`^"([^"]*)" submit custom billing request$`:                                   s.submitCustomBillingRequest,
		`^receives "([^"]*)" error message for create custom billing with "([^"]*)"$`: s.receivesErrorMessageForCreateCustomBillingWith,
		`^request for create custom billing with "([^"]*)"$`:                          s.requestForCreateCustomBillingWith,

		// Get billing items in order details
		`^"([^"]*)" create "([^"]*)" order with "([^"]*)" products successfully$`: s.createOrderSuccessfully,
		`^"([^"]*)" get bill items of "([^"]*)" order$`:                           s.getBillItemsOfOrderDetails,
		`^get bill items of order successfully$`:                                  s.getBillItemsOfOrderDetailsSuccessfully,

		// Update One Time Material
		`^prepare data for update order one time material with bill_status invoiced$`:  s.prepareForUpdateOneTimeMaterialWithStatusInvoiced,
		`^order one time material with bill_status invoiced is updated successfully$`:  s.updateOrderOneTimeMaterialWithStatusInvoicedSuccess,
		`^prepare data for cancel order one time material with bill_status ordered$`:   s.prepareForCancelOneTimeMaterialWithStatusOrdered,
		`^order one time material with bill_status ordered is cancelled successfully$`: s.cancelOrderOneTimeMaterialWithStatusOrderedSuccess,

		// Get Product List Of Order
		`^"([^"]*)" get product list of "([^"]*)" order with "([^"]*)" filter$`: s.getProductListOfOrder,
		`^get product list of order with "([^"]*)" response successfully$`:      s.checkProductListOfOrderResponse,

		// Get order product list of student billing
		`^create data for order "([^"]*)" "([^"]*)"$`:                                      s.prepareDataForGetOrderProductList,
		`^"([^"]*)" create "([^"]*)" orders data for get list order product successfully$`: s.createOrdersForOrderProduct,
		`^"([^"]*)" get list order product with "([^"]*)"$`:                                s.getOrderProductList,
		`^check response data of "([^"]*)" "([^"]*)" with "([^"]*)" successfully$`:         s.checkResponseOrderProduct,

		// Create bulk order
		`^prepare data for create bulk order$`: s.prepareDataForCreateBulkOrder,
		`^"([^"]*)" submit bulk order$`:        s.createBulkOrder,
		`^bulk order is created successfully$`: s.createBulkOrderSuccess,

		// Create Order Frequency-Base Package
		`^prepare data for create order frequency-base package$`:                                    s.prepareDataForCreateOrderFrequencyBasePackage,
		`^order frequency-base package is created successfully$`:                                    s.createOrderFrequencyBasePackageSuccess,
		`^request for create order frequency-base package with "([^"]*)"$`:                          s.prepareDataForCreateOrderFrequencyBasePackageWithInvalidRequest,
		`^receives "([^"]*)" error message for create order frequency-base package with "([^"]*)"$`: s.receivesErrMessageForCreateOrderFrequencyBasePackage,

		// Create Order Schedule-Base Package
		`^prepare data for create order schedule-base package$`:                                    s.prepareDataForCreateOrderScheduleBasePackage,
		`^order schedule-base package is created successfully$`:                                    s.createOrderScheduleBasePackageSuccess,
		`^request for create order schedule-base package with "([^"]*)"$`:                          s.requestForCreateOrderSchedulebasePackageWith,
		`^receives "([^"]*)" error message for create order schedule-base package with "([^"]*)"$`: s.receivesErrorMessageForCreateOrderSchedulebasePackageWith,

		// Void An Order With One Time Products
		`^"([^"]*)" create an order with one-time products successfully$`: s.createOrderWithOneTimeProductsSuccessfully,
		`^"([^"]*)" void an "([^"]*)" order with one-time-products$`:      s.voidOrderWithOneTimeProducts,
		`^void an order with one-time-products successfully$`:             s.voidOrderWithOneTimeProductsSuccessfully,
		`^"([^"]*)" void an "([^"]*)" order with out of version request$`: s.voidOrderWithOneTimeProductsOutOfVersion,
		`^void an order with one-time-products unsuccessfully$`:           s.updateOrderOutOfVersionUnsuccess,

		// Update One Time Fee
		`^prepare data for update order one time fee with bill_status invoiced$`:  s.prepareForUpdateOneTimeFeeWithStatusInvoiced,
		`^order one time fee with bill_status invoiced is updated successfully$`:  s.updateOrderOneTimeFeeWithStatusInvoicedSuccess,
		`^prepare data for cancel order one time fee with bill_status ordered$`:   s.prepareForCancelOneTimeFeeWithStatusOrdered,
		`^order one time fee with bill_status ordered is cancelled successfully$`: s.cancelOrderOneTimeFeeWithStatusOrderedSuccess,
		`^prepare data for update order one time fee of out version$`:             s.prepareForUpdateOneTimeFeeWithStatusInvoicedAndOutVersion,
		`^order one time fee with bill_status invoiced is not updated$`:           s.updateOrderOutOfVersionUnsuccess,

		// Update One Time Package
		`^update data for update order one time package$`:  s.prepareDataForUpdateOrderOneTimePackage,
		`^update data for cancel order one time package$`:  s.prepareDataForCancelOrderOneTimePackage,
		`^order one time package is updated successfully$`: s.updateOrderOneTimePackageSuccess,

		// Withdraw Recurring Material
		`^prepare data for withdraw order recurring material with "([^"]*)"$`: s.prepareDataForCreateOrderWithdrawRecurringMaterial,
		`^withdraw order recurring material is created successfully$`:         s.createOrderWithdrawRecurringMaterialSuccess,

		// Withdraw Recurring Fee
		`^prepare data for withdraw order recurring fee with "([^"]*)"$`:    s.prepareDataForCreateOrderWithdrawRecurringFee,
		`^withdraw order recurring fee is created "([^"]*)" for "([^"]*)"$`: s.createOrderWithdrawRecurringFeeSuccessFor,

		// Withdraw Schedule-Base Package
		`^prepare data for withdraw order schedule-base package with "([^"]*)"$`: s.prepareDataForCreateOrderWithdrawScheduleBasePackage,
		`^withdraw order schedule-base package is created successfully$`:         s.createOrderWithdrawScheduleBasePackageSuccess,

		// Withdraw Frequency-Base Package
		`^prepare data for withdraw order frequency-base package with "([^"]*)"$`: s.prepareDataForCreateOrderWithdrawFrequencyBasePackage,
		`^withdraw order frequency-base package is created successfully$`:         s.createOrderWithdrawFrequencyBasePackageSuccess,

		// Update Recurring Material
		`^prepare data for create order recurring material$`: s.prepareDataForCreateOrderRecurringMaterial,
		`^prepare data for update order recurring material$`: s.prepareDataForUpdateOrderRecurringMaterial,
		`^prepare data for cancel order recurring material$`: s.prepareDataForCancelOrderRecurringMaterial,
		`^update recurring material success$`:                s.updateOrderForRecurringMaterialSuccess,

		// Update Recurring Package
		`^prepare data for create order recurring package$`:                s.prepareDataForCreateOrderRecurringPackage,
		`^prepare data for update order recurring package$`:                s.prepareDataForUpdateOrderRecurringPackage,
		`^prepare data for cancel order recurring package$`:                s.prepareDataForCancelOrderRecurringPackage,
		`^update recurring package success$`:                               s.updateOrderForRecurringPackageSuccess,
		`^prepare data for update order recurring package out of version$`: s.prepareDataForUpdateOrderRecurringPackageOutOfVersion,
		`^update recurring package unsuccess with out version$`:            s.updateOrderOutOfVersionUnsuccess,

		// Cronjob Generate Billing Items
		`^next billing items are generated$`:                                                      s.nextBillingItemsAreGenerated,
		`^prepare data for scheduled generation of bill items recurring material with "([^"]*)"$`: s.prepareDataForScheduledGenerationOfBillItemsRecurringMaterialWith,
		`^order is created and next upcoming billing date is within 30 days$`:                     s.nextBillingDateIsWithin30Days,

		// Create Order - Multiple Products
		`^order of multiple one time products is created successfully$`:               s.createOrderOneTimeFeeSuccess,
		`^prepare data for creating order multiple one time products$`:                s.prepareDataForCreatingOrderMultipleOneTimeProducts,
		`^order of multiple recurring products is created successfully$`:              s.createOrderRecurringFeeSuccess,
		`^prepare data for creating order multiple recurring products$`:               s.prepareDataForCreatingOrderMultipleRecurringProducts,
		`^order of multiple one time and recurring products is created successfully$`: s.createOrderRecurringFeeSuccess,
		`^prepare data for creating order multiple one time and recurring products$`:  s.prepareDataForCreatingOrderMultipleOneTimeAndRecurringProducts,

		// Void An Order With Recurring Products
		`^"([^"]*)" create an "([^"]*)" order with recurring products successfully$`: s.createOrderWithRecurringProductsSuccessfully,
		`^void an order with recurring products$`:                                    s.voidOrderWithRecurringProducts,
		`^void an order with recurring products successfully$`:                       s.voidOrderWithRecurringProductsSuccessfully,
		`^void an order with recurring products out of version$`:                     s.voidOrderWithRecurringProductsOutOfVersion,
		`^void an order with recurring products unsuccessfully$`:                     s.updateOrderOutOfVersionUnsuccess,

		// Subscribe created order event log
		`^subscribe created order event log$`: s.subscribeCreatedOrderEventLog,

		// Publish order log event for enrollment
		`^prepare valid order request for enrollment$`:                            s.prepareDataValidOrderRequestForEnrollment,
		`^an event must be published and enrollment status is updated "([^"]*)"$`: s.eventPublishedSignalEnrollmentOrderSubmitted,

		// Publish order log event for withdrawal
		`^prepare valid order request for withdrawal$`:                       s.prepareDataValidOrderRequestForWithdrawal,
		`^an event must be published to signal withdrawal order submitted$`:  s.eventPublishedSignalWithdrawalOrderSubmitted,
		`^prepare valid order request for withdrawal with empty product$`:    s.prepareDataValidOrderRequestForWithdrawalWithEmptyProduct,
		`^an event must be published to signal voiding of withdrawal order$`: s.eventPublishedSignalVoidWithdrawalOrderSubmitted,
		`^"([^"]*)" void withdrawal order without products$`:                 s.voidWithdrawalOrderWithNoProducts,

		// Publish order log event for graduation
		`^prepare valid order request for graduation$`:                      s.prepareDataValidOrderRequestForGraduation,
		`^an event must be published to signal graduation order submitted$`: s.eventPublishedSignalGraduationOrderSubmitted,

		// Import Product Discount
		`^an product discount valid request payload with correct data with "([^"]*)"`:   s.anProductDiscountValidRequestPayloadWithCorrectDataWith,
		`^an product discount valid request payload with incorrect data with "([^"]*)"`: s.anProductDiscountValidRequestPayloadWithIncorrectDataWith,
		`^an product discount invalid request payload with "([^"]*)"$`:                  s.anProductDiscountInvalidRequestPayload,
		`^"([^"]*)" importing product discount$`:                                        s.importingProductDiscount,
		`^the valid product discount lines are imported successfully$`:                  s.theValidProductDiscountLinesAreImportedSuccessfully,
		`^the import product discount transaction is rolled back$`:                      s.theImportProductDiscountTransactionIsRolledBack,
		`^the invalid product discount lines are returned with error$`:                  s.theInvalidProductDiscountLinesAreReturnedWithError,

		// Update Order Reviewed Flag
		`^order reviewed flag updated "([^"]*)"$`:                                  s.orderReviewedFlagUpdated,
		`^"([^"]*)" request payload to update order reviewed flag$`:                s.requestPayloadToUpdateOrderReviewedFlag,
		`^an existing order from order records$`:                                   s.anExistingOrderFromOrderRecords,
		`^"([^"]*)" submitted the update order reviewed flag request$`:             s.submittedTheUpdateOrderReviewedFlagRequest,
		`^update order reviewed flag response success$`:                            s.updateOrderReviewedFlagResponseSuccess,
		`^"([^"]*)" request out of version payload to update order reviewed flag$`: s.requestOutOfVersionPayloadToUpdateOrderReviewedFlag,
		`^update order reviewed flag response unsuccess$`:                          s.updateOrderReviewedFlagResponseUnsuccess,

		// Update Student Product Status
		`^prepare data for "([^"]*)" order with valid effective date`:           s.prepareDataForOrderWithValidEffectiveDate,
		`^the scheduled job runs for "([^"]*)" on the effective date of order$`: s.theScheduledJobRunsOnTheEffectiveDateOfOrder,
		`^student product status changed to "([^"]*)"$`:                         s.studentProductStatusChangedToNewStatus,

		// Get unique product
		`^create data for order "([^"]*)" "([^"]*)" for unique product$`:               s.prepareDataForUniqueProductList,
		`^"([^"]*)" create "([^"]*)" orders data for get unique product successfully$`: s.createOrdersForUniqueProduct,
		`^"([^"]*)" get unique product$`:                                               s.getUniqueProductList,
		`^check unique product of "([^"]*)"$`:                                          s.checkResponseUniqueProduct,

		// Get unique product for bulk order
		`^create data for bulk order for unique product$`:                         s.prepareDataForUniqueProductListForBulkOrder,
		`^"([^"]*)" create bulk orders data for get unique product successfully$`: s.createBulkOrdersForUniqueProduct,
		`^"([^"]*)" get unique product for bulk order$`:                           s.getUniqueProductListForBulkOrder,
		`^list of unique products for bulk order were returned correctly$`:        s.checkResponseUniqueProductForBulkOrder,

		// Order-base student course subscription
		`^prepare data for withdraw order recurring package$`: s.prepareDataWithdrawalRecurringPackage,
		`^prepare data for graduate order recurring package$`: s.prepareDataGraduationRecurringPackage,
		`^package upserted to student package table$`:         s.packageUpsertedToStudentPackageTable,

		// Upload enrollment pdf file
		`^prepare enrollment pdf for upload$`:          s.prepareFileUpload,
		`^"([^"]*)" upload file$`:                      s.uploadEnrollmentPDF,
		`^"([^"]*)" get download url enrollment file$`: s.getDownloadUrlEnrollmentPDF,

		// Get export student billing data
		`^the organization "([^"]*)" has existing student billing data$`: s.theOrganizationHasExistingBankBranchData,
		`^"([^"]*)" export student billing data$`:                        s.theUserExportStudentBilling,
		`^the student billing CSV has a correct content$`:                s.theStudentBillingCSVHasCorrectContent,

		// LOA
		`^prepare data for LOA request with "([^"]*)"$`:                             s.prepareDataForCreateLOARequest,
		`^LOA request is created successfully$`:                                     s.createLOARequestSuccess,
		`^prepare data for LOA request with product pausable tag set to "([^"]*)"$`: s.prepareDataForCreateLOARequestWithPausableTag,
		`^product setting pausable tag "([^"]*)" validated successfully$`:           s.pausableTagValidatedSuccessfully,

		// Import student course
		`^a student course valid request payload$`: s.prepareDataForImportStudentCourse,
		`^"([^"]*)" importing student course$`:     s.importStudentCourse,

		// Get export master data
		`^data of "([^"]*)" is existing$`:                s.addDataForExportMasterData,
		`^"([^"]*)" export "([^"]*)" data successfully$`: s.theUserExportMasterData,
		`^the "([^"]*)" CSV has a correct content$`:      s.theMasterDataCSVHasCorrectContent,

		// Import student class
		`^a student class valid request payload$`:        s.prepareDataForImportStudentClass,
		`^"([^"]*)" importing student class for insert`:  s.importStudentClassForInsert,
		`^^"([^"]*)" importing student class for delete`: s.importStudentClassForDelete,

		// Resume products
		`^prepare data for resume order when student status is LOA$`: s.prepareDataForResumeProducts,
		`^paused products are resumed successfully$`:                 s.resumeProductsSuccess,

		// Get locations for creating order
		`^new locations$`:                                s.createLocations,
		`^new "([^"]*)" is granted to locations$`:        s.createUserAndAssignToLocations,
		`^getting locations for creating order$`:         s.getLocationsForCreatingOrder,
		`^check response data with "([^"]*)" locations$`: s.checkLocationsForCreatingOrder,

		// Create order with enrollment required products
		`^prepare data for create order with enrollment required tag set to to "([^"]*)" and student status set to "([^"]*)"$`: s.prepareEnrollmentRequiredProductOrder,
		`^permission to order enrollment required tag set to "([^"]*)" and student status set to "([^"]*)" is validated$`:      s.orderWithEnrollmentRequiredTagIsValidated,

		// Get product list
		`^create products to get product list$`:                              s.prepareDataForGetProductList,
		`^"([^"]*)" get product list after creating product with "([^"]*)"$`: s.getProductList,
		`^get product list successfully$`:                                    s.getProductListSuccessfully,

		// Get Enrolled Status
		`^prepare data for student enrollment status history$`:         s.prepareDataForGetOrgLevelStudentStatusValidRequest,
		`^"([^"]*)" get organization enrolled status$`:                 s.getOrgEnrollmentStatus,
		`^check valid data get organization enrolled status$`:          s.checkDataWhenGetOrgEnrollmentStatus,
		`^"([^"]*)" get student enrolled location$`:                    s.getStudentEnrolledLocation,
		`^check valid data get student enrolled location$`:             s.checkDataWhenGetStudentEnrolledLocation,
		`^"([^"]*)" get student enrollment status by location$`:        s.getStudentEnrollmentStatusByLocation,
		`^check valid data get student enrollment status by location$`: s.checkDataWhenGetStudentEnrollmentStatusByLocation,

		// Import Notification Date
		`^the valid notification date lines are imported successfully$`:                  s.theValidNotificationDateLinesAreImportedSuccessfully,
		`^a notification date valid request payload with "([^"]*)"$`:                     s.aNotificationDateValidRequestPayloadWith,
		`^"([^"]*)" importing notification date`:                                         s.importingNotificationDate,
		`^a notification date valid request payload with incorrect data with "([^"]*)"$`: s.aNotificationDateValidRequestPayloadWithIncorrectData,
		`^the import notification date transaction is rolled back$`:                      s.theImportNotificationDateTransactionIsRolledBack,
		`^the invalid notification date lines are returned with error$`:                  s.theInvalidNotificationDateLinesAreReturnedWithError,
		`^a notification date invalid "([^"]*)" request payload$`:                        s.aNotificationDateInvalidRequestPayload,

		// Manual modify student course
		`^prepare data for manual insert student course$`:        s.prepareDataForInsertStudentCourse,
		`"([^"]*)" submit manual modify student course request$`: s.manualModifyStudentCourse,

		// Kafka sync for student parents table on bob to fatima
		`^a record is inserted in student parent in bob$`:  s.aRecordIsInsertedInStudentParentInBob,
		`^the student parent must be recorded in payment$`: s.theStudentParentMustBeRecordedInPayment,

		// Discount auto update
		`^prepare data for create order recurring "([^"]*)"$`:                   s.prepareDataForRecurringProductForDiscountAutomation,
		`^student tagged for org level "([^"]*)" discount$`:                     s.studentTaggedOrgLevelDiscount,
		`^discount service sends data for discount update$`:                     s.discountServiceSendsDataForDiscountUpdate,
		`^recurring "([^"]*)" "([^"]*)" with discount is updated successfully$`: s.recurringProductDiscountUpdatedSuccessfully,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
