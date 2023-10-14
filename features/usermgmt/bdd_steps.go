package usermgmt

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
		// Database
		`^a database with shard id$`:                                              s.aDatabaseWithShardId,
		`^client acquires a connection$`:                                          s.clientAcquiresAConnection,
		`^the connection should have corresponding shard id in session variable$`: s.theConnectionShouldHaveCorrespondingShardIdInSessionVariable,
		`^client generate sharded id via database func$`:                          s.clientGenerateShardedIdViaDatabaseFunc,
		`^the client receive valid sharded id$`:                                   s.theClientReceiveValidShardedId,

		// Create student
		`^everything is OK$`:                                       s.everythingIsOK,
		`^health check endpoint called$`:                           s.healthCheckEndpointCalled,
		`^usermgmt should return "([^"]*)" with status "([^"]*)"$`: s.usermgmtShouldReturnWithStatus,
		`^"([^"]*)" cannot create that account$`:                   s.cannotCreateThatAccount,
		`^"([^"]*)" create new student account$`:                   s.createNewStudentAccount,
		`^new student account created success with student info$`:  s.newStudentAccountCreatedSuccessWithStudentInfo,
		`^new student account created success with student info and first name, last name, phonetic name$`: s.newStudentAccountCreatedSuccessWithStudentInfoAndFirstNameLastNamePhoneticName,
		`^only student info$`: s.onlyStudentInfo,
		`^only student info with first name last name and phonetic name$`:   s.studentInfoWithFirstNameLastNameAndPhoneticName,
		`^receives "([^"]*)" status code$`:                                  s.receivesStatusCode,
		`^student data missing "([^"]*)"$`:                                  s.studentDataMissing,
		`^student data with unknown student enrollment status$`:             s.studentDataWithUnknownStudentEnrollmentStatus,
		`^"([^"]*)" create new student account with invalid resource path$`: s.createNewStudentAccountWithInvalidResourcePath,
		`^student info with invalid locations "([^"]*)"$`:                   s.studentInfoWithInvalidLocations,
		`^a signed in "([^"]*)"$`:                                           s.signedAsAccount,
		`^a signed in as "([^"]*)" in "([^"]*)" organization$`:              s.aSignedInAsInOrganization,

		`^only student info with enrollment status string "([^"]*)"$`:                               s.onlyStudentInfoWithEnrollmentStatusString,
		`^"([^"]*)" in organization (\d+) create user (\d+)$`:                                       s.inOrganizationCreateUser,
		`^"([^"]*)" in organization (\d+) create user (\d+) with the same "([^"]*)" as user (\d+)$`: s.inOrganizationCreateUserWithTheSameAsUser,
		`^user (\d+) will be created successfully and belonged to organization (\d+)$`:              s.userWillBeCreatedSuccessfullyAndBelongedToOrganization,
		`^student data with some "([^"]*)" tags$`:                                                   s.studentProfileWithTags,
		`^student info with invalid tags "([^"]*)"$`:                                                s.studentProfileInvalidWithTags,

		// Update student
		`^student account data to update$`:                                                  s.studentAccountDataToUpdate,
		`^student account data to update with first name lastname and phonetic name$`:       s.studentAccountDataToUpdateWithFirstNameLastNameAndPhoneticName,
		`^"([^"]*)" cannot update student account$`:                                         s.cannotUpdateStudentAccount,
		`^student account is updated success$`:                                              s.studentAccountIsUpdatedSuccess,
		`^student account is updated success with first name last name and phonetic name$`:  s.studentAccountWithFirstNameLastNameAndPhoneticNameUpdatedSuccessfully,
		`^"([^"]*)" update student account$`:                                                s.updateStudentAccount,
		`^"([^"]*)" update student account that has "([^"]*)"$`:                             s.updateStudentAccountThatHas,
		`^"([^"]*)" updates student account with new student data missing "([^"]*)" field$`: s.updatesStudentAccountWithNewStudentDataMissingField,
		`^"([^"]*)" update student account does not exist$`:                                 s.updateStudentAccountDoesNotExist,
		`^"([^"]*)" update student email that exist in our system$`:                         s.updateStudentEmailThatExistInOurSystem,
		`^student account data to update has invalid locations "([^"]*)"$`:                  s.studentAccountDataToUpdateHasInvalidLocations,
		`^assign course package with "([^"]*)" to exist student$`:                           s.assignCoursePackageWithToExistStudent,
		`^"([^"]*)" user "([^"]*)" update student account$`:                                 s.userUpdateStudentAccount,
		`^student account data to update with enrollment status string "([^"]*)"$`:          s.studentAccountDataToUpdateWithEnrollmentStatusString,
		`^existed student data with some "([^"]*)" tags$`:                                   s.existedStudentProfileWithTags,
		`^update student info with "([^"]*)" tags$`:                                         s.updateStudentTags,
		`^student account is updated success with tags$`:                                    s.studentUpdateSuccessWithUserTags,
		`^student account data to update with parent info$`:                                 s.studentAccountDataToUpdateWithParentInfo,

		// retrieve student profile
		`^an invalid authentication token$`:                 s.anInvalidAuthenticationToken,
		`^an other student profile in DB$`:                  s.anOtherStudentProfileInDB,
		`^user retrieves student profile$`:                  s.userRetrievesStudentProfile,
		`^a signed in student$`:                             s.aSignedInStudent,
		`^teacher retrieves a "([^"]*)" student profile$`:   s.teacherRetrievesAStudentProfile,
		`^teacher retrieves the student profile$`:           s.teacherRetrievesTheStudentProfile,
		`^returns requested student profile$`:               s.returnsRequestedStudentProfile,
		`^return the requester\'s profile$`:                 s.returnsRequesterStudentProfile,
		`^returns empty student profile$`:                   s.returnsEmptyStudentProfile,
		`^returns student profile with correct grade info$`: s.returnsStudentProfileWithCorrectGradeInfo,

		// update user last login date
		`^a "([^"]*)" signed in user with "([^"]*)"$`:        s.aSignedInUserWith,
		`^user last login date "([^"]*)"$`:                   s.userLastLoginDate,
		`^user update last login date with "([^"]*)" value$`: s.userUpdateLastLoginDateWithValue,

		// Create parents
		`^"([^"]*)" create new parents$`:                                                 s.createNewParents,
		`^"([^"]*)" create new parents with invalid resource path`:                       s.createNewParentsWithInvalidResourcePath,
		`^"([^"]*)" create "([^"]*)" new parents$`:                                       s.createMultipleNewParents,
		`^new parents data$`:                                                             s.newParentsData,
		`^new "([^"]*)" parents data "([^"]*)"$`:                                         s.newMultipleParentsData,
		`^new parents were created successfully$`:                                        s.newParentsWereCreatedSuccessfully,
		`^parent data has empty or invalid "([^"]*)"$`:                                   s.parentDataHasEmptyOrInvalid,
		`^"([^"]*)" in organization (\d+) create parent (\d+)$`:                          s.inOrganizationCreateParent,
		`^"([^"]*)" in organization 2 create parent 2 with the same email as parent 1`:   s.inOrganizationCreateParentWithTheSameAsParent,
		`^parent (\d+) will be created successfully and belonged to organization (\d+)$`: s.parentWillBeCreatedSuccessfullyAndBelongedToOrganization,
		`^assign "([^"]*)" tag to parents$`:                                              s.assignTagsToParentData,

		// create staff
		`^generate a "([^"]*)" CreateStaffProfile and choose "([^"]*)" locations$`: s.generateACreateStaffProfile,
		`^"([^"]*)" create staff account$`:                                         s.createStaffAccount,
		`^new staff account was created successfully$`:                             s.newStaffAccountWasCreatedSuccessfully,

		// update staff profile
		`^profile of staff must be updated$`: s.profileOfStaffMustBeUpdated,
		`^staff update profile$`:             s.staffUpdateProfileCreatedProfile,
		`^a profile of staff with name: "([^"]*)"; user group type: "([^"]*)" and generate a "([^"]*)" UpdateStaffProfile and "([^"]*)" locations$`: s.aProfileOfStaffWithSpecificData,
		`^existed staff with "([^"]*)"$`: s.createRandomStaff,

		// Update parents
		`^"([^"]*)" cannot update parents$`:                      s.cannotUpdateParents,
		`^parent data to update has empty or invalid "([^"]*)"$`: s.parentDataToUpdateHasEmptyOrInvalid,
		`^parents data to update$`:                               s.parentsDataToUpdate,
		`^parents were updated successfully$`:                    s.parentsWereUpdatedSuccessfully,
		`^"([^"]*)" update new parents$`:                         s.updateNewParents,
		`^edit update parent data to blank in "([^"]*)"$`:        s.editParentDataToBlank,
		`^update parent data "([^"]*)"$`:                         s.updateParentData,

		// Reissue password
		`^"([^"]*)" reissues user\'s password$`:                              s.reissuesUsersPassword,
		`^returns "([^"]*)" status code$`:                                    s.CommonSuite.ReturnsStatusCode,
		`^user can sign in with the new password$`:                           s.userCanSignInWithTheNewPassword,
		`^the signed in user create "([^"]*)" user$`:                         s.theSignedInUserCreateUser,
		`^"([^"]*)" reissues user\'s password when missing "([^"]*)" field$`: s.reissuesUsersPasswordWhenMissingField,
		`^"([^"]*)" reissues user\'s password with non-existing user$`:       s.reissuesUsersPasswordWithNonexistingUser,
		`^the owner reissues user\'s password$`:                              s.theOwnerReissuesUsersPassword,

		// retrieve students profiles associated to parent account
		`^create handsome father and pretty mother as parent and the relationship with his children who're students at manabie$`: s.createFatherAndMotherAsAParentAndTheRelationshipWithTheirChildren,
		`^retrieve students profiles associated to each account$`:                                                                s.retrieveStudentsProfilesEachParentAccount,
		`^fetched students exactly associated to parent$`:                                                                        s.fetchedStudentsExactlyAssociatedToParent,
		`^returns the same students profiles$`:                                                                                   s.sameStudentProfiles,
		`^no students profiles are fetched$`:                                                                                     s.noStudentsProfilesAreFetched,
		`^remove relationship of student$`:                                                                                       s.removeRelationshipOfStudent,

		// Remove parents
		`^parents data to remove from student$`:                  s.parentsDataToRemove,
		`^parents were removed successfully$`:                    s.parentsWereRemoveSuccessfully,
		`^"([^"]*)" remove new parents with "([^"]*)"$`:          s.removeParentFromStudentWithConditions,
		`^parents data without relationship to remove$`:          s.parentsDataWithoutRelationshipToRemove,
		`^create data parent with (\d+) different students$`:     s.createDataParentWithMultipleStudents,
		`^receives "([^"]*)" status code and message "([^"]*)"$`: s.receiveCodeStatusAndMessage,
		`^parents was not removed in database$`:                  s.verifyParentInDatabase,

		// Upsert student course packages
		`^"([^"]*)" cannot upsert student course packages$`:                                      s.cannotUpsertStudentCoursePackages,
		`^upsert student course packages successfully$`:                                          s.upsertStudentCoursePackagesSuccessfully,
		`^upsert student course packages successfully with student package extra$`:               s.upsertStudentCoursePackagesSuccessfullyWithStudentPackageExtra,
		`^student exist in our system$`:                                                          s.studentExistInOurSystem,
		`^"([^"]*)" upsert student course packages$`:                                             s.upsertStudentCoursePackages,
		`^"([^"]*)" upsert student course packages with student package course extra$`:           s.upsertStudentCoursePackagesWithStudentPackageExtra,
		`^"([^"]*)" upsert student course packages with course invalid start date and end date$`: s.upsertStudentCoursePackagesWithCourseInvalidStartDateAndEndDate,
		`^"([^"]*)" upsert student course packages with "([^"]*)" invalid$`:                      s.upsertStudentCoursePackagesWithInvalid,
		`^"([^"]*)" upsert student course packages with only "([^"]*)"$`:                         s.upsertStudentCoursePackagesWithOnly,
		`^"([^"]*)" upsert student course packages with package id empty$`:                       s.upsertStudentCoursePackagesWithPackageIdEmpty,
		`^"([^"]*)" upsert student course packages with location ids empty$`:                     s.upsertStudentCoursePackagesWithLocationIdsEmpty,
		`^assign course package to exist student$`:                                               s.assignCoursePackageToExistStudent,
		`^assign student package with class empty to exist student$`:                             s.assignStudentPackageWithClassEmptyToExistStudent,

		// Import users from a tenant to another tenant
		`^admin import users from tenant (\d+) to tenant (\d+)$`: s.adminImportUsersFromTenantToTenant,
		`^users in tenant (\d+)$`:                                s.usersInTenant,
		`^users in tenant (\d+) has corresponding info$`:         s.usersInTenantHasCorrespondingInfo,
		`^users in tenant (\d+) still have valid info$`:          s.usersInTenantStillHaveValidInfo,

		// Import users from firebase auth to tenant
		`^admin import users from firebase auth to tenant in identity platform$`: s.adminImportUsersFromFirebaseAuthToTenantInIdentityPlatform,
		`^users in firebase auth$`:                             s.usersInFirebaseAuth,
		`^users in firebase auth still still have valid info$`: s.usersInFirebaseAuthStillStillHaveValidInfo,
		`^users in tenant has corresponding info$`:             s.usersInTenantHasCorrespondingInfo,

		// Job migrates users from firebase auth
		`^users in our system have been imported to firebase auth$`:             s.usersInOurSystemHaveBeenImportedToFirebaseAuth,
		`^system run job to migrate users from firebase auth$`:                  s.systemRunJobToMigrateUsersFromFirebaseAuth,
		`^info of users in firebase auth is still valid$`:                       s.infoOfUsersInFirebaseAuthIsStillValid,
		`^info of users in tenant of identity platform has corresponding info$`: s.infoOfUsersInTenantOfIdentityPlatformHasCorrespondingInfo,

		// Auth users importing operation
		`^system create auth profiles with valid info`:                                                                                           s.systemCreateAuthProfilesWithValidInfo,
		`^auth profiles are created successfully and users can use them to login in to system$`:                                                  s.authProfilesAreCreatedSuccessfullyAndUsersCanUseThemToLoginInToSystem,
		`^system create auth profiles "([^"]*)"$`:                                                                                                s.systemCreateAuthProfiles,
		`^system failed to create auth profiles and users can not use them to login in to system$`:                                               s.systemFailedToCreateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem,
		`^user already logged in with existing auth profile$`:                                                                                    s.userAlreadyLoggedInWithExistingAuthProfile,
		`^system update existing auth profiles with valid info "([^"]*)"$`:                                                                       s.systemUpdateExistingAuthProfilesWithValidInfoArg,
		`^auth profiles are updated successfully but user doesn\'t need to login in again, also they still can login in again with old profile$`: s.authProfilesAreUpdatedSuccessfullyButUserDoesntNeedToLoginInAgainAlsoTheyStillCanLoginInAgainWithOldProfile,
		`^auth profiles are updated successfully and user can use them login in, if already logged in, user have to login in again$`:             s.authProfilesAreUpdatedSuccessfullyAndUserCanUseThemLoginInIfAlreadyLoggedInUserHaveToLoginInAgain,
		`^auth profiles are updated successfully without changing password and users can use them to login in to system$`:                        s.authProfilesAreUpdatedSuccessfullyWithoutChangingPasswordAndUsersCanUseThemToLoginInToSystem,
		`^existing auth profiles in system$`:                                                                                                     s.existingAuthProfilesInSystem,
		`^system update existing auth profiles with valid info$`:                                                                                 s.systemUpdateExistingAuthProfilesWithValidInfo,
		`^system update existing auth profiles "([^"]*)"$`:                                                                                       s.systemUpdateExistingAuthProfiles,
		`^auth profiles are updated successfully and users can use them to login in to system$`:                                                  s.authProfilesAreUpdatedSuccessfullyAndUsersCanUseThemToLoginInToSystem,
		`^system failed to update auth profiles and users can not use them to login in to system$`:                                               s.systemFailedToUpdateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem,

		// Import school_info
		`^an school info valid request payload with "([^"]*)"$`:   s.anSchoolInfoValidRequestPayloadWith,
		`^an school info invalid "([^"]*)" request payload$`:      s.anSchoolInfoInvalidRequestPayload,
		`^"([^"]*)" importing school info$`:                       s.importingSchoolInfo,
		`^the valid school info lines are imported successfully$`: s.theValidSchoolInfoLinesAreImportedSuccessfully,
		`^the invalid school info lines are returned with error$`: s.theInvalidSchoolInfoLinesAreReturnedWithError,

		// search basic profile
		`^search basic profile request with student_ids and "([^"]*)"$`: s.searchBasicProfileRequestWithStudentIDsAnd,
		`^"([^"]*)" search basic profile$`:                              s.searchBasicProfile,
		`returns a list basic profile correctly`:                        s.returnAListBasicProfileCorrectly,
		`^prepare students data$`:                                       s.prepareStudentsData,

		// get basic profile
		`^user can not get basic profile$`:                        s.userCanNotGetBasicProfile,
		`^user get basic profile$`:                                s.userGetBasicProfile,
		`^user get basic profile with invalid "([^"]*)" request$`: s.userGetBasicProfileWithInvalidRequest,
		`^user receive basic profile$`:                            s.userReceiveBasicProfile,

		// job migrate delete student location org
		`^existing students have default location are removed location type org$`:       s.existingStudentsHaveDefaultLocationAreRemovedLocationTypeOrg,
		`^students with location type org in our system$`:                               s.studentsWithLocationTypeOrgInOurSystem,
		`^system run job to migrate delete student location org existed in our system$`: s.systemRunJobToMigrateDeleteStudentLocationOrgExistedInOurSystem,

		// update user profile
		`^the signed in user update profile$`:                                       s.theSignedInUserUpdateProfile,
		`^the signed in user update another user profile$`:                          s.theSignedInUserUpdateAnotherUserProfile,
		`the signed in user update user profile without mandatory "([^"]*)" field$`: s.theSignedInUserUpdateUserProfileWithoutMandatoryField,
		`^user update profile successfully$`:                                        s.userUpdateProfileSuccessfully,
		`^user cannot update user profile$`:                                         s.userCannotUpdateUserProfile,

		// job migrate update student enrollment status quit to withdrawn
		`^students have enrollment status "([^"]*)" are updated to "([^"]*)" with "([^"]*)"$`:                         s.studentsHaveEnrollmentStatusAreUpdatedToWith,
		`^students with enrollment original status in our system$`:                                                    s.studentsWithEnrollmentOriginalStatusInOurSystem,
		`^system run job to migrate update student enrollment status "([^"]*)" to "([^"]*)" in our system "([^"]*)"$`: s.systemRunJobToMigrateUpdateStudentEnrollmentStatusToInOurSystem,
		// upsert student comment
		`^a valid upsert student comment request with "([^"]*)"$`: s.aValidUpsertStudentCommentRequestWith,
		`^upsert comment for student$`:                            s.upsertCommentForStudent,
		`^BobDB must "([^"]*)" comment for student$`:              s.bobDBMustStoreCommentForStudent,

		// delete student's comments
		`^a signed in teacher$`:                                                s.aSignedInTeacher,
		`^the signed in user delete student\'s comment$`:                       s.theSignedInUserDeleteStudentComments,
		`^a student with some comments$`:                                       s.aStudentWithSomeComments,
		`^our systems have to store comment correctly$`:                        s.ourSystemHaveToStoreCommentsCorrectly,
		`^the signed in user delete student\'s comment with nil commentIds$`:   s.theSignedInUserDeleteStudentCommentsWithNilCmtIds,
		`^the signed in user delete student\'s comment but comment not exist$`: s.theSignedInUserDeleteStudentCommentsButCommentNotExist,

		// retrieve student's comments
		`^the signed in user retrieve student\'s comments$`:                   s.theSignedInUserRetrieveStudentComments,
		`^the signed in user retrieve student\'s comment with nil studentId$`: s.theSignedInUserRetrieveStudentCommentsWithNilStudentId,
		`^get comment belong to user\'s correctly$`:                           s.getCommentsBelongToUserCorrectly,

		// Validation check for gender when create user
		`^only student info with "([^"]*)"$`: s.onlyStudentInfoWith,

		// increase grade of students job
		`^list students with different grade and different school$`: s.listStudentsWithDifferentGradeAndDifferentSchool,
		`system run job to increase grade of students$`:             s.systemRunJobToIncreaseGradeOfStudents,
		`grade of students was increased by "([^"]*)" level$`:       s.gradeOfStudentsWasIncreasedByLevel,

		// import student
		`^a student valid request payload with "([^"]*)"$`:                               s.aStudentValidRequestPayloadWith,
		`^a student valid request payload home address with "([^"]*)"$`:                  s.aStudentValidRequestPayloadHomeAddressWith,
		`^a student valid request payload school history with "([^"]*)"$`:                s.aStudentValidRequestPayloadSchoolHistoryWith,
		`^"([^"]*)" importing student$`:                                                  s.importingStudent,
		`^the valid student lines with home address are imported successfully$`:          s.theValidStudentLinesWithHomeAddressAreImportedSuccessfully,
		`^the valid student lines with student phone number imported successfully$`:      s.theValidStudentLinesWithStudentPhoneNumberAreImportedSuccessfully,
		`^the valid student lines with school history are imported successfully$`:        s.theValidStudentLinesWithHomeAddressAreImportedSuccessfully,
		`^the valid student lines are imported successfully$`:                            s.theValidStudentLinesAreImportedSuccessfully,
		`^the invalid student lines are returned with error code "([^"]*)"$`:             s.theInvalidStudentLinesAreReturnedWithError,
		`^a student info invalid "([^"]*)" request payload$`:                             s.aStudentInvalidRequestPayload,
		`^a student valid request payload tag with "([^"]*)"$`:                           s.aStudentValidRequestTagWith,
		`^after import amount tag have students$`:                                        s.userTagsAreImported,
		`^a student valid request payload enrollment status history with "([^"]*)"$`:     s.aStudentValidRequestEnrollmentStatusHistoryWith,
		`^the valid student lines with enrollment status history imported successfully$`: s.enrollmentStatusHistoryAreImported,

		// import parent and assign to student
		`^"([^"]*)" import (\d+) parent\(s\) and assign to (\d+) student\(s\) with valid payload having "([^"]*)"$`: s.importingParentWithValidPayload,
		`^the valid parent lines are imported successfully$`:                                                        s.theValidParentLinesAreImportedSuccessfully,
		`^the invalid parent lines are returned with error code "([^"]*)"$`:                                         s.theInvalidParentLinesAreReturnedWithError,

		// create user group
		`^signed in user create "([^"]*)" user group$`:                      s.signedInAndCreateUserGroupWithValidityPayload,
		`^user group is created successfully$`:                              s.userGroupCreatedSuccessfully,
		`^user group after creating must be existed in database correctly$`: s.userGroupMustBeExistedInDB,

		// update user group
		`^a user group need to be updated$`:                                             s.userGroupNeedToBeUpdated,
		`^"([^"]*)" update user group successfully$`:                                    s.updateUserGroupSuccessfully,
		`^"([^"]*)" update user group with valid payload$`:                              s.userUpdateUserGroupWithValidPayload,
		`^"([^"]*)" can not update user group and receive status code "([^"]*)" error$`: s.userCanNotUpdateUserGroupAndReceiveStatusCodeError,
		`^"([^"]*)" update user group with "([^"]*)" invalid argument$`:                 s.userUpdateUserGroupWithInvalidArgument,
		`^"([^"]*)" update user group without argument "([^"]*)"$`:                      s.userUpdateUserGroupWithoutArgument,
		`^a user group was granted "([^"]*)" role need to be updated$`:                  s.createUserGroupWithRole,
		`^"([^"]*)" update user group with valid payload and grant "([^"]*)" role$`:     s.userUpdateUserGroupWithValidPayloadWithRoleName,
		`^all users of that user group have "([^"]*)"$`:                                 s.usersOfThatUserGroupHaveAUserGroupv1,

		// validate user login with user group
		`^check this user is able to access "([^"]*)" platform$`: s.checkThisUserIsAbleToAccessPlatform,
		`^user is "([^"]*)" to login this platform$`:             s.userHasPermissionToAccessPlatform,

		// task publish import user events
		`^generate import student event records$`:         s.generateImportStudentEventRecords,
		`^system run task to publish import user events$`: s.systemRunTaskToPublishImportUserEvents,
		`^downstream services consume the events$`:        s.downstreamServicesConsumeTheEvents,
		`^status of import user events get updated$`:      s.statusOfImportUserEventsGetUpdated,
		`^generate import parent event records$`:          s.generateImportParentEventRecords,

		// job migrate add default usergroup for student & parent
		`^some students and parents without user group$`:                         s.generateAmountStudentParentWithoutUserGroup,
		`^system run job to migrate add default usergroup for student & parent$`: s.systemRunMigrationJobAddDefaultUserGroup,
		`^previous students and parents have user group$`:                        s.assertPreviousStudentAndParentHasUserGroup,

		// job migrate student full name into last name
		`^some random student with full name only$`:                    s.someRandomStudentWithFullNameOnlyInDB,
		`^system run job to migrate student full name into last name$`: s.systemRunJobToMigrateStudentFullNameIntoLastName,
		`^full name migrated to last name successfully$`:               s.studentFullNameMigratedToLastNameSuccessfully,

		// job to migrate user phone number from users table into user_phone_number table
		`^some random "([^"]*)" with phone number$`:                                                       s.someRandomUserWithPhoneNumberInDB,
		`^The system runs a job to migrate the phone number of "([^"]*)" into user_phone_number's table$`: s.systemRunJobToMigrateUserPhoneNumber,
		`^The phone number of "([^"]*)" was successfully migrated to the user_phone_number table$`:        s.phoneNumberOfUserMigratedToUserPhoneNumberSuccessfully,

		// unleash enable/disable feature toggle
		`^"([^"]*)" Unleash feature with feature name "([^"]*)"$`: s.ToggleUnleashFeatureWithName,

		// job migrate assign user group to existed staff
		`^user group with role "([^"]*)"$`:                                                      s.createUserGroupWithRoleName,
		`^amount staff have "([^"]*)" old user group is "([^"]*)"$`:                             s.existedStaffInDBOfASchool,
		`^admin choose the previous user group assign to "([^"]*)" of staff and run migration$`: s.systemRunMigrationAssignUsergroupToSpecifyStaff,
		`^the specified staff must have the previous user group$`:                               s.userMustHaveUserGroup,

		// handle event creating org
		`^organization informations$`:                                      s.genOrganizationInfo,
		`^an "([^"]*)" create a new organization$`:                         s.signedInAndCreateOrg,
		`^default roles and permissions will be created for organization$`: s.roleAndPermissionMustBeExistedInDB,

		// create student with school histories
		`^student info with school histories request and valid "([^"]*)"$`:                s.studentInfoWithSchoolHistoriesRequest,
		`^new student account created success with school histories$`:                     s.newStudentAccountCreatedSuccessWithSchoolHistories,
		`^new student account created success with school histories have current school$`: s.newStudentAccountCreatedSuccessWithSchoolHistoriesHaveCurrentSchool,
		`^student info with school histories request and invalid "([^"]*)"$`:              s.studentInfoWithSchoolHistoriesInvalidRequest,

		// update student with school histories
		`^student info with school histories update request and valid "([^"]*)"$`:       s.studentInfoWithSchoolHistoriesUpdateValidRequest,
		`^student info with school histories update request and invalid "([^"]*)"$`:     s.studentInfoWithSchoolHistoriesUpdateInvalidRequest,
		`^student account updated success with school histories$`:                       s.studentAccountUpdatedSuccessWithSchoolHistories,
		`^student account updated success with school histories have current school$`:   s.studentAccountUpdatedSuccessWithSchoolHistoriesHaveCurrentSchool,
		`^student account updated success with school histories remove current school$`: s.studentAccountUpdatedSuccessWithSchoolHistoriesRemoveCurrentSchool,

		// migrate create user group
		`^some roles "([^"]*)" and locations to create user group$`:                                                              s.someRolesAndLocationsToCreateUserGroup,
		`^system run job to migrate create user group with userGroupName "([^"]*)", roles "([^"]*)" and organization "([^"]*)"$`: s.systemRunJobToMigrateCreateUserGroupWithUserGroupNameRolesAndOrganization,
		`^user group create successfully with userGroupName "([^"]*)", roles "([^"]*)" and organization "([^"]*)"$`:              s.userGroupCreateSuccessfullyWithUserGroupNameRolesAndOrganization,

		// update staff setting
		`^user update staff config$`:                s.userUpdateStaffConfig,
		`^a staff config with staff id: "([^"]*)"$`: s.aStaffConfigWith,

		// create student with user addresses
		`^student info with home addresses request and valid "([^"]*)"$`:   s.studentInfoWithHomeAddressesRequestValid,
		`^student info with home addresses request and invalid "([^"]*)"$`: s.studentInfoWithHomeAddressesInvalidRequest,
		`^new student account created success with home addresses$`:        s.newStudentAccountCreatedSuccessWithHomeAddresses,

		// update student with user addresses
		`^student info with home addresses update request and valid "([^"]*)"$`:   s.studentInfoWithHomeAddressesUpdateValidRequest,
		`^student info with home addresses update request and invalid "([^"]*)"$`: s.studentInfoWithHomeAddressesUpdateInvalidRequest,
		`^student account updated success with home addresses$`:                   s.studentAccountUpdatedSuccessWithHomeAddresses,

		// student phone number
		`^student info with student phone number and contact preference with "([^"]*)"$`:         s.studentInfoWithStudentPhoneNumberAndContactPreference,
		`^new student account created success with student phone number and contact preference$`: s.newStudentCreatedSuccessfullyWithStudentPhoneNumberAndContactPreference,

		// migrate current_grade to grade_id
		`^generate grade master$`:                               s.generateGradeMaster,
		`^list students with grade master$`:                     s.listStudentsWithGradeMaster,
		`^system run job to migrate current_grade to grade_id$`: s.systemRunJobToMigrateCurrentGradeToGradeID,
		`^grade of students was migrated to grade_id$`:          s.gradeOfStudentsWasMigratedToGradeID,

		// sync student
		`^data log split store correct "([^"]*)"$`:                                                           s.storeLogDataSplitWithCorrectStatus,
		`^jprep sync "([^"]*)" students with action "([^"]*)" and "([^"]*)" students with action "([^"]*)"$`: s.jprepSyncStudentsWithActionAndStudentsWithAction,
		`^these students must be store in our system$`:                                                       s.theseStudentsMustBeStoreInOurSystem,

		// sync enrollment status
		`^school admin "([^"]*)" to the student$`:                           s.schoolAdminCreateOrderWithOrderStatusAndOrderType,
		`^the enrollment status histories of student must "([^"]*)"$`:       s.checkEnrollmentStatusHistoriesOfStudentOrderFlow,
		`^students were upserted successfully$`:                             s.studentsWereUpsertedSuccessfully,
		`^enrollment history with "([^"]*)" status of student was deleted$`: s.deleteEnrollmentHistoryWithStatusOfStudent,
		`^school admin simulate the "([^"]*)" to (\d+) student$`:            s.simulateTheOrderEvent,

		// sync staff
		`^after the deleted staff were "([^"]*)"$`:                                                       s.jprepSyncSyncDeletedStaffWithAction,
		`^JPREP sync "([^"]*)" staffs with action "([^"]*)" and "([^"]*)" staffs with action "([^"]*)"$`: s.jprepSyncStaffsWithActionAndStaffsWithAction,
		`^these staffs must be store in our system$`:                                                     s.theseStaffsMustBeStoreInOurSystem,
		`^they login our system and "([^"]*)" get self-profile info$`:                                    s.theyLoginOurSystemAndGetSelfProfileInfo,

		// update student with student phone number
		`^update student info with student phone number and contact preference with "([^"]*)"$`: s.updateStudentInfoWithStudentPhoneNumberAndContactPreference,
		`^student account updated success with student phone number and contact preference$`:    s.studentAccountUpdatedSuccessWithStudentPhoneNumberAndContactPreference,
		`^student account updated success with student phone number id and contact preference$`: s.studentAccountUpdatedSuccessWithStudentPhoneNumberIDAndContactPreference,

		// create student with grade master
		`^student info with grade master request$`:                s.studentInfoWithGradeMasterRequest,
		`^student info with invalid grade master request$`:        s.studentInfoWithInvalidGradeMasterRequest,
		`^new student account created success with grade master$`: s.newStudentAccountCreatedSuccessWithGradeMaster,

		// update student with grade master
		`^student info with grade master update request$`:         s.studentInfoWithGradeMasterUpdateRequest,
		`^student info with invalid grade master update request$`: s.studentInfoWithInvalidGradeMasterUpdateRequest,
		`^new student account updated success with grade master$`: s.newStudentAccountUpdatedSuccessWithGradeMaster,

		// migrate job to assign locations to user
		`^some "([^"]*)" users are existed in Manabie system$`:                   s.existedUserWithoutLocation,
		`^we run migration specify "([^"]*)" and pick "([^"]*)" with "([^"]*)"$`: s.runMigrationLocationsForUsers,
		`^existed "([^"]*)" with "([^"]*)" must be assigned locations$`:          s.usersMustHaveLocation,

		// migrate job to set current value by grade
		`^generate school history without current school$`:                       s.generateSchoolHistoryWithoutCurrentSchool,
		`^system run job to migrate set current school by grade in our system$`:  s.systemRunJobToMigrateSetCurrentSchoolByGradeInOurSystem,
		`^existing school history with current school value set by grade value$`: s.existingSchoolHistoryWithCurrentSchoolValueSetByGradeValue,

		// job to generate API Keypair
		`^system run job to generate API Key with organization "([^"]*)"$`: s.systemRunJobToGenerateAPIKeyWithOrganization,
		`^API keypair is created successfully$`:                            s.apiKeyIsCreatedSuccessfully,

		// shamir verify signature
		`^an invalid VerifySignatureRequest with "([^"]*)"$`: s.anInvalidVerifySignatureRequestWith,
		`^a client verifies signature$`:                      s.aClientVerifiesSignature,
		`^a valid VerifySignatureRequest$`:                   s.aValidVerifySignatureRequest,

		// register student class
		`^"([^"]*)" register class for a student`:      s.signedInUserRegisterClassForAStudent,
		`student package class must store in database`: s.studentClassMustStoreInDatabase,

		// job check enrollment status date
		`^enrollment status outdate in our system$`:                                      s.enrollmentStatusOutdateInOurSystem,
		`^system run job to disable access path location for outdate enrollment status$`: s.systemRunJobToDisableAccessPathLocationForOutdateEnrollmentStatus,
		`^student no longer access location when access path removed$`:                   s.studentNoLongerAccessLocationWhenAccessPathRemoved,

		// OpenAPI for student
		`^school admin creates (\d+) students with "([^"]*)" by OpenAPI in folder "([^"]*)"$`:      s.createStudentsByOpenAPI,
		`^school admin updates (\d+) students with "([^"]*)" by OpenAPI$`:                          s.updateStudentsByOpenAPI,
		`^students were upserted successfully by OpenAPI$`:                                         s.studentsWereSuccessfullyByOpenAPI,
		`^student were created unsuccessfully by OpenAPI with code "([^"]*)" and field "([^"]*)"$`: s.studentsWereCreatedUnsuccessfullyWithCodeAndField,
		`^student were updated unsuccessfully by OpenAPI with code "([^"]*)" and field "([^"]*)"$`: s.studentsWereUpdatedUnsuccessfullyWithCodeAndField,
		`^students were upserted by OpenAPI with failed (\d+) rows and successful (\d+) rows$`:     s.studentsWereUpsertedByOpenAPIWithErrorsCollection,

		// upsert student
		`^school admin create a student with "([^"]*)" and "([^"]*)" by GRPC$`:                    s.createStudentByGRPC,
		`^school admin update a student with "([^"]*)"$`:                                          s.updateStudentByGRPC,
		`^students were upserted successfully by GRPC$`:                                           s.studentsWereUpsertedSuccessfullyByGRPC,
		`^students were upserted unsuccessfully by GRPC with "([^"]*)" code and "([^"]*)" field$`: s.studentsWereUpsertedUnsuccessfullyByGRPCWithCodeAndField,
		// import students v2
		`^school admin create (\d+) students with "([^"]*)" by import in folder "([^"]*)"$`:                        s.createStudentsByImport,
		`^school admin update (\d+) students with "([^"]*)" by import$`:                                            s.updateStudentsByImport,
		`^student were updated unsuccessfully by import with code "([^"]*)" and field "([^"]*)" at row "([^"]*)"$`: s.studentWereUpdatedUnsuccessfulByImportWithError,
		`^student were created unsuccessfully by import with code "([^"]*)" and field "([^"]*)" at row "([^"]*)"$`: s.studentWereCreatedUnsuccessfulByImportWithError,
		`^students were upserted successfully by import$`:                                                          s.studentsWereUpsertedSuccessfulByImport,
		`^students were imported with failed (\d+) rows and successful (\d+) rows$`:                                s.studentsWereImportedWithErrorsCollection,

		// migrate enrollment status
		`^setup data to migrate enrollment status$`:            s.setUpDataToMigrateEnrollmentStatus,
		`^run job migrate enrollment status$`:                  s.runJobMigrateEnrollmentStatus,
		`^check data job migration enrollment status correct$`: s.checkDataJobMigrationEnrollmentStatusCorrect,

		// upsert parent by OpenAPI
		`^school admin creates parents "([^"]*)" with "([^"]*)" by OpenAPI$`:                       s.createParentByOpenAPI,
		`^parents were created by OpenAPI successfully$`:                                           s.parentsWereByOpenAPISuccessfully,
		`^parents were created by OpenAPI unsuccessfully with "([^"]*)" code and "([^"]*)" field$`: s.parentsWereCreatedByOpenAPIUnsuccessfully,

		// deactivate users
		`^a "([^"]*)" has been created successfully$`:                             s.createUserByRole,
		`^staff "([^"]*)" this user$`:                                             s.staffUpdateUserStatusSuccessfully,
		`^this "([^"]*)" "([^"]*)" login to the system$`:                          s.checkLoginStatus,
		`^this "([^"]*)" uses the old credential and "([^"]*)" get self profile$`: s.userGetSelfProfileByOldToken,
		`^staff "([^"]*)" "([^"]*)" users$`:                                       s.staffTryUpdateUserStatus,

		// auto deactivate and reactivate students
		`school admin creates a student with "([^"]*)" "([^"]*)" status is active and "([^"]*)" "([^"]*)" status is inactive`: s.upsertStudentWithEnrollmentStatuses,
		`school admin sees student is "([^"]*)"`:                     s.assertUserActivation,
		`school admin has created a student being "([^"]*)"`:         s.upsertStudentWithActiveStatus,
		`school admin create "([^"]*)" by Orders function`:           s.syncOrderToDeactivateAndReactivateStudents,
		`school admin has created a student who will be "([^"]*)"`:   s.upsertStudentWithStatusWillBe,
		`school admin has created a student who has "([^"]*)"`:       s.upsertStudentWithStatusWillBe,
		`system run daily job to deactivate and reactivate students`: s.systemRunJobToDisableAccessPathLocationForOutdateEnrollmentStatus,

		// withus: sync data
		`^schedule job uploaded a tsv data file in "([^"]*)" bucket$`: s.tsvDataFileInBucket,
		`^data were synced successfully$`:                             s.dataWereSyncedSuccessfully,
		`^system run job to sync data from bucket$`:                   s.systemRunJobToSyncDataFromBucket,
		// `^tsv data file in "([^"]*)" bucket to update$`: s.tsvDataFileInBucketToUpdate,
		// `^system does not have tsv data file in "([^"]*)" bucket$`: s.systemDoesNotHaveTsvDataFileInBucket,

		// validate user IP
		`^school admin's IP is "([^"]*)" whitelist and the IP restriction feature is "([^"]*)"$`: s.setUpUserIPAndFeatureConfig,
		`^school admin validates the IP address$`:                                                s.validateUserIPAddress,
		`^school admin sees the IP address is "([^"]*)"$`:                                        s.assertUserIPValidation,

		// reallocate student enrollment status
		`^school admin can not reallocate the student\'s enrollment status$`:             s.schoolAdminCanNotReallocateTheStudentsEnrollmentStatus,
		`^school admin reallocate the student\'s enrollment status with "([^"]*)" data$`: s.schoolAdminReallocateTheStudentsEnrollmentStatus,
		`^student\'s enrollment status was reallocated successfully$`:                    s.studentsEnrollmentStatusWasReallocatedSuccessfully,

		// parent activation
		`^school admin sees parent "([^"]*)"$`:                                s.validationParentActivation,
		`^school admin adds more "([^"]*)" student to the parent by OpenAPI$`: s.addMoreStudentToParent,
		`^school admin creates a parent with (\d+) student\(s\) by OpenAPI$`:  s.createsParentByOpenAPI,
		`^school admin removes (\d+) student from the parent by OpenAPI$`:     s.removeStudentFromParent,

		// unleash manager
		`^a scenario requires "([^"]*)" with corresponding statuses: "([^"]*)"$`: s.aScenarioRequiresWithCorrespondingStatuses,
		`^"([^"]*)" must be locked and have corresponding statuses: "([^"]*)"$`:  s.mustBeLockedAndHaveCorrespondingStatuses,

		// get user login email
		`^a user gets auth info by "([^"]*)" and "([^"]*)"$`:     s.getUserAuthInfoByLoginEmailAndDomainName,
		`^user receives login email and tenant id successfully$`: s.userReceivesLoginEmailAndTenantIDSuccessfully,

		// staff reset password
		`^a user reset password by "([^"]*)" and "([^"]*)" in "([^"]*)"$`: s.userResetPasswordWithLoginEmailAndDomainName,
		`^user received reset password email in "([^"]*)"$`:               s.userReceivedEmailWithContent,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
