package timesheet

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

func initSteps(ctx *godog.ScenarioContext, s *Suite) {
	steps := map[string]interface{}{
		// Health Check
		`^everything is OK$`:                                        s.everythingIsOK,
		`^health check endpoint called$`:                            s.healthCheckEndpointCalled,
		`^timesheet should return "([^"]*)" with status "([^"]*)"$`: s.timesheetShouldHealthCheckReturnWithStatus,
		`^returns "([^"]*)" status code$`:                           s.CommonSuite.ReturnsStatusCode,
		`^returns "([^"]*)" status code for "([^"]*)"$`:             s.CommonSuite.ReturnsStatusCodeWithExamInfo,

		// Signed in
		`^"([^"]*)" signin system$`: s.SignedAsAccount,

		// Create timesheet
		`^have timesheet configuration is on`:                                       s.initMasterMgmtConfiguration,
		`^new timesheet data for current staff`:                                     s.newTimesheetDataForCurrentStaff,
		`^new timesheet data for other staff`:                                       s.newTimesheetDataForOtherStaff,
		`^user creates a new timesheet$`:                                            s.userCreateANewTimesheet,
		`^the timesheet is created "([^"]*)"$`:                                      s.theTimesheetIsCreated,
		`^new timesheet data for existing timesheet`:                                s.newTimesheetForExistingTimesheet,
		`^new timesheet data with other working hours and remark for current staff`: s.newTimesheetDataWithOtherWorkingHours,
		`^new data with "([^"]*)" for current staff`:                                s.newTimesheetDataWithInvalidRequest,

		// Update timesheet
		`^user update a timesheet$`: s.userUpdateTimesheet,
		`^new update "([^"]*)" for timesheet with other working hours data for current staff`: s.userUpdateInvalidArgsForTimesheet,
		`^new update timesheet with "([^"]*)" other working hours data for current staff`:     s.newUpdateTimesheetWithOWHsDataForCurrentStaff,
		`^new updated timesheet data with "([^"]*)" status for current staff$`:                s.newUpdatedTimesheetDataWithStatusForCurrentStaff,
		`^new updated timesheet data with "([^"]*)" status for other staff "([^"]*)"$`:        s.newUpdatedTimesheetDataWithStatusForOtherStaff,
		`^admin change user timesheet status to "([^"]*)"$`:                                   s.adminChangeUserTimesheetStatus,

		// Delete timesheet
		`^current staff deletes this timesheet$`:                       s.deletesThisTimesheet,
		`^an existing "([^"]*)" timesheet for current staff$`:          s.anExistingTimesheetForCurrentStaff,
		`^an existing "([^"]*)" timesheet for other staff "([^"]*)"$`:  s.anExistingTimesheetForOtherStaff,
		`^timesheet is deleted "([^"]*)"$`:                             s.timesheetIsDeleted,
		`^user deletes the timesheet for other staff$`:                 s.deletesThisTimesheet,
		`^timesheet has lesson records$`:                               s.timesheetHasLessonRecords,
		`^timesheet has "([^"]*)" other working hours records$`:        s.timesheetHasOtherWorkingHoursRecords,
		`^timesheet other working hours records is deleted "([^"]*)"$`: s.timesheetOtherWorkingHoursRecordsIsDeleted,
		`^timesheet has "([^"]*)" transport expenses records$`:         s.timesheetHasTransportExpensesRecords,
		`^timesheet transport expenses records is deleted "([^"]*)"$`:  s.timesheetTransportExpensesIsDeleted,

		// Import timesheet config
		`^"([^"]*)" importing timesheet config$`:                                        s.importingTimesheetConfig,
		`^a timesheet config valid request payload$`:                                    s.aTimesheetConfigValidRequestPayload,
		`^a timesheet config valid request payload with incorrect data with "([^"]*)"$`: s.aTimesheetConfigValidRequestPayloadWithIncorrectData,
		`^the valid timesheet config lines are imported successfully$`:                  s.theValidTimesheetConfigLinesAreImportedSuccessfully,
		`^the import timesheet config transaction is rolled back$`:                      s.theImportTimesheetConfigTransactionIsRolledBack,
		`^the invalid timesheet config lines are returned with error$`:                  s.theInvalidTimesheetConfigLinesAreReturnedWithError,
		`^a timesheet config invalid "([^"]*)" request payload$`:                        s.aTimesheetConfigInvalidRequestPayload,

		// Submit timesheet
		`^current staff submits this timesheet$`:                                  s.submitsThisTimesheet,
		`^timesheet status changed to submitted "([^"]*)"$`:                       s.timesheetStatusChangedToSubmitted,
		`^an existing "([^"]*)" timesheet with date "([^"]*)" for current staff$`: s.anExistingTimesheetWithDateForCurrentStaff,
		`^timesheet has lesson records with "([^"]*)"$`:                           s.createLessonRecords,
		`^user submits the timesheet for other staff$`:                            s.submitsThisTimesheet,

		// lesson mgmt
		`^user signed in as school admin$`:                               s.CommonSuite.ASignedInAsSchoolAdmin,
		`^enter a school$`:                                               s.enterASchool,
		`^have some centers$`:                                            s.someCenters,
		`^have some teacher accounts$`:                                   s.CreateTeacherAccounts,
		`^have 2 teacher accounts$`:                                      s.Create2TeacherAccounts,
		`^have 2 teacher accounts will be use for update lesson$`:        s.CreateTeacherAccountsForUpdateLesson,
		`^have some student accounts$`:                                   s.CommonSuite.CreateStudentAccounts,
		`^have some courses$`:                                            s.CommonSuite.SomeCourse,
		`^have some student subscriptions$`:                              s.CommonSuite.SomeStudentSubscriptions,
		`^have some medias$`:                                             s.CommonSuite.UpsertValidMediaList,
		`^cloned teacher to timesheet db`:                                s.clonedTeacherToTimesheetDB,
		`^user create a lesson$`:                                         s.UserCreateALessonWithAllRequiredFields,
		`^user create a lesson in lessonmgmt$`:                           s.UserCreateALessonWithAllRequiredFieldsInLessonmgmt,
		`^the lesson scheduling status was updated$`:                     s.TheLessonSchedulingStatusWasUpdated,
		`^user updates scheduling status in the lesson is "([^"]*)"$`:    s.userUpdatesStatusInTheLessonIsValue,
		`^the lesson teachers was updated$`:                              s.TheLessonTeacherWasUpdated,
		`^the lesson "([^"]*)" was updated$`:                             s.TheLessonFieldWasUpdated,
		`^user "([^"]*)" teacher in the lesson$`:                         s.userUpdateTeacherInTheLesson,
		`^user remove 1 and add 1 teacher in the lesson$`:                s.userUpdateNewTeacherInTheLesson,
		`^the user updates the lesson date to a different date$`:         s.userUpdateLessonDate,
		`^the user updates the lesson location to a different location$`: s.userUpdateLessonLocation,
		`user deletes a lesson$`:                                         s.userDeleteALesson,
		`^user signed in as teacher`:                                     s.CommonSuite.ASignedInTeacher,
		`^an existing lesson in lessonmgmt$`:                             s.UserCreateALessonWithAllRequiredFieldsInLessonmgmt,
		`^user creates recurring lesson$`:                                s.createRecurringLesson,
		`^user have created recurring lesson$`:                           s.createRecurringLesson,
		`^the lesson "([^"]*)" will locked$`:                             s.lockLessons,
		`^user deletes recurring lesson from "([^"]*)" with "([^"]*)"$`:  s.userDeleteLessonRecurring,
		`^user changed lesson end date to "([^"]*)"$`:                    s.userChangedEndDate,
		`^user update selected lesson by saving weekly recurrence$`:      s.updateLessonBySavingWeekly,

		// Delete timesheet lesson hours
		`^total "([^"]*)" timesheet lesson hours will be "([^"]*)"$`:                    s.checkTimesheetLessonHoursIsDeleted,
		`^timesheet will be deleted$`:                                                   s.checkTimesheetIsDeleted,
		`^timesheet cannot be deleted$`:                                                 s.checkTimesheetIsNotDeleted,
		`^"([^"]*)" timesheet recurring lesson will be deleted`:                         s.checkTimesheetRecurringLessonBeDeleted,
		`^"([^"]*)" other working hours records be added to existing timesheet$`:        s.addOtherWorkingHoursToTimesheet,
		`^"([^"]*)" transportation expenses be added to existing timesheet`:             s.addTransportationExpenseToTimesheet,
		`^"([^"]*)" other working hours records are still existing`:                     s.checkOtherWorkingHoursStillExisted,
		`^"([^"]*)" transportation expenses are still existing`:                         s.checkTransportationExpensesStillExisted,
		`^"([^"]*)" timesheet lesson hours should be "([^"]*)"`:                         s.checkTimesheetLessonHoursRecurringLessonBeDeleted,
		`^"([^"]*)" timesheet lesson hours are still existed`:                           s.checkTimesheetLessonHoursRemaining,
		`^"([^"]*)" timesheet cannot be deleted$`:                                       s.checkTimesheetRecurringLessonNotDeleted,
		`^"([^"]*)" other working hours records be added to timesheet lesson recurring`: s.addOtherWorkingHoursToTimesheetLessonRecurring,
		`^"([^"]*)" transportation expenses be added to timesheet lesson recurring`:     s.addTransportationExpenseToTimesheetLessonRecurring,

		// Auto create timesheet
		`^"([^"]*)" timesheet will be "([^"]*)"$`:                                              s.checkTimesheetIsCreated,
		`^"([^"]*)" timesheet lesson hours will be "([^"]*)"$`:                                 s.checkTimesheetLessonHoursIsCreated,
		`^current timesheet lesson hours is "([^"]*)"$`:                                        s.checkTimesheetLessonHoursIsValid,
		`timesheet have status "([^"]*)"$`:                                                     s.checkTimesheetStatus,
		`timesheet with old date was removed and timesheet with new date was created$`:         s.checkTimesheetsChangedWhenLessonDateChanged,
		`timesheet with old location was removed and timesheet with new location was created$`: s.checkTimesheetsChangedWhenLessonLocationChanged,

		// Approve timesheet
		`^current staff approves this timesheet$`:               s.approvesThisTimesheet,
		`^timesheet status changed to approve "([^"]*)"$`:       s.timesheetStatusChangedToApprove,
		`^user approves the timesheet for other staff$`:         s.approvesThisTimesheet,
		`^staff has an existing "([^"]*)" submitted timesheet$`: s.staffHasanExistingSubmittedTimesheet,
		`^each timesheets has lesson records with "([^"]*)"$`:   s.eachTimesheetsHasLessonRecordsWith,
		`^"([^"]*)" staff approves this timesheet$`:             s.staffApprovesThisTimesheet,

		// Send Timesheet Lesson Lock Flag
		`^timesheet send event lock lesson$`:                  s.timesheetSendEventLockLesson,
		`^timesheet event lock lesson published successfully`: s.timesheetEventLockLessonPublishedSuccessfully,

		// Cancel Approve Timesheet
		`^timesheet status approved changed to submitted "([^"]*)"$`: s.timesheetStatusApproveChangedToSubmitted,
		`^current staff cancel approve this timesheet$`:              s.cancelApproveThisTimesheet,
		`^user cancel approve the timesheet for other staff$`:        s.cancelApproveThisTimesheet,

		// Confirm Timesheet
		`^staff has an existing "([^"]*)" approve timesheet$`: s.staffHasAnExistingApproveTimesheet,
		`^"([^"]*)" staff confirms this timesheet$`:           s.staffConfirmsThisTimesheet,
		`^user confirms the timesheet for other staff$`:       s.confirmsThisTimesheet,
		`^timesheet statuses changed to confirm "([^"]*)"$`:   s.timesheetStatusesChangedToConfirm,
		`^current staff confirms this timesheet$`:             s.confirmsThisTimesheet,

		// Auto Create Timesheet Flag
		`^user update a auto create timesheet flag$`:            s.userUpsertAutoCreateTimesheetFlag,
		`^user config auto create flag "([^"]*)" for teachers$`: s.userUpsertAutoCreateTimesheetFlagForTeachers,
		`^new flag data with "([^"]*)" status$`:                 s.newUpsertAutoCreateTimesheetFlagData,
		`^flag status changed to "([^"]*)"$`:                    s.verifyFlagStatusAfterUpsert,

		// Auto Create Timesheet Flag Log
		`^a log record is inserted with status is "([^"]*)"$`: s.checkOnelogRecordIsInserted,
		`^count number "([^"]*)" status log record of user$`:  s.countLogRecordOfUser,

		`^The teacher have auto create flag is "([^"]*)"$`:                                    s.adminUpdateTeacherAutoCreateFlag,
		`^Admin create a future lesson in lessonmgmt for the teacher$`:                        s.UserCreateALessonWithFutureDateInLessonmgmt,
		`"([^"]*)" timesheet lesson hours will be "([^"]*)" with auto create flag "([^"]*)"$`: s.checkTimesheetLessonHoursIsCreatedWithFlag,
		`^School Admin update teacher auto create flag to "([^"]*)"$`:                         s.adminUpdateTeacherAutoCreateFlag,
		`^flag status changed to "([^"]*)" v2$`:                                               s.verifyFlagStatusAfterUpdate,
		`^flag in timesheet lesson hours changed to "([^"]*)"$`:                               s.flagInTimesheetLessonHoursChangeTo,

		// Create timesheet with transportation expense data
		`^new timesheet with transportation expenses$`:               s.newTimesheetWithTranportExpensesData,
		`^new transportation data with "([^"]*)" for current staff$`: s.newInvalidTransportExpenseDataRequest,

		// Update timesheet with transportation expense data
		`^new update "([^"]*)" for timesheet with transportation expenses data for current staff$`: s.userUpdateInvalidTransportExpensesArgsForTimesheet,
		`^new update timesheet with "([^"]*)" transportation expenses data for current staff$`:     s.newUpdateTimesheetWithTransportExpensesDataForCurrentStaff,

		// Cancel submission timesheet
		`^current staff cancel submits this timesheet$`:       s.cancelSubmitThisTimesheet,
		`^timesheet status changed to draft "([^"]*)"$`:       s.timesheetStatusChangedToDraft,
		`^user cancel submits the timesheet for other staff$`: s.cancelSubmitThisTimesheet,

		// Upsert staff transportation expense
		`^new insert staff transportation expense request with "([^"]*)"$`: s.newInsertStaffTransportationExpenseConfig,
		`^user upsert staff transportation expense config$`:                s.userUpsertStaffTransportationExpense,
		`^staff have "([^"]*)" transportation expense config value$`:       s.verifyStaffTransportationConfigNumberAfterUpsert,
		`^remove all staff old staff transportation expense records$`:      s.removeAllStaffTransportationExpense,
		`^new update staff transportation expense config request$`:         s.newUpdateStaffTransportationExpenseConfig,
		`^new upsert staff transportation expense config request$`:         s.newUpsertStaffTransportationExpenseConfig,
		`^new delete staff transportation expense config request$`:         s.newDeleteStaffTransportationExpenseConfig,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
