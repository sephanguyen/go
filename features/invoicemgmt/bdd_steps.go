package invoicemgmt

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
		// Health Check
		`^everything is OK$`:                                          s.everythingIsOK,
		`^health check endpoint called$`:                              s.healthCheckEndpointCalled,
		`^invoicemgmt should return "([^"]*)" with status "([^"]*)"$`: s.invoicemgmtShouldReturnWithStatus,
		// Kafka sync User Access Paths
		`^a user access path record is inserted in bob$`:                 s.aUserAccessPathRecordIsInsertedInBob,
		`^this user access path record must be recorded in invoicemgmt$`: s.thisUserAccessPathRecordedInInvoicemgmt,
		// Kafka sync User
		`^a user record is inserted in bob$`:                 s.aUserRecordIsInsertedInBob,
		`^this user record must be recorded in invoicemgmt$`: s.thisUserRecordMustBeRecordedInInvoicemgmt,
		// Kafka sync Location
		`^a location record is inserted in bob$`:                 s.aLocationRecordIsInsertedInBob,
		`^this location record must be recorded in invoicemgmt$`: s.thisLocationRecordMustBeRecordedInInvoicemgmt,
		// Kafka sync Student
		`^a student record is inserted in bob$`:                 s.aStudentRecordIsInsertedInBob,
		`^this student record must be recorded in invoicemgmt$`: s.thisStudentRecordMustBeRecordedInInvoicemgmt,
		// Kafka sync Bill item
		`^invoicemgmt bill item table will be updated$`:                      s.invoicemgmtBillItemTableWillBeUpdated,
		`^this bill item is sync to invoicemgmt$`:                            s.invoicemgmtBillItemTableWillBeUpdated,
		`^admin deletes this bill item record on fatima$`:                    s.adminDeletesThisBillItemRecordOnFatima,
		`^this bill item on invoicemgmt will be deleted$`:                    s.thisBillItemOnInvoicemgmtWillBeDeleted,
		`^admin inserts a bill item record to fatima with status "([^"]*)"$`: s.adminInsertsABillItemRecordToFatimaWithStatus,
		`^there is an existing bill item on fatima with status "([^"]*)"$`:   s.adminInsertsABillItemRecordToFatimaWithStatus,

		// Payment Update Bill Items Status Check
		`^payment endpoint is called to update these bill items status$`: s.paymentEndpointIsCalledToUpdateTheseBillItemsStatus,
		`^receives "([^"]*)" status code$`:                               s.receivesStatusCode,
		`^there is an existing bill items created on payment$`:           s.thereIsAnExistingBillItemsCreatedOnPayment,

		// Issue invoice
		`^"([^"]*)" issues invoice with "([^"]*)" payment method$`: s.issuesInvoiceWithPaymentMethod,
		`^invoice has draft invoice status$`:                       s.invoiceHasDraftInvoiceStatus,
		`^invoice status is updated to "([^"]*)" status$`:          s.invoiceStatusIsUpdatedToStatus,
		`^payment history is recorded with pending status$`:        s.paymentHistoryIsRecordedWithPendingStatus,
		`^there is an existing invoice$`:                           s.thereIsAnExistingInvoice,
		`^invoice ID is non-existing$`:                             s.invoiceIDIsNonexisting,
		`^no payment history is recorded$`:                         s.noPaymentHistoryIsRecorded,
		`^invoice has type "([^"]*)"$`:                             s.invoiceHasType,
		`^invoice exported tag is set to "([^"]*)"$`:               s.invoiceExportedTagIsSetTo,
		`^payment exported tag is set to "([^"]*)"$`:               s.paymentExportedTagIsSetTo,

		// Retrieve Invoice List
		`^"([^"]*)" logins Learner App$`:                                s.loginsLearnerApp,
		`^student has "([^"]*)" invoice records$`:                       s.studentHasInvoiceRecords,
		`^this student has "([^"]*)" invoice records$`:                  s.studentHasInvoiceRecords,
		`^this parent has an existing student$`:                         s.thisParentHasAnExistingStudent,
		`^parent is at the invoice list screen$`:                        s.parentIsAtTheInvoiceListScreen,
		`^parent selects this existing student$`:                        s.parentSelectsThisExistingStudent,
		`^records found with default limit are displayed successfully$`: s.recordsFoundWithDefaultLimitAreDisplayedSuccessfully,
		`^no records found displayed successfully$`:                     s.noRecordsFoundDisplayedSuccessfully,
		`^parent scrolls down to display all records$`:                  s.parentScrollsDownToDisplayAllRecords,
		`^all records found are displayed successfully$`:                s.allRecordsFoundAreDisplayedSuccessfully,
		`^parent has another existing student$`:                         s.parentHasAnotherExistingStudent,
		`^no invoice draft records found$`:                              s.noInvoiceDraftRecordsFound,

		// Create invoice
		`^there is a student that has bill item with status "([^"]*)"$`:                                        s.thereIsAStudentThatHasBillItemWithStatus,
		`^bill item exists in invoicemgmt database$`:                                                           s.billItemExistsInInvoicemgmtDatabase,
		`^generateInvoice endpoint is called to create multiple invoice$`:                                      s.generateInvoiceEndpointIsCalledToCreateMultipleInvoice,
		`^there are "([^"]*)" students that has "([^"]*)" bill item with status "([^"]*)" and type "([^"]*)"$`: s.thereAreStudentsThatHasBillItemWithStatusAndType,
		`^there are "([^"]*)" student draft invoices created successfully$`:                                    s.thereArestudentDraftInvoicesCreatedSuccessfully,
		`^invoice bill item is created$`:                                                                       s.invoiceBillItemIsCreated,
		`^there are no errors in response$`:                                                                    s.thereAreNoErrorsInResponse,
		`^there are "([^"]*)" response error$`:                                                                 s.thereAreResponseError,
		`^invoice data is present in the response with count "([^"]*)"$`:                                       s.invoiceDataIsPresentInTheResponseWithCount,
		`^"([^"]*)" bill item has review required tag$`:                                                        s.billItemHasReviewRequiredTag,
		`^there is an error and no invoice in the response$`:                                                   s.thereIsAnErrorAndNoInvoiceInTheResponse,

		// Retrieve invoice info
		`^invoice has "([^"]*)" status with "([^"]*)" bill items count$`: s.invoiceHasStatusWithBillItemsCount,
		`^logged-in user views an invoice$`:                              s.loggedinUserViewsAnInvoice,
		`^receives "([^"]*)" bill items count$`:                          s.receivesBillItemsCount,

		// Void invoice
		`^"([^"]*)" logins to backoffice app$`:                                        s.loginsToBackofficeApp,
		`^there is an existing invoice with "([^"]*)" invoice status with bill item$`: s.thereIsAnExistingInvoiceWithInvoiceStatusWithBillItem,
		`^bill item has "([^"]*)" previous status$`:                                   s.billItemHasPreviousStatus,
		`^has billing date "([^"]*)" today$`:                                          s.hasBillingDateToday,
		`^there is "([^"]*)" payment history$`:                                        s.thereIsPaymentHistory,
		`^admin voids an invoice with "([^"]*)" remarks$`:                             s.adminVoidsAnInvoiceWithRemarks,
		`^admin voids an invoice with "([^"]*)" remarks using v2 endpoint$`:           s.adminVoidsAnInvoiceWithRemarksUsingV2Endpoint,
		`^invoice has "([^"]*)" invoice status$`:                                      s.invoiceHasInvoiceStatus,
		`^bill item has "([^"]*)" bill item status$`:                                  s.billItemHasBillItemStatus,
		`^latest payment record has "([^"]*)" payment status$`:                        s.latestPaymentRecordHasPaymentStatus,
		`^latest payment record has "([^"]*)" payment status and amount zero$`:        s.latestPaymentRecordHasPaymentStatusAndAmountZero,
		`^action log record is recorded$`:                                             s.actionLogRecordIsRecorded,

		// Approve payment
		`^admin approves payment with "([^"]*)" remarks$`:                      s.adminApprovesPaymentWithRemarks,
		`^admin approves payment with "([^"]*)" remarks without payment date$`: s.adminApprovesPaymentWithRemarksWithoutPaymentDate,
		`^action log record is recorded with "([^"]*)" action log type$`:       s.actionLogRecordIsRecordedWithActionLogType,
		`^bill item has "([^"]*)" final price value$`:                          s.billItemHasFinalPriceValue,
		`^bill item has "([^"]*)" adjustment price value$`:                     s.billItemHasAdjustmentPriceValue,
		`^invoice outstanding_balance set to "([^"]*)"$`:                       s.invoiceOutstandingBalanceSetTo,
		`^invoice amount_paid set to "([^"]*)"$`:                               s.invoiceAmountPaidSetTo,
		`^invoice amount_refunded set to "([^"]*)"$`:                           s.invoiceAmountRefundedSetTo,
		`^admin adds "([^"]*)" invoice adjustment with "([^"]*)" amount$`:      s.addsInvoiceAdjustmentWithAmount,

		// Cancel invoice
		`^admin cancels an invoice with "([^"]*)" remarks$`: s.adminCancelsAnInvoiceWithRemarks,
		`^action log has failed action$`:                    s.actionLogHasFailedAction,

		// Scheduled Invoice Checker
		`^the organizations "([^"]*)" have "([^"]*)" students with "([^"]*)" "([^"]*)" bill items$`: s.theOrganizationsHaveStudentWithBillItems,
		`^there is scheduled invoice to be run at day "([^"]*)" for these organizations "([^"]*)"$`: s.thereIsScheduledInvoicetoBeRunAtDayForTheseOrganizations,
		`^there is no scheduled invoice to be run at day "([^"]*)"$`:                                s.thereIsNoScheduledInvoiceToBeRunAtDay,
		`^the InvoiceScheduleChecker endpoint was called at day "([^"]*)"$`:                         s.theInvoiceScheduledCheckerEndpointWasCalledAtDay,
		`^there are correct number of students invoice generated in organizations "([^"]*)"$`:       s.thereAreCorrectNumberOfStudentInvoiceGeneratedInOrganization,
		`^the scheduled invoice status is updated to "([^"]*)"$`:                                    s.theScheduledInvoiceStatusIsUpdatedTo,
		`^a history of scheduled invoice was saved$`:                                                s.aHistoryOfScheduledInvoiceWasSaved,
		`^there are no invoice scheduled student was saved$`:                                        s.thereAreNoInvoiceScheduledStudentWasSaved,
		`^there are no students invoice generated in organizations "([^"]*)"$`:                      s.thereAreNoStudentsInvoiceGeneratedInOrganizations,
		`^the organizations "([^"]*)" have students with bill items with error$`:                    s.theOrganizationsHaveStudentWithBillItemsWithError,
		`^there are invoice scheduled student saved$`:                                               s.thereAreInvoiceScheduledStudentSaved,
		`^only student billed bill items are invoiced$`:                                             s.onlyStudentBilledBillItemsAreInvoiced,
		`^a bill item of these organization "([^"]*)" has adjustment price "([^"]*)"$`:              s.aBillItemOfTheseOrganizationHasAdjustmentPrice,
		`^all bill item with review required tag was skipped$`:                                      s.allBillItemWithReviewRequiredTagWasSkipped,
		`^the InvoiceScheduleChecker endpoint was called at day "([^"]*)" concurrently$`:            s.theInvoiceScheduledCheckerEndpointWasCalledAtDayConcurrently,
		`^only one response has OK status code and others have error$`:                              s.onlyOneResponseHasOKStatusAndOthersHaveError,
		`^"([^"]*)" bill item created after the cutoff date$`:                                       s.billItemCreatedAfterTheCutoffDate,
		`^all bill item created after the cutoff date was skipped$`:                                 s.allBillItemCreatedAfterCutoffDateWasSkipped,

		// Bulk Issue Invoice
		`^there are existing invoices with "([^"]*)" status$`:                                                            s.thereAreExistingInvoicesWithStatus,
		`^"([^"]*)" issues invoices in bulk with payment method "([^"]*)" and "([^"]*)" "([^"]*)"$`:                      s.issuesInvoicesInBulkWithPaymentMethodAnd,
		`^invoices status is updated to "([^"]*)" status$`:                                                               s.invoicesStatusIsUpdatedToStatus,
		`^one invoice ID is added to the request but is non-existing$`:                                                   s.oneInvoiceIDIsAddedToTheRequestButIsNonexisting,
		`^action log record for each invoice is recorded with "([^"]*)" action log type$`:                                s.actionLogRecordForEachInvoiceIsRecordedWithActionLogType,
		`^one invoice has negative total amount$`:                                                                        s.oneInvoiceHasNegativeTotalAmount,
		`^"([^"]*)" issues invoices in bulk with payment method "([^"]*)" and "([^"]*)" due date after expiry date$`:     s.issuesInvoicesInBulkWithPaymentMethodAndDueDateAfterExpiryDate,
		`^these invoice for students have default payment method "([^"]*)"$`:                                             s.theseInvoiceForStudentsHaveDefaultPaymentMethod,
		`^there are pending payment records for students created with payment method "([^"]*)" and "([^"]*)" "([^"]*)"$`: s.thereArePendingPaymentRecordsForStudentsCreatedWithPaymentMethodAnd,
		`^invoices exported tag is set to "([^"]*)"$`:                                                                    s.invoicesExportedTagIsSetTo,
		`^payments exported tag is set to "([^"]*)"$`:                                                                    s.paymentsExportedTagIsSetTo,

		// Import invoice schedule
		`^"([^"]*)" signed-in user imports invoice schedule file with "([^"]*)" file content type$`: s.signedinUserImportsInvoiceScheduleFileWithFileContentType,
		`^"([^"]*)" from "([^"]*)" org in "([^"]*)" country logins to backoffice app$`:              s.fromOrgInCountryLoginsToBackofficeApp,
		`^error list is empty$`: s.errorListIsEmpty,
		`^import schedule reflects in the DB based on "([^"]*)" file content type$`: s.importScheduleReflectsInTheDBBasedOnFileContentType,
		`^there is no existing import schedule$`:                                    s.thereIsNoExistingImportSchedule,
		`^receives "([^"]*)" import error$`:                                         s.receivesImportError,
		`^there is an existing import schedule$`:                                    s.thereIsAnExistingImportSchedule,
		`^error list is correct$`:                                                   s.errorListIsCorrect,
		`^imported invoice schedules are converted in "([^"]*)"$`:                   s.importedInvoiceSchedulesAreConvertedIn,
		`^the scheduled date is one day ahead of invoice_date$`:                     s.scheduledDateIsOneDayAheadOfInvoiceDate,

		// Kafka sync discount
		`^a discount record is inserted in fatima$`:              s.aDiscountRecordIsInsertedInFatima,
		`^this discount record must be recorded in invoicemgmt$`: s.thisDiscountRecordMustBeRecordedInInvoicemgmt,

		// Cronjob Invoice Scheduled
		`^cronjob run InvoiceScheduleChecker endpoint today$`: s.cronjobRunFuncImportInvoiceChecker,

		// Create Payment Request
		`there are "([^"]*)" existing "([^"]*)" payments with payment method "([^"]*)"`:         s.thereAreExistingPayments,
		`partner has existing convenience store master record`:                                  s.partnerHasExistingConvenienceStore,
		`admin is at create payment request modal`:                                              s.adminIsAtCreatePaymentRequestModal,
		`admin chooses "([^"]*)" as payment method`:                                             s.adminChoosesAsPaymentMethod,
		`admin adds payment due date from at day "([^"]*)" and due date until at day "([^"]*)"`: s.adminAddsPaymentDueDateFromAndPaymentDueDateUntil,
		`admin clicks save create payment request`:                                              s.adminClicksSaveCreatePaymentRequest,
		`there are "([^"]*)" payment file with correct file name saved on database`:             s.thereArePaymentFileWithCorrectFileNameSavedOnDatabase,
		`the payments are associated to a payment request file`:                                 s.thePaymentsAreAssociatedToAPaymentRequestFile,
		`^this partner bank record limit is "([^"]*)"$`:                                         s.thisPartnerBankRecordLimitIs,

		`there are "([^"]*)" payment file associated to a payment request`: s.thereArePaymentFileAssociatedToAPaymentRequest,
		`the payments and invoices isExported field was set to "([^"]*)"`:  s.thePaymentsAndInvoicesIsExportedFieldWasSetTo,
		`a payment is already exported`:                                    s.aPaymentIsAlreadyExported,
		`^students has payment detail and billing address$`:                s.studentHasPaymentDetailAndBillingAddress,
		`^there is an existing bank mapped to partner bank$`:               s.thereIsAnExistingBankMappedToPartnerBank,
		`^students has payment and bank account detail$`:                   s.studentsHasPaymentAndBankAccountDetail,
		`^admin adds payment due date at day "([^"]*)"$`:                   s.adminAddsPaymentDueDate,
		`^there are banks mapped to different partner bank$`:               s.thereAreBanksMappedToDifferentPartnerBank,
		`^students have bank account in either of the banks$`:              s.studentsHaveBankAccountInEitherOfTheBanks,
		`^students have new customer code history record$`:                 s.studentsHaveNewCustomerCodeHistoryRecord,
		`^the invoices have invoice adjustment with amount "([^"]*)"$`:     s.theInvoicesHaveInvoiceAdjustmentWithAmount,

		// Import Partner Bank Master
		`^a request payload file with "([^"]*)" partner bank "([^"]*)" record$`: s.aRequestPayloadFileWithPartnerBankRecord,
		`^imports partner bank "([^"]*)" records$`:                              s.importsPartnerBankRecords,
		`^imports invalid partner bank records$`:                                s.importsInvalidPartnerBankRecords,
		`^archives "([^"]*)" partner bank records$`:                             s.archivesPartnerBankRecords,
		`^partner bank csv is imported successfully$`:                           s.partnerBankCsvIsImportedSuccessfully,
		`^partner bank csv is archived successfully$`:                           s.partnerBankCsvIsArchivedSuccessfully,
		`^partner bank csv "([^"]*)" record is imported unsuccessfully$`:        s.partnerBankCsvIsImportedUnsuccessfully,

		// Download Payment File
		`these payments already belong to payment request file with payment method "([^"]*)"`: s.thesePaymentsAlreadyBelongToPaymentRequestFileWithPaymentMethod,
		`admin is at create payment request table`:                                            s.adminIsAtCreatePaymentRequestTable,
		`admin select and downloads the payment request file`:                                 s.adminSelectAndDownloadsThePaymentRequestFile,
		`the data byte returned is not empty`:                                                 s.theDataByteReturnedIsNotEmpty,
		`the payment request file has a correct CSV format`:                                   s.thePaymentRequestFileHasACorrectCSVFormat,
		`partner has existing partner bank`:                                                   s.partnerHasExistingPartnerBank,
		`the payment request file has a correct bank TXT format`:                              s.thePaymentRequestFileHasACorrectBankTXTFormat,
		`send a download file request with empty file ID`:                                     s.sendADownloadFileRequestWithEmptyFileID,
		`there is a payment file that has no associated payments`:                             s.thereIsAPaymentFileThatHasNoAssociatedPayments,

		// Download Payment Validation File
		`^there is existing bulk payment validation record for "([^"]*)"$`:  s.thereIsExistingBulkPaymentValidationRecordFor,
		`^this record consists of "([^"]*)" payment "([^"]*)" validated$`:   s.thisRecordConsistsOfPaymentValidated,
		`^another "([^"]*)" payment "([^"]*)" validated$`:                   s.thisRecordConsistsOfPaymentValidated,
		`^admin is at payment validation screen$`:                           s.adminIsAtPaymentValidationScreen,
		`^selects the existing bulk payment validation record to download$`: s.selectsTheExistingBulkPaymentValidationRecordToDownload,
		`^has response payment data with "([^"]*)" correct records$`:        s.hasResponsePaymentDataWithCorrectRecords,
		`^has response validation date$`:                                    s.hasResponseValidationDate,

		// Kafka sync prefecture
		`^a prefecture record is inserted in bob$`:                 s.aPrefectureRecordIsInsertedInBob,
		`^this prefecture record must be recorded in invoicemgmt$`: s.thisPrefectureRecordMustBeRecordedInInvoicemgmt,

		// Bulk payment validation
		`^there are "([^"]*)" preexisting number of existing invoices with "([^"]*)" status$`:                    s.thereArePreexistingNumberOfExistingInvoicesWithStatus,
		`^receives expected result with correct DB records based on "([^"]*)" file content type$`:                s.receivesExpectedResultWithCorrectDBRecordsBasedOnFileContentType,
		`^"([^"]*)" signed-in user uploads the payment file for "([^"]*)" payment method$`:                       s.signedinUserUploadsThePaymentFileForPaymentMethod,
		`^there are existing payments for those invoices for "([^"]*)" payment method with "([^"]*)" status$`:    s.thereAreExistingPaymentsForThoseInvoicesForPaymentMethodWithStatus,
		`^there is an existing payment file for "([^"]*)" payment method$`:                                       s.thereIsAnExistingPaymentFileForPaymentMethod,
		`^has "([^"]*)" file content type with payment date "([^"]*)" for successful payments$`:                  s.hasFileContentTypeWithTransferredDateForSuccessfulPayments,
		`^has duplicate payment records with "([^"]*)" date and "([^"]*)" result code sequence on payment file$`: s.hasDuplicatePaymentRecordsWithDateAndResultCodeSequenceOnPaymentFile,
		`^receives expected record for duplicate payment with "([^"]*)" result code$`:                            s.receivesExpectedRecordForDuplicatePaymentWithResultCode,
		`^these existing payments have existing result code "([^"]*)"$`:                                          s.theseExistingPaymentsHaveExistingResultCode,

		// Upsert student payment info
		`^an existing student with student payment "([^"]*)" info$`:                           s.anExistingStudentWithBillingOrBankAccountInfo,
		`^"([^"]*)" create a billing information that "([^"]*)" for existing student$`:        s.createABillingInformationThatForExistingStudent,
		`^"([^"]*)" create a bank account that "([^"]*)" for existing student$`:               s.createABankAccountThatForExistingStudent,
		`^"([^"]*)" update with new billing information that "([^"]*)" for existing student$`: s.updateWithNewBillingInformationThatForExistingStudent,
		`^"([^"]*)" update with new bank account that "([^"]*)" for existing student$`:        s.updateWithNewBankAccountThatForExistingStudent,
		`^"([^"]*)" billing information$`:                                                     s.billingInformation,
		`^"([^"]*)" bank account`:                                                             s.bankAccount,
		`^this student bank account is verified$`:                                             s.thisStudentBankAccountIsVerified,
		`^this student billing address was removed$`:                                          s.thisStudentBillingAddressWasRemoved,
		`^this student payment method was removed$`:                                           s.thisStudentPaymentMethodWasRemoved,
		`^the student default payment method was set to "([^"]*)"$`:                           s.theStudentDefaultPaymentMethodWasSetTo,

		// Kafka sync order
		`^an order record is inserted into fatima$`:           s.anOrderRecordIsInsertedIntoFatima,
		`^this order record must be recorded in invoicemgmt$`: s.thisOrderRecordMustBeRecordedInInvoicemgmt,

		// Export Invoice Schedule
		`^the organization "([^"]*)" has existing "([^"]*)" import invoice schedules in "([^"]*)"$`:        s.theOrganizationHasExistingImportInvoiceSchedulesIn,
		`^the organization "([^"]*)" has no existing invoice schedule$`:                                    s.theOrganizationHasNoExistingInvoiceSchedule,
		`^admin export the invoice schedule data$`:                                                         s.adminExportTheInvoiceScheduleData,
		`^the invoice schedule CSV has a correct content with invoice date in default timezone "([^"]*)"$`: s.theInvoiceScheduleCSVHasACorrectContent,
		`^the invoice schedule CSV only contains the header record$`:                                       s.theInvoiceScheduleCSVOnlyContainsHeaderRecord,

		// Export Bank
		`^the organization "([^"]*)" has existing "([^"]*)" bank data$`: s.theOrganizationHasExistingBankData,
		`^the organization "([^"]*)" has no existing bank$`:             s.theOrganizationHasNoExistingBank,
		`^admin export the bank data$`:                                  s.adminExportTheBankData,
		`^the bank CSV has a correct content$`:                          s.theBankCSVHasACorrectContent,
		`^the bank CSV only contains the header record$`:                s.theBankCSVOnlyContainsHeaderRecord,

		// Export Bank Mapping
		`^the organization "([^"]*)" has existing bank mappings$`:   s.theOrganizationHasExistingBankMappings,
		`^the organization "([^"]*)" has no existing bank mapping$`: s.theOrganizationHasNoExistingBankMapping,
		`^admin export the bank mapping data$`:                      s.adminExportsTheBankMappingData,
		`^the bank mapping CSV has a correct content$`:              s.theBankMappingCSVHasACorrectContent,
		`^the bank mapping CSV only contains the header record$`:    s.theBankMappingCSVOnlyContainsTheHeaderRecord,

		// Export Bank Branch
		`^the organization "([^"]*)" has existing "([^"]*)" bank branch data$`: s.theOrganizationHasExistingBankBranchData,
		`^the user export bank branch data$`:                                   s.theUserExportBankBranchData,
		`^the bank branch CSV has a correct content$`:                          s.theBankBranchCSVHasCorrectContent,
		`^the organization "([^"]*)" has no existing bank branch data$`:        s.theOrganizationHasNoExistingBankBranch,
		`^the bank branch CSV only contains the header record$`:                s.theBankBranchCSVOnlyContainsTheHeaderRecord,

		// Create Invoice From Order
		`^there are "([^"]*)" existing students$`: s.thereAreExistingStudents,
		`^each of these students have "([^"]*)" orders with status "([^"]*)" and "([^"]*)" review required tag$`: s.eachOfTheseStudentsHaveOrdersWithStatusAndReviewRequiredTag,
		`^each of these orders have "([^"]*)" bill items with status "([^"]*)"$`:                                 s.eachOfTheseOrdersHaveBillItemWithStatus,
		`^admin selects the order list$`:                                        s.adminSelectsTheOrderList,
		`^submits the create invoice from order request$`:                       s.submitsTheCreateInvoiceFromOrderRequest,
		`^each invoice have "([^"]*)" bill items$`:                              s.eachInvoiceHaveBillItems,
		`^these bill items have "([^"]*)" billing status$`:                      s.theseBillItemsHaveBillingStatus,
		`^"([^"]*)" billing items of order have billing date at day "([^"]*)"$`: s.billingItemsOfOrderHaveBillingDateAtDay,
		`^there is invoice date scheduled at day "([^"]*)"$`:                    s.thereIsInvoiceDateScheduledAtDay,
		`^each invoice has correct total amount and outstanding balance$`:       s.eachInvoiceHasCorrectTotalAmountAndOutstandingBalance,
		`^"([^"]*)" billing items of order have adjustment price "([^"]*)"$`:    s.billingItemsOfOrderHaveAdjustmentPrice,

		// Update Student Payment Method
		`^there is an existing student that "([^"]*)"$`:                 s.theOrganizationHasExistingBankBranchData,
		`^student bank account is set to "([^"]*)" status$`:             s.studentBankAccountIsSetToStatus,
		`^student payment detail has "([^"]*)" payment method$`:         s.studentPaymentDetailHasPaymentMethod,
		`^updates "([^"]*)" payment method of the student$`:             s.updatesPaymentMethodOfTheStudent,
		`^student payment method is updated successfully to "([^"]*)"$`: s.studentPaymentMethodIsUpdatedSuccesfullyTo,

		// Retrieve Student Payment Method
		`^an existing student with default payment method "([^"]*)"$`:            s.anExistingStudentWithDefaultPaymentMethod,
		`^the RetrieveStudentPaymentMethod endpoint is called for this student$`: s.theRetrieveStudentPaymentMethodEndpointIsCalledForThisStudent,
		`^"([^"]*)" payment method for this student is retrieve successfully`:    s.paymentMethodForThisStudentIsRetrieveSuccessfully,
		`^an existing student with no student payment method$`:                   s.anExistingStudentWithNoStudentPaymentMethod,
		`^empty payment method for this student is retrieve successfully$`:       s.emptyPaymentMethodForThisStudentIsretrieveSuccessfully,
		`^a non existing student record$`:                                        s.aNonExistingStudentRecord,

		// Kafka sync user basic info
		`^a user basic info record is inserted in bob$`:                 s.aUserBasicInfoRecordIsInsertedInBob,
		`^this user basic info record must be recorded in invoicemgmt$`: s.thisUserBasicInfoRecordMustBeRecordedInInvoicemgmt,

		// Issue invoice V2
		`^"([^"]*)" issues invoice using v2 endpoint$`:                                                             s.issueInvoiceUsingV2Endpoint,
		`^this invoice has "([^"]*)" payment with "([^"]*)" status$`:                                               s.thisInvoiceHasPaymentWithStatus,
		`^"([^"]*)" issues invoice using v2 endpoint with "([^"]*)" payment method and dates "([^"]*)" "([^"]*)"$`: s.issuesInvoiceUsingV2EndpointWithPaymentMethodAndDates,

		// Bulk Issue Invoice V2
		`^"([^"]*)" bulk issue invoices using v2 endpoint with payment method "([^"]*)" and "([^"]*)" "([^"]*)"$`: s.bulkIssueInvoicesUsingV2EndpointWithPaymentMethod,
		`^invoices has "([^"]*)" payment with "([^"]*)" status$`:                                                  s.invoicesHasPaymentWithStatus,
		`^one invoice has zero total amount$`:                                                                     s.oneInvoiceHasZeroTotalAmount,

		// Invoice Adjustment
		`^this invoice has "([^"]*)" total amount$`:                                                       s.thisInvoiceHasTotalAmount,
		`^adds "([^"]*)" invoice adjustment with "([^"]*)" amount$`:                                       s.addsInvoiceAdjustmentWithAmount,
		`^edits "([^"]*)" existing invoice adjustment with "([^"]*)" amount updated to "([^"]*)" amount$`: s.editsExistingInvoiceAdjustmentThatHasAmountUpdatedToAmount,
		`^apply the adjustment on the invoice$`:                                                           s.applyTheAdjustmentOnTheInvoice,
		`^invoice total subtotal and outstanding balance are correctly updated to "([^"]*)" amount$`:      s.invoiceTotalSubtotalAndOutstandingBalanceAreCorrectlyUpdatedToAmount,
		`^there are "([^"]*)" created invoice adjustment with "([^"]*)" amount$`:                          s.thereAreCreatedInvoiceAdjustmentWithAmount,
		`^deletes "([^"]*)" existing invoice adjustment with "([^"]*)" amount$`:                           s.deletesExistingInvoiceAdjustmentThatHasAmount,

		// Download payment file from cloud storage
		`^there is an existing payment file in cloud storage$`:                                         s.thereIsAnExistingPaymentFileInCloudStorage,
		`^there is a payment request file with payment method "([^"]*)" that is not in cloud storage$`: s.thereIsAPaymentRequestFileWithPaymentMethodThatIsNotInCloudStorage,
		`^"([^"]*)" billing items of students are adjustment billing type$`:                            s.billingItemsOfStudentsAreAdjustmentBillingType,

		// Unleash toggle
		`^unleash feature flag is "([^"]*)" with feature name "([^"]*)"$`: s.UnleashSuite.ToggleUnleashFeatureWithName,

		// Create Payment Request With Gcloud
		`^these "([^"]*)" payment file are saved and uploaded successfully$`: s.thesePaymentFileAreSavedAndUploadedSuccessfully,
		`^these "([^"]*)" payments belong to a bulk payment$`:                s.thesePaymentsBelongsToABulkPayment,
		`^"([^"]*)" bulk payment record is updated to exported status$`:      s.thereIsBulkPaymentStatusUpdatedToExported,

		// Upload Existing Payment Request File
		`^admin is logged-in back office on organization "([^"]*)"$`:               s.adminIsLoggedInBackOfficeOnOrganization,
		`^these payments belong to old payment request files$`:                     s.thesePaymentsBelongToOldPaymentRequestFiles,
		`^an admin runs the upload payment request file job script$`:               s.anAdminRunsTheUploadPaymentRequestFileJobScript,
		`^the file_url of payment files is not empty$`:                             s.theFileURLofPaymentFileIsNotEmpty,
		`^these payment files are uploaded successfully$`:                          s.thesePaymentFilesAreUploadedSuccessfully,
		`^the payment request files have a correct format$`:                        s.thePaymentRequestFilesHaveACorrectFormat,
		`^there is "([^"]*)" payment request file that has no associated payment$`: s.thereIsPaymentRequestFileThatHasNoAssociatedPayment,
		`^the upload payment request file script returns an error$`:                s.theUploadPaymentRequestFileScriptReturnsError,
		`^the upload payment request file script has no error$`:                    s.theUploadPaymentRequestFileScriptHasNoError,

		// Cancel Invoice Payment V2
		`^admin cancels an invoice with "([^"]*)" remarks using v2 endpoint$`:                     s.adminCancelsAnInvoiceWithRemarksUsingV2Endpoint,
		`^invoice remains "([^"]*)" invoice status$`:                                              s.invoiceHasInvoiceStatus,
		`^action log record is recorded with "([^"]*)" action and "([^"]*)" remarks$`:             s.actionLogRecordIsRecordedWithActionAndRemarks,
		`^this payment has exported status "([^"]*)"$`:                                            s.thisPaymentHasExportedStatus,
		`^there is "([^"]*)" payment history with "([^"]*)" payment method$`:                      s.thereIsPaymentHistoryWithPaymentMethod,
		`^belongs in bulk with "([^"]*)" other "([^"]*)" payments with "([^"]*)" payment method$`: s.belongsInBulkWithOtherPaymentsWithStatus,
		`^bulk payment record has "([^"]*)" status$`:                                              s.bulkPaymentRecordHasStatus,

		// Add Payment
		`^admin adds payment to invoice$`:                                                  s.adminAddsPaymentToInvoice,
		`^sets payment method to "([^"]*)" in add payment request$`:                        s.setsPaymentMethodToInAddPaymentRequest,
		`^sets due date to "([^"]*)" and expiry date to "([^"]*)" in add payment request$`: s.setsDueDateToAndExpiryDateToInAddPaymentRequest,
		`^sets amount same with invoice outstanding balance in add payment request$`:       s.setsAmountSameWithInvoiceOutstandingBalance,
		`^admin submits the add payment form with remarks "([^"]*)"$`:                      s.adminSubmitsTheAddPaymentFormWithRemarks,
		`^this student bank account is not verified$`:                                      s.thisStudentBankAccountIsNotVerified,
		`^this student has payment and bank account detail$`:                               s.thisStudentHasPaymentAndBankAccountDetail,

		// Approve invoice payment v2
		`^there is an existing invoice with "([^"]*)" status$`:                              s.thereIsAnExistingInvoiceWithStatus,
		`^admin submits the approve payment form with remarks "([^"]*)" using v2 endpoint$`: s.adminSubmitsTheApprovePaymentFormWithRemarksUsingV2Endpoint,
		`^admin already requested payment with amount same on invoice outstanding balance$`: s.adminAlreadyRequestedPaymentWithAmountSameOnInvoiceOutstandingBalance,
		`^admin sets the approve payment form with "([^"]*)" payment date$`:                 s.adminSetsTheApprovePaymentFormWithPaymentDate,
		`^this payment has payment "([^"]*)" payment method$`:                               s.thisPaymentHasPaymentMethod,
		`^invoice amount paid is equal to payment amount$`:                                  s.invoiceAmountPaidIsEqualToPaymentAmount,
		`^invoice has zero outstanding balance$`:                                            s.invoiceHasZeroOutstandingBalance,
		`^admin added the requested payment on the invoice$`:                                s.addedTheRequestedPaymentOnTheInvoice,
		`^latest payment record has receipt date today$`:                                    s.latestPaymentRecordHasReceiptDateToday,

		// Refund Invoice
		`^admin refunds an invoice$`:                                                    s.adminRefundsAnInvoice,
		`^sets refund method "([^"]*)" in refund invoice request$`:                      s.setsRefundMethodInRefundMethodRequest,
		`^sets amount same with invoice outstanding balance in refund invoice request$`: s.setsAmountSameWithInvoiceOutstandingBalanceInRefundMethodRequest,
		`^admin submits the refund invoice form with remarks "([^"]*)"$`:                s.adminSubmitsTheRefundInvoiceFormWithRemarks,
		`^sets amount to "([^"]*)" in refund invoice request$`:                          s.setsAmountToInRefundInvoiceRequest,

		// Payment Data Migration
		`^there are "([^"]*)" migrated invoices with "([^"]*)" status$`:                  s.thereAreMigratedInvoicesWithStatus,
		`^this payment csv file has payment data with "([^"]*)" payment status$`:         s.thisPaymentCsvFileHasPaymentDataWithPaymentStatus,
		`^imports the payment csv file$`:                                                 s.importsThePaymentCsvFile,
		`^payment csv file is imported successfully$`:                                    s.paymentCsvFileIsImportedSuccessfully,
		`^there are payment records with correct invoice created successfully$`:          s.thereArePaymentRecordsWithCorrectInvoiceCreatedSuccessfully,
		`^there is a payment csv file with "([^"]*)" payment method for these invoices$`: s.thereIsAPaymentCsvFileWithPaymentMethodForTheseInvoices,
		`^payment csv file is imported unsuccessfully$`:                                  s.paymentCsvFileIsImportedUnsuccessfully,
		`^response has error on "([^"]*)" payment status that should be "([^"]*)"$`:      s.responseHasErrorOnPaymentStatusThatShouldBe,
		`^payment csv file contains invalid students$`:                                   s.paymentCsvFileContainsInvalidStudents,
		`^response has error for invalid payment student$`:                               s.responseHasErrorForInvalidPaymentStudent,

		// Invoice Data Migration
		`^there is invoice CSV file for these students$`:                                                    s.thereIsInvoiceCSVFileForTheseStudents,
		`^admin imports invoice migration data$`:                                                            s.adminImportsInvoiceMigrationData,
		`^there are "([^"]*)" invoices of students migrated successfully$`:                                  s.thereAreInvoicesOfStudentsMigratedSuccessfully,
		`^migrated invoices have correct amount based on its status$`:                                       s.migratedInvoicesHaveCorrectAmountBasedOnItsStatus,
		`^migrated invoice have saved reference number and migrated_at$`:                                    s.migratedInvoicesHaveSavedReferenceNumberAndMigratedAt,
		`^there are no error lines in import invoice response$`:                                             s.thereAreNoErrorLinesInImportInvoiceResponse,
		`^there is invoice CSV file for non existing students$`:                                             s.thereIsInvoiceCSVFileForNonExistingStudents,
		`^there are error lines in import invoice response$`:                                                s.thereAreErrorLinesInImportInvoiceResponse,
		`^there is invoice CSV file for these students with invalid amount$`:                                s.thereIsInvoiceCSVFileForTheseStudentsWithInvalidAmount,
		`^there are "([^"]*)" students that have "([^"]*)" bill items migrated with "([^"]*)" total price$`: s.thereAreStudentsThatHaveBillItemsMigratedWithTotalPrice,

		// Migrate Invoice Bill Item
		`^there are "([^"]*)" existing invoice and bill_item that have the same invoice reference$`: s.thereAreExistingInvoiceAndBillItemThatHaveSameReference,
		`^an admin runs the migrate invoice bill item job script$`:                                  s.adminRunsTheMigrateInvoiceBillItemScript,
		`^the migrate invoice bill item script has no error$`:                                       s.migrateInvoiceBillItemScriptHasNoError,
		`^the invoice bill items were successfully migrated$`:                                       s.invoiceBillItemsWereSuccessfullyMigrated,
		`^the migrated invoice bill item have the same reference$`:                                  s.migratedInvoiceBillItemHaveTheSameReference,
		`^an admin runs the migrate invoice bill item job script with "([^"]*)"$`:                   s.adminRunsTheMigrateInvoiceBillItemScriptWith,
		`^the migrate invoice bill item script returns error$`:                                      s.migrateInvoiceBillItemScriptReturnsError,

		// Upsert student payment info with action log changes
		`^request "([^"]*)" updates on student payment info with "([^"]*)" information$`:                             s.requestUpdatesOnStudentPaymentInfoWithInformation,
		`^admin updates the student payment information$`:                                                            s.adminUpdatesTheStudentPaymentInformation,
		`^student payment information updated successfully with "([^"]*)" student payment detail action log record$`: s.studentPaymentInformationUpdatedSuccessfullyWithActionLogRecord,
		`^no student payment detail action log recorded$`:                                                            s.noStudentPaymentDetailActionLogRecorded,
		`^admin updates student payment "([^"]*)" info with same information$`:                                       s.adminUpdatesStudentPaymentInformationWithSameInformation,

		// Retrieve Bulk Student Payment Method
		`^there are existing "([^"]*)" students with "([^"]*)" default payment method$`: s.thereAreExistingStudentsWithDefaultPaymentMethod,
		`^the RetrieveBulkStudentPaymentMethod endpoint is called for these students$`:  s.theRetrieveBulkStudentPaymentMethodEndpointIsCalledForThisStudent,
		`^payment methods for these students are retrieve successfully$`:                s.paymentMethodsForTheseStudentsAreRetrieveSuccessfully,

		// Auto set convenience store payment method
		`^there is an event create student request with user address info$`: s.thereIsAnEventCreateStudentRequestWithUserAddressInfo,
		`^yasuo send the create student event request$`:                     s.yasuoSendTheCreateStudentEventRequest,
		`^student payment detail record is successfully created$`:           s.studentPaymentDetailRecordIsSuccessfullyCreated,
		`^student billing address record is successfully created$`:          s.studentBillingAddressRecordIsSuccessfullyCreated,
		`^no student billing address record created$`:                       s.noStudentBillingAddressRecordCreated,
		`^no student payment detail record created$`:                        s.noStudentPaymentDetailRecordCreated,
		`^billing address is the same as user address$`:                     s.billingAddressIsTheSameAsUserAddress,
		`^invoicemgmt internal config "([^"]*)" is "([^"]*)"$`:              s.invoicemgmtInternalConfigIs,

		// Bulk Add Payment
		`^"([^"]*)" bulk add payment for these invoices with payment method "([^"]*)" and "([^"]*)" "([^"]*)"$`: s.bulkAddPaymentForTheseInvoicesWithPaymentMethod,
		`^there are no payments for these invoices$`:                                                            s.thereAreNoPaymentsForTheseInvoices,
		`^these invoices has "([^"]*)" type$`:                                                                   s.theseInvoicesHasType,
		`^another "([^"]*)" preexisting number of existing invoices with "([^"]*)" status$`:                     s.thereArePreexistingNumberOfExistingInvoicesWithStatus,
		`^bulk payment record is created successfully with payment method "([^"]*)"$`:                           s.bulkPaymentRecordIsCreatedSuccessfullyWithPaymentMethod,

		// Update Billing Address Info In Update Student Event Message
		`^there is an existing student with "([^"]*)" billing address and "([^"]*)" payment detail$`: s.thereIsAnExistingStudentWithBillingAddressAndPaymentDetail,
		`^yasuo send the update student event request with "([^"]*)"$`:                               s.yasuoSendTheUpdateStudentEventWith,
		`^student payer name successfully updated$`:                                                  s.studentPayerNameSuccessfullyUpdated,
		`^student billing address record is successfully updated$`:                                   s.studentBillingAddressRecordSuccessfullyUpdated,
		`^student payer name is not updated$`:                                                        s.studentPayerNameIsNotUpdated,
		`^student default payment method including payer name is successfully removed$`:              s.studentPaymentMethodIncludingPayerNameSuccessfullyRemoved,
		`^student billing address record is successfully removed$`:                                   s.studentBillingAddressRecordSuccessfullyRemoved,
		`^the default payment method of this student is "([^"]*)"$`:                                  s.theDefaultPaymentMethodOfThisStudentIs,
		`^this student has bank account with verification "([^"]*)" status$`:                         s.thisStudentHasBankAccountWithVerificationStatus,

		// Bank OpenAPI
		`^this student is included on bank OpenAPI valid payload$`:             s.thisStudentIsIncludedOnBankOpenAPIPayload,
		`^there are existing bank and bank branch$`:                            s.thereAreExistingBankAndBankBranch,
		`^bank info of the student was upserted successfully by OpenAPI$`:      s.bankInfoOfTheStudentWasUpsertedSuccessfullyByOpenAPI,
		`^admin submits the bank OpenAPI payload$`:                             s.adminSubmitsTheBankOpenAPIPayload,
		`^receives failed "([^"]*)" response code from OpenAPI$`:               s.receivesFailedResponseCodeFromOpenAPI,
		`^this student is included on bank OpenAPI invalid "([^"]*)" payload$`: s.studentIsIncludedOnBankOpenAPIInvalidPayload,
		`^admin already setup an api user$`:                                    s.adminAlreadySetupAnAPIUser,

		// Bulk Payment Validate V2
		`^has result code category "([^"]*)" on its file content$`:                                s.hasResultCodeCategoryOnItsFileContent,
		`^has "([^"]*)" payment date on its file content$`:                                        s.hasPaymentDateOnItsFileContent,
		`^there are "([^"]*)" number of existing "([^"]*)" invoices with total "([^"]*)" amount$`: s.thereAreNumberOfExistingInvoicesWithTotalAmount,
		`^payments have "([^"]*)" result code with correct expected result$`:                      s.paymentsHaveResultCodeWithCorrectExpectedResult,
		`^has amount mismatched on its file content$`:                                             s.hasAmountMismatchedOnItsFileContent,
		`^there is a payment that is not match in our system$`:                                    s.thereIsAPaymentThatIsNotMatchInOurSystem,
		`^has "([^"]*)" payment date on the request$`:                                             s.hasPaymentDateOnItsFileContent,

		// Bulk Cancel Payment
		`^admin cancel the bulk payment$`:                                                  s.adminCancelTheBulkPayment,
		`^bulk payment record status is updated to "([^"]*)"$`:                             s.bulkPaymentRecordStatusIsUpdatedTo,
		`^each payments has "([^"]*)" payment status$`:                                     s.eachPaymentsHasPaymentStatus,
		`^this bulk payment has status "([^"]*)"$`:                                         s.thisBulkPaymentHasStatus,
		`^only pending payments were updated to "([^"]*)" payment status$`:                 s.onlyPendingPaymentsWereUpdatedToPaymentStatus,
		`^only pending payments invoice are recorded with "([^"]*)" action log type$`:      s.onlyPendingPaymentsInvoiceAreRecordedWithActionLogType,
		`^no invoice action log with "([^"]*)" action log type recorded for each invoice$`: s.noInvoiceActionLogWithActionLogTypeRecordedForEachInvoice,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
