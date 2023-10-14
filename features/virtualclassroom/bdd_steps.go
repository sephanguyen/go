package virtualclassroom

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
		`^everything is OK$`:             s.everythingIsOK,
		`^health check endpoint called$`: s.healthCheckEndpointCalled,
		`^virtual classroom should return "([^"]*)" with status "([^"]*)"$`: s.virtualClassroomShouldReturnWithStatus,
		`^returns "([^"]*)" status code$`:                                   s.returnsStatusCode,
		// background data
		`^enter a school$`:    s.enterASchool,
		`^have some centers$`: s.someCenters,

		// common suite
		`^"([^"]*)" signin system$`:                        s.signedAsAccountV2,
		`^user signed in as school admin$`:                 s.aSignedInAsSchoolAdmin,
		`^user signed in as teacher$`:                      s.aSignedInTeacherV2,
		`^have some teacher accounts$`:                     s.CreateTeacherAccounts,
		`^have some student accounts$`:                     s.CreateStudentAccounts,
		`^have "([^"]*)" student accounts$`:                s.CreateStudentNumberAccounts,
		`^have parent accounts for students$`:              s.CreateParentAccountsForStudents,
		`^student and parent accounts have device tokens$`: s.UpdateDeviceTokenForLeanerUser,
		`^an existing "([^"]*)" user signin system$`:       s.anExistingUserSigninSystem,

		`^have "([^"]*)" student accounts with "([^"]*)" first name and "([^"]*)" last name$`: s.CreateStudentNumberAccountsWithName,
		`^have some teacher accounts with "([^"]*)" first name and "([^"]*)" last name$`:      s.CreateTeacherAccountsWithName,

		`^a signed in admin$`:               s.CommonSuite.ASignedInAdmin,
		`^a signed in student$`:             s.CommonSuite.ASignedInStudent,
		`^a random number$`:                 s.CommonSuite.ARandomNumber,
		`^have some courses$`:               s.CommonSuite.SomeCourse,
		`^have some student subscriptions$`: s.CommonSuite.SomeStudentSubscriptions,
		`^have some medias$`:                s.CommonSuite.UpsertValidMediaList,

		// unleash
		`^"([^"]*)" Unleash feature with feature name "([^"]*)"$`: s.UnleashSuite.ToggleUnleashFeatureWithName,

		// modify virtual classroom state - material
		`^an existing a virtual classroom session$`:                s.anExistingVirtualClassroomWithWait,
		`^user join a virtual classroom session in bob$`:           s.userJoinVirtualClassRoomInBob,
		`^user join a virtual classroom session$`:                  s.userJoinVirtualClassRoomInVirtualClassroom,
		`^user signed as student who belong to lesson$`:            s.userSignedAsStudentWhoBelongToLesson,
		`^user stop sharing material in virtual classroom$`:        s.UserStopSharingMaterialInVirtualClassRoom,
		`^user stop sharing material in virtual classroom in bob$`: s.UserStopSharingMaterialInVirtualClassRoomInBob,

		// material video
		`^user share a material with type is video in virtual classroom$`:                  s.userShareAMaterialWithTypeIsVideoInVirtualClassRoom,
		`^user get current material state of a virtual classroom session is video$`:        s.userGetCurrentMaterialStateOfVirtualClassRoomIsVideo,
		`^user share a material with type is video in virtual classroom in bob$`:           s.userShareAMaterialWithTypeIsVideoInVirtualClassRoomInBob,
		`^user get current material state of a virtual classroom session is video in bob$`: s.userGetCurrentMaterialStateOfVirtualClassRoomIsVideoInBob,

		// material pdf
		`^user share a material with type is pdf in a virtual classroom session$`: s.userShareAMaterialWithTypeIsPdfInVirtualClassRoom,
		`^user get current material state of a virtual classroom session is pdf$`: s.userGetCurrentMaterialStateOfVirtualClassRoomIsPdf,

		// material audio
		`^user share a material with type is audio in virtual classroom$`:                  s.userShareAMaterialWithTypeIsAudioInVirtualClassRoom,
		`^user get current material state of a virtual classroom session is audio$`:        s.userGetCurrentMaterialStateOfVirtualClassRoomIsAudio,
		`^user share a material with type is audio in virtual classroom in bob$`:           s.userShareAMaterialWithTypeIsAudioInVirtualClassRoomInBob,
		`^user get current material state of a virtual classroom session is audio in bob$`: s.userGetCurrentMaterialStateOfVirtualClassRoomIsAudioInBob,

		// material empty
		`^user get current material state of a virtual classroom session is empty$`:        s.userGetCurrentMaterialStateOfLiveLessonRoomIsEmpty,
		`^user get current material state of a virtual classroom session is empty in bob$`: s.userGetCurrentMaterialStateOfLiveLessonRoomIsEmptyInBob,

		`^have a uncompleted log with "([^"]*)" joined attendees, "([^"]*)" times getting room state, "([^"]*)" times updating room state and "([^"]*)" times reconnection$`: s.haveAUncompletedVirtualClassRoomLog,

		`^user fold a learner\'s hand in a virtual classroom session$`:        s.userFoldALearnersHandInVirtualClassRoom,
		`^user fold hand all learner$`:                                        s.UserFoldHandAllLearner,
		`^user get all learner\'s hands up states who all have value is off$`: s.userGetAllLearnersHandsUpStatesWhoAllHaveValueIsOff,
		`^user get hands up state$`:                                           s.userGetHandsUpState,
		`^user raise hand in a virtual classroom session$`:                    s.UserRaiseHandInVirtualClassRoom,
		`^user hand off in a virtual classroom session$`:                      s.UserHandOffInVirtualClassRoom,

		`^user enables annotation learners in a virtual classroom session$`:      s.UserEnableAnnotationInVirtualClassRoom,
		`^user disables annotation learners in a virtual classroom session$`:     s.UserDisableAnnotationInVirtualClassRoom,
		`^user disables all annotation learners in a virtual classroom session$`: s.UserDisableAllAnnotationInVirtualClassRoom,
		`^user get annotation state$`:                                            s.userGetAnnotationState,
		`^all annotation state is disable$`:                                      s.AllAnnotationStateIsDisable,

		`^user start polling in a virtual classroom session$`:                                                      s.UserStartPollingInVirtualClassroom,
		`^user start polling with "([^"]*)" options and "([^"]*)" correct answers in a virtual classroom session$`: s.UserStartPollingInVirtualClassroomWithNumOption,
		`^user submit the answer "([^"]*)" for polling$`:                                                           s.UserSubmitPollingAnswerInVirtualClassroom,
		`^user stop polling in a virtual classroom session$`:                                                       s.UserStopPollingInVirtualClassroom,
		`^user end polling in a virtual classroom session$`:                                                        s.UserEndPollingInVirtualClassroom,
		`^user get current polling state of a virtual classroom session started$`:                                  s.userGetCurrentPollingStateOfVirtualClassroomStarted,
		`^user get current polling state of a virtual classroom session stopped$`:                                  s.userGetCurrentPollingStateOfVirtualClassroomStopped,
		`^user get current polling state of a virtual classroom session is empty$`:                                 s.userGetCurrentPollingStateOfVirtualClassroomIsEmpty,
		`^user get polling answer state$`:                                                                          s.userGetPollingAnswerState,
		`^user zoom whiteboard at in a virtual classroom session$`:                                                 s.userZoomWhiteboardInVirtualClassRoom,
		`^user get zoom whiteboard state$`:                                                                         s.userGetWhiteboardZoomState,
		`^user get zoom whiteboard state with the default value$`:                                                  s.userGetWhiteboardZoomStateWithDefaultValue,
		`^user start to recording$`:                                                                                s.userStartRecording,
		`^start recording state is updated$`:                                                                       s.startRecordingStateIsUpdated,
		`^user get start recording state$`:                                                                         s.userGetStartRecordingState,
		`^user stop recording$`:                                                                                    s.userStopRecording,
		`^stop recording state is updated$`:                                                                        s.stopRecordingStateIsUpdated,
		`^user get stop recording state$`:                                                                          s.userGetStopRecordingState,
		`^recorded videos are saved$`:                                                                              s.recordedVideosAreSaved,
		`^user get recorded videos on BO lesson detail$`:                                                           s.userGetRecordedVideos,
		`^must return a list recorded video$`:                                                                      s.mustReturnAListRecordedVideo,
		`^user download each recorded videos$`:                                                                     s.userDownloadEachRecordedVideo,
		`^user deletes a lesson$`:                                                                                  s.userDeleteALesson,
		`^Lessonmgmt must push msg "([^"]*)" subject "([^"]*)" to nats$`:                                           s.CommonSuite.LessonmgmtMustPushMsgSubjectToNats,
		`^media and recorded video will be deleted in db and cloud storage$`:                                       s.recordedVideoWillBeDeletedInDBAndCloudStorage,
		`^agora callback$`:                                                                                         s.AgoraCallback,
		`^a valid Agora signature in its header$`:                                                                  s.aValidAgoraSignatureInItsHeader,
		`^a request exit recording service$`:                                                                       s.requestExistRecordingService,
		`^user "([^"]*)" share polling in a virtual classroom session$`:                                            s.UserSharePollingInVirtualClassroom,
		`^user get current polling state of a virtual classroom session "([^"]*)" share polling$`:                  s.userGetCurrentPollingStateOfVirtualClassroomSharePolling,
		`^user end the live lesson$`:                                                                               s.userEndTheLiveLesson,
		`^user end the live lesson in bob$`:                                                                        s.userEndTheLiveLessonInBob,

		// join live lesson in virtual classroom
		`^"([^"]*)" receives room ID and other tokens$`: s.userReceivesRoomIDAndOtherTokens,

		// leave live lesson in virtual classroom
		`^user leaves the current virtual classroom session$`: s.userLeaveVirtualClassRoomInVirtualClassroom,

		// retrieve whiteboard token
		`^user retrieves whiteboard token$`:            s.userRetrievesWhiteboardToken,
		`^user receives room ID and whiteboard token$`: s.userReceivesRoomIDAndWhiteboardToken,
		`^lesson does not have existing room ID$`:      s.lessonDoesNotHaveExistingRoomID,

		// modify room state - spotlight
		`^user "([^"]*)" a spotlight user$`:   s.userSpotlightAUser,
		`^user gets correct spotlight state$`: s.userGetsCorrectSpotlightState,

		// prepare to publish
		`^current lesson has "([^"]*)" streaming learner$`: s.currentLessonHasStreamingLearner,
		`^user prepares to publish$`:                       s.userPreparesToPublish,
		`^user gets "([^"]*)" publish status$`:             s.userGetsPublishStatus,
		`^current lesson "([^"]*)" streaming learner$`:     s.currentLessonIncludeStreamingLearner,

		// unpublish
		`^user unpublish$`:                       s.userUnpublish,
		`^user gets "([^"]*)" unpublish status$`: s.userGetsUnpublishStatus,

		// get live lessons
		`^existing a virtual classroom sessions$`:                                        s.existingVirtualClassrooms,
		`^user gets live lessons$`:                                                       s.userGetsLiveLessonsWithoutFilter,
		`^user gets live lessons with filters$`:                                          s.userGetsLiveLessonsWithFilters,
		`^user gets live lessons with paging only$`:                                      s.userGetsLiveLessonsWithPagingOnly,
		`^"([^"]*)" receives live lessons$`:                                              s.userReceivesLiveLessons,
		`^"([^"]*)" receives live lessons that matches with the filters in the request$`: s.userReceivesLiveLessonsThatMatchesWithTheFilters,

		// get list of learners in lesson
		`^students have enrollment status$`:                   s.studentsHaveEnrollmentStatus,
		`^user gets list of learners in lesson$`:              s.userGetsListOfLearnersInLesson,
		`^returns a list of students with enrollment status$`: s.returnsAListOfStudentsWithEnrollmentStatus,

		// get list of learners from lessons
		`^user gets list of learners from lessons$`: s.userGetsListOfLearnersFromLessons,
		`^returns a list of students$`:              s.returnsAListOfStudents,

		// modify chat learner state
		`^user "([^"]*)" chat of learners in a virtual classroom session$`: s.userUpdatesChatOfLearnersInVirtualClassroom,
		`^user gets learners chat permission to "([^"]*)"$`:                s.userGetsLearnersChatPermission,
		`^user gets learners chat permission to "([^"]*)" with wait$`:      s.userGetsLearnersChatPermissionWithWait,

		// modify virtual classroom - upsert session time
		`^user modifies the session time in the live lesson$`: s.userModifiesTheSessionTimeInTheLiveLesson,
		`^user gets session time in the live lesson$`:         s.userGetsSessionTimeInTheLiveLesson,

		// nats consumer - create live lesson room
		`^user creates a virtual classroom session$`: s.userCreatesAVirtualClassroomSession,
		`^lesson has an existing room ID with wait$`: s.lessonHasExistingRoomIDWithWait,
		`^lesson has an existing room ID$`:           s.lessonHasExistingRoomID,

		// upcoming live lesson notification
		`^user creates a live lesson with start time after "([^"]*)" for newly created students$`:     s.createLiveLessonWithInterval,
		`^students have the country "([^"]*)"$`:                                                       s.updateStudentsCountryTo,
		`^wait for upcoming live lesson notification cronjob to run$`:                                 s.waitForCronJobToSendLiveLessonNotificationEvent,
		`^live lesson participants should receive notifications with message for "([^"]*)" interval$`: s.participantsShouldReceiveNotificationsWithProperIntervalMessage,

		// live room - join
		`^user joins a new live room$`:                              s.userJoinsANewLiveRoom,
		`^user joins an existing live room$`:                        s.userJoinsAnExistingLiveRoom,
		`^"([^"]*)" receives channel and room ID and other tokens$`: s.userReceivesChannelAndRoomIDAndOtherTokens,
		`^student who is part of the live room$`:                    s.studentWhoIsPartOfTheLiveRoom,

		// live room - leave
		`^user leaves the current live room$`: s.userLeavesTheCurrentLiveRoom,

		// live room - end
		`^user end the live room$`: s.userEndTheLiveRoom,

		// live room - get state
		`^user gets live room state$`:                 s.userGetsLiveRoomState,
		`^live room state is in default empty state$`: s.liveRoomStateIsInDefaultEmptyState,

		// live room - log
		`^have an uncompleted live room log with "([^"]*)" joined attendees, "([^"]*)" times getting room state, "([^"]*)" times updating room state and "([^"]*)" times reconnection$`: s.haveAnUncompletedLiveRoomLog,
		`^have a completed live room log with "([^"]*)" joined attendees, "([^"]*)" times getting room state, "([^"]*)" times updating room state and "([^"]*)" times reconnection$`:    s.haveACompletedLiveRoomLog,

		// live room - annotation
		`^user "([^"]*)" learners annotation in the live room$`:                             s.userModifiesLearnersAnnotationInTheLiveRoom,
		`^user gets the expected annotation state in the live room$`:                        s.userGetsExpectedAnnotationStateLiveRoom,
		`^user gets the expected annotation state in the live room is "([^"]*)"$`:           s.userGetsExpectedAnnotationStateLiveRoomIs,
		`^user gets the expected annotation state in the live room is "([^"]*)" with wait$`: s.userGetsExpectedAnnotationStateLiveRoomIsWithWait,
		`^user disables all annotation in the live room$`:                                   s.userDisablesAllAnnotationInTheLiveRoom,

		// live room - chat permission
		`^user "([^"]*)" learners chat permission in the live room$`:                       s.userModifiesLearnersChatPermissionInTheLiveRoom,
		`^user gets the expected chat permission state in the live room$`:                  s.userGetsExpectedChatPermissionStateLiveRoom,
		`^user gets the expected chat permission in the live room is "([^"]*)"$`:           s.userGetsExpectedChatPermissionStateLiveRoomIs,
		`^user gets the expected chat permission in the live room is "([^"]*)" with wait$`: s.userGetsExpectedChatPermissionStateLiveRoomIsWithWait,

		// live room - polling
		`^user start polling with "([^"]*)" options and "([^"]*)" correct answers in the live room$`: s.userStartPollingInTheLiveRoomWithNumOption,
		`^user start polling in the live room$`:                                                      s.userStartPollingInTheLiveRoom,
		`^user stop polling in the live room$`:                                                       s.userStopPollingInTheLiveRoom,
		`^user end polling in the live room$`:                                                        s.userEndPollingInTheLiveRoom,
		`^user "([^"]*)" sharing the polling in the live room$`:                                      s.userSharingThePollingInTheLiveRoom,
		`^user submit the answer "([^"]*)" in the live room polling$`:                                s.userSubmitPollingAnswerInTheLiveRoomPolling,

		`^user get current polling state in the live room has started$`:                        s.userGetCurrentPollingStateInTheLiveRoomHasStarted,
		`^user get current polling state in the live room has stopped$`:                        s.userGetCurrentPollingStateInTheLiveRoomHasStopped,
		`^user get current polling state in the live room is empty$`:                           s.userGetCurrentPollingStateInTheLiveRoomIsEmpty,
		`^user get polling answer state in the live room$`:                                     s.userGetPollingAnswerStateInTheLiveRoom,
		`^user get current polling state of the live room containing "([^"]*)" share polling$`: s.userGetCurrentPollingStateOfTheLiveRoomContainingSharePolling,

		// live room - user hand
		`^user "([^"]*)" hand in the live room$`:                            s.userModifiesHandInTheLiveRoom,
		`^user get hands up state in the live room$`:                        s.userGetHandsUpStateInTheLiveRoom,
		`^user get all learner\'s hands up states to off in the live room$`: s.userGetAllLearnersHandsUpStatesToOffInTHeLiveRoom,

		// live room - spotlight
		`^user "([^"]*)" a spotlighted user in the live room$`: s.userSpotlightAUserInTheLiveRoom,
		`^user gets correct spotlight state in the live room$`: s.userGetsCorrectSpotlightStateInTheLiveRoom,
		`^user gets empty spotlight in the live room$`:         s.userGetsEmptySpotlightInTheLiveRoom,

		// live room - whiteboard
		`^user zoom whiteboard in the live room$`:                                s.userZoomWhiteboardInTheLiveRoom,
		`^user gets whiteboard zoom state in the live room$`:                     s.userGetsWhiteboardZoomStateInTheLiveRoom,
		`^user gets whiteboard zoom state in the live room with default values$`: s.userGetsWhiteboardZoomStateDefaultInTheLiveRoom,

		// live room - share material
		`^user share a material with type is "([^"]*)" in the live room$`:  s.userShareAMaterialWithTypeInTheLiveRoom,
		`^user stop sharing material in the live room$`:                    s.userStopSharingMaterialInTheLiveRoom,
		`^user gets current material state of the live room is "([^"]*)"$`: s.userGetsCurrentMaterialStateOfLiveRoom,
		`^user gets empty current material state in the live room$`:        s.userGetsEmptyCurrentMaterialStateOfLiveRoom,

		// live room - get whiteboard token
		`^user gets whiteboard token for a new channel$`:             s.userGetsWhiteboardTokenForANewChannel,
		`^user gets whiteboard token for an existing channel$`:       s.userGetsWhiteboardTokenForAnExistingChannel,
		`^user receives whiteboard token and other channel details$`: s.userReceivesWhiteboardTokenAndOtherChannelDetails,
		`^the existing live room has no whiteboard room ID$`:         s.existingLiveRoomHasNoWhiteboardRoomID,

		// live room - recording
		`^user starts recording in the live room$`:                s.userStartsRecordingInTheLiveRoom,
		`^user starts recording in the live room only$`:           s.userStartsRecordingInTheLiveRoomOnly,
		`^user stops recording in the live room$`:                 s.userStopRecordingInTheLiveRoom,
		`^user stops recording in the live room only$`:            s.userStopRecordingInTheLiveRoomOnly,
		`^user gets the live room state recording has "([^"]*)"$`: s.userGetsTheLiveRoomStateRecordingHas,
		`^recorded videos are saved in the live room$`:            s.recordedVideosAreSavedInTheLiveRoom,

		// live room - prepare publish
		`^current live room has max streaming learner$`:                                      s.currentLiveRoomHasMaxStreamingLearner,
		`^user prepares to publish in the live room$`:                                        s.userPreparesToPublishInTheLiveRoom,
		`^user gets "([^"]*)" publish status in the live room$`:                              s.userGetsPublishStatusInTheLiveRoom,
		`^current live room "([^"]*)" streaming learner and gets "([^"]*)" streaming count$`: s.currentLiveRoomHasStreamingLearner,

		// live room - unpublish
		`^user unpublish in the live room$`:                       s.userUnpublishInTheLiveRoom,
		`^user gets "([^"]*)" unpublish status in the live room$`: s.userGetsUnpublishInTheLiveRoom,

		// live room - upsert session time
		`^user modifies the session time in the live room$`: s.userModifiesTheSessionTimeInTheLiveRoom,
		`^user gets session time in the live room$`:         s.userGetsSessionTimeInTheLiveRoom,

		// zego cloud - get auth token
		`^user gets authentication token for zegocloud$`:          s.userGetsAuthenticationTokenForZegoCloud,
		`^user gets authentication token for zegocloud using v2$`: s.userGetsAuthenticationTokenForZegoCloudUsingV2,
		`^user receives authentication token$`:                    s.userReceivesAuthenticationToken,
		`^user receives authentication token from v2$`:            s.userReceivesAuthenticationTokenFromV2,

		// zego cloud - get chat config
		`^user gets chat config for zegocloud$`: s.userGetsChatConfigForZegoCloud,
		`^user receives chat configurations$`:   s.userReceivesChatConfiguration,

		// get lessons - teacher web
		`^an existing set of lessons$`:                                       s.anExistingSetOfLessonsWithWait,
		`^user gets list of lessons on page "([^"]*)" with limit "([^"]*)"$`: s.userGetsListOfLessonsOnPageWithLimit,
		`^user gets list of lessons using "([^"]*)" and "([^"]*)"$`:          s.userGetsListOfLessonsUsingTimeCompareAndLookup,
		`^user gets list of lessons using "([^"]*)" with "([^"]*)"$`:         s.userGetsListOfLessonsUsingFilter,
		`^returns a list of lessons with the correct page info$`:             s.returnsAListOfLessonsWithTheCorrectPageInfo,
		`^returns a list of lessons based from the correct time$`:            s.returnsAListOfLessonsWithTheCorrectTime,
		`^returns a list of correct lessons based from the filter$`:          s.returnsAListOfCorrectLessonsBasedFromTheFilter,

		// stress test local
		`^user signed in as school admin for e2e$`: s.signedAdminAccountWithE2E,
		`^has a center for stress test$`:           s.hasACenterForStressTest,
		`^has a course for stress test$`:           s.hasACourseForStressTest,
		`^has a student for stress test$`:          s.hasAStudentForStressTest,
		`^has a teacher for stress test$`:          s.hasATeacherForStressTest,
		`^an existing set of "([^"]*)" lessons$`:   s.anExistingSetOfNumberLessonsWithWait,

		// chat service - get conversation id
		`^user already has existing private conversations with other "([^"]*)" "([^"]*)" users$`: s.useAlreadyHasExistingPrivateGroupOfConversationsWithUsers,
		`^user gets public conversation ID$`:                                                     s.userGetsPublicConversationID,
		`^user gets the same public conversation ID$`:                                            s.userGetsTheSamePublicConversationID,
		`^user gets private conversation ID with a "([^"]*)" user$`:                              s.userGetsPrivateConversationIDwithAUser,
		`^user gets the same private conversation ID with a "([^"]*)" user$`:                     s.userGetsTheSamePrivateConversationIDwithAUser,
		`^user gets non-empty conversation ID$`:                                                  s.userGetsNonEmptyConversationID,
		`^user gets the expected private conversation ID$`:                                       s.userGetsExpectedPrivateConversationID,
		`^user gets the expected public conversation ID$`:                                        s.userGetsExpectedPublicConversationID,

		// chat service - get private conversation ids
		`^user already has existing private conversation with one of the student accounts$`: s.userAlreadyHasExistingPrivConvWithOneOfTheStudentAcc,
		`^user gets private conversation IDs$`:                                              s.userGetsPrivateConversationIDsStep,
		`^user gets private conversation IDs again$`:                                        s.userGetsPrivateConversationIDsAgainStep,
		`^user gets non-empty private conversation IDs$`:                                    s.userGetsNonEmptyPrivateConversationIDs,
		`^user gets the expected private conversation IDs$`:                                 s.userGetsExpectedPrivateConversationIDs,
		`^user gets the expected one private conversation ID$`:                              s.userGetsTheExpectedOnvePrivateConversationID,

		// get user information
		`^user gets user information$`:              s.userGetsUserInformation,
		`^user receives expected user information$`: s.userReceivesExpectedUserInformation,

		// get classdo link
		`^has a ClassDo account$`:                            s.CommonSuite.HasAClassDoAccount,
		`^an existing lesson with a ClassDo link and owner$`: s.anExistingLessonWithAClassDoAccount,
		`^user gets the ClassDo link of a lesson$`:           s.userGetsALessonWithAClassDoAccount,
		`^returns the expected ClassDo link$`:                s.returnsTheExpectedClassDoLink,
	}

	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
