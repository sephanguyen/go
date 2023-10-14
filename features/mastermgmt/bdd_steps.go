package mastermgmt

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/helper"
	golibs_constants "github.com/manabie-com/backend/internal/golibs/constants"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^everything is OK$`:                                         s.everythingIsOK,
		`^health check endpoint called$`:                             s.healthCheckEndpointCalled,
		`^mastermgmt should return "([^"]*)" with status "([^"]*)"$`: s.mastermgmtShouldReturnWithStatus,
		`^a random number$`:                                          s.aRandomNumber,
		`^a random number in range (\d+)$`:                           s.ARandomNumberInRange,
		`^a generate school$`:                                        s.aGenerateSchool,
		`^some centers$`:                                             s.someCenters,
		`^some location types$`:                                      s.seedLocationTypes,
		`^some course types$`:                                        s.someCourseTypes,
		`^"([^"]*)" signin system$`:                                  s.signedAsAccountV2,
		// master data course
		`^user upsert courses "([^"]*)" data with (\d+) locations and teaching method "([^"]*)"$`: s.userUpsertCoursesDataWithLocationsAndTeachingMethod,
		`^course access paths already exist in DB with (\d+) locations$`:                          s.courseAccessPathsExistInDB,
		`^return a status code "([^"]*)"$`:                                                        s.returnAStatusCode,
		`^returns "([^"]*)" status code$`:                                                         s.CommonSuite.ReturnsStatusCode,
		`^some student accounts with school id$`:                                                  s.CreateStudentAccounts,
		`^some student subscription with "([^"]*)" existed in DB$`:                                s.someStudentSubscriptionExistedInDB,
		`^a list of lesson_student_subscription_access_path are existed in DB$`:                   s.aListStudentSubscriptionAccessPathExistedInDB,
		`^return an error "([^"]*)" message$`:                                                     s.returnAErrorMessage,
		`^location have state "([^"]*)"$`:                                                         s.locationRemovedHaveState,
		// master data course: teaching method
		`^course saved DB correct with teaching method "([^"]*)"$`: s.checkCourseInfoInDB,

		// course and subjects

		`^courses updated with correct subjects$`:       s.checkUpdatedCoursesAndSubjects,
		`^user upsert courses with subjects "([^"]*)"$`: s.upsertCoursesWithSubjects,
		`^some subjects$`: s.seedSubjects,

		// Create organization
		`^user create new organization$`:                            s.createNewOrganizationData,
		`^location type and location default created successfully$`: s.locationTypeAndLocationDefaultCreatedSuccessfully,
		`^organization data has invalid domain name "([^"]*)"$`:     s.organizationDataHasInvalidDomainName,

		// location
		`^a list of location types are existed in DB$`:                    s.aListOfLocationTypesInDB,
		`^a list of locations are existed in DB$`:                         s.aListOfLocationsInDB,
		`^user retrieve location types$`:                                  s.retrieveLocationTypes,
		`^user retrieve location types v2$`:                               s.retrieveLocationTypesV2,
		`^user retrieve locations$`:                                       s.retrieveLocations,
		`^must return a correct list of locations$`:                       s.mustReturnCorrectLocations,
		`^must return a correct list of location types$`:                  s.mustReturnCorrectLocationTypes,
		`^returns unArchived location types$`:                             s.verifyLocationTypes,
		`^a list of locations with variant types are existed in DB$`:      s.aListOfLocationsVariantTypesInDB,
		`^user retrieve lowest level of locations with filter "([^"]*)"$`: s.retrieveLowestLevelLocations,
		`^must return lowest level of locations with filter "([^"]*)"$`:   s.mustReturnLowestLevelLocations,
		`import location logs were inserted successfully`:                 s.mustStoreLocationImportLogs,
		// import location type
		`^some locations with location type "([^"]*)" and parentName "([^"]*)"$`: s.someLocationsWithLocationTypes,
		`^admin update parent of location_type "([^"]*)"$`:                       s.adminUpdateParentOfLocationType,
		`^a location type valid request payload with "([^"]*)"$`:                 s.aLocationTypeValidRequestPayloadWith,
		`^importing location type$`:                                              s.importingLocationType,
		`^the invalid location type lines are returned with error$`:              s.theInvalidLocationTypeLinesAreReturnedWithError,
		`^the valid location type lines are imported successfully$`:              s.theValidLocationTypeLinesAreImportedSuccessfully,
		`^a location type invalid "([^"]*)" request payload$`:                    s.aLocationTypeInvalidRequestPayload,
		`^admin import a location type with "([^"]*)", "([^"]*)", "([^"]*)"$`:    s.importLocationTypeOtherSchool,
		`^returns (\d+) location_type failed$`:                                   s.returnNumberLocationTypeFailed,
		`^location type "([^"]*)" with parentName "([^"]*)" still exist in DB$`:  s.locationTypeWithParentExistInDB,
		`^import location type logs were inserted successfully$`:                 s.mustStoreLocationTypeImportLogs,
		// import master data - location
		`^a location valid request payload with "([^"]*)", "([^"]*)"$`: s.aLocationValidRequestPayloadWith,
		`^importing location$`:                                           s.importingLocation,
		`^the invalid location lines are returned with error$`:           s.theInvalidLocationLinesAreReturnedWithError,
		`^the valid location lines are imported successfully$`:           s.theValidLocationLinesAreImportedSuccessfully,
		`^a location invalid "([^"]*)" request payload$`:                 s.aLocationInvalidRequestPayload,
		`^a location type value "([^"]*)"$`:                              s.aLocationTypeValues,
		`^had "([^"]*)" record of the locations database$`:               s.aListOfLocationsInDB,
		`^admin import a location with "([^"]*)", "([^"]*)", "([^"]*)"$`: s.importLocationOtherSchool,
		`^returns (\d+) location failed$`:                                s.returnNumberLocationFailed,

		// get location tree
		`^locations with children existed$`:     s.PrepareLocationWithChildren,
		`^user gets location tree$`:             s.GetLocationTree,
		`^must return a correct location tree$`: s.VerifyLocationTree,

		// Create organization
		`^new organization data$`:                            s.newOrganizationData,
		`^"([^"]*)" create new organization$`:                s.createNewOrganization,
		`^new organization were created successfully$`:       s.newOrganizationWereCreatedSuccessfully,
		`^organization data has empty or invalid "([^"]*)"$`: s.organizationDataHasEmptyOrInvalid,
		`^"([^"]*)" user can not create organization$`:       s.userCannotCreateOrganization,

		`^Mastermgmt must push msg "([^"]*)" subject "([^"]*)" to nats$`: s.mastermgmtMustPushMsgSubjectToNats,
		`^mastermgnt have to publish subject "([^"]*)"$`:                 s.checkEventPublisher,
		// Class mgmt
		`^have some courses$`:                     s.CommonSuite.SomeCourse,
		`^user import classes by csv file$`:       s.importClass,
		`^the valid classes was created$`:         s.createdClassProperly,
		`^return error of invalid classes$`:       s.returnErrorOfInvalidClass,
		`^a valid classes payload$`:               s.validClassesPayload,
		`^a valid and invalid classes payload$`:   s.validAndInvalidlassesPayload,
		`^have a class$`:                          s.insertClass,
		`^admin update class$`:                    s.updateClass,
		`^class have updated successfully$`:       s.updatedClassProperly,
		`^admin delete class$`:                    s.deleteClass,
		`^class have deleted successfully$`:       s.deletedClassProperly,
		`^a list of class are existed in DB$`:     s.aListOfClassInDB,
		`^user retrieve classes$`:                 s.retrieveClassesByIDs,
		`^must return a correct list of classes$`: s.mustReturnCorrectClasses,

		// sync class member
		`^user add a package by course with student package extra for a student$`: s.userAddCourseWithStudentPackageExtraForAStudent,
		`^server must store correct class members$`:                               s.classMemberStoredInDB,

		// import grade
		`^user import grades by csv file$`:            s.importsGrades,
		`^the valid grades were updated$`:             s.checkImportedGrades,
		`^returns error of "([^"]*)" invalid grades$`: s.checkGradeImportErrors,
		`^a valid grades payload$`:                    s.prepareValidGradesPayload,
		`^a "([^"]*)" invalid grades payload$`:        s.prepareInValidGradesPayload,

		// get configuration
		`^some existing configurations in DB$`:                                                      s.someExistingConfigurationsInDB,
		`^configurations are returned all items$`:                                                   s.returnCorrectConfigurations,
		`^locations configurations are returned all items "([^"]*)"$`:                               s.returnCorrectLocationConfigurations,
		`^user gets configurations with "([^"]*)" keyword at page "(\d+)" and limit "(\d+)" items$`: s.getConfigurations,
		`^user gets configurations with "([^"]*)" keyword$`:                                         s.getAllConfigurations,
		`^service gets configurations with "([^"]*)" keyword$`:                                      s.getAllConfigurationsByService,
		`^user gets locations configurations with "([^"]*)"$`:                                       s.getLocationConfigurations,
		`^paginated configurations are returned at page "(\d+)" and limit "(\d+)" items$`:           s.returnPaginatedConfigurations,
		`^configurations value existed on DB$`:                                                      s.configurationsExistedOnDB,
		`^location configurations value existed on DB$`:                                             s.initLocationConfigInDB,
		`^location configurations v2 value "([^"]*)" existed on DB$`:                                s.initLocationConfigV2InDB,
		`^locations configurations are returned with "([^"]*)"$`:                                    s.locationsConfigurationsAreReturnedWith,
		`^user gets locations configurations with "([^"]*)" locations$`:                             s.userGetsLocationsConfigurationsWithLocations,

		// init configuration
		`^any org and config key in DB$`:                                              s.anyOrgAndConfigkeyInDB,
		`^a new org inserted in to DB$`:                                               s.aNewOrgInsertedIntoDB,
		`^a new "([^"]*)" config key inserted in to DB$`:                              s.aNewConfigKeyInsertedIntoDB,
		`^new values of the new "([^"]*)" config key are added for all existing org$`: s.newConfigValueAddedForExistingOrg,
		`^new values of all existing config key are added for the new org$`:           s.newConfigValueAddedForNewOrg,

		//audit configuration value change
		`^a "([^"]*)" config key was inserted in DB$`:                            s.createConfigurationKey,
		`^update the value of any "([^"]*)" configuration$`:                      s.updateValueOfConfiguration,
		`^the change be captured in the audit table of "([^"]*)" configuration$`: s.checkTheAuditLogRecorded,

		// version control
		`^a invalid version request$`:              s.aInvalidVersionRequest,
		`^user verify version$`:                    s.userVerifyVersion,
		`^a request with lower version$`:           s.aRequestWithLowerVersion,
		`^a request with valid version$`:           s.aRequestWithValidVersion,
		`^return false in message$`:                s.returnFalseInMessage,
		`^a request with lower version "([^"]*)"$`: s.aRequestWithLowerVersion,

		// export grades
		`^some grades existed in DB$`:                 s.gradesExistedInDB,
		`^returns grades in csv with Ok status code$`: s.returnsGradesInCsv,
		`^user export grades$`:                        s.exportGrades,

		// export subjects
		`^some subjects existed in DB$`:                 s.subjectsExistedInDB,
		`^returns subjects in csv with Ok status code$`: s.returnsSubjectsInCsv,
		`^user export subjects$`:                        s.exportSubjects,

		// import subjects
		`^user import subjects by csv file$`:            s.importsSubjects,
		`^the valid subjects were updated$`:             s.checkImportedSubjects,
		`^returns error of "([^"]*)" invalid subjects$`: s.checkSubjectImportErrors,
		`^a valid subjects payload$`:                    s.prepareValidSubjectsPayload,
		`^a "([^"]*)" invalid subjects payload$`:        s.prepareInValidSubjectsPayload,

		// import course access path
		`^have some course access paths$`:                         s.seedCourseAccessPaths,
		`^seeded (\d+) courses$`:                                  s.seedSomeCourses,
		`^user import course access paths by csv file$`:           s.importCourseAccessPaths,
		`^the valid course access paths were updated$`:            s.checkImportedCAP,
		`^returns error of "([^"]*)" invalid course access paths`: s.checkCAPImportErrors,
		`^a valid course access paths payload$`:                   s.prepareValidCAPPayload,
		`^a "([^"]*)" invalid course access paths payload$`:       s.prepareInValidCAPPayload,

		// export course access paths
		`^returns course access paths in csv with Ok status code$`: s.checkCourseAccessPathCSV,
		`^user export course access paths`:                         s.exportCourseAccessPaths,

		// export courses
		`^courses existed in DB$`:                      s.coursesExistedInDB,
		`^returns courses in csv with Ok status code$`: s.returnsCoursesInCsv,
		`^user export courses$`:                        s.exportCourses,

		// import courses
		`^user import courses by csv file$`:            s.importCourses,
		`^the valid courses were updated$`:             s.checkUpdatedCourses,
		`^returns error of "([^"]*)" invalid courses$`: s.checkImportCourseCSVErrors,
		`^a valid courses payload$`:                    s.prepareValidCoursesPayload,
		`^a "([^"]*)" invalid courses payload$`:        s.prepareInValidCoursesPayload,

		// import course types
		`^user import course types by csv file$`:            s.importCourseTypes,
		`^the valid course types were updated$`:             s.checkUpdatedCourseTypes,
		`^returns error of "([^"]*)" invalid course types$`: s.checkImportCourseTypeCSVErrors,
		`^a valid course types payload$`:                    s.prepareValidCourseTypesPayload,
		`^a "([^"]*)" invalid course types payload$`:        s.prepareInValidCourseTypesPayload,

		// import location type v2
		`^user import location type by csv file$`:            s.importLocationTypeV2,
		`^the valid location type were updated$`:             s.checkUpdatedLocationTypes,
		`^returns error of "([^"]*)" invalid location type$`: s.checkImportLocationTypeCSVErrors,
		`^a valid location type payload$`:                    s.prepareValidLocationTypePayload,
		`^a "([^"]*)" invalid location type payload$`:        s.prepareInvalidLocationTypePayload,

		// import location v2
		`^user import location by csv file$`:            s.importLocationV2,
		`^the valid location were updated$`:             s.checkUpdatedLocation,
		`^returns error of "([^"]*)" invalid location$`: s.checkImportLocationCSVErrors,
		`^a valid location payload$`:                    s.prepareValidLocationPayload,
		`^a "([^"]*)" invalid location payload$`:        s.prepareInvalidLocationPayload,

		// export classes
		`^classes existed in DB$`:                      s.classesExistedInDB,
		`^returns classes in csv with Ok status code$`: s.returnsClassesInCsv,
		`^user export classes$`:                        s.exportClasses,

		// export locations
		`^some locations existed in DB$`:                 s.locationsExistedInDB,
		`^returns locations in csv with Ok status code$`: s.returnsLocationsInCsv,
		`^user export locations$`:                        s.exportLocations,

		// export location types
		`^returns "([^"]*)" location types in csv with Ok status code$`: s.returnsLocationTypesInCsv,
		`^user export location types$`:                                  s.exportLocationTypes,

		// appsmith
		`^user gets appsmith page info by slug$`: s.getPageInfoBySlug,
		`^returns corresponding appsmith page$`:  s.returnCorrectAppsmithPage,

		// get courses by ids
		`^user get courses by "([^"]*)" ids$`:               s.getCoursesByIDs,
		`^must return a correct list of "([^"]*)" courses$`: s.mustReturnCorrectCourses,

		// track appsmith data
		`^a request with a "([^"]*)" header$`:             s.withAppsmithTrackRequest,
		`^the track endpoint is called$`:                  s.trackEndpointIsCalled,
		`^returns "([^"]*)" and "([^"]*)" response data$`: s.returnTrackResponse,

		// execute custom entity
		`^school admin execute script$`: s.adminExecuteCustomScript,

		// unleash client
		`^"([^"]*)" Unleash feature with feature name "([^"]*)"$`: s.UnleashSuite.ToggleUnleashFeatureWithName,

		// import academic year
		`^a valid academic calendar payload`:                   s.validAcademicCalendarCsvPayload,
		`^user import academic calendar by csv file`:           s.importAcademicCalendar,
		`^have academic year`:                                  s.addAcademicYear,
		`^an invalid academic calendar payload`:                s.invalidAcademicCalendarCsvPayload,
		`^user try to export academic calendar`:                s.exportAcademicCalendar,
		`returns academic calendar in csv with Ok status code`: s.returnsAcademicCalendarInCsv,

		// import working hours
		`^a valid working hours payload`:         s.validWorkingHoursCsvPayload,
		`^user import working hours by csv file`: s.importWorkingHours,
		`^an invalid working hours payload`:      s.invalidWorkingHoursCsvPayload,

		// import time slots
		`^a valid time slots payload`:         s.validTimeSlotCsvPayload,
		`^user import time slots by csv file`: s.importTimeSlot,
		`^an invalid time slots payload`:      s.invalidTimeSlotCsvPayload,

		// retrieve locations for academic
		`^user retrieve locations for academic calendar`: s.retrieveLocationsForAcademic,

		// schedule class
		`^user has scheduled class to reserve class`:                          s.hasScheduledClassToReserveClass,
		`^user schedule class to reserve class again with other class`:        s.scheduleClassToReserveClassAgain,
		`reserve class must be stored correct on db`:                          s.reserveClassMustStoreInDatabaseCorrect,
		`^user call func wrapper register student class`:                      s.callFuncWrapperRegisterStudentClass,
		`student package class with reserve class must be stored in database`: s.studentClassMustStoreInDatabase,
		`^user retrieve scheduled class info`:                                 s.retrieveScheduledClass,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

