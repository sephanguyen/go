package discount

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
		// Max discount selection
		`^prepare data for max discount selection with "([^"]*)"$`: s.prepareDataForMaxDiscountSelection,
		`^"([^"]*)" added "([^"]*)" to student$`:                   s.tagIsAddedToStudent,
		`^system selects max discount for student$`:                s.systemSelectsMaxDiscount,
		`^event is received for update product$`:                   s.eventIsReceivedForUpdateProduct,

		// Import discount tag
		`^an discount tag valid request payload with "([^"]*)"$`:                     s.anDiscountTagValidRequestPayloadWith,
		`^"([^"]*)" importing discount tag$`:                                         s.importingDiscountTag,
		`^the invalid discount tag lines are returned with error$`:                   s.theInvalidDiscountTagLinesAreReturnedWithError,
		`^an discount tag valid request payload with incorrect data with "([^"]*)"$`: s.anDiscountTagValidRequestPayloadWithIncorrectData,
		`^the import discount tag transaction is rolled back$`:                       s.theImportDiscountTagTransactionIsRolledBack,
		`^the valid discount tag lines are imported successfully$`:                   s.theValidDiscountTagLinesAreImportedSuccessfully,
		`^an discount tag invalid "([^"]*)" request payload$`:                        s.anDiscountTagInvalidRequestPayload,
		`^receives "([^"]*)" status code$`:                                           s.receivesStatusCode,

		// Retrieve Active Student Discount Tag
		`^there is an existing discount master data with discount tag$`:                  s.thereIsAnExistingDiscountMasterDataWithDiscountTag,
		`^there is a student that has "([^"]*)" user discount tag "([^"]*)" records$`:    s.thereIsAStudentThatHasUserDiscountTagRecords,
		`^"([^"]*)" retrieves user discount tag for this student$`:                       s.retrievesUserDiscountTagForThisStudent,
		`^this user discount tag has "([^"]*)" start date and "([^"]*)" end date$`:       s.thisUserDiscountTagHasStartDateAndEndDate,
		`^user discount tag "([^"]*)" records are retrieved successfully$`:               s.userDiscountTagRecordsAreRetrievedSuccessfully,
		`^there is no user discount tag records retrieved$`:                              s.thereIsNoUserDiscountTagRecordsRetrieved,
		`^there is a non existing student record$`:                                       s.thereIsANonExistingStudentRecord,
		`^a valid request payload for retrieve user discount tag with date today$`:       s.aValidRequestPayloadForRetrieveUserDiscountTagWithDateToday,
		`^there is an invalid "([^"]*)" payload request for retrieve user discount tag$`: s.thereIsAnInvalidPayloadRequestForRetrieveUserDiscountTag,

		// Upsert Student Discount Tag
		`^there is an existing student with active products$`:                          s.thereIsAnExistingStudentWithActiveProducts,
		`^this student has "([^"]*)" user discount tag "([^"]*)" records$`:             s.thisStudentHasUserDiscountTagRecords,
		`^a request payload for upsert user discount tag$`:                             s.aRequestPayloadForUpsertUserDiscountTag,
		`^upserts user discount tag "([^"]*)" records for this student$`:               s.upsertsUserDiscountTagRecordsForThisStudent,
		`^apply the upsert discount tags on the student$`:                              s.applyTheUpsertDiscountTagsOnTheStudent,
		`^"([^"]*)" logins to backoffice app$`:                                         s.loginsToBackofficeApp,
		`^this student has correct user discount tag "([^"]*)" records$`:               s.thisStudentHasCorrectUserDiscountTagRecords,
		`^this student has no user discount tag records$`:                              s.thisStudentHasNoUserDiscountTagRecords,
		`^there is an invalid "([^"]*)" payload request for upsert user discount tag$`: s.thereIsAnInvalidPayloadRequestForUpsertUserDiscountTag,

		// Import product group
		`^a product group valid request payload with correct "([^"]*)" data$`:           s.productGroupValidRequestPayloadWithCorrectData,
		`^"([^"]*)" importing product group$`:                                           s.importingProductGroup,
		`^the valid product group lines with "([^"]*)" data are imported successfully$`: s.theValidProductGroupLinesWithDataAreImportedSuccessfully,
		`^a product group valid request payload with incorrect "([^"]*)" data$`:         s.aProductGroupValidRequestPayloadWithIncorrectData,
		`^the import product group transaction is rolled back$`:                         s.theImportProductGroupTransactionIsRolledBack,
		`^a product group invalid "([^"]*)" request payload$`:                           s.aProductGroupInvalidRequestPayload,

		// Import product group mapping
		`^a product group mapping valid request payload with correct "([^"]*)" data$`:           s.productGroupMappingValidRequestPayloadWithCorrectData,
		`^"([^"]*)" importing product group mapping$`:                                           s.importingProductGroupMapping,
		`^the valid product group mapping lines with "([^"]*)" data are imported successfully$`: s.theValidProductGroupMappingLinesWithDataAreImportedSuccessfully,
		`^a product group mapping invalid request payload with incorrect "([^"]*)" data$`:       s.aProductGroupMappingInvalidRequestPayloadWithIncorrectData,
		`^the import product group mapping transaction is rolled back$`:                         s.theImportProductGroupMappingTransactionIsRolledBack,
		`^a product group mapping invalid "([^"]*)" request payload$`:                           s.aProductGroupMappingInvalidRequestPayload,

		`^prepare data for sibling discount automation for "([^"]*)"$`: s.prepareDataForSiblingDiscountAutomationForCase,
		`^service sent order info stream$`:                             s.serviceReceivedOrderInfoStream,
		`^student is "([^"]*)" for sibling discount automation$`:       s.studentTrackedForSiblingDiscount,

		// Import Package Discount Setting
		`^a package discount setting payload with "([^"]*)" data$`:                                 s.aPackageDiscountSettingPayloadWithData,
		`^"([^"]*)" importing package discount setting$`:                                           s.importingPackageDiscountSetting,
		`^the valid package discount setting lines with "([^"]*)" data are imported successfully$`: s.theValidPackageDiscountSettingLinesWithDataAreImportedSuccessfully,
		`^a package discount setting request payload with incorrect "([^"]*)" data$`:               s.aPackageDiscountSettingRequestPayloadWithIncorrectData,
		`^the import package discount setting transaction is rolled back$`:                         s.theImportPackageDiscountSettingTransactionIsRolledBack,
		`^a package discount setting invalid "([^"]*)" request payload$`:                           s.aPackageDiscountSettingInvalidRequestPayload,

		// Import Package Discount Course Mapping
		`^a package discount course mapping payload with "([^"]*)" data$`:                                 s.aPackageDiscountCourseMappingPayloadWithData,
		`^"([^"]*)" imports package discount course mapping$`:                                             s.importsPackageDiscountCourseMapping,
		`^the valid package discount course mapping lines with "([^"]*)" data are imported successfully$`: s.theValidPackageDiscountCourseMappingLinesWithDataAreImportedSuccessfully,
		`^a package discount course mapping request payload with incorrect "([^"]*)" data$`:               s.aPackageDiscountCourseMappingRequestPayloadWithIncorrectData,
		`^the import package discount course mapping transaction is rolled back$`:                         s.theImportPackageDiscountCourseMappingTransactionIsRolledBack,
		`^a package discount course mapping invalid "([^"]*)" request payload$`:                           s.aPackageDiscountCourseMappingInvalidRequestPayload,

		// Get export master data
		`^data of "([^"]*)" is existing$`:                s.addDataForExportMasterData,
		`^"([^"]*)" export "([^"]*)" data successfully$`: s.theUserExportMasterData,
		`^the "([^"]*)" CSV has a correct content$`:      s.theMasterDataCSVHasCorrectContent,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
