package calendar

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
		// health Check
		`^everything is OK$`:                                       s.everythingIsOK,
		`^health check endpoint called$`:                           s.healthCheckEndpointCalled,
		`^calendar should return "([^"]*)" with status "([^"]*)"$`: s.calendarShouldReturnWithStatus,

		// signed in steps
		`^user signed in as school admin$`: s.aSignedInAsSchoolAdmin,
		`^signed as "([^"]*)" account$`:    s.SignedAsAccountV2,
		`^a signed in admin$`:              s.CommonSuite.ASignedInAdmin,
		`^a signed in student$`:            s.CommonSuite.ASignedInStudent,
		`^a random number$`:                s.CommonSuite.ARandomNumber,
		`^returns "([^"]*)" status code$`:  s.CommonSuite.ReturnsStatusCode,

		// common background steps
		`^enter a school$`:                  s.enterASchool,
		`^have some locations$`:             s.someExistingLocations,
		`^have some teacher accounts$`:      s.CommonSuite.CreateTeacherAccounts,
		`^have some student accounts$`:      s.CommonSuite.CreateStudentAccounts,
		`^have some grades$`:                s.CommonSuite.CreateSomeGrades,
		`^have some classrooms$`:            s.CommonSuite.CreateClassrooms,
		`^have some courses$`:               s.CommonSuite.SomeCourse,
		`^have some student subscriptions$`: s.CommonSuite.SomeStudentSubscriptions,
		`^have some medias$`:                s.CommonSuite.UpsertValidMediaList,

		// unleash
		`^"([^"]*)" Unleash feature with feature name "([^"]*)"$`: s.UnleashSuite.ToggleUnleashFeatureWithName,

		// Background for date info
		`^an existing location "([^"]*)" in DB$`:                                                             s.anExistingLocationInCalendarDB,
		`^an existing date type "([^"]*)" in DB$`:                                                            s.anExistingDateTypeInDB,
		`^an existing date info for date "([^"]*)" and location "([^"]*)"$`:                                  s.anExistingDateInfoForDateAndLocation,
		`^a date type "([^"]*)"$`:                                                                            s.existingDateTypes,
		`^a date "([^"]*)", location "([^"]*)", date type "([^"]*)", open time "([^"]*)", status "([^"]*)"$`: s.existingDateInfos,

		// Get Calendar Info
		`^user get calendar with filter "([^"]*)", "([^"]*)", "([^"]*)"$`: s.getDateInfoByDurations,
		`^must return all date info by location$`:                         s.returnsDateInfoByDurations,

		// Duplicate date info
		`^admin choose "([^"]*)", "([^"]*)"$`:                                  s.userChooseTheDate,
		`^duplicate date info with condition "([^"]*)", "([^"]*)", "([^"]*)"$`: s.duplicateDateInfo,

		// Upsert Date Info
		`^user creates a date info for date "([^"]*)" and location "([^"]*)"$`: s.userCreatesADateInfoForDateAndLocation,
		`^date info is created successfully$`:                                  s.dateInfoIsCreatedSuccessfully,
		`^user updates date info for date "([^"]*)" and location "([^"]*)"$`:   s.userUpdatesDateInfoForDateAndLocation,
		`^date info is updated successfully$`:                                  s.dateInfoIsCreatedSuccessfully,

		// scheduler
		`^user creates a scheduler "([^"]*)", "([^"]*)", "([^"]*)"$`: s.createSchedulerFromScenario,
		`^user has created a scheduler$`:                             s.randomScheduler,
		`^scheduler has been added to the database$`:                 s.existedScheduler,
		`^scheduler has been updated to the database$`:               s.updatedScheduler,
		`^user update scheduler$`:                                    s.updateScheduler,

		// migrate data
		`^a list of location records has been added to the bob database$`:      s.createAListOfLocationsInBobDB,
		`^location data on calendar db updated successfully$`:                  s.checkSyncLocationData,
		`^a list of location type records has been added to the bob database$`: s.createAListOfLocationTypesInBobDB,
		`^location type data on calendar db updated successfully$`:             s.checkSyncLocationTypeData,

		// get list staff
		`^user get list staff$`:                                  s.getListStaffByLocation,
		`^a list of staff already exists in DB$`:                 s.aListOfStaffCreated,
		`^a list of staff with user group already exists in DB$`: s.aListOfStaffWithUserGroupCreated,
		`^a list correct staff is returned$`:                     s.aListCorrectStaffReturned,
		`^an empty list staff is returned$`:                      s.anEmptyListStaffReturned,

		// export day info
		`^returns day infos in csv with Ok status code$`: s.returnsDayInfosInCsv,
		`^user export day infos$`:                        s.exportDayInfo,

		// get lesson detail on calendar
		`^an existing "([^"]*)" lesson$`:             s.anExistingLesson,
		`^user get lesson detail on calendar$`:       s.getLessonDetailOnCalendar,
		`^the lesson detail matches lesson created$`: s.lessonDetailMatchesLessonCreated,

		// get lesson ids for bulk status update
		`^some existing "([^"]*)" lessons$`:                     s.someExistingStatusLessons,
		`^user get lesson IDs for bulk "([^"]*)"$`:              s.userGetLessonIDsforBulkAction,
		`^returned lesson IDs for bulk "([^"]*)" are expected$`: s.returnedLessonIDsForBulkActionAreExpected,
	}

	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