//nolint:gocyclo
func (s *suite) mastermgmtMustPushMsgSubjectToNats(ctx context.Context, msg, subject string) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)

	var (
		// handler  func(msg *stan.Msg)
		foundChn = make(chan struct{}, 1)
	)

	switch subject {
	case golibs_constants.SubjectSyncLocationUpserted:
		if msg == "UpsertLocation" {
			timer := time.NewTimer(time.Minute)
			defer timer.Stop()

			select {
			case <-stepState.FoundChanForJetStream:
				return StepStateToContext(ctx, stepState), nil
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	case golibs_constants.SubjectSyncLocationTypeUpserted:
		if msg == "UpsertLocationType" {
			timer := time.NewTimer(time.Minute)
			defer timer.Stop()

			select {
			case <-stepState.FoundChanForJetStream:
				return StepStateToContext(ctx, stepState), nil
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	case golibs_constants.SubjectMasterMgmtClassUpserted:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		for {
			select {
			case message := <-stepState.FoundChanForJetStream:
				switch message.(type) {
				case *mpb.EvtClass_CreateClass_:
					if msg == "CreateClass" {
						return StepStateToContext(ctx, stepState), nil
					}
				case *mpb.EvtClass_UpdateClass_:
					if msg == "UpdateClass" {
						return StepStateToContext(ctx, stepState), nil
					}
				case *mpb.EvtClass_DeleteClass_:
					if msg == "DeleteClass" {
						return StepStateToContext(ctx, stepState), nil
					}
				default:
					return StepStateToContext(ctx, stepState), errors.New("message type unknown")
				}
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	}

	timer := time.NewTimer(time.Minute * 6)
	defer timer.Stop()

	select {
	case <-foundChn:
		return StepStateToContext(ctx, stepState), nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}
