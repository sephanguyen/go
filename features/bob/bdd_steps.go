package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	golibs_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^waiting for "([^"]*)"$`:   s.waitingFor,
		`^"([^"]*)" signin system$`: s.signedAsAccountV2,

		`^an invalid authentication token$`:                          s.anInvalidAuthenticationToken,
		`^returns "([^"]*)" status code$`:                            s.returnsStatusCode,
		`^a valid authentication token with ID already exist in DB$`: s.aValidAuthenticationTokenWithIDAlreadyExistInDB,
		`^a valid authentication token with tenant$`:                 s.aValidAuthenticationTokenWithTenant,
		`^Bob must create an "([^"]*)" subscription`:                 s.bobMustCreateAnSubscription,
		`^student inputs activation code "([^"]*)"$`:                 s.studentInputsActivationCode,
		`^the student "([^"]*)" is "([^"]*)"$`:                       s.theStudentIs,
		`^a signed in admin$`:                                        s.aSignedInAdmin,

		`^a signed in student$`:                                        s.aSignedInStudent,
		`^user retrieves student profile with token id$`:               s.userRetrievesStudentProfileWithTokenId,
		`^status with error detail type "([^"]*)", subject "([^"]*)"$`: s.statusWithErrorDetailTypeSubject,

		// upload service
		`^user retrieves student profile$`:                             s.userRetrievesStudentProfile,
		`^user retrieves student profile v1$`:                          s.userRetrievesStudentProfileV1,
		`^an assigned student$`:                                        s.anAssignedStudent,
		`^returns requested student profile$`:                          s.returnsRequestedStudentProfile,
		`^returns requested student profile v1$`:                       s.returnsRequestedStudentProfileV1,
		`^an unassigned student$`:                                      s.anUnassignedStudent,
		`^returns empty list of student profile$`:                      s.returnsEmptyListOfStudentProfile,
		`^teacher retrieves a "([^"]*)" student profile v(\d+)$`:       s.teacherRetrievesAStudentProfileV1,
		`^student retrieves his\/her own profile$`:                     s.studentRetrievesHisherOwnProfile,
		`^returns his\/her own profile with status "([^"]*)"$`:         s.returnsHisherOwnProfile,
		`^an other student profile in DB$`:                             s.anOtherStudentProfileInDB,
		`^an identity platform account with existed account in db$`:    s.anIdentifyPlatformAccountWithExistedAccountInDB,
		`^student with onTrial = "([^"]*)" and "([^"]*)" billingDate$`: s.studentWithOnTrialAndBillingDate,

		// Question
		`^an unauthenticated user$`: s.anUnauthenticatedUser,

		// `^one "([^"]*)" image$`:                                 s.oneImage,
		// `^upload "([^"]*)" chunk size and "([^"]*)" file size$`: s.uploadChunkSizeAndFileSize,
		// `^url must be contain "([^"]*)"$`:                       s.urlMustBeContain,
		// `^bob must store image in s(\d+)$`:                      s.bobMustStoreImageInS,

		// master data
		`^upload preset_study_plan file$`:                     s.adminUploadPreset_study_planFile,
		`^bob must store all preset_study_plan data in db\.$`: s.bobMustStoreAllDataInDB,

		// preset study plans
		`^his owned student UUID$`:  s.hisOwnedStudentUUID,
		`^an invalid student UUID$`: s.anInvalidStudentUUID,

		// store device token
		`^user try to store device token$`:                                      s.userTryToStoreDeviceToken,
		`^a valid device token$`:                                                s.aValidDeviceToken,
		`^Bob must store the user\'s device token$`:                             s.bobMustStoreTheUsersDeviceToken,
		`^a device token with device_token empty$`:                              s.aDeviceTokenWithDevice_tokenEmpty,
		`^Bob must publish event to user_device_token channel$`:                 s.BobMustPublishEventToUserDeviceTokenChannel,
		`^Bob must store the user\'s device token to user_device_tokens table$`: s.bobMustStoreTheUsersDeviceTokenToUser_device_tokensTable,

		`^a user comment for his student$`:          s.aUserCommentForHisStudent,
		`^user upsert comment for his student$`:     s.userUpsertCommentForHisStudent,
		`^Bob must store comment for student$`:      s.bobMustStoreCommentForStudent,
		`^user retrieve comment for student$`:       s.userRetrieveCommentForStudent,
		`^Bob must return all comment for student$`: s.bobMustReturnAllCommentForStudent,
		`^valid comment for student in DB$`:         s.validCommentForStudentInDB,

		// Basic Profile
		`^a list user valid in db$`:                            s.aListUserValidInDB,
		`^search basic profile "([^"]*)" filter$`:              s.searchBasicProfile,
		`^update a student name$`:                              s.updateAStudent,
		`^returns a list basic profile according search_text$`: s.searchBasicProfileMustReturnCorrectlyWithSearchText,
		`^returns a list basic profile correctly$`:             s.searchBasicProfileMustReturnCorrectly,

		`^user updates profile$`:                       s.userUpdatesProfile,
		`^a valid updates profile request$`:            s.aValidUpdatesProfileRequest,
		`^Bob must records student\'s profile update$`: s.bobMustRecordsStudentsProfileUpdate,
		`^Bob must not update student\'s profile$`:     s.bobMustNotUpdateStudentsProfile,
		`^a FindStudentRequest with "([^"]*)" phone$`:  s.aFindStudentRequestWithPhone,
		`^user retrieves profile student by phone$`:    s.userRetrievesProfileStudentByPhone,
		`^returns student profile own profile$`:        s.returnsStudentProfileOwnProfile,

		`^user updated his own "([^"]*)" profile$`: s.userUpdatedHisOwnProfile,
		`^user updated other "([^"]*)" profile$`:   s.userUpdatedOtherProfile,

		`^user ask Bob to do update$`:         s.userAskBobToDoUpdate,
		`^Bob must update "([^"]*)" profile$`: s.bobMustUpdateProfile,

		`^a profile of "([^"]*)" user with usergroup: "([^"]*)", name: "([^"]*)", phone: "([^"]*)", email: "([^"]*)", school: (\d+)$`: s.aProfileOfUserWithUsergroupNamePhoneEmailSchool,
		`^a signed in user "([^"]*)" with school: (\d+)$`:          s.aSignedInUserWithSchool,
		`^event "([^"]*)" must be published to "([^"]*)" channel$`: s.eventMustBePublishedToChannel,
		`^profile of "([^"]*)" user must be updated$`:              s.profileOfUserMustBeUpdated,
		`^user update profile$`:                                    s.userUpdateProfile,

		`^student retrieves grade map$`: s.studentRetrievesGradeMap,
		`^Bob return all grade map$`:    s.bobReturnAllGradeMap,

		`^a user get client version$`:                   s.aUserGetClientVersion,
		`^bob must returns client version from config$`: s.bobMustReturnsClientVersionFromConfig,

		`^user retrieves his own profile$`:                        s.userRetrievesHisOwnProfile,
		`^Bob must returns user own profile$`:                     s.bobMustReturnsUserOwnProfile,
		`^Bob must push msg "([^"]*)" subject "([^"]*)" to nats$`: s.bobMustPushMsgSubjectToNats,
		`^a random number$`:                                       s.ARandomNumber,
		`^a random number in range (\d+)$`:                        s.ARandomNumberInRange,

		`^a school name "([^"]*)", country "([^"]*)", city "([^"]*)", district "([^"]*)"$`: s.aSchoolNameCountryCityDistrict,
		`^admin inserts schools$`: s.adminInsertsSchools,

		`^(\d+) valid preset study plan$`:            s.validPresetStudyPlan,
		`^user upsert preset study plan$`:            s.userUpsertPresetStudyPlan,
		`^Bob must store all preset study plan$`:     s.bobMustStoreAllPresetStudyPlan,
		`^returns a list of stored study plan$`:      s.returnsAListOfStoredStudyPlan,
		`^(\d+) valid preset study plan without Id$`: s.validPresetStudyPlanWithoutId,

		`^Bob must expire all "([^"]*)" subscription$`: s.bobMustExpireAllSubscription,

		`^a valid teacher profile$`:                    s.aValidTeacherProfile,
		`^a invalid teacher profile$`:                  s.aInvalidTeacherProfile,
		`^user retrieves teacher profile$`:             s.userRetrievesTeacherProfile,
		`^Bob must returns teacher profile$`:           s.bobMustReturnsTeacherProfile,
		`^Bob must returns teacher profile not found$`: s.bobMustReturnsTeacherProfileNotFound,

		`^a CreateClassRequest$`: s.aCreateClassRequest,
		`^CreateClassRequest has grade is "([^"]*)" and subject is "([^"]*)"$`: s.createClassRequestHasGradeIsAndSubjectIs,
		`^user create a class`:                            s.userCreateAClass,
		`^a signed in teacher$`:                           s.aSignedInTeacherV1,
		`^Bob must create class from CreateClassRequest$`: s.bobMustCreateClassFromCreateClassRequest,
		`^class must have (\d+) member is "([^"]*)" and is owner "([^"]*)" and status "([^"]*)"$`:     s.classMustHaveMemberIsAndIsOwnerAndStatus,
		`^a valid name in CreateClassRequest$`:                                                        s.aValidNameInCreateClassRequest,
		`^a (\d+) "([^"]*)" ownerId with school id is (\d+) in CreateClassRequest$`:                   s.aOwnerIdWithSchoolIdIsInCreateClassRequest,
		`^a "([^"]*)" schoolId in CreateClassRequest$`:                                                s.aSchoolIdInCreateClassRequest,
		`^default config for "([^"]*)" has "([^"]*)" is "([^"]*)"$`:                                   s.defaultConfigForHasIs,
		`^class must has "([^"]*)" is "([^"]*)"$`:                                                     s.classMustHasIs,
		`^this school has config "([^"]*)" is "([^"]*)", "([^"]*)" is "([^"]*)", "([^"]*)" is (\d+)$`: s.thisSchoolHasConfigIsIsIs,

		`^a EditClassRequest with class name is "([^"]*)"$`: s.aEditClassRequestWithClassNameIs,
		`^user edit a class$`:                               s.userEditAClass,
		`^a "([^"]*)" classId in EditClassRequest$`:         s.aClassIdInEditClassRequest,
		`^Bob must update class in db$`:                     s.bobMustUpdateClassInDB,

		`^a JoinClassRequest$`:                        s.aJoinClassRequest,
		`^user join a class$`:                         s.userJoinAClass,
		`^a "([^"]*)" classCode in JoinClassRequest$`: s.aClassCodeInJoinClassRequest,

		`^(\d+) "([^"]*)" join class with school id is (\d+)$`:        s.joinClassWithSchoolIdIs,
		`^(\d+) "([^"]*)" join class with school name "([^"]*)"$`:     s.joinClassWithSchoolName,
		`^a RemoveMemberRequest with class in school id is (\d+)$`:    s.aRemoveMemberRequestWithClassInSchoolIdIs,
		`^a RemoveMemberRequest with class in school name "([^"]*)"$`: s.aRemoveMemberRequestWithClassInSchoolName,
		`^user remove member from class$`:                             s.userRemoveMemberFromClass,
		`^a "([^"]*)" in RemoveMemberRequest$`:                        s.aInRemoveMemberRequest,
		`^Bob must store activity logs "([^"]*)" class member$`:       s.bobMustStoreActivityLogsClassMember,

		`^a valid token of current teacher$`: s.aValidTokenOfCurrentTeacher,

		`^a RetrieveClassMemberRequest$`:                                                  s.aRetrieveClassMemberRequest,
		`^user retrieve class member$`:                                                    s.userRetrieveClassMember,
		`^a valid classId in RetrieveClassMemberRequest$`:                                 s.aValidClassIDInRetrieveClassMemberRequest,
		`^returns (\d+) student\(s\) and (\d+) teacher\(s\) RetrieveClassMemberResponse$`: s.returnsStudentsAndTeachersRetrieveClassMemberResponse,
		`^returns list of student profile with empty "([^"]*)"$`:                          s.returnsListOfStudentProfileWithEmpty,

		`^student join teacher current class$`: s.studentJoinTeacherCurrentClass,

		// `^teacher edit topic assignment from his class$`:                                                               s.teacherEditTopicAssignmentFromHisClass,
		// `^a valid teacher edit topic assignment request with start date "([^"]*)" today and end date "([^"]*)" today$`: s.aValidTeacherEditTopicAssignmentRequestWithStartDateTodayAndEndDateToday,
		// `^assignment start date and due date must be changed$`:                                                         s.assignmentStartDateAndDueDateMustBeChanged,
		// `^teacher edit assignment$`:                                                                                    s.teacherEditAssignment,

		`^student retrieves courses$`:                        s.studentRetrievesCourses,
		`^a list of courses are existed in DB of "([^"]*)"$`: s.aListOfCoursesAreExistedInDBOf,
		`^returns a list of courses of "([^"]*)"$`:           s.returnsAListOfCoursesOf,
		`^student retrieves assigned courses$`:               s.studentRetrievesAssignedCourses,

		`^everything is OK$`:                                  s.everythingIsOK,
		`^health check endpoint called$`:                      s.healthCheckEndpointCalled,
		`^bob should return "([^"]*)" with status "([^"]*)"$`: s.bobShouldReturnWithStatus,

		`^a "([^"]*)" userId GetBasicProfileRequest$`:                  s.aUserIdGetBasicProfileRequest,
		`^user retrieves basic profile$`:                               s.userRetrievesBasicProfile,
		`^user retrieves basic profile with missing metadata$`:         s.userRetrievesBasicProfileWithMissingMetadata,
		`^Bob must returns (\d+) basic profile$`:                       s.bobMustReturnsBasicProfile,
		`^user cannot retrieves basic profile when missing "([^"]*)"$`: s.userCannotRetrievesBasicProfileWhenMissing,

		`^a "([^"]*)" userId RetrieveBasicProfileRequest$`:   s.aUserIdRetrieveBasicProfileRequest,
		`^a user retrieves basic profile$`:                   s.aUserRetrievesBasicProfile,
		`^a user retrieves basic profile without metadata$`:  s.aUserRetrievesBasicProfileWithoutMetadata,
		`^a Bob must returns (\d+) basic profile$`:           s.aBobMustReturnsBasicProfile,
		`^a user retrieves "([^"]*)" basic profile request$`: s.aUserRetrieveBasicProfileRequest,

		// region Notification
		`^a valid "([^"]*)" request$`:    s.aValidRequest,
		`^an invalid "([^"]*)" request$`: s.anInvalidRequest,
		`^request "([^"]*)" has page "([^"]*)" and limit "([^"]*)" and type "([^"]*)"$`: s.requestHasPageAndLimitAndType,
		`^user try to make "([^"]*)" request$`:                                          s.userTryToMakeRequest,
		// endregion notification
		`^user check an "([^"]*)" with empty value$`:     s.userCheckAnWithEmptyValue,
		`^user check an "([^"]*)" that "([^"]*)" in DB$`: s.userCheckAnThatInDB,
		`^that basic profile$`:                           s.thatBasicProfile,

		`^a signed in "([^"]*)" with school: (\d+)$`: s.aSignedInWithSchool,
		`^user updated "([^"]*)" profile with user group: "([^"]*)", name: "([^"]*)", phone: "([^"]*)", email: "([^"]*)", school: (\d+)$`: s.userUpdatedProfileWithUserGroupNamePhoneEmailSchool,
		`^a signed in "([^"]*)" with school name "([^"]*)"$`:                                                                              s.aSignedInWithSchoolName,

		`^a signed in of current "([^"]*)"$`: s.aSignedInOfCurrent,

		`^a student in class$`: s.aStudentInClass,
		// update class code
		`^create a class with school id and expired at "([^"]*)"$`:             s.createAClassWithSchoolIdIsAndExpiredAt,
		`^create a class with school name "([^"]*)" and expired at "([^"]*)"$`: s.createAClassWithSchoolNameAndExpiredAt,
		`^a UpdateClassCodeRequest with "([^"]*)" class id$`:                   s.aUpdateClassCodeRequestWithClassId,
		`^user updates a class$`:                                               s.userUpdatesAClass,
		`^Bob must update class code$`:                                         s.bobMustUpdateClassCode,

		`^a AddClassMemberRequest$`:                                                        s.aAddClassMemberRequest,
		`^a "([^"]*)" classCode in AddClassMemberRequest$`:                                 s.aClassCodeInAddClassMemberRequest,
		`^user add a class member$`:                                                        s.userAddAClassMember,
		`^a "([^"]*)" userID of school id is (\d+) in AddClassMemberRequest$`:              s.aUserIDOfSchoolIdIsInAddClassMemberRequest,
		`^a "([^"]*)" userID of school name "([^"]*)" in AddClassMemberRequest$`:           s.aUserIDOfSchoolNameInAddClassMemberRequest,
		`^student subscription must has "([^"]*)" is "([^"]*)" with plan id is "([^"]*)"$`: s.studentSubscriptionMustHasIsWithPlanIdIs,

		`^upload invalid preset_study_plan file$`:             s.uploadInvalidPreset_study_planFile,
		`^user retrieves student profile of classMembers$`:    s.userRetrievesStudentProfileOfClassMembers,
		`^user retrieves student profile of classMembers v1$`: s.userRetrievesStudentProfileOfClassMembersV1,

		// upsert_topic permission

		`^JoinClassResponse must return "([^"]*)"$`: s.joinClassResponseMustReturn,

		`^a teacher with valid lesson$`:                           s.aTeacherWithValidLesson,
		`^teacher retrieve stream token$`:                         s.teacherRetrieveStreamToken,
		`^student retrieve stream token$`:                         s.studentRetrieveStreamToken,
		`^a student with valid lesson$`:                           s.aStudentWithValidLesson,
		`^a student with valid lesson which has no room id$`:      s.aStudentWithValidLessonWhichHasNoRoomID,
		`^student retrieves assigned live courses$`:               s.studentRetrievesAssignedLiveCourses,
		`^student retrieves courses with ids$`:                    s.studentRetrievesCoursesWithIds,
		`^returns a list of courses from requested ids$`:          s.returnsAListOfCoursesFromRequestedIds,
		`^return lessons must have correct topic$`:                s.returnLessonsMustHaveCorrectTopic,
		`^return lessons must have correct teacher profile$`:      s.returnLessonsMustHaveCorrectTeacherProfile,
		`^return lessons must have correct status$`:               s.returnLessonsMustHaveCorrectStatus,
		`^student retrieve live lesson with invalid time period$`: s.studentRetrieveLiveLessonWithInvalidTimePeriod,
		`^Bob must return "([^"]*)" live lesson for student$`:     s.bobReturnResultLiveLessonForStudent,
		`^Bob must return "([^"]*)" live lesson for teacher$`:     s.bobReturnResultLiveLessonForTeacher,
		`^teacher retrieve live lesson with invalid time period$`: s.teacherRetrieveLiveLessonWithInvalidTimePeriod,

		`^a list of lessons are existed in DB of "([^"]*)" with start time "([^"]*)" and end time "([^"]*)"$`:   s.aListOfLessonsAreExistedInDBOfWithStartTimeAndEndTime,
		`^student retrieve live lesson with start time "([^"]*)" and end time "([^"]*)"$`:                       s.studentRetrieveLiveLessonWithStartTimeAndEndTime,
		`^student retrieve live lesson by courseID "([^"]*)" with start time "([^"]*)" and end time "([^"]*)"$`: s.studentRetrieveLiveLessonByCourseWithStartTimeAndEndTime,
		`^teacher retrieve live lesson with start time "([^"]*)" and end time "([^"]*)"$`:                       s.teacherRetrieveLiveLessonWithStartTimeAndEndTime,
		`^teacher retrieve live lesson by courseID "([^"]*)" with start time "([^"]*)" and end time "([^"]*)"$`: s.teacherRetrieveLiveLessonByCourseWithStartTimeAndEndTime,

		`^a list of lessons from "([^"]*)" to "([^"]*)" of school are existed in DB$`:                                                                                    s.aListOfLessonOfSchoolAreExistedInDB,
		`^admin retrieve live lesson with page "([^"]*)" and "([^"]*)"$`:                                                                                                 s.adminRetrieveLiveLesson,
		`^Bob must return list of "([^"]*)" lessons from "([^"]*)" to "([^"]*)" and page limit "([^"]*)" with next page offset "([^"]*)" and pre page offset "([^"]*)"$`: s.bobMustReturnListLesson,
		`^a list of lessons of school (\d+) are existed in DB$`:                                                                                                          s.aListOfLessonsOfSchoolAreExistedInDB,
		`^admin retrieve live lesson with filter "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`:                                                                     s.adminRetrieveLiveLessonWithFilter,
		`^Bob must return correct list lesson with "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)" "([^"]*)"$`:                                                                   s.bobMustReturnCorrectListLessonWith,

		// lesson management
		`^a signed in "([^"]*)" with school random id$`:                                                                                                                             s.aSignedInWithRandomID,
		`^a list of lessons management tab "([^"]*)" from "([^"]*)" to "([^"]*)" of school are existed in DB$`:                                                                      s.aListOfLessonManagementOfSchoolAreExistedInDB,
		`^admin retrieve live lesson management tab "([^"]*)" with page "([^"]*)" and "([^"]*)"$`:                                                                                   s.adminRetrieveLiveLessonManagement,
		`^Bob must return list of "([^"]*)" lessons management from "([^"]*)" to "([^"]*)" and page limit "([^"]*)" with next page offset "([^"]*)" and pre page offset "([^"]*)"$`: s.bobMustReturnListLessonManagement,
		`^a list of lessons management of school (\d+) are existed in DB$`:                                                                                                          s.aListOfLessonsManagementOfSchoolAreExistedInDB,

		`^admin retrieve live lesson management tab "([^"]*)" with filter "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`: s.adminRetrieveLiveLessonManagementWithFilter,
		`^Bob must return correct list lesson management tab "([^"]*)" with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)"$`:          s.bobMustReturnCorrectListLessonManagementWith,

		// import master data - location
		`^had "([^"]*)" record of the locations database$`:                s.aListOfLocationsInDB,
		`^user retrieve locations$`:                                       s.retrieveLocations,
		`^user retrieve lowest level of locations with filter "([^"]*)"$`: s.retrieveLowestLevelLocations,
		`^a list of locations with are existed in DB$`:                    s.aListOfLocationsInDB,
		`^a list of locations with variant types are existed in DB$`:      s.aListOfLocationsVariantTypesInDB,
		`^must return a correct list of locations$`:                       s.mustReturnCorrectLocations,
		`^must return lowest level of locations with filter "([^"]*)"$`:   s.mustReturnLowestLevelLocations,
		// import master data - location type
		`^a list of location types are existed in DB$`:   s.aListOfLocationTypesInDB,
		`^user retrieve location types$`:                 s.retrieveLocationTypes,
		`^must return correct a list of location types$`: s.mustReturnCorrectLocationTypes,

		`^a lesson$`:                                      s.aLesson,
		`^some centers$`:                                  s.someCenters,
		`^some courses with school id$`:                   s.CreateLiveCourse,
		`^some student subscriptions$`:                    s.someStudentSubscriptions,
		`^the lesson is updated$`:                         s.theLessonIsUpdated,
		`^user updates "([^"]*)" in the existing lesson$`: s.userUpdatesFieldInTheLesson,

		// lesson reports
		`^a list of lesson reports from "([^"]*)" to "([^"]*)" of school (\d+) are existed in DB$`: s.aListOfLessonReportsOfSchoolAreExistedInDB,
		`^user retrieve partner domain type "([^"]*)"$`:                                            s.userGetPartnerDomain,
		`^user create lesson report$`:                                                              s.UserCreateIndividualLessonReport,
		`^a form\'s config for "([^"]*)" feature with school id$`:                                  s.AFormConfigForFeature,
		`^Bob have a new lesson report$`:                                                           s.BobHaveANewLessonReport,
		`^Bob have a lesson report$`:                                                               s.BobHaveANewLessonReport,
		`^Bob have a lesson report with lesson is locked$`:                                         s.BobHaveANewLessonReportWithLessonIsLocked,
		`^user submit a new lesson report$`:                                                        s.UserSubmitANewLessonReport,
		`^"([^"]*)" must have event from "([^"]*)" to "([^"]*)"$`:                                  s.mustHaveEventFromStatusToAfterStatus,
		`^Bob have a new draft lesson report$`:                                                     s.BobHaveANewDraftLessonReport,
		`^Bob have a draft lesson report$`:                                                         s.BobHaveANewDraftLessonReport,
		`^Bob have a draft lesson report with lesson is locked$`:                                   s.BobHaveANewDraftLessonReportWithLessonIsLocked,
		`^user save a new draft lesson report$`:                                                    s.UserSaveANewDraftLessonReport,
		`^user submit to update a lesson report$`:                                                  s.UserSubmitToUpdateALessonReport,
		`^user save to update a draft lesson report$`:                                              s.UserSaveToUpdateADraftLessonReport,
		`^a lesson report$`:                                                                        s.UserSubmitANewLessonReport,
		`^the lesson scheduling status updates to "([^"]*)"$`:                                      s.LessonSchedulingStatusUpdatesTo,
		`^updates scheduling status in the lesson is "([^"]*)"$`:                                   s.userUpdatesStatusInTheLessonIsValue,
		`^lesson is locked by timesheet$`:                                                          s.lessonIsLockedByTimesheet,
		// lesson report - multi version form config
		`^user save a new draft lesson report with multi version feature name is "([^"]*)"$`: s.UserSaveANewDraftLessonReportWithMultiVersionFeatureNameIs,
		`^user submit a new lesson report with multi version feature name is "([^"]*)"$`:     s.UserSubmitANewLessonReportWithMultiVersionFeatureNameIs,

		// take the quiz v2
		`^teacher end one of the live lesson$`:                                s.teacherEndOneOfTheLiveLesson,
		`^the ended lesson must have status completed$`:                       s.theEndedLessonMustHaveStatusCompleted,
		`^teacher end one of the live lesson v1$`:                             s.teacherEndOneOfTheLiveLessonV1,
		`^the ended lesson must have status completed v1$`:                    s.theEndedLessonMustHaveStatusCompletedV1,
		`^bob must update lesson end at time v1$`:                             s.bobMustUpdateLessonEndAtTimeV1,
		`^student retrieves assigned courses with retrieve course end point$`: s.studentRetrievesAssignedCoursesWithRetrieveCourseEndPoint,
		`^student join lesson$`:                                               s.studentJoinLesson,
		`^student join lesson with v1 API$`:                                   s.studentJoinLessonV1,
		`^teacher join lesson$`:                                               s.teacherJoinLesson,
		`^teacher join lesson with v1 API$`:                                   s.teacherJoinLessonV1,
		`^a teacher with invalid lesson$`:                                     s.aTeacherWithInvalidLesson,
		`^a teacher from same school with valid lesson$`:                      s.aTeacherFromSameSchoolWithValidLesson,
		`^bob must update lesson end at time$`:                                s.bobMustUpdateLessonEndAtTime,
		`^returns valid information for broadcast$`:                           s.returnsValidInformationForBroadcast,
		`^returns valid information for broadcast with v1 API$`:               s.returnsValidInformationForBroadcastV1,
		`^returns valid information for subscribe$`:                           s.returnsValidInformationForSubscribe,
		`^returns a list of courses in current class$`:                        s.returnsAListOfCoursesInCurrentClass,
		`^student retrieves courses in current class$`:                        s.studentRetrievesCoursesInCurrentClass,
		`^student retrieves live courses with status "([^"]*)"$`:              s.studentRetrievesLiveCoursesWithStatus,
		`^student\'s class is removed from course$`:                           s.studentsClassIsRemovedFromCourse,
		`^returns empty list of course$`:                                      s.returnsEmptyListOfCourse,
		`^student retrieves assigned live courses belong to current class$`:   s.studentRetrievesAssignedLiveCoursesBelongToCurrentClass,
		`^student must receive lesson room id$`:                               s.StudentMustReceiveLessonRoomId,
		`^student must receive lesson room id same above$`:                    s.MustReceiveLessonRoomIDAfterJoinLessonWhichSameCurrentRoomID,
		`^student must receive lesson room id and tokens$`:                    s.studentMustReceiveLessonRoomIDAndTokens,
		`^student join lesson again$`:                                         s.studentJoinLesson,
		`^retrieve whiteboard token$`:                                         s.retrieveWhiteboardToken,
		`^receive whiteboard token$`:                                          s.receiveWhiteboardToken,

		`^a list of lesson are existed in DB$`:     s.aListOfLessonAreExistedInDB,
		`^a student in live lesson background$`:    s.aStudentInLiveLessonBackground,
		`^student leave lesson$`:                   s.studentLeaveLesson,
		`^student leave lesson for other student$`: s.studentLeaveLessonForOtherStudent,

		`^user create brightcove upload url for video "([^"]*)"$`: s.userCreateBrightcoveUploadUrlForVideo,
		`^bob must return a video upload url$`:                    s.bobMustReturnAVideoUploadUrl,
		`^user finish brightcove upload url for video "([^"]*)"$`: s.userFinishBrightcoveUploadUrlForVideo,

		// retrieve books

		`^Bob must record all media list$`: s.bobMustRecordAllMediaList,
		`^user upsert valid media list$`:   s.userUpsertValidMediaList,
		`^Bob must return all media$`:      s.bobMustReturnAllMedia,
		`^student has multiple media$`:     s.studentHasMultipleMedia,
		`^student retrieve media by ids$`:  s.studentRetrieveMediaByIds,

		`^bob must return correct class ids$`: s.bobMustReturnCorrectClassIds,
		`^user retrieve class by ids$`:        s.userRetrieveClassByIds,

		// exchange token
		`^a user exchange token$`:                                      s.aUserExchangeToken,
		`^our system need to do return valid token$`:                   s.ourSystemNeedToDoReturnValidToken,
		`^our system need to do return valid school admin token$`:      s.ourSystemNeedToDoReturnValidSchoolAdminToken,
		`^an school admin profile in DB$`:                              s.anSchoolAdminProfileInDB,
		`^"([^"]*)" in "([^"]*)" logged in$`:                           s.inLoggedIn,
		`^"([^"]*)" in keycloak logged in$`:                            s.inKeycloakLoggedIn,
		`^"([^"]*)" receives valid exchanged token$`:                   s.receivesValidExchangedToken,
		`^"([^"]*)" uses id token to exchanges token with our system$`: s.usesIdTokenToExchangesTokenWithOurSystem,
		`^system init default values for auth info in "([^"]*)"$`:      s.systemInitDefaultValuesForAuthInfoIn,
		`^the initialized values must be valid$`:                       s.theInitializedValuesMustBeValid,

		`^a list of media$`:                                 s.aListOfMedia,
		`^user converts media to image$`:                    s.userConvertsMediaToImage,
		`^media conversion tasks must be created$`:          s.mediaConversionTasksMustBeCreated,
		`^a list of media conversion tasks$`:                s.aListOfMediaConversionTasks,
		`^finished conversion tasks must be updated$`:       s.finishedConversionTasksMustBeUpdated,
		`^our system receives a finished conversion event$`: s.ourSystemReceivesAFinishedConversionEvent,

		`^a list of media which attached to a lesson$`:               s.aListOfMediaWhichAttachedToALesson,
		`^the list of media match with response medias$`:             s.theListOfMediaMatchWithResponseMedias,
		`^user get lesson medias$`:                                   s.userGetLessonMedias,
		`^user get lesson medias and returns "([^"]*)" status code$`: s.userGetLessonMediasAndReturnsStatusCode,

		`^a ListCoursesRequest message "([^"]*)" school$`:                                s.aListCoursesRequestMessageSchool,
		`^a ListCoursesRequest with search keywork "([^"]*)"$`:                           s.aListCoursesRequestKeyword,
		`^returns courses in ListCoursesResponse matching filter of ListCoursesRequest$`: s.returnsCoursesInListCoursesResponseMatchingFilterOfListCoursesRequest,
		`^user list courses$`: s.userListCourses,

		`^above teacher belong to "([^"]*)" school$`: s.aboveTeacherBelongToSchool,

		// list course by locations
		`^a ListCoursesByLocationsRequest message "([^"]*)" school and keyword "([^"]*)" with (\d+) locations$`: s.aListCoursesbyLocationRequestMessageSchool,
		`^user list courses by locations$`:                                              s.userListCoursesByLocations,
		`^a list of course_access_paths are existed in DB$`:                             s.listCourseAccessPathExistedInDB,
		`^locations of courses matching with course_access_paths with (\d+) locations$`: s.locationsOfCoursesMatchingWithCourseAccessPath,

		`^a lesson with some lesson members$`:    s.aLessonWithSomeLessonMembers,
		`^returns a list of students$`:           s.returnsAListOfStudents,
		`^teacher list students in that lesson$`: s.teacherListStudentsInThatLesson,
		// exchange custom token
		`^a client exchange custom token$`:                  s.aClientExchangeCustomToken,
		`^our system need to returns a valid custom token$`: s.ourSystemNeedToDoReturnValidCustomToken,
		`^our system need to returns error$`:                s.ourSystemNeedToDoReturnError,
		`^user get media of non existed lesson$`:            s.userGetMediaOfNonExistedLesson,
		`^empty media result$`:                              s.emptyMediaResult,

		`^student list students in that lesson$`: s.studentListStudentsInThatLesson,
		// prepare publish
		`^a valid lesson in database$`:                                                          s.aValidLessonInDB,
		`^some valid learners in database$`:                                                     s.someValidLearnersInDatabase,
		`^a "([^"]*)" number of stream of the lesson$`:                                          s.aNumberOfStreamOfTheLesson,
		`^new record indicating that the learner is publishing an upload stream in the lesson$`: s.newRecordIndicatingThatTheLearnerIsPublishingAnUploadStreamInTheLesson,
		`^the arbitrary learner does not publishing any uploading stream in the lesson$`:        s.theAbitraryLearnerDoesNotPublishingAnyUploadingStreamInTheLesson,
		`^the learner prepare to publish$`:                                                      s.theLearnerPrepareToPublish,
		`^the number of stream of the lesson have to increasing$`:                               s.theNumberOfStreamOfTheLessonHaveToIncreasing,
		`^returns the response "([^"]*)"$`:                                                      s.returnsPreparePublishTheResponse,
		`^the learner is not allowed to publish any uploading stream in the lesson$`:            s.theLearnerIsNotAllowedToPublishAnyUploadingStreamInTheLesson,
		`^the number of stream of the lesson have to unchanged$`:                                s.theNumberOfStreamOfTheLessonHaveToNoChange,
		`^an arbitrary learner publishing an uploading stream in the lesson$`:                   s.anArbitraryLearnerPublishingAnUploadingStreamInTheLesson,
		`^the learner is still publishing an uploading stream in the lesson$`:                   s.theLearnerIsStillPublishingAnUploadingStreamInTheLesson,

		`^new record indicating that the "([^"]*)" learner is publishing an upload stream in the lesson$`:              s.newRecordIndicatingThatTheLearnerIsPublishingAnUploadStreamInTheLesson,
		`^returns "([^"]*)" status for both requests$`:                                                                 s.returnsStatusForBothRequests,
		`^the "([^"]*)" learner currently does not publish any uploading stream in the lesson$`:                        s.theLearnerCurrentlyDoesNotPublishAnyUploadingStreamInTheLesson,
		`^the lesson\'s learner counter have to increasing two$`:                                                       s.theLessonsLearnerCounterHaveToIncreasingTwo,
		`^two learners prepare to publish in concurrently$`:                                                            s.twoLearnersPrepareToPublishInConcurrently,
		`^new record indicating that the either first learner or second is publishing an upload stream in the lesson$`: s.newRecordIndicatingThatOneOfThemPublishing,
		`^returns "([^"]*)" for the user who is granted$`:                                                              s.returnsForTheUserWhoIsGranted,
		`^returns the response "([^"]*)" for the user who is rejected$`:                                                s.returnsTheResponseForTheUserWhoIsRejected,
		`^the lesson\'s learner counter have to maximum$`:                                                              s.theLessonsLearnerCounterHaveToMaximum,
		`^returns "([^"]*)" for the one$`:                                                                              s.returnsStatusForBothRequests,
		`^returns the response "([^"]*)" for another$`:                                                                 s.returnsTheResponseForTheUserWhoIsRejected,
		`^the learner prepare to publish twice in concurrently$`:                                                       s.theLearnersPrepareToPublishTwiceInConcurrently,
		`^a learner prepared publish in the lesson$`:                                                                   s.theLearnerPrepareToPublish,
		`^no record indicating that the "([^"]*)" learner is unpublish an upload stream in the lesson$`:                s.noRecordIndicatingThatTheLearnerIsUnpublishAnUploadStreamInTheLesson,
		`^the learner unpublish$`:                                                                                      s.theLearnerUnpublish,
		`^unpublish returns the response "([^"]*)"$`:                                                                   s.unpublishReturnsTheResponse,

		`^the number of stream of the lesson have to equal to "([^"]*)" number$`:                               s.theNumberOfStreamOfTheLessonHaveToEqualToNumber,
		`^two learner prepared publish in the lesson$`:                                                         s.twoLearnerPreparedPublishInTheLesson,
		`^no record indicating that the "([^"]*)" learner prepared to publish an upload stream in the lesson$`: s.noRecordIndicatingThatTheLearnerPreparedToPublishAnUploadStreamInTheLesson,
		`^record indicating the two learner prepared publish in the lesson$`:                                   s.recordIndicatingTheTwoLearnerPreparedPublishInTheLesson,
		`^unpublish returns "([^"]*)" status code for both requests$`:                                          s.returnsStatusForBothRequests,
		`^two learners unpublish in concurrently$`:                                                             s.twoLearnersUnpublishInConcurrently,
		`^returns OK for the one$`:                                                                             s.returnsOKForTheOne,
		`^the learner unpublish twice in concurrently$`:                                                        s.theLearnerUnpublishTwiceInConcurrent,
		`^unpublish returns the response "([^"]*)" for another$`:                                               s.unpublishReturnsTheResponseForAnother,
		`^get streaming learners$`:                                                                             s.getStreamingLearners,
		`^"([^"]*)" learner prepared publish in the lesson$`:                                                   s.learnerPreparedPublishInTheLesson,
		`^return "([^"]*)" learner ids, who are currently uploading in the lesson$`:                            s.returnLearnerIdsWhoAreCurrentlyUploadingInTheLesson,
		`^the number of stream of the lesson have to become "([^"]*)"$`:                                        s.theNumberOfStreamOfTheLessonHaveToBecome,
		`^the number of stream of the lesson have to decrease "([^"]*)"$`:                                      s.theNumberOfStreamOfTheLessonHaveToDecrease,
		`^a lesson with arbitrary number of student publishing$`:                                               s.aLessonWithArbitraryNumberOfStudentPublishing,
		`^students publish and unpublish as the same time$`:                                                    s.studentsPublishAndUnpublishAsTheSameTime,
		`^the number of publishing students must be record correctly$`:                                         s.theNumberOfPublishingStudentsMustBeRecordCorrectly,
		`^current student assigned to above lessons$`:                                                          s.currentStudentAssignedToAboveLessons,

		`^returns response of list courses have to correctly$`:          s.returnsResponseOfListCoursesHaveToCorrectly,
		`^some courses have icon url$`:                                  s.someCoursesHaveIconUrl,
		`^returns response for student list courses have to correctly$`: s.returnsResponseForStudentListCoursesHaveToCorrectly,

		`^all courses are belong to "([^"]*)" academicYear$`: s.allCoursesAreBelongToAcademicYear,
		`^returns no courses in ListCoursesResponse$`:        s.returnsNoCoursesInListCoursesResponse,

		`^a signed in user has a expiration time "([^"]*)" and a prefix name "([^"]*)"$`: s.aSignedInUserHasAExpirationTimeAndAPrefixName,
		`^file storage must store file if presigned url not yet expired$`:                s.fileStorageMustStoreFileIfPresignedUrlNotYetExpired,
		`^user get url to upload file$`:                                                  s.userGetUrlToUploadFile,
		`^return a presigned url to upload file and a expiration time "([^"]*)"$`:        s.returnAPresignedUrlToUploadFileAndAExpirationTime,
		`^return a status code "([^"]*)"$`:                                               s.returnAStatusCode,
		`^upload a file via a presigned url$`:                                            s.uploadAFileViaAPresignedUrl,
		`^user wait a interval "([^"]*)"$`:                                               s.userWaitAInterval,

		// offline learning

		`^our system have to returns a list of students correctly$`: s.ourSystemHaveToReturnsAListOfStudentsCorrectly,
		`^some students has been removed from the lesson$`:          s.someStudentsHasBeenRemovedFromTheLesson,

		`^a lesson with some lesson members given their name$`: s.aLessonWithSomeLessonMembersWithTheirName,

		`^a signed in user as a teacher$`:              s.aSignedInTeacher,
		`^our system returns class members correctly$`: s.ourSystemReturnsClassMembersCorrectly,
		`^some class members$`:                         s.someMembersInSomeClasses,
		`^some class$`:                                 s.generateTwoClasses,
		`^the teacher gets "([^"]*)" class members$`:   s.theTeacherRetrievesClassMembers,

		`^a student with some comments$`:                           s.aStudentWithSomeComments,
		`^our systems have to store comment correctly$`:            s.ourSystemsHaveToStoreCommentCorrectly,
		`^the teacher delete student\'s comment$`:                  s.theTeacherDeleteStudentsComment,
		`^a teacher gives some comments for student$`:              s.aTeacherGivesSomeCommentsForStudent,
		`^our system have to response retrieve comment correctly$`: s.ourSystemHaveToResponseRetrieveCommentCorrectly,
		`^the teacher retrieves comment for student$`:              s.theTeacherRetrievesCommentForStudent,

		`^our system need to do return valid "([^"]*)" exchanged token$`: s.ourSystemNeedToDoReturnValidExchangedToken,

		// delete_live_lesson.feature
		`^user signed as school admin$`: s.userSignedAsSchoolAdmin,
		`^user creates live lesson with "([^"]*)", "([^"]*)", teachers, courses, learners and start time, end time in the future$`: s.userCreateLiveLessonWithStartTimeEndTimeAtFuture,
		`^user deletes the live lesson$`:        s.userDeletesTheLiveLesson,
		`^user no longer sees the live lesson$`: s.userNoLongerSeesTheLiveLesson,
		`^user creates live lesson with "([^"]*)", "([^"]*)", "([^"]*)", teachers, courses learners and end time in the future$`: s.userCreateLiveLessonWithEndTimeAtFuture,
		`^user creates live lesson with "([^"]*)", "([^"]*)", teachers, courses, learners and start time, end time in the past$`: s.userCreateLiveLessonWithEndTimeAtPast,
		`^user can not delete the live lesson$`: s.userCanNotDeleteTheLiveLesson,

		// `^user create a learning objective$`: s.userCreateALearningObjective,
		// `^user upsert a topic$`:              s.useUpsertATopic,
		`^the live lesson has room_id$`: s.liveLessonHasRoomId,

		// update_live_lesson.feature
		`^some live courses with school id$`:                                 s.CreateLiveCourse,
		`^some medias$`:                                                      s.CreateMedias,
		`^some student accounts with school id$`:                             s.CreateStudentAccounts,
		`^some teacher accounts with school id$`:                             s.CreateTeacherAccounts,
		`^user signed as admin$`:                                             s.aSignedInAdmin,
		`^user signed in as admin$`:                                          s.aSignedInAdmin,
		`^an existing live lesson$`:                                          s.anExistingLiveLesson,
		`^an existing live lesson with status "([^"]*)"$`:                    s.anExistingLiveLessonWithStatus,
		`^user updates "([^"]*)" in the live lesson$`:                        s.userUpdatesFieldInTheLiveLesson,
		`^the live lesson is updated$`:                                       s.theLiveLessonIsUpdated,
		`^user updates the live lesson with start time later than end time$`: s.userUpdatesTheLiveLessonWithStartTimeLaterThanEndTime,
		`^user updates the live lesson with missing "([^"]*)"$`:              s.userUpdatesTheLiveLessonWithMissingField,
		`^the live lesson is not updated$`:                                   s.theLiveLessonIsNotUpdated,
		`^a new parent profile in DB$`:                                       s.aNewParentProfileInDB,

		`^a user signed in as a parent$`: s.aUserSignedInAsAParent,
		`^user calls "([^"]*)" API$`:     s.userCallsAPI,
		`^create handsome father as a parent and the relationship with his children who\'re students at manabie$`: s.createHandsomeFatherAsAParentAndTheRelationshipWithHisChildrenWhoreStudentsAtManabie,
		`^multiple students profile in DB$`:                         s.multipleStudentsProfileInDB,
		`^a signed in parent$`:                                      s.UserSignedInParent,
		`^retrieve students profiles associated to parent account$`: s.retrieveStudentsProfilesAssociatedToParentAccount,
		`^fetched students exactly associated to parent$`:           s.fetchedStudentsExactlyAssociatedToParent,

		`^generate presign url to put object$`:                  s.generatePresignUrlToPutObject,
		`^generate resumable upload url$`:                       s.generateResumableUploadUrl,
		`^return presign put object url$`:                       s.returnPresignPutObjectUrl,
		`^return resumable upload url$`:                         s.returnResumableUploadUrl,
		`^a file information to generate put object url$`:       s.aFileInformationToGeneratePutObjectUrl,
		`^a file information to generate resumable upload url$`: s.aFileInformationToGenerateResumableUploadUrl,
		`^the file can be uploaded using the returned url$`:     s.theFileCanBeUploadedUsingTheReturnedUrl,

		// live lesson room state
		`^user signed as teacher$`: s.SignedInTeacher,
		`^a live lesson$`:          s.aLiveLesson,
		`^user get current material state of live lesson room is empty$`:      s.userGetCurrentMaterialStateOfLiveLessonRoomIsEmpty,
		`^user get current material state of live lesson room is pdf$`:        s.userGetCurrentMaterialStateOfLiveLessonRoomIsPdf,
		`^user get current material state of live lesson room is video$`:      s.userGetCurrentMaterialStateOfLiveLessonRoomIsVideo,
		`^user share a material with type is pdf in live lesson room$`:        s.userShareAMaterialWithTypeIsPdfInLiveLessonRoom,
		`^user share a material with type is video in live lesson room$`:      s.userShareAMaterialWithTypeIsVideoInLiveLessonRoom,
		`^user signed as student who belong to lesson$`:                       s.userSignedAsStudentWhoBelongToLesson,
		`^user get all learner\'s hands up states who all have value is off$`: s.userGetAllLearnersHandsUpStatesWhoAllHaveValueIsOff,
		`^user get hands up state$`:                                           s.userGetHandsUpState,
		`^user raise hand in live lesson room$`:                               s.userRaiseHandInLiveLessonRoom,
		`^user "([^"]*)" chat of learners in a live lesson room$`:             s.userUpdatesChatOfLearnersInLiveLessonRoom,
		`^user gets learners chat permission to "([^"]*)"$`:                   s.userGetsLearnersChatPermission,

		`^live lesson is not recording$`:                                         s.liveLessonIsNotRecording,
		`^user get current recording live lesson permission to start recording$`: s.userGetCurrentRecordingLiveLessonPermissionToStartRecording,
		`^user request recording live lesson$`:                                   s.userRequestRecordingLiveLesson,
		`^"([^"]*)" user signed as teacher$`:                                     s.UserSignedInTeacher,

		// generate audio files and upload to cloud storage
		// update user last login date
		`^a "([^"]*)" signed in user with "([^"]*)"$`:        s.aSignedInUserWith,
		`^user last login date "([^"]*)"$`:                   s.userLastLoginDate,
		`^user update last login date with "([^"]*)" value$`: s.userUpdateLastLoginDateWithValue,

		// verify app version
		`^user verify app version receive force update request$`: s.userVerifyAppVersionReceiveForceUpdateRequest,
		`^user check app version$`:                               s.userCheckAppVersion,
		`^verify app version request$`:                           s.verifyAppVersionRequest,
		`^verify app version request missing "([^"]*)"$`:         s.verifyAppVersionRequestMissing,
		`^verify app version request with "([^"]*)"$`:            s.verifyAppVersionRequestWith,

		`create subscription to receive messages "([^"]*)"$`: s.createSubscriptionToReceiveMsg,

		// delete_lesson.feature
		`^user deletes a lesson$`:                                      s.userDeleteALesson,
		`^user no longer sees any lesson report belong to the lesson$`: s.userNoLongerSeesTheLessonReport,
		`^user no longer sees the lesson$`:                             s.userNoLongerSeesTheLesson,

		// Sync Student Subscriptions - test upsert
		`^a "([^"]*)" number of existing students$`:        s.aNumberOfExistingStudents,
		`^a "([^"]*)" number of existing courses$`:         s.aNumberOfExistingCourses,
		`^assigning course packages to existing students$`: s.assignCoursePackagesToExistingStudents,
		`^sync student subscription successfully$`:         s.syncStudentSubscriptionSuccessfully,
		// sync student Subscription - test update
		`^an existing course$`:  s.anExistingCourse,
		`^an existing student$`: s.anExistingStudent,
		`^edit course package with new start at "([^"]*)" and end at "([^"]*)"$`:                    s.editCoursePackageTime,
		`^sync student subscription with new start at "([^"]*)" and end at "([^"]*)" successfully$`: s.syncStudentSubscriptionWithNewStartAtAndEndAtSuccessfully,

		// get postgres users info
		`^a request to get postgres users info$`:                 s.aRequestToGetPostgresUserInfo,
		`^call get postgres users info API$`:                     s.callGetPostgresUserInfoAPI,
		`^postgres users info data must contains "([^"]*)"$`:     s.postgresUserInfoDataMustContain,
		`^a valid postgres user info key$`:                       s.aValidPostgresUserInfoKey,
		`^an invalid base64 format postgres user info key$`:      s.anInvalidBase64FormatPostgresUserInfoKey,
		`^an invalid RSA format postgres user info key$`:         s.anInvalidRSAFormatPostgresUserInfoKey,
		`^an invalid postgres user info key$`:                    s.anInvalidPostgresUserInfoKey,
		`^a request to get postgres privilege$`:                  s.aRequestToGetPostgresPrivilege,
		`^call get postgres privilege by API$`:                   s.callGetPostgresPrivilegeAPI,
		`^postgres privilege info data must contains "([^"]*)"$`: s.postgresMustHavePrivilege,

		// unleash client
		`^"([^"]*)" Unleash feature with feature name "([^"]*)"$`: s.ToggleUnleashFeatureWithName,
		`^a signed in as "([^"]*)"$`:                              s.aSignedIn,
		`^a signed in "([^"]*)"$`:                                 s.aSignedIn,
	}

	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

