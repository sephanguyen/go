package constant

const (
	UserGroupStudent     = "USER_GROUP_STUDENT"
	UserGroupAdmin       = "USER_GROUP_ADMIN"
	UserGroupSchoolAdmin = "USER_GROUP_SCHOOL_ADMIN"
	RowScanError         = "row.Scan: %w"

	OptimisticLockingEntityVersionMismatched = "optimistic_locking_entity_version_mismatched" // User performs an action with an outdated entity

	MissingCourseInfoBillItem             = "missing_course_info_in_bill_item_of_order"                             // User creates an Order (of any type) with one of the Orders' Package bill items missing course info
	BillItemHasNoSchedulePeriodID         = "bill_item_missing_billing_schedule_period_id"                          // User creates an Order (of any type)
	BillItemHasNoSchedulePeriodIDDebugMsg = "Billing item index %d is missing BillingSchedulePeriodId"              // with one of the Orders' Package bill items missing billing_schedule_period_id
	InconsistentDiscountID                = "inconsistent_discount_id_between_order_item_and_bill_item"             // User creates an Order (of any type)
	InconsistentDiscountIDDebugMsg        = "Mismatch discount in billing item = %s vs discount in order item = %s" // with inconsistent bill item's discount and order item's discount

	InconsistentTax = "inconsistent_tax_between_tax_in_product_and_tax_in_bill_item" // User creates an Order (of any type) with inconsistent bill item's tax and product's tax

	DiscountIsNotAvailable             = "discount_is_not_available"                             // User creates an Order (of any type) when it's outside of a Discount's available range
	DiscountAmountsAreNotEqual         = "discount_amount_mismatch"                              // User creates an Order (of any type)
	DiscountAmountsAreNotEqualDebugMsg = "Discount amount is wrong actual = %v vs expected = %v" // with mismatched discount_amount between BE and FE

	UpdateLikeOrdersInvalidEffectiveDate = "update_like_order_invalid_effective_date" // Check usage
	UpdateLikeOrdersMissingBillItem      = "update_like_order_missing_billing_item"   // Check usage

	// User creates an Enrollment order for students that are not eligible
	InvalidStudentEnrollmentStatus                = "invalid_student_enrollment_status"
	InvalidStudentEnrollmentStatusAlreadyEnrolled = "invalid_student_enrollment_status_already_enrolled"
	InvalidStudentEnrollmentStatusOnLOA           = "invalid_student_enrollment_status_on_loa"
	InvalidStudentEnrollmentStatusUnavailable     = "invalid_student_enrollment_status_unavailable"
	InvalidEffectedDateForWithdrawalAndGraduate   = "invalid_effected_date_when_withdrawal_and_graduate"

	IncorrectFinalPrice                     = "Incorrect final price in bill item of product with id %v actual = %v vs expected = %v"
	IncorrectProductPrice                   = "Incorrect price of product with id %v actual = %v vs expected = %v"
	MissingAdjustmentPriceWhenUpdatingOrder = "Missing adjustment price in bill item with product %v when update order"
	CourseItemMissingSlotField              = "Course item for slot base is missing slot field"
	CourseHasSlotGreaterThanMaxSlot         = "Course with id %s has slot greater than max slot allowed for the course"
	UnableToUpdateProductDueToPendingOrder  = "Unable to update the product as it has a pending withdrawal/graduate/loa order."

	NoDataInCsvFile                           = "no data in csv file"
	UnableToParseAssociatedProductsByFee      = "unable to parse associated products by fee: %s"
	UnableToParseAssociatedProductsByMaterial = "unable to parse associated products by material: %s"
	UnableToParseBillingRatioItem             = "unable to parse billing ratio item: %s"
	UnableToParseStudentCourse                = "unable to parse student course: %s"
	UnableToParseBillingSchedulePeriodItem    = "unable to parse billing schedule period item: %s"
	UnableToParseBillingScheduleItem          = "unable to parse billing schedule item: %s"
	UnableToParseDiscountItem                 = "unable to parse discount item: %s"
	UnableToParseLeavingReasonItem            = "unable to parse leaving reason item: %s"
	UnableToParseTaxItem                      = "unable to parse tax item: %s"
	UnableToParseNotificationDate             = "unable to parse notification date item: %s"

	ErrorWhenGettingProductPriceWithEmptyQuantity = "Error when get product price of product %v with empty quantity"
	ErrorWhenGettingProRatingPriceOfNoneQuantity  = "Getting prorating price of none quantity recurring product have err %v"
	ErrorWhenGettingProductPrice                  = "Error when get product price of product %v with error %s"

	GetAllQuery = `SELECT %s FROM %s`

	LocationReadPermission = "master.location.read"
	OrderWritePermission   = "payment.order.write"

	ProductSettingDefaultIsPausable                   = true
	ProductSettingDefaultIsEnrollmentRequired         = false
	ProductSettingDefaultIsAddedToEnrollmentByDefault = false
	ProductSettingDefaultIsOperationFee               = false

	DuplicateCourses              = "duplicated_courses"
	DuplicatedAssociate           = "duplicated_associate_product"
	ProductPriceNotExistOrUpdated = "product_price_not_exist_or_updated"

	// UserCantAccessThisCourse error from manual upsert student course
	UserCantAccessThisCourse             = "user_cant_access_this_course"
	DuplicateCourseByManualError         = "duplicate_course_by_manual_error"
	UpdateTimeStudentCourseByManualError = "update_time_student_course_by_manual_error"
	UpsertStudentCourseByManualError     = "upsert_student_course_by_manual_error"
)
