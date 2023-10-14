package lessonmgmt

import (
	"context"
	"fmt"
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
		`^everything is OK$`:                                          s.everythingIsOK,
		`^health check endpoint called$`:                              s.healthCheckEndpointCalled,
		`^lesson mgmt should return "([^"]*)" with status "([^"]*)"$`: s.lessonMgmtShouldReturnWithStatus,

		`^a signed in admin$`:       s.CommonSuite.ASignedInAdmin,
		`^a signed in student$`:     s.CommonSuite.ASignedInStudent,
		`^a random number$`:         s.CommonSuite.ARandomNumber,
		`^"([^"]*)" signin system$`: s.signedAsAccountV2,
		`^a school name "([^"]*)", country "([^"]*)", city "([^"]*)", district "([^"]*)"$`: s.CommonSuite.ASchoolNameCountryCityDistrict,
		`^admin inserts schools$`:                      s.CommonSuite.AdminInsertsSchools,
		`^a list of valid topics$`:                     s.CommonSuite.AListOfValidTopics,
		`^admin inserts a list of valid topics$`:       s.CommonSuite.AdminInsertsAListOfValidTopics,
		`^a signed in teacher$`:                        s.aSignedInTeacherV2,
		`^a CreateClassRequest$`:                       s.CommonSuite.ACreateClassRequest,
		`^a "([^"]*)" schoolId in CreateClassRequest$`: s.CommonSuite.ASchoolIdInCreateClassRequest,
		`^a valid name in CreateClassRequest$`:         s.CommonSuite.AValidNameInCreateClassRequest,
		`^this school has config "([^"]*)" is "([^"]*)", "([^"]*)" is "([^"]*)", "([^"]*)" is (\d+)$`: s.CommonSuite.ThisSchoolHasConfigIsIsIs,
		`^returns "([^"]*)" status code$`:                                                         s.CommonSuite.ReturnsStatusCode,
		`^Bob must create class from CreateClassRequest$`:                                         s.CommonSuite.BobMustCreateClassFromCreateClassRequest,
		`^class must has "([^"]*)" is "([^"]*)"$`:                                                 s.CommonSuite.ClassMustHasIs,
		`^class must have (\d+) member is "([^"]*)" and is owner "([^"]*)" and status "([^"]*)"$`: s.CommonSuite.ClassMustHaveMemberIsAndIsOwnerAndStatus,
		`^his owned student UUID$`:                                                                s.CommonSuite.HisOwnedStudentUUID,
		`^a JoinClassRequest$`:                                                                    s.CommonSuite.AJoinClassRequest,
		`^user create a class`:                                                                    s.CommonSuite.UserCreateAClass,
		`^a "([^"]*)" classCode in JoinClassRequest$`:                                             s.CommonSuite.AClassCodeInJoinClassRequest,
		`^user join a class$`:                                                                     s.CommonSuite.UserJoinAClass,
		`^a list of courses are existed in DB of "([^"]*)"$`:                                      s.CommonSuite.AListOfCoursesAreExistedInDBOf,
		`^a list of lessons are existed in DB of "([^"]*)" with start time "([^"]*)" and end time "([^"]*)"$`:     s.CommonSuite.AListOfLessonsAreExistedInDBOfWithStartTimeAndEndTime,
		`^teacher retrieve live lesson with start time "([^"]*)" and end time "([^"]*)"$`:                         s.teacherRetrieveLiveLessonWithStartTimeAndEndTime,
		`^student retrieve live lesson with start time "([^"]*)" and end time "([^"]*)"$`:                         s.studentRetrieveLiveLessonWithStartTimeAndEndTime,
		`^teacher end one of the live lesson$`:                                                                    s.teacherEndOneOfTheLiveLesson,
		`^bob must update lesson end at time$`:                                                                    s.bobMustUpdateLessonEndAtTime,
		`^the ended lesson must have status completed$`:                                                           s.theEndedLessonMustHaveStatusCompleted,
		`^Bob must push msg "([^"]*)" subject "([^"]*)" to nats$`:                                                 s.CommonSuite.BobMustPushMsgSubjectToNats,
		`^Lessonmgmt must push msg "([^"]*)" subject "([^"]*)" to nats$`:                                          s.CommonSuite.LessonmgmtMustPushMsgSubjectToNats,
		`^a list of locations with are existed in DB$`:                                                            s.aListOfLocationsInDB,
		`^current student assigned to above lessons$`:                                                             s.currentStudentAssignedToAboveLessons,
		`^"([^"]*)" retrieve live lesson by courseID "([^"]*)" with start time "([^"]*)" and end time "([^"]*)"$`: s.retrieveLiveLessonByCourseWithStartTimeAndEndTime,
		// retrieve live lesson with location_ids
		`^a list of lessons are existed in DB of "([^"]*)" with start time "([^"]*)" and end time "([^"]*)" and location id "([^"]*)"$`:     s.aListOfLessonsAreExistedInDBOfWithStartTimeAndEndTimeAndLocationID,
		`^"([^"]*)" retrieve live lesson with start time "([^"]*)" and end time "([^"]*)" and location id "([^"]*)"$`:                       s.userRetrieveLiveLessonWithStartTimeAndEndTimeAndLocationID,
		`^Bob must return "([^"]*)" live lesson for "([^"]*)" with location id "([^"]*)"$`:                                                  s.bobReturnResultLiveLessonForStudentWithLocationID,
		`^Bob must return "([^"]*)" live lesson for student$`:                                                                               s.bobMustReturnPbLiveLessonForStudent,
		`^Bob must return "([^"]*)" live lesson for teacher$`:                                                                               s.bobReturnResultLiveLessonForTeacher,
		`^"([^"]*)" retrieve live lesson by courseID "([^"]*)" with start time "([^"]*)" and end time "([^"]*)" and location id "([^"]*)"$`: s.userRetrieveLiveLessonByCourseWithStartTimeAndEndTimeAndLocationID,
		// TODO: remove old steps
		`^user signed in admin$`:           s.CommonSuite.ASignedInAdmin,
		`^user signed in as school admin$`: s.aSignedInAsSchoolAdmin,
		`^user signed in as teacher$`:      s.aSignedInTeacherV2,

		`^enter a school$`:             s.enterASchool,
		`^have some centers$`:          s.someCenters,
		`^have some teacher accounts$`: s.CreateTeacherAccounts,
		`^have some student accounts$`: s.CreateStudentAccounts,
		`^have some grades$`:           s.CommonSuite.CreateSomeGrades,
		`^have some classrooms$`:       s.CommonSuite.CreateClassrooms,

		`^have some courses$`:                                                s.CommonSuite.SomeCourse,
		`^have some student subscriptions$`:                                  s.CommonSuite.SomeStudentSubscriptions,
		`^have some medias$`:                                                 s.CommonSuite.UpsertValidMediaList,
		`^user creates a new lesson with all required fields$`:               s.UserCreateALessonWithAllRequiredFieldsWithSub,
		`^user creates a new lesson with all required fields in lessonmgmt$`: s.UserCreateALessonWithAllRequiredFieldsWithSubInLessonmgmt,
		`^an existing lesson$`:                                               s.UserCreateALessonWithAllRequiredFields,
		`^the system already has "([^"]*)" lessons in the database$`:         s.CreateLessons,
		`^the system already has "([^"]*)" lessons in the database with student attendance status "([^"]*)"$`: s.CreateLessonsWithAttendanceStatus,
		`^an existing lesson in lessonmgmt$`:                            s.UserCreateALessonWithAllRequiredFieldsInLessonmgmt,
		`^an existing live lesson$`:                                     s.UserCreateALiveLessonWithAllRequiredFields,
		`^the lesson was updated$`:                                      s.TheLessonWasUpdated,
		`^the lesson scheduling status was updated$`:                    s.TheLessonSchedulingStatusWasUpdated,
		`^the lesson was created$`:                                      s.TheLessonWasCreated,
		`^the lesson was created in lessonmgmt$`:                        s.TheLessonWasCreatedInLessonmgmt,
		`^the lesson was "([^"]*)" in lessonmgmt$`:                      s.TheLessonWasInLessonmgmt,
		`^user updates "([^"]*)" in the lesson$`:                        s.userUpdatesFieldInTheLesson,
		`^user updates scheduling status in the lesson is "([^"]*)"$`:   s.userUpdatesStatusInTheLessonIsValue,
		`^user updates the lesson with start time later than end time$`: s.userUpdateALessonWithStartTimeLaterThanEndTime,
		`^user updates the lesson with missing "([^"]*)"$`:              s.UserUpdatesCurrentLessonWithMissingField,
		`^user bulk updates scheduling status with action "([^"]*)"$`:   s.userBulkUpdatesStatusWithAction,
		`^student and teacher name must be updated correctly$`:          s.StudentTeacherNameMustBeCorrect,
		// bulk update scheduling status
		`^a list of existing lessons with scheduling status "([^"]*)" and teaching method "([^"]*)"$`: s.UserCreateLessonsWithSchedulingStatusAndTeachingMethod,
		`^the lessons scheduling status are updated correctly to "([^"]*)"$`:                          s.lessonsSchedulingStatusAreUpdatedCorrectlyTo,
		// group teaching method
		`^user creates a new lesson with "([^"]*)" teaching method and all required fields$`:               s.UserCreateALessonWithTeachingMethodAndAllRequiredFields,
		`^user creates a new lesson with "([^"]*)" teaching method and all required fields in lessonmgmt$`: s.UserCreateALessonWithTeachingMethodAndAllRequiredFieldsInLessonmgmt,
		`^a class with id prefix "([^"]*)" and a course with id prefix "([^"]*)"$`:                         s.aClassWithIDPrefixAndACourseWithIDPrefix,
		`^the lesson have course\'s teaching time info$`:                                                   s.TheLessonsHaveCorrectCourseTeachingTimeInfo,
		`^register some course\'s teaching time$`:                                                          s.RegisterSomeCourseTeachingTime,

		`^user get list student subscriptions$`:               s.userGetListStudentSubscriptions,
		`^got list student subscriptions$`:                    s.gotListStudentSubscriptions,
		`^user get list student subscriptions in lessonmgmt$`: s.userGetListStudentSubscriptionsInLessonmgmt,
		`^got list student subscriptions in lessonmgmt$`:      s.gotListStudentSubscriptionsInLessonmgmt,

		// delete_lesson.feature
		`^user deletes a lesson$`:                                      s.userDeleteALesson,
		`^user no longer sees any lesson report belong to the lesson$`: s.userNoLongerSeesTheLessonReport,
		`^user no longer sees the lesson$`:                             s.userNoLongerSeesTheLesson,

		`^user still sees lesson report belong to the lesson$`: s.userStillSeesTheLessonReport,
		`^user still sees the lesson$`:                         s.userStillSeesTheLesson,

		`^some centers$`:                         s.someCenters,
		`^some teacher accounts with school id$`: s.CommonSuite.CreateTeacherAccounts,
		`^some student accounts with school id$`: s.CommonSuite.CreateStudentAccounts,
		`^some courses with school id$`:          s.CommonSuite.SomeCourse,
		`^some student subscriptions$`:           s.CommonSuite.SomeStudentSubscriptions,

		// Sync Student Course
		`^an existing student$`: s.anExistingStudent,
		`^assigning "([^"]*)" course packages and location (\d+) to existing students$`: s.assignCoursePackagesWithStateToExistingStudents,
		`^sync student subscription successfully$`:                                      s.syncStudentSubscriptionSuccessfully,
		`^an existing "([^"]*)" lesson of location (\d+)$`:                              s.anExistingTypeLesson,
		`admin create a lesson report`:                                                  s.adminCreateALessonReport,
		`^edit "([^"]*)" course package with location (\d+)$`:                           s.editCoursePackageLocation,
		`^lesson member state is "([^"]*)"$`:                                            s.lessonMemberDeletedAtState,
		`^lesson report state is "([^"]*)"$`:                                            s.lessonReportDeletedAtState,
		`^"([^"]*)" Unleash feature with feature name "([^"]*)"$`:                       s.UnleashSuite.ToggleUnleashFeatureWithName,

		// live lesson room's state
		`user get current material state of live lesson room is empty$`:  s.userGetCurrentMaterialStateOfLiveLessonRoomIsEmpty,
		`^user get current material state of live lesson room is pdf$`:   s.userGetCurrentMaterialStateOfLiveLessonRoomIsPdf,
		`^user get current material state of live lesson room is video$`: s.userGetCurrentMaterialStateOfLiveLessonRoomIsVideo,
		`^user share a material with type is pdf in live lesson room$`:   s.UserShareAMaterialWithTypeIsPdfInLiveLessonRoom,
		`^user share a material with type is video in live lesson room$`: s.UserShareAMaterialWithTypeIsVideoInLiveLessonRoom,
		`^user signed as student who belong to lesson$`:                  s.userSignedAsStudentWhoBelongToLesson,
		`^user stop sharing material in live lesson room$`:               s.UserStopSharingMaterialInLiveLessonRoom,

		`^user fold a learner\'s hand in live lesson room$`:                   s.userFoldALearnersHandInLiveLessonRoom,
		`^user fold hand all learner$`:                                        s.UserFoldHandAllLearner,
		`^user get all learner\'s hands up states who all have value is off$`: s.userGetAllLearnersHandsUpStatesWhoAllHaveValueIsOff,
		`^user get hands up state$`:                                           s.userGetHandsUpState,
		`^user raise hand in live lesson room$`:                               s.UserRaiseHandInLiveLessonRoom,
		`^user hand off in live lesson room$`:                                 s.UserHandOffInLiveLessonRoom,

		`^user enables annotation learners in the live lesson room$`:  s.UserEnableAnnotationInLiveLessonRoom,
		`^user disables annotation learners in the live lesson room$`: s.UserDisableAnnotationInLiveLessonRoom,
		`^user get annotation state$`:                                 s.userGetAnnotationState,

		`^user start polling in live lesson room$`:                      s.UserStartPollingInLiveLessonRoom,
		`^user submit the answer "([^"]*)" for polling$`:                s.UserSubmitPollingAnswerInLiveLessonRoom,
		`^user stop polling in live lesson room$`:                       s.UserStopPollingInLiveLessonRoom,
		`^user end polling in live lesson room$`:                        s.UserEndPollingInLiveLessonRoom,
		`^user get current polling state of live lesson room started$`:  s.userGetCurrentPollingStateOfLiveLessonRoomStarted,
		`^user get current polling state of live lesson room stopped$`:  s.userGetCurrentPollingStateOfLiveLessonRoomStopped,
		`^user get current polling state of live lesson room is empty$`: s.userGetCurrentPollingStateOfLiveLessonRoomIsEmpty,
		`^user get polling answer state$`:                               s.userGetPollingAnswerState,

		`^live lesson is not recording$`:                                         s.liveLessonIsNotRecording,
		`^user get current recording live lesson permission to start recording$`: s.userGetCurrentRecordingLiveLessonPermissionToStartRecording,
		`^user have no current recording live lesson permission$`:                s.userHaveNoCurrentRecordingLiveLessonPermissionToStartRecording,
		`^user request recording live lesson$`:                                   s.UserRequestRecordingLiveLesson,
		`^"([^"]*)" user signed in as teacher$`:                                  s.CommonSuite.ASignedInTeacherWithOrdinalNumberInTeacherList,
		`^user stop recording live lesson$`:                                      s.UserStopRecordingLiveLesson,
		`^live lesson is still recording$`:                                       s.liveLessonIsStillRecording,

		`^user join live lesson$`:                   s.userJoinLiveLesson,
		`^returns valid information for broadcast$`: s.returnsValidInformationForBroadcast,
		`^have a uncompleted log with "([^"]*)" joined attendees, "([^"]*)" times getting room state, "([^"]*)" times updating room state and "([^"]*)" times reconnection$`: s.haveAUncompletedVirtualClassRoomLog,
		`^have a completed log with "([^"]*)" joined attendees, "([^"]*)" times getting room state, "([^"]*)" times updating room state and "([^"]*)" times reconnection$`:   s.haveACompletedVirtualClassRoomLog,
		`^returns valid information for student\'s broadcast$`: s.returnsValidInformationForStudentBroadcast,
		`^user end live lesson$`:                               s.EndLiveLesson,
		`^user enable spotlight for student$`:                  s.enableSpotlight,
		`^user get spotlighted user$`:                          s.userGetSpotlightedUser,
		`^user disable spotlight for student$`:                 s.disableSpotlight,

		`^a signed in "([^"]*)" with a school$`: s.aSignedInWithRandomID,
		`^a list of lessons management tab "([^"]*)" from "([^"]*)" to "([^"]*)" of school are existed in DB$`:                                                                                                          s.aListOfLessonManagementOfSchoolAreExistedInDB,
		`^a list of lessons management of school (\d+) are existed in DB$`:                                                                                                                                              s.aListOfLessonsManagementOfSchoolAreExistedInDB,
		`^Lessonmgmt must return list of "([^"]*)" lessons management from "([^"]*)" to "([^"]*)" and page limit "([^"]*)" with next page offset "([^"]*)" and pre page offset "([^"]*)"$`:                              s.lessonmgmtMustReturnListLessonManagement,
		`^a signed in "([^"]*)" with school: (\d+)$`:                                                                                                                                                                    s.CommonSuite.ASignedInWithSchool,
		`^Lessonmgmt must return correct list lesson management tab "([^"]*)" with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`: s.lessonmgmtMustReturnCorrectListLessonManagementWith,
		`^a course type with id "([^"]*)" and school id (\d+)$`:                                                                                                                                                         s.aCourseTypeWithIDAndSchoolID,

		// Get Lesson By ID
		`^user get detail lesson$`:                 s.RetrieveLessonByID,
		`^the lesson detail match lesson created$`: s.LessonMatchWithLessonCreated,

		// get students list by lesson id
		`^student list students in that lesson$`: s.studentListInThatLesson,
		`^returns a list of students$`:           s.returnsAListOfStudents,

		// get medias by lesson
		`^teacher get medias of lesson$`:                 s.teacherGetMediaOfLesson,
		`^the list of media match with response medias$`: s.theListOfMediaMatchWithResponseMedias,

		// students subscriptions
		`^(\d+) student subscriptions of school (\d+) are existed in DB with enrollment status "([^"]*)" and range date from "([^"]*)" to "([^"]*)"$`:                                                                                       s.aListStudentSubscriptionsAreExistedInDBWithEnrollmentStatus,
		`^(\d+) student subscriptions of school (\d+) are existed in DB with "([^"]*)" and "([^"]*)"$`:                                                                                                                                      s.aListStudentSubscriptionsAreExistedInDB,
		`^(\d+) student subscriptions of school (\d+) are existed in DB with "([^"]*)" and "([^"]*)" using user basic info table$`:                                                                                                          s.studentSubscriptionsOfSchoolAreExistedInDBWithAndUsingUserBasicInfoTable,
		`^admin retrieve student subscriptions limit (\d+), offset (\d+), lesson date "([^"]*)" with filter "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)" in lessonmgmt$`:                                             s.adminRetrieveStudentSubscriptionsInLessonmgmt,
		`^admin retrieve student subscriptions limit (\d+), offset (\d+), lesson date "([^"]*)" with filter "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)"$`:                                                                      s.adminRetrieveStudentSubscriptions,
		`^Bob must return list of (\d+) student subscriptions from "([^"]*)" to "([^"]*)" and page limit (\d+) with next page offset "([^"]*)" and pre page offset "([^"]*)" and around lesson date "([^"]*)" and filter "([^"]*)"$`:        s.bobMustReturnListStudentSubscriptions,
		`^Lessonmgmt must return list of (\d+) student subscriptions from "([^"]*)" to "([^"]*)" and page limit (\d+) with next page offset "([^"]*)" and pre page offset "([^"]*)" and around lesson date "([^"]*)" and filter "([^"]*)"$`: s.lessonmgmtMustReturnListStudentSubscriptions,
		`^a list of lesson_student_subscription_access_path are existed in DB$`:                                                                                                                                                             s.aListStudentSubscriptionAccessPathExistedInDB,

		// lesson management
		`a date "([^"]*)", location "([^"]*)", date type "([^"]*)", open time "([^"]*)", status "([^"]*)", resource path "([^"]*)" are existed in DB$`:                                                               s.aDateInfoExistedInDB,
		`^user updated lesson location "([^"]*)", start time "([^"]*)", end time "([^"]*)"$`:                                                                                                                         s.userUpdatedLocationAndLessonTime,
		`^a signed in "([^"]*)" with school random id$`:                                                                                                                                                              s.aSignedInWithRandomID,
		`^admin get live lesson management tab "([^"]*)" with page "([^"]*)" and "([^"]*)"$`:                                                                                                                         s.adminRetrieveLiveLessonManagement,
		`^admin get live lesson management tab "([^"]*)" with filter "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`: s.adminRetrieveLiveLessonManagementWithFilterOnBob,
		`^Bob must return list of "([^"]*)" lessons management from "([^"]*)" to "([^"]*)" and page limit "([^"]*)" with next page offset "([^"]*)" and pre page offset "([^"]*)"$`:                                  s.bobMustReturnListLessonManagement,
		`^Bob must return correct list lesson management tab "([^"]*)" with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`:     s.bobMustReturnCorrectListLessonManagementWith,
		// update
		`^user updates "([^"]*)" in the lesson lessonmgmt$`:                           s.userUpdatesFieldInTheLessonLessonmgmt,
		`^the lesson was updated in lessonmgmt$`:                                      s.TheLessonWasUpdatedInLessonmgmt,
		`^user updates the lesson with start time later than end time in lessonmgmt$`: s.userUpdateALessonWithStartTimeLaterThanEndTimeInLessonmgmt,
		`^user updates the lesson with missing "([^"]*)" in lessonmgmt$`:              s.UserUpdatesCurrentLessonWithMissingFieldInLessonmgmt,

		// lesson report
		`^a form\'s config for "([^"]*)" feature with school id$`: s.AFormConfigForFeature,
		`^a lesson report$`: s.UserSubmitANewLessonReport,
		`^the new saved draft lesson report existed in DB`: s.LessonmgmtHaveANewDraftLessonReport,
		`^the new submitted lesson report existed in DB`:   s.LessonmgmtHaveANewSubmittedLessonReport,
		`^user has been granted "([^"]*)" permission`:      s.userHasBeenGrantedPermission,
		// lesson report - individual
		`^user saves a new draft individual lesson report$`:       s.UserSaveDraftLessonReportIndividual,
		`^user submits the created individual lesson report$`:     s.UserSubmitLessonReportIndividual,
		`^user saves to update a draft individual lesson report$`: s.UserSaveDraftLessonReportIndividual,
		// lesson report - group
		`^user saves a new draft group lesson report$`:       s.UserSaveDraftLessonReportGroup,
		`^user submits the created group lesson report$`:     s.UserSubmitLessonReportGroup,
		`^user saves to update a draft group lesson report$`: s.UserSaveDraftLessonReportGroup,
		// attendance info
		`^students have attendance info "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`: s.StudentHaveAttendanceInfo,
		// update scheduling status
		`^"([^"]*)" must have event from "([^"]*)" to "([^"]*)"$`: s.mustHaveEventFromStatusToAfterStatus,
		`^the lesson scheduling status updates to "([^"]*)"$`:     s.LessonSchedulingStatusUpdatesTo,
		`^updates scheduling status in the lesson is "([^"]*)"$`:  s.updatesStatusInTheLessonIsValue,

		`^user creates recurring lesson$`:                                                                       s.createRecurringLesson,
		`^the recurring lesson was created successfully$`:                                                       s.hasRecurringLesson,
		`^the recurring lesson is "([^"]*)"$`:                                                                   s.theRecurringLessonIs,
		`^existing lesson event subscriber$`:                                                                    s.ValidateLessonCreatedSubscriptionInLessonmgmt,
		`^submit many lesson reports from many lessons recurring "([^"]*)"$`:                                    s.UserSubmitANewLessonReportFromLessonRecurring,
		`^user deletes recurring lesson from "([^"]*)" with "([^"]*)"$`:                                         s.userDeleteLessonRecurring,
		`^user no longer sees any lessons report belong to the lessons from many lessons recurring "([^"]*)"$`:  s.userNoLongerSeesTheLessonReportRecurring,
		`^user no longer sees the lessons from many lessons recurring "([^"]*)"$`:                               s.userNoLongerSeesTheLessonRecurring,
		`^user still sees the lesson from many lessons recurring "([^"]*)"$`:                                    s.userStillSeesTheLessonRecurring,
		`^user still sees lesson report belong to the lesson from many lessons recurring "([^"]*)"$`:            s.userStillSeesTheLessonReportRecurring,
		`^user have created recurring lesson$`:                                                                  s.createRecurringLesson,
		`^user update selected lesson by saving "([^"]*)"$`:                                                     s.updateLessonBySaving,
		`^user have created recurring lesson with zoom info$`:                                                   s.createRecurringLessonWithZoomInfo,
		`^user update selected lesson by saving weekly recurrence$`:                                             s.updateLessonBySavingWeekly,
		`^user save "([^"]*)" selected lesson by saving weekly recurrence$`:                                     s.saveLessonByStatus,
		`^user changed lesson time to "([^"]*)" , "([^"]*)" and "([^"]*)"$`:                                     s.userChangedLessonTimeTo,
		`^user changed lesson location to "([^"]*)" and "([^"]*)"$`:                                             s.userChangedLocationTo,
		`^the selected lesson & all of the followings were updated and link with new recurring chain$`:          s.selectedAndFollowingLessonUpdated,
		`^the locked lessons were not updated "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" or deleted$`:           s.checkLockedLessonInThisAndFollowingLesson,
		`^end date of old chain was updated$`:                                                                   s.haveUpdatedEndDate,
		`^user changed lesson general info to "([^"]*)"$`:                                                       s.userChangedLessonInfoTo,
		`^the selected lesson & all of the followings were updated$`:                                            s.selectedAndFollowingLessonUpdated,
		`^user update selected lesson by saving only this$`:                                                     s.updateLessonBySavingOnlyThis,
		`^the selected lesson was updated$`:                                                                     s.selectedLessonUpdated,
		`^the selected lesson still keep in old chain$`:                                                         s.selectedLessonKeepChain,
		`^the selected lesson was updated in new scheduler$`:                                                    s.selectedLessonLeaveChain,
		`^admin retrieve live lesson management tab "([^"]*)" with page "([^"]*)" and "([^"]*)" on lessonmgmt$`: s.adminRetrieveLessonManagementOnLessonmgmt,
		`^admin retrieve live lesson management tab "([^"]*)" with filter "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" on lessonmgmt$`: s.adminRetrieveLiveLessonManagementWithFilterOnLessonmgmt,
		`^user creates a "([^"]*)" lesson with "([^"]*)" in "([^"]*)"$`: s.userCreateLesson,
		`^user creates "([^"]*)" recurring lesson with "([^"]*)"$`:      s.createRecurringLessonWithMissingFields,
		`^user edit lesson by saving "([^"]*)" in "([^"]*)"$`:           s.userUpdateBySavingTo,
		`^user changed lesson general info$`:                            s.updateLessonRequestDefault,

		// student course slot info sync
		`^a message is published to student course event sync$`: s.aMessageIsPublishedToStudentCourseEventSync,
		`^receive student course slot info successfully$`:       s.receiveStudentCourseSlotInfoSuccessfully,

		`^user creates a new lesson with "([^"]*)", "([^"]*)", and other required fields in lessonmgmt$`: s.UserCreateANewLessonWithDateLocationAndOtherRequiredFieldsInLessonmgmt,
		`^user creates recurring lesson with "([^"]*)" and "([^"]*)" until "([^"]*)"$`:                   s.createRecurringLessonWithDateAndLocationUntilEndDate,
		`^recurring lessons will include "([^"]*)" and skip "([^"]*)"$`:                                  s.recurringLessonsWillInclude,
		`^returns some lesson dates$`: s.returnSomeLessonDates,
		`^user changed lesson time to "([^"]*)", "([^"]*)", "([^"]*)" and location "([^"]*)"$`:                    s.userChangedLessonTimeAndLocation,
		`^user creates a new lesson with student attendance info "([^"]*)", "([^"]*)", "([^"]*)", and "([^"]*)"$`: s.UserCreateANewLessonWithStudentAttendanceInfoInLessonMgmt,
		`^an existing lesson with student attendance info "([^"]*)", "([^"]*)", "([^"]*)", and "([^"]*)"$`:        s.AnExistingLessonWithStudentAttendanceInfo,
		`^user updates lesson student attendance info to "([^"]*)", "([^"]*)", "([^"]*)", and "([^"]*)"$`:         s.UserUpdatesLessonStudentAttendanceInfoTo,
		`^user marks student as reallocate$`:       s.MarkStudentAsReallocate,
		`^the attendance info is correct$`:         s.TheAttendanceInfoIsCorrect,
		`^the attendance info is updated$`:         s.TheAttendanceInfoIsUpdated,
		`^student attendance status is "([^"]*)"$`: s.TheStudentAttendanceStatusIs,
		`^locks lesson$`:                           s.LocksLesson,

		`^user creates a new lesson with "([^"]*)" classrooms$`:     s.UserCreateANewLessonWithClassrooms,
		`^user creates recurring lesson with "([^"]*)" classrooms$`: s.CreateRecurringLessonWithClassrooms,
		`^an existing lesson with classroom$`:                       s.AnExistingLessonWithClassroom,
		`^an existing recurring lesson with classroom$`:             s.AnExistingRecurringLessonWithClassroom,
		`^user updates lesson classroom with "([^"]*)" record$`:     s.UserUpdatesLessonClassroomWithRecord,
		`^user changed lesson with "([^"]*)" classroom$`:            s.UserChangedLessonWithClassroom,
		`^the classrooms are "([^"]*)" in the lesson$`:              s.TheClassroomsAreInTheLesson,
		`^the classrooms are "([^"]*)" in the recurring lesson$`:    s.TheClassroomsAreInTheRecurringLesson,
		`^the lesson classrooms are "([^"]*)" in the lesson$`:       s.TheClassroomsAreInTheLesson,
		`^the lesson classrooms are "([^"]*)"$`:                     s.TheLessonClassroomsAreUpdated,
		`^the selected lesson classroom is "([^"]*)"$`:              s.TheSelectedLessonClassroomIsUpdated,
		`^the other lessons classroom are "([^"]*)"$`:               s.TheOtherLessonsClassroomAre,

		`^an existing "([^"]*)" timesheet for current staff$`:  s.anExistingTimesheetForCurrentStaff,
		`^timesheet has lesson records with "([^"]*)"$`:        s.createLessonRecords,
		`^current staff approves this timesheet$`:              s.approvesThisTimesheet,
		`^timesheet status changed to approve "([^"]*)"$`:      s.timesheetStatusChangedToApprove,
		`^update lock lesson successfully$`:                    s.updateLockLessonSuccessfully,
		`^the lesson is locked "([^"]*)"$`:                     s.lockLesson,
		`^the lesson has start_time is "([^"]*)" will locked$`: s.lockLessonAt,
		`^the lesson "([^"]*)" will locked$`:                   s.lockLessons,

		`^have some student subscriptions with "([^"]*)" and "([^"]*)"$`: s.CommonSuite.SomeStudentSubscriptionsWithParams,
		`^user changed lesson student info to "([^"]*)", "([^"]*)"$`:     s.userChangedStudentInfoTo,

		`^user change status to "([^"]*)" by saving "([^"]*)"$`: s.userChangedStatusTo,

		// assigned student list
		`a list of students from "([^"]*)" to "([^"]*)" purchase package (\d+) days with method "([^"]*)" of location "([^"]*)" are existed in DB`:                                             s.aListOfPurchasedStudentAreExistedInDB,
		`^admin get assigned student list from "([^"]*)" tab on page "([^"]*)" with page limit of "([^"]*)"$`:                                                                                  s.adminGetAssignedStudentListNoFilter,
		`^admin get assigned student list from "([^"]*)" tab with page limit of "([^"]*)" and filters "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`: s.adminGetAssignedStudentListWithFilter,
		`^must return assigned student list with correct total and offset values$`:                                                                                                             s.mustReturnCorrectAssignedStudentList,
		`^must return correct assigned student list with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`:                                                                    s.mustReturnCorrectAssignedStudentListWithFilters,
		`^users get student attendance$`:                       s.getStudentAttendance,
		`^users get student attendance with filter "([^"]*)"$`: s.getStudentAttendanceWithFilter,

		// student change class subscription
		`^have some classes assign to courses$`:   s.createClassesAndAssignToCourses,
		`^have some lessons assign to classes$`:   s.createSomeLessonsToClass,
		`^a student add courses$`:                 s.studentJoinSomeClass,
		`^student must join to lessons of class$`: s.checkStudentJoinClass,
		`^student change other class$`:            s.studentChangeOtherClass,

		// update student course duration
		`^user added course to student$`:                                                      s.addCourseToStudent,
		`^user updates student course duration$`:                                              s.updateStudentCourseDuration,
		`^inactive student was removed from lesson$`:                                          s.checkInactiveStudent,
		`^student change duration$`:                                                           s.studentChangeDurationClass,
		`^student must leave to lessons have a start time less than class end time duration$`: s.checkStudentLeaveClass,

		// for retrieve lessons on calendar
		`^user creates a set of lessons for "([^"]*)"$`:                                                           s.UserCreatesASetOfLessonsFor,
		`^a list of lessons are existing$`:                                                                        s.aListOfLessonsAreExisting,
		`^user retrieves lessons on "([^"]*)" from "([^"]*)" to "([^"]*)"$`:                                       s.UserRetrievesLessonsOnFromTo,
		`^lessons retrieved are within the date range "([^"]*)" to "([^"]*)"$`:                                    s.LesonsRetrievedAreWithinTheDateRangeTo,
		`^the lessons first and last date matches with "([^"]*)" and "([^"]*)"$`:                                  s.TheLessonsFirstAndLastDateMatchesWithAnd,
		`^user retrieves lessons on "([^"]*)" with filter "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`: s.UserRetrievesLessonsOnWithFilter,
		`^lessons retrieved on calendar with filter are correct$`:                                                 s.LessonsRetrievedOnCalendarWithFilterAreCorrect,
		// steps from usermgmt
		`^create new student account$`:                                                s.createNewStudentAccount,
		`^only student info with first name last name and phonetic name$`:             s.studentInfoWithFirstNameLastNameAndPhoneticName,
		`^student account data to update with first name lastname and phonetic name$`: s.studentAccountDataToUpdateWithFirstNameLastNameAndPhoneticName,
		`^update student account$`:                                                    s.updateStudentAccount,
		`^assign student to a student subscription$`:                                  s.assignStudentToStudentSubscriptions,
		`^assign student to a lesson$`:                                                s.assignStudentToALesson,
		`^student name is updated correctly`:                                          s.studentNameIsUpdatedCorrectly,

		// for replicate data from calendar db
		`^user creates a scheduler to the calendar database$`:         s.createSchedulerToCalendarDB,
		`^scheduler data in bob db has synced successfully$`:          s.schedulerSynced,
		`^user creates a list of date info to the calendar database$`: s.createDateInfoToCalendarDB,
		`^date info data in bob db has synced successfully$`:          s.dateInfoSynced,

		// for replicate data from bob to lessonmgmt db
		`^lesson data in lessonmgmt db has synced successfully$`:     s.lessonsSyncedToLessonmgmt,
		`^lesson teachers in lessonmgmt db has synced successfully$`: s.lessonTeachersSyncedToLessonmgmt,
		`^lesson members in lessonmgmt db has synced successfully$`:  s.lessonMembersSyncedToLessonmgmt,

		`^user retrieve partner domain type "([^"]*)"$`: s.userGetPartnerDomain,

		// for reallocate student
		`^user retrieve all students that pending reallocation$`: s.retrieveStudentsPendingReallocation,
		`^return all student reallocate correctly$`:              s.returnCorrectReallocateStudent,
		`^some student assigned with reallocate status$`:         s.createLessonWithReallocateStudent,
		`^have some student subscriptions v2$`:                   s.SomeStudentSubscriptions,

		// for replicate student course data from fatima to bob db
		`^prepares data for create one time package$`:                   s.PrepareRequestForCreateOrderOneTimePackage,
		`^user creates a student course in order management$`:           s.UserCreatesAStudentCourseInOrderManagement,
		`^student course data in bob database has synced successfully$`: s.StudentCourseDataInBobDBSyncSuccessfully,
		`^have zoom account owner$`:                                     s.UpsertValidZoomAccount,
		`^user creates a new lesson with zoom link$`:                    s.UserCreateALessonZoomWithAllRequiredFields,
		`^user creates a new lesson with class do link$`:                s.UserCreateALessonClassDoWithAllRequiredFields,
		`^user gets the lesson detail with class do link$`:              s.UserGetsALessonClassDo,

		// export classroom
		`^returns classrooms in csv with "([^"]*)" columns$`: s.returnsClassroomsInCsv,
		`^user export classrooms$`:                           s.exportClassrooms,

		// classrooms
		`^user gets classrooms of locations "([^"]*)"$`: s.userGetClassroomsOfLocations,
		`^the list classrooms of these locations$`:      s.theListClassroomOfLocations,

		// import classroom
		`^a valid classrooms payload$`:                          s.avalidClassroomRequestPayload,
		`^importing classrooms$`:                                s.importingClassrooms,
		`^the valid classroom lines are imported successfully$`: s.theValidClassroomLinesAreImportedSuccessfully,
		`^an invalid classrooms "([^"]*)" request payload$`:     s.anInvalidClassroomRequestPayload,
		`^the invalid classrooms must returned with error$`:     s.theInvalidClassroomMustReturnedError,

		// for student attendance tab
		`^the list lesson members have returned correctly$`: s.returnCorrectStudentAttendance,

		// import/export lessons
		`^user download sample csv file to import lesson$`:        s.downloadLessonTemplate,
		`^returns a lesson template csv with columns: "([^"]*)"$`: s.returnsLessonCSVTemplate,
		`^a valid lessons payload$`:                               s.avalidLessonRequestPayload,
		`^a valid lessons payload v2$`:                            s.avalidLessonRequestPayloadV2,
		`^importing lessons$`:                                     s.importingLessons,
		`^the valid lessons lines are imported successfully$`:     s.theValidLessonsLinesAreImportedSuccessfully,
		`^an invalid lessons "([^"]*)" request payload$`:          s.anInvalidLessonRequestPayload,
		`^the invalid lessons must returned with error$`:          s.theInvalidLessonMustReturnedError,

		// import zoom account
		`^the zoom account request payload with "([^"]*)"$`: s.theZoomAccountRequestPayloadWith,
		`^importing zoom account$`:                          s.importingZoomAccount,

		// import course location schedule
		`^the course location schedule request payload with "([^"]*)"$`: s.theCourseLocationScheduleRequestPayloadWith,
		`^importing course location schedule$`:                          s.importingCourseLocationSchedule,
		`^user export course location schedule`:                         s.exportCourseLocationSchedule,
		`^returns course location schedule in csv with Ok status code`:  s.returnsCourseLocationScheduleInCsv,

		//export zoom account
		`^have some zoom account`:                            s.haveSomeZoomAccounts,
		`^user export zoom accounts$`:                        s.exportZoomAccounts,
		`^returns zoom accounts in csv with Ok status code$`: s.returnsZoomAccountInCsv,

		`^user export course with teaching time info$`:    s.exportCoursesTeachingTime,
		`^returns course with teaching time info in csv$`: s.returnsCourseTeachingTimeInCsv,
		`^a valid course teaching time payload$`:          s.aValidCourseTeachingTimePayload,
		`^importing course teaching time$`:                s.importCourseTeachingTime,

		`^insert student class member$`:                      s.insertStudentToClassMember,
		`^get student\'s courses and classes$`:               s.GetStudentsCoursesAndClasses,
		`^must get correct courses and classes of students$`: s.MustGetCorrectCoursesAndClassOfStudents,

		`^user get students by email or name: "([^"]*)"$`: s.GetStudentsManyByEmailOrName,
		`^got list students by email or name: "([^"]*)"$`: s.GotListStudentsByEmailOrName,

		// import classdo account
		`^user imports ClassDo accounts with "([^"]*)" data$`: s.userImportsClassDoAccountsWithData,
		`^user imports ClassDo accounts with delete action$`:  s.userImportsClassDoAccountsWithDeleteAction,
		`^ClassDo accounts are "([^"]*)" in the database$`:    s.classDoAccountsAreInTheDatabase,

		// export classdo account
		`^have some imported ClassDo accounts$`: s.haveSomeImportedClassDoAccounts,
		`^user export ClassDo accounts$`:        s.exportClassDoAccounts,

		// get classdo account
		`^has a ClassDo account$`:         s.CommonSuite.HasAClassDoAccount,
		`^user gets Class Do by User ID$`: s.userGetsClassDoByUserID,
		`^got expected ClassDo account$`:  s.userGotExpectedClassDoAccount,
	}
	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

func isMatchLessonID(expectedID, actualID string) error {
	if expectedID != actualID {
		return fmt.Errorf("expected lesson %s but got %s", expectedID, actualID)
	}
	return nil
}

func contextWithToken(s *Suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}
