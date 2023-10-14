package tom

import (
	"github.com/cucumber/godog"
)

func initLiveLessonChatSuite(previousMap map[string]interface{}, ctx *godog.ScenarioContext, s *suite) {
	mergedMap := map[string]interface{}{
		// lesson chat
		`^a teacher joins lesson creating new lesson session$`:                                                          s.aTeacherJoinsLessonCreatingNewLessonSession,
		`^teacher sees correct info calling LiveLessonConversationDetail$`:                                              s.teacherSeesCorrectInfoCallingLiveLessonConversationDetail,
		`^the second teacher sees "([^"]*)" messages with content "([^"]*)" calling LiveLessonConversationMessages$`:    s.theSecondTeacherSeesMessagesWithContentCallingLiveLessonConversationMessages,
		`^teacher sees correct latest message calling LiveLessonConversationDetail$`:                                    s.teacherSeesCorrectLatestMessageCallingLiveLessonConversationDetail,
		`^the first teacher sends "([^"]*)" message with content "([^"]*)" to live lesson chat$`:                        s.theFirstTeacherSendsMessageWithContentToLiveLessonChat,
		`^the "([^"]*)" in lesson receives "([^"]*)" message with type "([^"]*)" with content "([^"]*)"$`:               s.theInLessonReceivesMessageWithTypeWithContent,
		`^the second teacher sees "([^"]*)" messages calling LiveLessonConversationMessages$`:                           s.theSecondTeacherSeesMessagesCallingLiveLessonConversationMessages,
		`^a second teacher joins lesson refreshing lesson session$`:                                                     s.aSecondTeacherJoinsLessonRefreshingLessonSession,
		`^a second teacher joins lesson without refreshing lesson session$`:                                             s.aSecondTeacherJoinsLessonWithoutRefreshingLessonSession,
		`^The second teacher in lesson seen the conversation$`:                                                          s.theSecondTeacherInLessonSeenTheConversation,
		`^the second teacher sees "([^"]*)" status calling LiveLessonConversationDetail$`:                               s.theSecondTeacherSeesStatusCallingLiveLessonConversationDetail,
		`^students join lesson without refreshing lesson session$`:                                                      s.studentsJoinLessonWithoutRefreshingLessonSession,
		`^a student sends "([^"]*)" message with content "([^"]*)" to live lesson chat$`:                                s.aStudentSendsMessageWithContentToLiveLessonChat,
		`^students in lesson seen the conversation$`:                                                                    s.studentsInLessonSeenTheConversation,
		`^students sees "([^"]*)" status calling LiveLessonConversationDetail$`:                                         s.studentsSeesStatusCallingLiveLessonConversationDetail,
		`^teacher sees empty latest message calling LiveLessonConversationDetail$`:                                      s.teacherSeesEmptyLatestMessageCallingLiveLessonConversationDetail,
		`^teacher sees "([^"]*)" status calling LiveLessonConversationDetail$`:                                          s.teacherSeesStatusCallingLiveLessonConversationDetail,
		`^a teacher sends "([^"]*)" message with content "([^"]*)" to live lesson chat$`:                                s.aTeacherSendsMessageWithContentToLiveLessonChat,
		`^the "([^"]*)" in lesson receive silent notification with content "([^"]*)"$`:                                  s.theInLessonReceiveSilentNotificationWithContent,
		`^a second teacher joins lesson but not subscribe stream$`:                                                      s.aSecondTeacherJoinsLessonButNotSubscribeStream,
		`^"([^"]*)" sends "([^"]*)" message to the live lesson private conversation with content "([^"]*)"$`:            s.sendsMessageToThePrivateConversationWithContent,
		`^"([^"]*)" sees "([^"]*)" messages with content "([^"]*)" when get live lesson private conversation messages$`: s.userSeeMessageWithContentInPrivateConversation,
		`^"([^"]*)" refresh live lesson session for private conversation$`:                                              s.userRefreshLiveLessonSessionForPrivateConversation,
		`^multiple teacher create new live lesson private conversations with a student$`:                                s.multipleTeacherCreatePrivateConversationsWithAStudent,
		`^multiple teacher sends "([^"]*)" message to the live lesson private conversation with content "([^"]*)"$`:     s.multipleTeacherSendMessageToPrivateConversations,
		`^"([^"]*)" sees "([^"]*)" messages with content "([^"]*)" when get messages in all private conversations$`:     s.userSeeMessageWithContentInAllPrivateConversation,
		`live lesson conversation and all private conversations have latest start time with the same date`:              s.verifyAllConversationsHaveTheSameLatestStartTime,

		// lesson events
		`^a lesson conversation with "([^"]*)" teachers and "([^"]*)" students$`:                                                  s.aLessonConversationWithTeachersAndStudents,
		`^a valid "([^"]*)" id in JoinLesson$`:                                                                                    s.aValidIDInJoinLesson,
		`^tom "([^"]*)" add above user to this lesson conversation$`:                                                              s.tomAddAboveUserToThisLessonConversation,
		`^a ConversationByLessonRequest$`:                                                                                         s.aConversationByLessonRequest,
		`^a "([^"]*)" in ConversationByLessonRequest$`:                                                                            s.aInConversationByLessonRequest,
		`^a user get all conversation of lesson$`:                                                                                 s.aUserGetAllConversationOfLesson,
		`^tom must return (\d+) conversation of lesson$`:                                                                          s.tomMustReturnConversationOfLesson,
		`^a EvtLesson with message "([^"]*)"$`:                                                                                    s.aEvtLessonWithMessage,
		`^bob send event EvtLesson$`:                                                                                              s.bobSendEventEvtLesson,
		`^tom must create conversation for all lesson$`:                                                                           s.tomMustCreateConversationForAllLesson,
		`^a EvtLesson with message CreateLesson with (\d+) students$`:                                                             s.aEvtLessonWithMessageCreateLessonWithStudents,
		`^tom must create conversation member for student in CreateLesson$`:                                                       s.tomMustCreateConversationMemberForStudentInCreateLesson,
		`^bob send LeaveLesson for one of previous "([^"]*)"$`:                                                                    s.bobSendLeaveLessonForOneOfPrevious,
		`^tom must remove teacher from conversation$`:                                                                             s.tomMustRemoveTeacherFromConversation,
		`^tom must not remove student from conversation$`:                                                                         s.tomMustNotRemoveStudentFromConversation,
		`^bob send UpdateLesson with "([^"]*)" new student and without previous students$`:                                        s.bobSendUpdateLessonWithNewStudentAndWithoutPreviousStudents,
		`^tom must correctly store only latest students in lesson conversation$`:                                                  s.tomMustCorrectlyStoreOnlyLatestStudentsInLessonConversation,
		`^bob send UpdateLesson with "([^"]*)" new student and without "([^"]*)" previous students$`:                              s.bobSendUpdateLessonWithNewStudentAndWithoutPreviousStudents,
		`^yasuo send EventSyncUserCourse inserting "([^"]*)" students and deleting "([^"]*)" prevous student for current lesson$`: s.yasuoSendEventSyncUserCourseInsertingStudentsAndDeletingPrevousStudentForCurrentLesson,
		`^teacher see deleted message in lesson chat$`:                                                                            s.teacherSeeDeletedChatInLessonChat,
	}
	applyMergedSteps(ctx, previousMap, mergedMap)
}
