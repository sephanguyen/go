package managing

import (
	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/internal/bob/entities"
)

func initStepForBobServiceFeature(s *suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^an authenticated admin$`:                                    s.anAuthenticatedAdmin,
		`^a school$`:                                                  s.aSchool,
		`^a valid class information$`:                                 s.aValidClassInformation,
		`^an admin creates a new class$`:                              s.anAdminCreatesANewClass,
		`^that class must be created successfully$`:                   s.thatClassMustBeCreatedSuccessfully,
		`^a chat room associated with that class must be created$`:    s.aChatRoomAssociatedWithThatClassMustBeCreated,
		`^a valid class information with multiple teachers assigned$`: s.aValidClassInformationWithMultipleTeachersAssigned,

		`^a signed in student$`:               s.aSignedInStudent,
		`^Bob returns "([^"]*)" status code$`: s.bobSuite.ReturnsStatusCode,
		`^Tom must store message "([^"]*)" with type "([^"]*)" in this conversation$`: s.tomMustStoreMessage,
		`^a signed in admin$`: s.bobSuite.ASignedInAdmin,

		// user update profile
		`^a signed in "([^"]*)" with school: (\d+)$`: s.bobSuite.ASignedInWithSchool,
		`^a request update "([^"]*)" profile with user group: "([^"]*)", name: "([^"]*)", phone: "([^"]*)", email: "([^"]*)", school: (\d+)$`: s.bobSuite.UserUpdatedProfileWithUserGroupNamePhoneEmailSchool,
		`^user ask Bob to do update$`:                                     s.bobSuite.UserAskBobToDoUpdate,
		`^Bob must update "([^"]*)" profile$`:                             s.bobSuite.BobMustUpdateProfile,
		`^Bob must publish event to user_device_token channel$`:           s.bobMustPublishEventToUser_device_tokenChannel,
		`^Tom must record new user_device_tokens with updated name info$`: s.tomMustInsertInfoToTableUserDeviceTokensWithUpdatedName,
		// store device token
		`^a valid device token$`:                                          s.bobSuite.AValidDeviceToken,
		`^user try to store device token$`:                                s.bobSuite.UserTryToStoreDeviceToken,
		`^Bob must store the user's device token`:                         s.bobSuite.BobMustStoreTheUsersDeviceToken,
		`^Tom must record new user_device_tokens with device_token info$`: s.tomMustRecordNewUserDeviceTokenWithDeviceTokenInfo,

		// update student profile
		`^a valid updates profile request$`:            s.bobSuite.AValidUpdatesProfileRequest,
		`^his owned student UUID$`:                     s.bobSuite.HisOwnedStudentUUID,
		`^user updates profile$`:                       s.bobSuite.UserUpdatesProfile,
		`^Bob must records student\'s profile update$`: s.bobSuite.BobMustRecordsStudentsProfileUpdate,
		`^Tom must record new user_device_tokens with message type \*pb\.UpdateProfileRequest$`: s.tomMustRecordNewUserDeviceTokenWithMessageTypeIsUpdateProfileRequest,
		`^a school name "([^"]*)", country "([^"]*)", city "([^"]*)", district "([^"]*)"$`:      s.bobSuite.ASchoolNameCountryCityDistrict,
		`^admin inserts schools$`: s.bobSuite.AdminInsertsSchools,

		// create a class
		`^a random number$`:                                         s.bobSuite.ARandomNumber,
		`^a signed in teacher$`:                                     s.bobSuite.ASignedInTeacher,
		`^a CreateClassRequest$`:                                    s.bobSuite.ACreateClassRequest,
		`^a "([^"]*)" schoolId in CreateClassRequest$`:              s.bobSuite.ASchoolIdInCreateClassRequest,
		`^a valid name in CreateClassRequest$`:                      s.bobSuite.AValidNameInCreateClassRequest,
		`^default config for "([^"]*)" has "([^"]*)" is "([^"]*)"$`: s.bobSuite.DefaultConfigForHasIs,
		`^user create a class`:                                      s.bobSuite.UserCreateAClass,
		`^Bob must create class from CreateClassRequest$`:           s.bobSuite.BobMustCreateClassFromCreateClassRequest,
		`^class must has "([^"]*)" is "([^"]*)"$`:                   s.bobSuite.ClassMustHasIs,
		`^class must have (\d+) member is "([^"]*)" and is owner "([^"]*)" and status "([^"]*)"$`:     s.bobSuite.ClassMustHaveMemberIsAndIsOwnerAndStatus,
		`^Bob must push msg "([^"]*)" subject "([^"]*)" to nats$`:                                     s.bobMustPushMsgSubjectToNats,
		`^a (\d+) "([^"]*)" ownerId with school id is (\d+) in CreateClassRequest$`:                   s.aOwnerIdWithSchoolIdIsInCreateClassRequest,
		`^this school has config "([^"]*)" is "([^"]*)", "([^"]*)" is "([^"]*)", "([^"]*)" is (\d+)$`: s.bobSuite.ThisSchoolHasConfigIsIsIs,

		// edit a class
		`^admin create a class with school name "([^"]*)" and expired at "([^"]*)"$`: s.bobSuite.CreateAClassWithSchoolNameAndExpiredAt,
		`^a EditClassRequest with class name is "([^"]*)"$`:                          s.bobSuite.AEditClassRequestWithClassNameIs,
		`^a "([^"]*)" classId in EditClassRequest$`:                                  s.bobSuite.AClassIdInEditClassRequest,
		`^user edit a class$`:                                s.bobSuite.UserEditAClass,
		`^Bob must update class in db$`:                      s.bobSuite.BobMustUpdateClassInDB,
		`^a signed in "([^"]*)" with school name "([^"]*)"$`: s.bobSuite.ASignedInWithSchoolName,
		`^a teacher who is owner current class$`:             s.aTeacherWhoIsOwnerCurrentClass,
		// join a class
		`^these schools were inserted by admin$`: s.bobSuite.AdminInsertsSchools,
		//`^some package plan available$`:                                                    s.bobSuite.SomePackagePlanAvailable,
		`^a class with school name "([^"]*)" and expired at "([^"]*)"$`: s.aClassWithSchoolNameAndExpiredAt,
		`^a JoinClassRequest$`:                                    s.bobSuite.AJoinClassRequest,
		`^a "([^"]*)" classCode in JoinClassRequest$`:             s.bobSuite.AClassCodeInJoinClassRequest,
		`^a user signed in "([^"]*)" with school name "([^"]*)"$`: s.aUserSignedInWithSchoolName,
		`^user join a class$`:                                     s.userJoinAClass,
		`^JoinClassResponse must return "([^"]*)"$`:               s.bobSuite.JoinClassResponseMustReturn,
		//`^student profile show "([^"]*)" plan$`:                                            s.bobSuite.StudentProfileShowPlan,
		`^student subscription must has "([^"]*)" is "([^"]*)" with plan id is "([^"]*)"$`:                                         s.bobSuite.StudentSubscriptionMustHasIsWithPlanIdIs,
		`^Tom must record message join class of current user$`:                                                                     s.tomMustRecordMessageJoinClassOfCurrentUser,
		`^a class created by current teacher with config "([^"]*)" is "([^"]*)", "([^"]*)" is "([^"]*)", "([^"]*)" is  "([^"]*)"$`: s.aClassWithThisConfigCreatedByCurrentTeacher,
		`^Eureka must add new class member$`:                                                                                       s.eurekaMustAddNewClassMember,

		// leave a class
		`^current school has config "([^"]*)" is "([^"]*)", "([^"]*)" is "([^"]*)", "([^"]*)" is (\d+)$`: s.bobSuite.ThisSchoolHasConfigIsIsIs,
		`^current student join class by using this class code$`:                                          s.currentStudentJoinToClassByUsingThisClassCode,
		`^Tom must record message leave class on this class conversation with kicked is "([^"]*)"$`:      s.tomMustRecordMessageLeaveClassOnThisClassConversation,
		`^current teacher join class by using this class code$`:                                          s.currentTeacherJoinToClassByUsingThisClassCode,
		`^Eureka must remove class members$`:                                                             s.eurekaMustRemoveClassMember,

		// remove class member
		`^(\d+) "([^"]*)" join class with school name "([^"]*)"$`:     s.joinClassWithSchoolName,
		`^a valid token of current teacher$`:                          s.bobSuite.AValidTokenOfCurrentTeacher,
		`^a RemoveMemberRequest with class in school name "([^"]*)"$`: s.bobSuite.ARemoveMemberRequestWithClassInSchoolName,
		`^a "([^"]*)" in RemoveMemberRequest$`:                        s.bobSuite.AInRemoveMemberRequest,
		`^user remove member from class$`:                             s.bobSuite.UserRemoveMemberFromClass,
		`^Bob must store activity logs "([^"]*)" class member$`:       s.bobSuite.BobMustStoreActivityLogsClassMember,

		// student join lesson
		`^a list of courses are existed in DB of "([^"]*)"$`:    s.bobSuite.AListOfCoursesAreExistedInDBOf,
		`^a student with valid lesson$`:                         s.aStudentWithValidLesson,
		`^student join lesson$`:                                 s.studentJoinLesson,
		`^student join lesson with v1 API$`:                     s.studentJoinLessonV1,
		`^student must receive lesson room id$`:                 s.bobSuite.StudentMustReceiveLessonRoomId,
		`^student must receive lesson room id and tokens$`:      s.bobSuite.StudentMustReceiveLessonRoomIdAndTokens,
		`^Tom must record new lesson conversation`:              s.tomMustRecordNewLessonConversation,
		`^Tom must record message join lesson of current user$`: s.tomMustStoreMessageJoinLesson,

		// teacher join lesson
		`^a teacher from same school with valid lesson$`: s.aTeacherFromSameSchoolWithValidLesson,
		`^teacher join lesson$`:                          s.teacherJoinLesson,
		`^returns valid information for broadcast$`:      s.bobSuite.ReturnsValidInformationForBroadcast,

		// leave live lesson
		`^student leave lesson$`: s.bobSuite.StudentLeaveLesson,
		`^Tom must change conversation lesson status of this student to inactive$`: s.tomMustChangeConversationLessonStatusOfThisUserToInactive,
		`^Tom must record message leave lesson of this student$`:                   s.tomMustRecordMessageLeaveLesson,
	}
	return steps
}

type BobStepState struct {
	CurrentSchoolID   int32
	CurrentStudentId  string
	CurrentTeacherID  string
	Schools           []*entities.School
	Random            string
	ClassOwnersID     []string
	ClassStudentsID   []string
	UserIDsLeaveClass []string
}

func (s *suite) newBobSuite() {
	s.bobSuite = &bob.Suite{}
	s.bobSuite.DB = s.bobDB
	s.bobSuite.Conn = s.bobConn
	s.bobSuite.JSM = s.jsm
	s.bobSuite.ZapLogger = s.ZapLogger
	s.bobSuite.ShamirConn = s.shamirConn
	s.bobSuite.ApplicantID = s.ApplicantID
}
