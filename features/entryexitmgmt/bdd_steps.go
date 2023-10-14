package entryexitmgmt

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
		`^everything is OK$`:                                            s.everythingIsOK,
		`^health check endpoint called$`:                                s.healthCheckEndpointCalled,
		`^entryexitmgmt should return "([^"]*)" with status "([^"]*)"$`: s.entryexitmgmtShouldReturnWithStatus,

		// Create Entry Exit Manual Record of Student
		`^"([^"]*)" creates "([^"]*)" record of this student in "([^"]*)"$`: s.createsRecordOfThisStudentIn,
		`^new entry exit record is created successfully$`:                   s.newEntryExitRecordIsCreatedSuccessfully,
		`^receives "([^"]*)" status code$`:                                  s.receivesStatusCode,
		`^"([^"]*)" creates invalid "([^"]*)" request$`:                     s.createsInvalidRequest,
		`^"([^"]*)" "([^"]*)" notify parents checkbox$`:                     s.notifyParentsCheckbox,

		// Delete Entry exit record of the student
		`^"([^"]*)" deletes that record of this student$`: s.deletesThatRecordOfThisStudent,

		// Event User
		`^a EvtUser with message "([^"]*)"$`: s.aEvtUserWithMessage,
		`^yasuo send event EvtUser$`:         s.yasuoSendEventEvtUser,
		`^student must have qrcode$`:         s.studentMustHaveQrcode,

		// Scan QR Code
		`^there is an existing student$`:                                                            s.thereIsAnExistingStudent,
		`^student has "([^"]*)" record$`:                                                            s.studentHasRecord,
		`^student parent has existing device$`:                                                      s.studentParentHasExistingDevice,
		`^student scans qrcode with "([^"]*)" date in time zone "([^"]*)"$`:                         s.studentScansQrcodeWithDateInTimeZone,
		`^student scans qrcode "([^"]*)" with "([^"]*)" date in time zone "([^"]*)"$`:               s.studentScansQrcodeRequestCountWithDateInTimeZone,
		`^student has no multiple record$`:                                                          s.studentHasNoMultipleRecord,
		`^student just scanned qrcode with "([^"]*)" date in time zone "([^"]*)"$`:                  s.studentScansQrcodeWithDateInTimeZone,
		`^student scans again$`:                                                                     s.studentScansAgain,
		`^scan returns "([^"]*)" status code$`:                                                      s.scanReturnsStatusCode,
		`^"([^"]*)" touch type is recorded$`:                                                        s.touchTypeIsRecorded,
		`^parent receives notification status "([^"]*)"$`:                                           s.parentReceivesNotificationStatus,
		`^student has "([^"]*)" parent$`:                                                            s.studentHasParent,
		`^name of the student is displayed on welcome screen$`:                                      s.nameOfTheStudentIsDisplayedOnWelcomeScreen,
		`^"([^"]*)" logins to backoffice app$`:                                                      s.loginsToBackofficeApp,
		`^student scans qrcode with "([^"]*)" date in time zone "([^"]*)" with invalid encryption$`: s.studentScansQrcodeWithDateInTimeZoneWithInvalidEncryption,

		// Update Entry Exit record of the student
		`^entry exit record is updated successfully$`:                             s.entryExitRecordIsUpdatedSuccessfully,
		`^"([^"]*)" updates the "([^"]*)" record of this student in "([^"]*)"$`:   s.updatesTheRecordOfThisStudentIn,
		`^"([^"]*)" updates the "([^"]*)" record with invalid "([^"]*)" request$`: s.updatesTheExistingRecordWithInvalidRequest,

		// Generate Batch QR Codes
		`^a qrcode request payload with "([^"]*)" student ids$`: s.aQrcodeRequestPayloadWithStudentIds,
		`^"([^"]*)" generates qrcode for these student ids$`:    s.generatesQrcodeForTheseStudentIds,
		`^response has no errors$`:                              s.responseHasNoErrors,
		`^student has "([^"]*)" qr version$`:                    s.studentHasQrVersion,
		`^student should have updated qrcode version$`:          s.studentShouldHaveUpdatedQrVersion,

		// Retrieve Entry Exit record of the student
		`^records found with default limit are displayed successfully$`: s.recordsFoundWithDefaultLimitAreDisplayedSuccessfully,
		`^parent is at the entry exit records screen$`:                  s.parentIsAtTheEntryExitScreen,
		`^"([^"]*)" logins Learner App$`:                                s.loginsLearnerApp,
		`^parent selects this existing student$`:                        s.parentSelectsThisExistingStudent,
		`^student has "([^"]*)" entry and exit "([^"]*)" record$`:       s.studentHasEntryAndExitRecord,
		`^parent checks the filter for records "([^"]*)"$`:              s.parentChecksTheFilterForRecords,
		`^parent scrolls down to display all records$`:                  s.parentScrollsDownToDisplayAllRecords,
		`^all records found are displayed successfully$`:                s.allRecordsFoundAreDisplayedSuccessfully,
		`^no records found displayed successfully$`:                     s.noRecordsFoundDisplayedSuccessfully,
		`^parent has another existing student$`:                         s.parentHasAnotherExistingStudent,
		`^this student has "([^"]*)" entry and exit record "([^"]*)"$`:  s.studentHasEntryAndExitRecord,
		`^this parent has an existing student$`:                         s.thisParentHasAnExistingStudent,

		// Retrieve QR Code of the student
		`^this student has "([^"]*)" qr code record$`:        s.thisStudentHasQrCodeRecord,
		`^student logins on Learner App$`:                    s.studentLoginsOnLearnerApp,
		`^student is at the My QR Code screen$`:              s.studentIsAtTheMyQRCodeScreen,
		`^student requested qr code with "([^"]*)" payload$`: s.studentRequestedQrCodeWithPayload,
		`^student qr code is displayed "([^"]*)"$`:           s.StudentQRCodeIsDisplayed,

		// Kafka Sync Tests
		`^a user record is inserted in bob$`:                                s.aUserRecordIsInsertedInBob,
		`^this user record must be recorded in entryexitmgmt$`:              s.thisUserRecordMustBeRecordedInEntryExitMgmt,
		`^a student record is inserted in bob$`:                             s.aStudentRecordIsInsertedInBob,
		`^this student record must be recorded in entryexitmgmt$`:           s.thisStudentRecordMustBeRecordedInEntryExitMgmt,
		`^a location record is inserted in bob$`:                            s.aLocationRecordIsInsertedInBob,
		`^this location record must be recorded in entryexitmgmt$`:          s.thisLocationRecordMustBeRecordedInEntryExitMgmt,
		`^a user access paths record is inserted in bob$`:                   s.aUserAccessPathRecordIsInsertedInBob,
		`^this user access paths record must be recorded in entryexitmgmt$`: s.thisUserAccessPathRecordedInEntryExitMgmt,
		`^a student_parent record is inserted in bob$`:                      s.aStudentParentsRecordIsInsertedInBob,
		`^this student_parent record must be recorded in entryexitmgmt$`:    s.thisStudentParentsRecordedInEntryExitMgmt,
		`^a user basic info record is inserted in bob$`:                     s.aUserBasicInfoRecordIsInsertedInBob,
		`^this user basic info record must be recorded in entryexitmgmt$`:   s.thisUserBasicInfoMustRecordedInEntryExitMgmt,
		`^a grade record is inserted in mastermgmt$`:                        s.aGradeRecordIsInsertedInMastermgmt,
		`^this grade record must be recorded in entryexitmgmt$`:             s.thisGradeRecordMustBeRecordedInEntryExitMgmt,

		// Unleash toggle
		`^unleash feature flag is "([^"]*)" with feature name "([^"]*)"$`: s.UnleashSuite.ToggleUnleashFeatureWithName,
		// Internal Config toggle
		`^entryexitmgmt internal config "([^"]*)" is "([^"]*)"$`: s.entryExitmgmtInternalConfigIs,
		`^student must have no qrcode$`:                          s.studentMustHaveNoQrcode,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