func (s *suite) returnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

func (s *suite) ReturnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	return s.returnsStatusCode(ctx, arg1)
}

func (s *suite) signedCtx(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}

//nolint:gocyclo
func (s *suite) bobMustPushMsgSubjectToNats(ctx context.Context, msg, subject string) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)

	// t, _ := jwt.ParseString(stepState.AuthToken)

	var (
		// handler  func(msg *stan.Msg)
		foundChn = make(chan struct{}, 1)
	)

	switch subject {
	// case constants.SubjectStudentLearning:
	// 	handler = func(m *stan.Msg) {
	// 		r := &pb.StudentLearning{}
	// 		err := r.Unmarshal(m.Data)
	// 		if err != nil {
	// 			return
	// 		}

	// 		parts := strings.Split(msg, "-")
	// 		if r.Event == pb.StudentLearningEvent(pb.StudentLearningEvent_value[parts[1]]) && r.StudentId == t.Subject() {
	// 			if stepState.CurrentPromotionID == r.PromotionId {
	// 				foundChn <- struct{}{}
	// 			}
	// 		}
	// 	}

	// case "student_learning_time":
	// 	handler = func(m *stan.Msg) {
	// 		r := &pb.StudentEventLogRequest{}
	// 		err := r.Unmarshal(m.Data)
	// 		if err != nil {
	// 			return
	// 		}
	// 		req := stepState.Request.(*pb.StudentEventLogRequest)
	// 		if len(req.StudentEventLogs) == len(r.StudentEventLogs) {
	// 			foundChn <- struct{}{}
	// 		}
	// 	}
	case golibs_constants.SubjectClassUpserted:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		for {
			select {
			case message := <-stepState.FoundChanForJetStream:
				switch message := message.(type) {
				case *pb.EvtClassRoom_CreateClass_:
					if msg == "CreateClass" {
						if message.CreateClass.ClassId == stepState.CurrentClassID && message.CreateClass.ClassName != "" {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_EditClass_:
					if msg == "EditClass" {
						if message.EditClass.ClassId == stepState.CurrentClassID && message.EditClass.ClassName != "" {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_JoinClass_:
					if msg == "JoinClass" {
						if message.JoinClass.ClassId == stepState.CurrentClassID {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_LeaveClass_:
					if msg == "LeaveClass" {
						return StepStateToContext(ctx, stepState), nil
					}
					if strings.Contains(msg, "LeaveClass") {
						if message.LeaveClass.ClassId == stepState.CurrentClassID && len(message.LeaveClass.UserIds) != 0 {
							if strings.Contains(msg, fmt.Sprintf("-is_kicked=%v", message.LeaveClass.IsKicked)) {
								return StepStateToContext(ctx, stepState), nil
							}
						}
					}
				case *pb.EvtClassRoom_ActiveConversation_:
					active := msg == "ActiveConversation"
					if message.ActiveConversation.ClassId == stepState.CurrentClassID && message.ActiveConversation.Active == active {
						return StepStateToContext(ctx, stepState), nil
					}
				}
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	// case constants.SubjectClassEventNats:
	// 	handler = func(m *stan.Msg) {
	// 		r := &pb.EvtClassRoom{}
	// 		err := r.Unmarshal(m.Data)
	// 		if err != nil {
	// 			return
	// 		}

	// 		switch r.Message.(type) {
	// 		case *pb.EvtClassRoom_CreateClass_:
	// 			if msg == "CreateClass" {
	// 				if r.GetCreateClass().ClassId == stepState.CurrentClassID && r.GetCreateClass().ClassName != "" {
	// 					foundChn <- struct{}{}
	// 				}
	// 			}
	// 		case *pb.EvtClassRoom_EditClass_:
	// 			if msg == "EditClass" {
	// 				if r.GetEditClass().ClassId == stepState.CurrentClassID && r.GetEditClass().ClassName != "" {
	// 					foundChn <- struct{}{}
	// 				}
	// 			}
	// 		case *pb.EvtClassRoom_JoinClass_:
	// 			if msg == "JoinClass" {
	// 				if r.GetJoinClass().ClassId == stepState.CurrentClassID {
	// 					foundChn <- struct{}{}
	// 				}
	// 			}
	// 		case *pb.EvtClassRoom_LeaveClass_:
	// 			if strings.Contains(msg, "LeaveClass") {
	// 				if r.GetLeaveClass().ClassId == stepState.CurrentClassID && len(r.GetLeaveClass().UserIds) != 0 {
	// 					if strings.Contains(msg, "-is_kicked=true") && !r.GetLeaveClass().IsKicked {
	// 						return
	// 					}

	// 					if strings.Contains(msg, "-is_kicked=false") && r.GetLeaveClass().IsKicked {
	// 						return
	// 					}

	// 					foundChn <- struct{}{}
	// 				}
	// 			}
	// 		case *pb.EvtClassRoom_ActiveConversation_:
	// 			active := msg == "ActiveConversation"
	// 			if r.GetActiveConversation().ClassId == stepState.CurrentClassID && r.GetActiveConversation().Active == active {
	// 				foundChn <- struct{}{}
	// 			}
	// 		}
	// 	}
	case constants.SubjectAllocateStudentQuestionAfter10SecondsNats:
	case constants.SubjectAllocateStudentQuestionAfter30SecondsNats:
	// case constants.SubjectAllocateStudentQuestionAfter60SecondsNats:
	// 	handler = func(m *stan.Msg) {
	// 		r := &pb.RetryWithDelayEvent{}
	// 		err := r.Unmarshal(m.Data)
	// 		if err != nil {
	// 			return
	// 		}
	// 		if r.GetEvtAllocateStudentQuestion().StudentQuestionId == stepState.CurrentQuestionID {
	// 			foundChn <- struct{}{}
	// 		}
	// 	}
	case golibs_constants.SubjectLessonCreated:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch message.(type) {
			case *pb.EvtLesson_CreateLessons_:
				if msg == "CreateLessons" {
					return StepStateToContext(ctx, stepState), nil
				}
			}
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	case golibs_constants.SubjectLessonUpdated:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch message.(type) {
			case *pb.EvtLesson_UpdateLesson_:
				if msg == "UpdateLesson" {
					return StepStateToContext(ctx, stepState), nil
				}
			case *pb.EvtLesson_EndLiveLesson_:
				if msg == "EndLiveLesson" {
					return StepStateToContext(ctx, stepState), nil
				}
			}
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	case golibs_constants.SubjectLesson:
		timer := time.NewTimer(time.Minute * 1)
		defer timer.Stop()
		select {
		case message := <-stepState.FoundChanForJetStream:
			switch message.(type) {
			case *pb.EvtLesson_UpdateLesson_:
				if msg == "UpdateLesson" {
					return StepStateToContext(ctx, stepState), nil
				}
			case *pb.EvtLesson_EndLiveLesson_:
				if msg == "EndLiveLesson" {
					return StepStateToContext(ctx, stepState), nil
				}
			}
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}

	// case constants.SubjectLessonEventNats:
	// 	handler = func(m *stan.Msg) {
	// 		r := &pb.EvtLesson{}
	// 		err := r.Unmarshal(m.Data)
	// 		if err != nil {
	// 			return
	// 		}
	// 		switch r.Message.(type) {
	// 		case *pb.EvtLesson_EndLiveLesson_:
	// 			if msg == "EndLiveLesson" {
	// 				if r.GetEndLiveLesson().LessonId == stepState.CurrentLessonID && r.GetEndLiveLesson().UserId == stepState.CurrentTeacherID {
	// 					foundChn <- struct{}{}
	// 				}
	// 			}
	// 		case *pb.EvtLesson_CreateLessons_:
	// 			if msg == "CreateLessons" &&
	// 				cmp.Equal(stepState.StudentIds, r.GetCreateLessons().Lessons[0].GetLearnerIds()) {
	// 				foundChn <- struct{}{}
	// 				return
	// 			}

	// 		case *pb.EvtLesson_UpdateLesson_:
	// 			if msg == "UpdateLesson" {
	// 				switch req := stepState.Request.(type) {
	// 				case *bpb.UpdateLiveLessonRequest:
	// 					if req.GetId() == r.GetUpdateLesson().LessonId && cmp.Equal(req.LearnerIds, r.GetUpdateLesson().LearnerIds) {
	// 						foundChn <- struct{}{}
	// 					}
	// 				case *bpb.UpdateLessonRequest:
	// 					learnerIDs := make([]string, 0, len(req.StudentInfoList))
	// 					for _, studentInfo := range req.StudentInfoList {
	// 						learnerIDs = append(learnerIDs, studentInfo.StudentId)
	// 					}
	// 					if req.GetLessonId() == r.GetUpdateLesson().LessonId && cmp.Equal(learnerIDs, r.GetUpdateLesson().LearnerIds) {
	// 						foundChn <- struct{}{}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	case golibs_constants.SubjectLearningObjectivesCreated:
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()

		select {
		case <-stepState.FoundChanForJetStream:
			return StepStateToContext(ctx, stepState), nil
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}

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
	}

	// subscription, err := s.Bus.Subscribe(subject, handler, stan.StartAtTime(stepState.RequestSentAt), stan.AckWait(time.Second))
	// if err != nil {
	// 	return StepStateToContext(ctx, stepState), fmt.Errorf("cannot subscribe to NATS: %v", err)
	// }

	// defer func() {
	// 	subscription.Unsubscribe()
	// 	subscription.Close()
	// }()

	timer := time.NewTimer(time.Minute * 6)
	defer timer.Stop()

	select {
	case <-foundChn:
		return StepStateToContext(ctx, stepState), nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}

func (s *suite) createSubscriptionToReceiveMsg(ctx context.Context, subject string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{nats.StartTime(time.Now()), nats.ManualAck(), nats.AckWait(2 * time.Second)},
	}

	stepState.FoundChanForJetStream = make(chan interface{}, 1)

	jetStreamHandler := func(ctx context.Context, data []byte) (bool, error) {
		r := &npb.EventLearningObjectivesCreated{}
		if err := proto.Unmarshal(data, r); err != nil {
			return false, err
		}
		if len(r.LearningObjectives) > 0 {
			stepState.FoundChanForJetStream <- stepState
			return false, nil
		}
		return true, fmt.Errorf("expect r.LearningObjectives > 0")
	}

	sub, err := s.JSM.Subscribe(subject, opts, jetStreamHandler)
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) BobMustPushMsgSubjectToNats(ctx context.Context, msg, subject string) (context.Context, error) {
	return s.bobMustPushMsgSubjectToNats(ctx, msg, subject)
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}

func (s *suite) aSignedInOfCurrent(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch user {
	case "student":
		stepState.AuthToken, err = generateValidAuthenticationToken(stepState.CurrentStudentID, user)
		stepState.CurrentUserID = stepState.CurrentStudentID
	case "teacher":
		stepState.AuthToken, err = generateValidAuthenticationToken(stepState.CurrentTeacherID, user)
		stepState.CurrentUserID = stepState.CurrentTeacherID
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) SignedInAdmin(ctx context.Context) (context.Context, error) {
	return s.aSignedInAdmin(ctx)
}

func (s *suite) SignedInSchoolAdmin(ctx context.Context) (context.Context, error) {
	return s.aSignedInSchoolAdmin(ctx)
}

func (s *suite) SignedInTeacher(ctx context.Context) (context.Context, error) {
	return s.aSignedInTeacher(ctx)
}

func (s *suite) UserSignedInTeacher(ctx context.Context, i string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var teacherID string
	switch i {
	case "first":
		teacherID = stepState.TeacherIDs[0]
	case "second":
		teacherID = stepState.TeacherIDs[1]
	case "current":
		teacherID = stepState.CurrentTeacherID
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not get account for teacher %s", i)
	}
	var err error
	stepState.AuthToken, err = s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentTeacherID = teacherID
	stepState.CurrentUserID = teacherID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSignedInParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	parentID := stepState.CurrentParentID

	var err error
	userGroup := entities.UserGroupParent
	stepState.AuthToken, err = s.generateExchangeToken(parentID, entities.UserGroupParent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = parentID
	stepState.CurrentUserGroup = userGroup

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInAdmin(ctx context.Context) (context.Context, error) {
	id := s.newID()
	var err error
	stepState := StepStateFromContext(ctx)
	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(entities_bob.UserGroupAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, constant.UserGroupAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken:%w", err)
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupAdmin
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInSchoolAdmin(ctx context.Context) (context.Context, error) {
	id := s.newID()
	var err error
	stepState := StepStateFromContext(ctx)
	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(entities_bob.UserGroupSchoolAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, constant.UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken:%w", err)
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupSchoolAdmin
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ASignedInAdmin(ctx context.Context) (context.Context, error) {
	return s.aSignedInAdmin(ctx)
}

func (s *suite) aValidUserInEureka(ctx context.Context, id, newgroup, oldGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := rand.Int()
	var now pgtype.Timestamptz
	now.Set(time.Now())
	u := entities_bob.User{}
	database.AllNullEntity(&u)
	u.ID = database.Text(id)
	u.LastName.Set(fmt.Sprintf("valid-user-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num))
	u.Country.Set(pb.COUNTRY_VN.String())
	u.Group.Set(oldGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	u.ResourcePath.Set("1")

	gr := &entities_bob.Group{}
	database.AllNullEntity(gr)
	gr.ID.Set(oldGroup)
	gr.Name.Set(oldGroup)
	gr.UpdatedAt.Set(time.Now())
	gr.CreatedAt.Set(time.Now())
	fieldNames, _ := gr.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	stmt := fmt.Sprintf("INSERT INTO groups (%s) VALUES(%s) ON CONFLICT DO NOTHING", strings.Join(fieldNames, ","), placeHolders)
	if _, err := s.EurekaDB.Exec(ctx, stmt, database.GetScanFields(gr, fieldNames)...); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert group error: %v", err)
	}
	ctx = s.setFakeClaimToContext(context.Background(), u.ResourcePath.String, oldGroup)

	ugroup := &entity.UserGroupV2{}
	database.AllNullEntity(ugroup)
	ugroup.UserGroupID.Set(idutil.ULIDNow())
	ugroup.UserGroupName.Set("name")
	ugroup.UpdatedAt.Set(time.Now())
	ugroup.CreatedAt.Set(time.Now())
	ugroup.ResourcePath.Set("1")

	ugMember := &entity.UserGroupMember{}
	database.AllNullEntity(ugMember)
	ugMember.UserID.Set(u.ID)
	ugMember.UserGroupID.Set(ugroup.UserGroupID.String)
	ugMember.CreatedAt.Set(time.Now())
	ugMember.UpdatedAt.Set(time.Now())
	ugMember.ResourcePath.Set("1")

	uG := entities_bob.UserGroup{
		UserID:   u.ID,
		GroupID:  database.Text(oldGroup),
		IsOrigin: database.Bool(true),
	}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt

	role := &entity.Role{}
	database.AllNullEntity(role)
	role.RoleID.Set(idutil.ULIDNow())
	role.RoleName.Set(newgroup)
	role.CreatedAt.Set(time.Now())
	role.UpdatedAt.Set(time.Now())
	role.ResourcePath.Set("1")

	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	grantedRole.RoleID.Set(role.RoleID.String)
	grantedRole.UserGroupID.Set(ugroup.UserGroupID.String)
	grantedRole.GrantedRoleID.Set(idutil.ULIDNow())
	grantedRole.CreatedAt.Set(time.Now())
	grantedRole.UpdatedAt.Set(time.Now())
	grantedRole.ResourcePath.Set("1")

	if _, err := database.Insert(ctx, &u, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user error: %v", err)
	}

	if _, err := database.Insert(ctx, &uG, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group error: %v", err)
	}
	if _, err := database.Insert(ctx, ugroup, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.Insert(ctx, ugMember, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.Insert(ctx, role, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}
	if _, err := database.Insert(ctx, grantedRole, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if u.Group.String == constant.UserGroupStudent {
		stepState.CurrentStudentID = u.ID.String
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setFakeClaimToContext(ctx context.Context, resourcePath string, userGroup string) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
			UserGroup:    userGroup,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claims)
}
