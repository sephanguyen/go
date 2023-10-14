package tom

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	tomconst "github.com/manabie-com/backend/internal/tom/constants"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type chatInfo struct {
	name      string
	id        string
	courses   []string
	chatType  string
	student   *upb.Student
	studentID string
	parents   []*upb.Parent
	parentIDs []string
	replied   bool
}

func (c chatInfo) getOneUserID() string {
	if c.chatType == "Student" {
		return c.student.UserProfile.UserId
	}
	return c.parents[0].UserProfile.UserId
}

func initStudentParentTeacherSuite(previousMap map[string]interface{}, ctx *godog.ScenarioContext, s *suite) {
	mergedMap := map[string]interface{}{
		// student teachers
		`^a chat between a student and "([^"]*)" teachers$`:       s.aChatBetweenAStudentAndTeachers,
		`^student receives sent message$`:                         s.studentReceivesSentMessage,
		`^student receives notification$`:                         s.studentReceivesNotification,
		`^a teacher sends "([^"]*)" item with content "([^"]*)"$`: s.aTeacherSendsItemWithContent,
		`^teachers receive sent message$`:                         s.teachersReceiveSentMessage,
		`^student and teachers are present$`:                      s.studentAndTeachersArePresent,
		`^student is not present$`:                                s.studentIsNotPresent,
		`^student is present$`:                                    s.studentIsPresent,
		`^student send message "([^"]*)"$`:                        s.studentSendMessage,
		`^student sends "([^"]*)" item with content "([^"]*)"$`:   s.studentSendsItemWithContent,
		`^teachers are not present$`:                              s.teachersAreNotPresent,
		`^teachers are present$`:                                  s.teachersArePresent,
		`^other teachers receive sent message$`:                   s.otherTeachersReceiveSentMessage,
		`^student seen the message$`:                              s.studentSeenTheMessage,
		`^student go to messages on learner app$`:                 s.studentGoToMessagesOnLearnerApp,
		`^the screen displays (\d+) student chat$`:                s.theScreenDisplaysStudentChat,
		`^student can see (\d+) conversations$`:                   s.studentCanSeeConversations,

		// parent teachers
		`^account for parent of these kids is created$`:                                               s.accountForParentOfTheseKidsIsCreated,
		`^chats are created for parent$`:                                                              s.chatsAreCreatedForParent,
		`^each parent chat has name assigned to the kids\' name$`:                                     s.eachParentChatHasNameAssignedToTheKidsName,
		`^"([^"]*)" student-teacher chats$`:                                                           s.studentTeacherChats,
		`^a student conversation$`:                                                                    s.aStudentConversation,
		`^another parent account is created with event "([^"]*)"$`:                                    s.anotherParentAccountIsCreatedWithEvent,
		`^chats between a parent and "([^"]*)" teachers to manage "([^"]*)" kids$`:                    s.chatsBetweenAParentAndTeachersToManageKids,
		`^this parent is added in these chats$`:                                                       s.thisParentIsAddedInTheseChats,
		`^a chat between "([^"]*)" parents and "([^"]*)" teachers$`:                                   s.aChatBetweenParentsAndTeachers,
		`^a parent sends "([^"]*)" item with content "([^"]*)"$`:                                      s.aParentSendsItemWithContent,
		`^other parents receive sent message$`:                                                        s.otherParentsReceiveSentMessage,
		`^parents are present$`:                                                                       s.parentsArePresent,
		`^parents and teachers are present$`:                                                          s.parentsAndTeachersArePresent,
		`^parents receive sent message$`:                                                              s.parentsReceiveSentMessage,
		`^parents are not present$`:                                                                   s.parentsAreNotPresent,
		`all parents are added into chat groups$`:                                                     s.allParentsAreAddedIntoChatGroups,
		`^"([^"]*)" parents account exist before "([^"]*)" student accounts are created$`:             s.parentsAccountExistBeforeStudentAccountsAreCreated,
		`^a parent seen the message$`:                                                                 s.aParentSeenTheMessage,
		`^teachers see the message has been read$`:                                                    s.teachersSeeTheMessageHasBeenRead,
		`^parent can view "([^"]*)" chats on learner app$`:                                            s.parentCanViewChatsOnLearnerApp,
		`^parents receive notification$`:                                                              s.parentsReceiveNotification,
		`^"([^"]*)" device tokens exist in DB$`:                                                       s.deviceTokensExistInDB,
		`^location configurations conversation value "([^"]*)" existed on DB$`:                        s.initLocationConversationConfigInDB,
		`"([^"]*)" conversation is disabled in location configurations table with location "([^"]*)"`: s.disableChatInLocationConfigTable,
		`^a student and conversation created$`:                                                        s.aStudentAndConversationCreated,
		`^a parent and conversation created$`:                                                         s.aParentAndConversationCreated,
		`^teacher join conversation of "([^"]*)"$`:                                                    s.teacherJoinConversationOf,
	}
	applyMergedSteps(ctx, previousMap, mergedMap)
}

func (s *suite) resourcePathOfSchoolIsApplied(ctx context.Context, schoolname string) (context.Context, error) {
	switch schoolname {
	case "Manabie":
		s.filterSuiteState.defaultLocationID = constants.ManabieOrgLocation
		s.CommonSuite.DefaultLocationID = constants.ManabieOrgLocation
		s.CommonSuite.DefaultLocationTypeID = ManabieOrgLocationType
		s.schoolID = fmt.Sprint(constants.ManabieSchool)
		ctx = contextWithResourcePath(ctx, strconv.Itoa(constants.ManabieSchool))
	default:
		return ctx, fmt.Errorf("unknown school %s", schoolname)
	}
	return ctx, nil
}

func (s *suite) teachersDeviceTokensIsExistedInDB(ctx context.Context) (context.Context, error) {
	for _, teacherID := range s.teachersInConversation {
		ctx, err := s.insertDeviceToken(ctx, teacherID, fmt.Sprintf("teacher-%s", teacherID))
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) insertDeviceToken(ctx context.Context, userID string, userName string) (context.Context, error) {
	repo := &repositories.UserDeviceTokenRepo{}
	if err := repo.Upsert(ctx, s.DB, &entities.UserDeviceToken{
		UserID:            pgtype.Text{String: userID, Status: pgtype.Present},
		Token:             pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
		AllowNotification: pgtype.Bool{Bool: true, Status: pgtype.Present},
		UserName:          pgtype.Text{String: userName, Status: pgtype.Present},
	}); err != nil {
		return ctx, err
	}
	return ctx, nil
}
func (s *suite) deviceTokensExistInDB(ctx context.Context, people string) (context.Context, error) {
	var userIDs []string
	switch people {
	case "parents":
		userIDs = s.parentIDs
	case "students in lesson":
		userIDs = s.LessonChatState.studentsInLesson
	case "second teacher in lesson":
		userIDs = append(userIDs, s.LessonChatState.secondTeacher)
	default:
		panic(fmt.Errorf("unsupported arg: %s", people))
	}
	for _, id := range userIDs {
		ctx, err := s.insertDeviceToken(ctx, id, fmt.Sprintf("%s-%s", people, id))
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) parentCanViewChatsOnLearnerApp(ctx context.Context, numChat int) (context.Context, error) {
	children := s.childrenIDs
	if len(children) != numChat {
		return ctx, fmt.Errorf("has %d children while expecting %d chats to return", len(children), numChat)
	}
	userID := s.singleParentID
	// _, par, err := s.CommonSuite.CreateStudentWithParent(ctx, nil, nil)
	// if err != nil {
	// 	return ctx, err
	// }
	// userID = par.UserProfile.UserId
	// ctx, err := s.aValidUser(ctx, constants.ManabieSchool, withID(userID), withRole(cpb.UserGroup_USER_GROUP_PARENT.String()))
	// if err != nil {
	// 	return ctx, err
	// }
	token, err := s.generateExchangeToken(userID, cpb.UserGroup_USER_GROUP_PARENT.String(), applicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	return ctx, try.Do(func(attempt int) (bool, error) {
		res, err := pb.NewChatServiceClient(s.Conn).ConversationList(
			metadata.AppendToOutgoingContext(contextWithValidVersion(context.Background()), "token", token),
			&pb.ConversationListRequest{
				Limit: 100,
			},
		)
		if err != nil {
			return false, err
		}
		if len(res.GetConversations()) != numChat {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("expect %d chats, %d returned", numChat, len(res.GetConversations()))
		}
		checkList := map[string]bool{}
		for _, child := range children {
			checkList[child] = false
		}
		for _, item := range res.GetConversations() {
			if item.ConversationType != pb.CONVERSATION_PARENT {
				return false, fmt.Errorf("expect conversation parent, got %s", item.ConversationType.String())
			}
			checked, ok := checkList[item.GetStudentId()]
			if !ok {
				return false, fmt.Errorf("not expecting student with id %s to appear in current parent chat", item.GetStudentId())
			}
			if checked {
				return false, fmt.Errorf("student with id %s has more than one parent chat", item.GetStudentId())
			}
			checkList[item.GetStudentId()] = true
		}

		for studentID, check := range checkList {
			if !check {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("parent chat for student %s is not returned", studentID)
			}
		}
		return false, nil
	})
}
func (s *suite) theScreenDisplaysStudentChat(ctx context.Context, numChat int) (context.Context, error) {
	recentRes := s.Response.(*pb.ConversationListResponse)
	if len(recentRes.GetConversations()) != numChat {
		return ctx, fmt.Errorf("expect %d chats, %d returned", numChat, len(recentRes.GetConversations()))
	}
	studentID := s.studentID
	for _, item := range recentRes.GetConversations() {
		if item.ConversationType != pb.CONVERSATION_STUDENT {
			return ctx, fmt.Errorf("expect conversation student, got %s", item.ConversationType.String())
		}
		if item.StudentId != studentID {
			return ctx, fmt.Errorf("conversation returned with student id %s not equal to current student id %s", item.StudentId, studentID)
		}
	}

	return ctx, nil
}
func (s *suite) studentGoToMessagesOnLearnerApp(ctx context.Context) (context.Context, error) {
	userID := s.studentID
	token, err := s.genStudentToken(userID)
	if err != nil {
		return ctx, err
	}

	res, err := pb.NewChatServiceClient(s.Conn).ConversationList(
		contextWithToken(context.Background(), token),
		&pb.ConversationListRequest{
			Limit: 100,
		},
	)
	if err != nil {
		return ctx, err
	}

	s.Response = res
	return ctx, nil
}
func (s *suite) studentSeenTheMessage(ctx context.Context) (context.Context, error) {
	req := &pb.SeenMessageRequest{
		ConversationId: s.conversationID,
	}
	userID := s.studentID

	token, err := s.generateExchangeToken(userID, cpb.UserGroup_USER_GROUP_STUDENT.String(), applicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return ctx, err
	}

	_, err = pb.NewChatServiceClient(s.Conn).
		SeenMessage(contextWithToken(context.Background(), token), req)

	return ctx, err
}
func (s *suite) aParentSeenTheMessage(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SeenMessageRequest{
		ConversationId: s.conversationID,
	}
	userID := s.parentIDs[0]

	// ctx, err := s.aValidUser(ctx2, constants.ManabieSchool, withID(userID), withRole(cpb.UserGroup_USER_GROUP_PARENT.String()))
	// if err != nil {
	// 	return ctx, fmt.Errorf("generate valid user %w", err)
	// }
	token, err := s.generateExchangeToken(userID, cpb.UserGroup_USER_GROUP_PARENT.String(), applicantID, constants.ManabieSchool, s.ShamirConn)

	if err != nil {
		return ctx, fmt.Errorf("generateExchangeToken %w", err)
	}

	_, err = pb.NewChatServiceClient(s.Conn).
		SeenMessage(contextWithToken(ctx2, token), req)
	if err != nil {
		return ctx, fmt.Errorf("SeenMessage: %w", err)
	}
	return ctx, nil
}
func (s *suite) teachersSeeTheMessageHasBeenRead(ctx context.Context) (context.Context, error) {
	g := new(errgroup.Group)
	for _, userID := range s.teachersInConversation {
		stream, exist := s.SubV2Clients[userID]
		if !exist {
			return ctx, fmt.Errorf("stream for user %s does not exist", userID)
		}
		g.Go(func() error {
			defer func() {
				err := stream.CloseSend()
				if err != nil {
					s.ZapLogger.Error(fmt.Sprintf("error closing stream: %s", err))
				}
			}()
			var newMsg *pb.MessageResponse

			for try := 0; try < 10; try++ {
				resp, err := stream.Recv()
				if err == io.EOF {
					return fmt.Errorf("received eof before try runs out")
				}
				newMsg = resp.GetEvent().GetEventNewMessage()
				if newMsg == nil {
					continue
				}
				if newMsg.Type == pb.MESSAGE_TYPE_SYSTEM && newMsg.Content == tpb.CodesMessageType_CODES_MESSAGE_TYPE_SEEN_CONVERSATION.String() {
					break
				}
			}
			if newMsg == nil {
				return fmt.Errorf("received nil after 10 tries")
			}
			if newMsg.GetConversationId() != s.conversationID {
				return fmt.Errorf("expting seen message for conversation %s returned, got %s", s.conversationID, newMsg.GetConversationId())
			}
			return nil
		})
	}
	return ctx, nil
}
func (s *suite) findCreatedParentChat(ctx context.Context, studentID string, parentIDs []string, schoolID string, locations []string) (chatID, chatName string, err error) {
	if err := try.Do(func(attempt int) (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()
		query := `SELECT c.name,c.conversation_id FROM conversation_students cs LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
    WHERE cs.student_id = $1 AND owner = $2 AND c.status= 'CONVERSATION_STATUS_NONE' AND cs.conversation_type = 'CONVERSATION_PARENT'`
		if err := s.DB.QueryRow(ctx2, query, studentID, schoolID).Scan(&chatName, &chatID); err != nil {
			time.Sleep(2 * time.Second)
			return attempt < 10, fmt.Errorf("s.DB.QueryRow: %w", err)
		}
		return false, nil
	}); err != nil {
		return "", "", err
	}

	checkConversationMember := `SELECT count(*) FROM conversation_members cm WHERE cm.conversation_id =$1`
	err = try.Do(func(attempt int) (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()
		var count int
		err := s.DB.QueryRow(ctx2, checkConversationMember, chatID).Scan(&count)
		if err != nil {
			return attempt < 5, err
		}

		if count != len(parentIDs) {
			time.Sleep(2 * time.Second)
			return attempt < 10, fmt.Errorf("only %d members found in chat, expecting %d", count, len(parentIDs))
		}
		return false, nil
	})
	if err != nil {
		return "", "", err
	}
	// check if location is correct
	checkLocs := "select count(*) from conversation_locations cl where cl.conversation_id=$1 and location_id=ANY($2)"
	var count pgtype.Int8
	if err := s.DB.QueryRow(ctx, checkLocs, chatID, database.TextArray(locations)).Scan(&count); err != nil {
		return "", "", err
	}
	if int(count.Int) != len(locations) {
		return "", "", fmt.Errorf("conversation %s has %d locations among %v expected locations", chatID, count.Int, locations)
	}
	return chatID, chatName, err
}

func (s *suite) allParentsAreAddedIntoChatGroups(ctx context.Context) (context.Context, error) {
	locations := []string{s.filterSuiteState.defaultLocationID}
	schoolID := s.getSchool()
	for _, childrenID := range s.childrenIDs {
		chatid, chatname, err := s.findCreatedParentChat(ctx, childrenID, s.parentIDs, strconv.Itoa(int(schoolID)), locations)
		if err != nil {
			return ctx, err
		}
		s.parentChats[childrenID] = chatInfo{
			name: chatname,
			id:   chatid,
		}
	}
	return ctx, nil
}
func (s *suite) parentsAccountExistBeforeStudentAccountsAreCreated(ctx context.Context, numParents int, numKids int) (context.Context, error) {
	s.filterSuiteState.defaultLocationID = constants.ManabieOrgLocation
	parentsIDs := []string{}
	for i := 0; i < numKids; i++ {
		newID := idutil.ULIDNow()
		s.childrenIDs = append(s.childrenIDs, newID)
	}
	for i := 0; i < numParents; i++ {
		newID := idutil.ULIDNow()
		parentsIDs = append(parentsIDs, newID)
		ctx, err := s.publishParentChildrensEvent(ctx, newID, "CreateParent")
		if err != nil {
			return ctx, err
		}
	}
	for _, childrenID := range s.childrenIDs {
		s.parentIDs = parentsIDs
		request := &upb.EvtUser{
			Message: &upb.EvtUser_CreateStudent_{
				CreateStudent: &upb.EvtUser_CreateStudent{
					StudentId:   childrenID,
					StudentName: pseudoNameForParentChat(childrenID, langEng),
					SchoolId:    resourcePathFromCtx(ctx),
					LocationIds: []string{constants.ManabieOrgLocation},
				},
			},
		}
		data, err := proto.Marshal(request)
		if err != nil {
			return ctx, err
		}
		_, err = s.JSM.TracedPublishAsync(ctx, "nats.TracedPublishAsync", constants.SubjectUserCreated, data)
		if err != nil {
			return ctx, fmt.Errorf("s.JSM.TracedPublishAsync: %w", err)
		}
	}

	return ctx, nil
}
func (s *suite) parentsAreNotPresent(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
func (s *suite) parentsAndTeachersArePresent(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx, s.parentsArePresent, s.teachersArePresent)
}
func (s *suite) parentsArePresent(ctx context.Context) (context.Context, error) {
	toks, err := s.generateExchangeTokens(s.parentIDs, cpb.UserGroup_USER_GROUP_PARENT.String(), applicantID, s.getSchool(), s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	return s.makeUsersSubscribeV2Ctx(ctx, s.parentIDs, toks)
}
func (s *suite) aParentSendsItemWithContent(ctx context.Context, msgType string, msgContent string) (context.Context, error) {
	userID := s.parentIDs[0]
	s.parentWhoSentMessage = userID
	return s.userSendsItemWithContent(ctx, msgType, msgContent, userID, cpb.UserGroup_USER_GROUP_PARENT)
}
func (s *suite) parentsReceiveSentMessage(ctx context.Context) (context.Context, error) {
	return s.usersReceiveSentMessage(ctx, s.parentIDs)
}
func (s *suite) otherParentsReceiveSentMessage(ctx context.Context) (context.Context, error) {
	otherParents := []string{}
	for _, parentID := range s.parentIDs {
		if parentID != s.parentWhoSentMessage {
			otherParents = append(otherParents, parentID)
		}
	}
	return s.usersReceiveSentMessage(ctx, otherParents)
}
func (s *suite) aChatBetweenParentsAndTeachers(ctx context.Context, numParent int, numTeacher int) (context.Context, error) {
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32ResourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	locations := []string{constants.ManabieOrgLocation}
	schoolID := strconv.Itoa(constants.ManabieSchool)
	stu, par, err := s.createStudentParentByAPI(ctx, locations)
	if err != nil {
		return ctx, err
	}
	ctx, chatinfos, err := s.findStudentParentChat(ctx, stu.UserProfile.UserId, par.UserProfile.UserId, schoolID, locations, stu, par)
	if err != nil {
		return ctx, err
	}
	s.parentChats[stu.UserProfile.UserId] = chatinfos[1]

	stuID := stu.UserProfile.UserId
	s.studentID = stuID
	s.childrenIDs = append(s.childrenIDs, stuID)
	s.conversationID = chatinfos[1].id
	s.chatName = chatinfos[1].name
	parIDs := []string{par.UserProfile.UserId}
	for i := 0; i < numParent-1; i++ {
		par, err := s.CommonSuite.CreateParentForStudent(ctx, stuID)
		if err != nil {
			return ctx, err
		}
		parIDs = append(parIDs, par.UserProfile.UserId)
	}
	s.parentIDs = parIDs
	_, _, err = s.findCreatedParentChat(ctx, stuID, parIDs, schoolID, locations)
	if err != nil {
		return ctx, err
	}
	for i := 0; i < numTeacher; i++ {
		prof, tok, err := s.CommonSuite.CreateTeacher(ctx)
		if err != nil {
			return ctx, err
		}
		s.teachersInConversation = append(s.teachersInConversation, prof.StaffId)
		req := &tpb.JoinConversationsRequest{
			ConversationIds: []string{
				s.conversationID,
			},
		}

		err = try.Do(func(attempt int) (bool, error) {
			_, err = tpb.NewChatModifierServiceClient(s.Conn).
				JoinConversations(contextWithToken(context.Background(), tok), req)
			if err != nil {
				time.Sleep(1 * time.Second)
				return attempt < 5, err
			}
			return false, nil
		})

		if err != nil {
			return ctx, err
		}
	}
	return ctx, err
}

func (s *suite) checkParentAddedIntoChats(ctx context.Context, parentIDs []string) (context.Context, error) {
	for _, newlyAddedParent := range parentIDs {
		for _, chatInfo := range s.parentChats {
			chatID := chatInfo.id
			parentID := newlyAddedParent
			err := try.Do(func(attempt int) (bool, error) {
				time.Sleep(2 * time.Second)

				ctx2, cancel := context.WithTimeout(ctx, time.Second*10)
				defer cancel()

				var count int
				checkConversationMember := `SELECT count(*) FROM conversation_members cm WHERE cm.conversation_id =$1 AND cm.user_id =$2`
				err := s.DB.QueryRow(ctx2, checkConversationMember, chatID, parentID).Scan(&count)
				if err != nil {
					return false, err
				}
				if count > 0 {
					return false, nil
				}
				return attempt < 5, fmt.Errorf("parent is not yet added")
			})
			if err != nil {
				return ctx, err
			}
		}
	}
	return ctx, nil
}
func (s *suite) thisParentIsAddedInTheseChats(ctx context.Context) (context.Context, error) {
	newlyAddedParent := s.additionalParentID
	if newlyAddedParent == "" {
		return ctx, fmt.Errorf("can't find newly added parent id")
	}
	return s.checkParentAddedIntoChats(ctx, []string{newlyAddedParent})
}

func (s *suite) chatsBetweenAParentAndTeachersToManageKids(ctx context.Context, numberOfTeacher int, numberOfKids int) (context.Context, error) {
	teacherIDs := []string{}
	kidIDs := []string{}
	for i := 0; i < numberOfTeacher; i++ {
		teacherIDs = append(teacherIDs, idutil.ULIDNow())
	}

	for i := 0; i < numberOfKids; i++ {
		ctx, err := godogutil.MultiErrChain(ctx,
			s.createStudentConversation,
			s.multipleTeachersJoinConversation, teacherIDs,
		)
		if err != nil {
			return ctx, err
		}
		kidIDs = append(kidIDs, s.studentID)
		s.studentID = ""
	}

	s.childrenIDs = kidIDs
	ctx, err := godogutil.MultiErrChain(ctx,
		s.accountForParentOfTheseKidsIsCreated,
		s.chatsAreCreatedForParent,
		s.eachParentChatHasNameAssignedToTheKidsName,
	)
	if err != nil {
		return ctx, err
	}

	s.teachersInConversation = teacherIDs
	return ctx, nil
}

func (s *suite) anotherParentAccountIsCreatedWithEvent(ctx context.Context, evttype string) (context.Context, error) {
	newParentID := idutil.ULIDNow()
	s.parentIDs = append(s.parentIDs, newParentID)
	s.additionalParentID = newParentID
	ctx, err := s.publishParentChildrensEvent(ctx, newParentID, evttype)
	return ctx, err
}

func (s *suite) chatsAreCreatedForParent(ctx context.Context) (context.Context, error) {
	for _, childrenID := range s.childrenIDs {
		studentID := childrenID
		var chatName, chatID string
		if err := try.Do(func(attempt int) (bool, error) {
			ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			query := `SELECT c.name,c.conversation_id FROM conversation_students cs LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
    WHERE cs.student_id = $1 AND owner = $2  AND c.status= 'CONVERSATION_STATUS_NONE' AND cs.conversation_type = 'CONVERSATION_PARENT'`
			if err := s.DB.QueryRow(ctx2, query, studentID, resourcePathFromCtx(ctx)).Scan(&chatName, &chatID); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					time.Sleep(2 * time.Second)
					return attempt < 5, err
				}
				return false, err
			}
			return false, nil
		}); err != nil {
			return ctx, err
		}
		s.parentChats[studentID] = chatInfo{
			name: chatName,
			id:   chatID,
		}
	}

	return ctx, nil
}
func (s *suite) eachParentChatHasNameAssignedToTheKidsName(ctx context.Context) (context.Context, error) {
	if len(s.childrenIDs) != len(s.parentChats) {
		return ctx, fmt.Errorf("expect chats created for parent (%d) equal number of kid (%d)", len(s.parentChats), len(s.childrenIDs))
	}
	var invalid pgtype.Int8
	// find all student/parent pairs having different name
	if err := s.DB.QueryRow(ctx, `
select count(*) from conversations c1 left join conversation_students cs1
on cs1.conversation_id=c1.conversation_id
left join conversation_students cs2 on cs1.student_id=cs2.student_id left join conversations c2 
on cs2.conversation_id=c2.conversation_id
where c1.conversation_type='CONVERSATION_STUDENT' and c1.name!=c2.name and cs1.student_id=ANY($1) 
	`, s.childrenIDs).Scan(&invalid); err != nil {
		return ctx, err
	}
	if invalid.Int > 0 {
		return ctx, fmt.Errorf("exist student chat name different from parent chat name in ids %v", s.childrenIDs)
	}
	return ctx, nil
}
func (s *suite) studentTeacherChats(ctx context.Context, studentNum int) (context.Context, error) {
	childrenIDs := []string{}
	teacherID := idutil.ULIDNow()
	for i := 0; i < studentNum; i++ {
		ctx, err := godogutil.MultiErrChain(ctx,
			s.createStudentConversation,
			s.multipleTeachersJoinConversation, []string{teacherID},
		)
		if err != nil {
			return ctx, err
		}
		childrenIDs = append(childrenIDs, s.studentID)
		s.studentID = ""
	}
	s.childrenIDs = childrenIDs
	return ctx, nil
}
func pseudoNameForStudentChat(studentID string, language string) string {
	nameGenLock.Lock()
	defer nameGenLock.Unlock()
	name, ok := idToName[studentID]
	if !ok {
		dictionary := studentNames[language]
		h := fnv.New32a()
		h.Write([]byte(studentID))
		indx := int(h.Sum32()) % len(dictionary)
		idToName[studentID] = dictionary[indx]
		return dictionary[indx]
	}
	return name
}
func pseudoNameForParentChat(studentID string, language string) string {
	nameGenLock.Lock()
	defer nameGenLock.Unlock()
	name, ok := idToName[studentID]
	if !ok {
		dictionary := studentNames[language]
		h := fnv.New32a()
		h.Write([]byte(studentID))
		indx := int(h.Sum32()) % len(dictionary)
		idToName[studentID] = dictionary[indx]
		return dictionary[indx]
	}
	return name
}

func (s *suite) publishParentChildrensEvent(ctx context.Context, parentID string, evttype string) (context.Context, error) {
	for _, studentID := range s.childrenIDs {
		var (
			data []byte
			err  error
			subj string
		)
		switch evttype {
		case "CreateParent":
			subj = constants.SubjectUserCreated
			request := &bpb.EvtUser{
				Message: &bpb.EvtUser_CreateParent_{
					CreateParent: &bpb.EvtUser_CreateParent{
						StudentId:   studentID,
						StudentName: pseudoNameForParentChat(studentID, langEng),
						SchoolId:    s.schoolID,
						ParentId:    parentID,
					},
				},
			}
			data, err = proto.Marshal(request)
		case "ParentAssignedToStudent":
			subj = constants.SubjectUserUpdated
			request := &upb.EvtUser{
				Message: &upb.EvtUser_ParentAssignedToStudent_{
					ParentAssignedToStudent: &upb.EvtUser_ParentAssignedToStudent{
						ParentId:  parentID,
						StudentId: studentID,
					},
				},
			}
			data, err = proto.Marshal(request)
		default:
			return ctx, fmt.Errorf("invalid event type %s", evttype)
		}

		if err != nil {
			return ctx, err
		}
		_, err = s.JSM.TracedPublishAsync(ctx, "publishParentChildrenEvent", subj, data)
		if err != nil {
			return ctx, fmt.Errorf("s.JSM.TracedPublishAsync: %w", err)
		}
	}
	return ctx, nil
}
func (s *suite) accountForParentOfTheseKidsIsCreated(ctx context.Context) (context.Context, error) {
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32ResourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	first := s.childrenIDs[0]
	par, err := s.CommonSuite.CreateParentForStudent(ctx, first)
	if err != nil {
		return ctx, err
	}
	s.singleParentID = par.UserProfile.UserId
	for _, children := range s.childrenIDs[1:] {
		err := s.CommonSuite.UpdateStudentParent(ctx, children, par.UserProfile.UserId, par.UserProfile.Email)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) ensureUserHasDeviceToken(ctx context.Context, userID string) (context.Context, error) {
	rand.Seed(time.Now().UnixNano())

	repo := &repositories.UserDeviceTokenRepo{}
	if err := repo.Upsert(ctx, s.DB, &entities.UserDeviceToken{
		UserID:            pgtype.Text{String: userID, Status: pgtype.Present},
		Token:             pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
		AllowNotification: pgtype.Bool{Bool: true, Status: pgtype.Present},
		UserName:          pgtype.Text{String: userID, Status: pgtype.Present},
	}); err != nil {
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) ensureUserDevicePingCtx(ctx context.Context, userID string, token string, durationSec int) (context.Context, error) {
	c := pb.NewChatServiceClient(s.Conn)

	t, _ := jwt.ParseString(token)
	r := &repositories.OnlineUserRepo{}
	since := time.Now().Add(-5 * time.Second)
	err := try.Do(func(attempt int) (retry bool, err error) {
		users, err := r.OnlineUserDBRepo.Find(ctx, s.DB, database.TextArray([]string{t.Subject()}), pgtype.Timestamptz{Time: since, Status: 2})
		if err != nil {
			return false, err
		}

		if len(users) == 0 {
			time.Sleep(1 * time.Second)
			return attempt < 5, fmt.Errorf("not found user online")
		}

		return false, nil
	})
	if err != nil {
		s.ZapLogger.Sugar().Infof("manual check user_online user_id = %s since %s", t.Subject(), since.String())
		return ctx, err
	}

	streamV2, ok := s.SubV2Clients[t.Subject()]
	if !ok {
		return ctx, errors.New("user did not subscribe before")
	}

	sessionID := ""
	for try := 0; try < 4; try++ {
		resp, err := streamV2.Recv()
		if err != nil {
			return ctx, err
		}

		if resp.Event.GetEventPing() == nil {
			continue
		}
		sessionID = resp.Event.GetEventPing().SessionId
		break
	}
	if sessionID == "" {
		return ctx, fmt.Errorf("attempt to get session id from upstream but failed")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := c.PingSubscribeV2(
				metadata.AppendToOutgoingContext(metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)),
				&pb.PingSubscribeV2Request{
					SessionId: sessionID,
				})
			if err != nil {
				return
			}

			time.Sleep(time.Duration(durationSec) * time.Second)
		}
	}()

	return ctx, nil
}
func (s *suite) aChatBetweenAStudentAndTeachers(ctx context.Context, numberOfTeacher int) (context.Context, error) {
	teacherIDs := []string{}
	for i := 0; i < numberOfTeacher; i++ {
		teacherIDs = append(teacherIDs, idutil.ULIDNow())
	}
	return godogutil.MultiErrChain(ctx,
		s.createStudentConversation,
		s.multipleTeachersJoinConversation, teacherIDs,
	)
}

func (s *suite) aChatBetweenAStudentAndTeachersWithUserGroups(ctx context.Context, numberOfTeacher int) (context.Context, error) {
	teacherCount := []string{}
	for i := 0; i < numberOfTeacher; i++ {
		teacherCount = append(teacherCount, idutil.ULIDNow())
	}
	return godogutil.MultiErrChain(ctx,
		s.createStudentConversation,
		s.multipleTeachersWithUserGroupsJoinConversation, teacherCount,
	)
}

func (s *suite) aChatBetweenAStudentWithLocationsAndTeachersWithUserGroups(ctx context.Context, numberOfTeacher int) (context.Context, error) {
	teacherCount := []string{}
	for i := 0; i < numberOfTeacher; i++ {
		teacherCount = append(teacherCount, idutil.ULIDNow())
	}
	return godogutil.MultiErrChain(ctx,
		s.createStudentConversationWithLocations,
		s.multipleTeachersWithUserGroupsJoinConversation, teacherCount,
	)
}

func (s *suite) studentIsPresent(ctx context.Context) (context.Context, error) {
	userIDs := []string{s.studentID}
	token, err := s.genStudentToken(s.studentID)
	if err != nil {
		return ctx, err
	}
	return s.makeUsersSubscribeV2Ctx(ctx, userIDs, []string{token})
}
func (s *suite) makeUsersSubscribeV2Ctx(ctx context.Context, userIDs []string, tokens []string) (context.Context, error) {
	for idx, id := range userIDs {
		token := tokens[idx]
		ctx2, cancel := context.WithCancel(context.Background())

		subClient, err := pb.NewChatServiceClient(s.Conn).SubscribeV2(contextWithToken(ctx2, token), &pb.SubscribeV2Request{})
		if err != nil {
			cancel()
			return ctx, err
		}
		// if old stream already exist, cancel it and overwrite
		if oldStream, ok := s.SubV2Clients[id]; ok {
			oldStream.cancel()
		}
		s.SubV2Clients[id] = cancellableStream{subClient, cancel}

		ctx, err = godogutil.MultiErrChain(ctx,
			s.ensureUserHasDeviceToken, id,
			s.ensureUserDevicePingCtx, id, token, 1,
		)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) studentAndTeachersArePresent(ctx context.Context) (context.Context, error) {
	// check subscribe v2
	userIDs := []string{s.studentID}
	token, err := s.generateExchangeToken(s.studentID, cpb.UserGroup_USER_GROUP_STUDENT.String(), applicantID, s.getSchool(), s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	tokens := []string{token}
	userIDs = append(userIDs, s.teachersInConversation...)
	teachertokens, err := s.generateExchangeTokens(s.teachersInConversation, cpb.UserGroup_USER_GROUP_TEACHER.String(), applicantID, s.getSchool(), s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	tokens = append(tokens, teachertokens...)

	return s.makeUsersSubscribeV2Ctx(ctx, userIDs, tokens)
}
func (s *suite) multipleTeachersJoinConversation(ctx context.Context, teacherIDs []string) (context.Context, error) {
	for _, teacherID := range teacherIDs {
		s.teacherID = teacherID
		ctx, err := godogutil.MultiErrChain(ctx,
			s.aValidToken, "current teacher",
			s.teacherJoinConversations,
			s.returnsStatusCode, "OK",
			s.teacherMustBeMemberOfConversations,
		)
		if err != nil {
			return ctx, err
		}
		s.teachersInConversation = append(s.teachersInConversation, s.teacherID)
		s.teacherTokens[s.teacherID] = s.TeacherToken
		s.teacherID = ""
	}
	return ctx, nil
}

func (s *suite) multipleTeachersWithUserGroupsJoinConversation(ctx context.Context, teacherIDs []string) (context.Context, error) {
	for _, teacherID := range teacherIDs {
		s.teacherID = teacherID
		ctx, err := godogutil.MultiErrChain(ctx,
			s.aValidToken, "teacher with user groups",
			s.teacherJoinConversations,
			s.returnsStatusCode, "OK",
			s.teacherMustBeMemberOfConversations,
		)
		if err != nil {
			return ctx, err
		}
		s.teachersInConversation = append(s.teachersInConversation, s.teacherID)
		s.teacherTokens[s.teacherID] = s.TeacherToken
		s.teacherID = ""
	}
	return ctx, nil
}

func (s *suite) aTeacherSendsItemWithContent(ctx context.Context, msgType, msgContent string) (context.Context, error) {
	userID := s.teachersInConversation[0]
	s.teacherWhoSentMessage = userID
	return s.userSendsItemWithContent(ctx, msgType, msgContent, userID, cpb.UserGroup_USER_GROUP_TEACHER)
}
func (s *suite) getSchool() int64 {
	if s.commonState.schoolID != "" {
		intSchool, _ := strconv.ParseInt(s.commonState.schoolID, 10, 64)
		return intSchool
	}
	if s.schoolID != "" {
		intSchool, _ := strconv.ParseInt(s.schoolID, 10, 64)

		return intSchool
	}

	return constants.ManabieSchool
}
func (s *suite) userSendsItemWithContent(ctx context.Context, msgType string, msgContent string, userID string, userGroup cpb.UserGroup) (context.Context, error) {
	school := s.getSchool()
	sendMsgReq := &pb.SendMessageRequest{
		ConversationId: s.conversationID,
		LocalMessageId: idutil.ULIDNow(),
	}
	switch msgType {
	case "text":
		sendMsgReq.Type = pb.MESSAGE_TYPE_TEXT
		sendMsgReq.Message = msgContent
	case "image":
		sendMsgReq.Type = pb.MESSAGE_TYPE_IMAGE
		sendMsgReq.UrlMedia = msgContent
	case "file":
		sendMsgReq.Type = pb.MESSAGE_TYPE_FILE
		sendMsgReq.UrlMedia = msgContent
	}
	s.sentMessages = append(s.sentMessages, sendMsgReq)
	// s.sentMsgType = sendMsgReq.Type

	token, err := s.generateExchangeToken(userID, userGroup.String(), applicantID, school, s.ShamirConn)
	if err != nil {
		return ctx, err
	}

	resp, err := pb.NewChatServiceClient(s.Conn).SendMessage(contextWithToken(context.Background(), token), sendMsgReq)

	if resp != nil {
		s.messageID = resp.MessageId
		s.studentMessageID = resp.MessageId
	}

	return ctx, err
}
func (s *suite) studentSendsItemWithContent(ctx context.Context, msgType string, msgContent string) (context.Context, error) {
	studentID := s.studentID
	userGroup := cpb.UserGroup_USER_GROUP_STUDENT
	return s.userSendsItemWithContent(ctx, msgType, msgContent, studentID, userGroup)
}
func (s *suite) studentReceivesSentMessage(ctx context.Context) (context.Context, error) {
	return s.usersReceiveSentMessage(ctx, []string{s.studentID})
}
func (s *suite) parentsReceiveNotification(ctx context.Context) (context.Context, error) {
	for _, id := range s.parentIDs {
		ctx, err := s.userWithIDAndRoleReceivesNotificationFromConversation(ctx, id, cpb.UserGroup_USER_GROUP_PARENT.String(), false, tpb.ConversationType_CONVERSATION_PARENT.String())
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) userWithIDAndRoleReceivesNotificationFromConversation(ctx context.Context, userID string, role string, silent bool, convType string) (context.Context, error) {
	var deviceToken string
	row := s.DB.QueryRow(ctx, "SELECT token FROM user_device_tokens WHERE user_id = $1", userID)
	if err := row.Scan(&deviceToken); err != nil {
		return ctx, fmt.Errorf("error finding user device token: %w", err)
	}

	token, err := s.generateExchangeToken(userID, role, applicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	return ctx, try.Do(func(attempt int) (bool, error) {
		sentMsg := s.sentMessages[len(s.sentMessages)-1]
		var expectContent string
		switch sentMsg.Type {
		case pb.MESSAGE_TYPE_TEXT:
			expectContent = sentMsg.Message
		case pb.MESSAGE_TYPE_FILE:
			fallthrough
		case pb.MESSAGE_TYPE_IMAGE:
			expectContent = tomconst.MessagingFileNotificationContent
		}

		var expectedTitle string
		findConversationQuery := `SELECT c.name
FROM conversations c
LEFT JOIN conversation_members cm ON c.conversation_id=cm.conversation_id
WHERE cm.user_id=$1
  AND cm.role=$2
  AND c.conversation_type=$3
  AND cm.status='CONVERSATION_STATUS_ACTIVE'`
		row = s.DB.QueryRow(ctx, findConversationQuery, userID, role, convType)
		if err := row.Scan(&expectedTitle); err != nil {
			return false, fmt.Errorf("error finding conversation for user %s role %s: %w", userID, role, err)
		}

		resp, err := pb.NewChatServiceClient(s.Conn).RetrievePushedNotificationMessages(
			contextWithToken(ctx, token),
			&pb.RetrievePushedNotificationMessageRequest{
				DeviceToken: deviceToken,
			})
		if err != nil {
			return false, err
		}
		if len(resp.Messages) == 0 {
			return attempt < 10, fmt.Errorf("wrong node")
		}
		gotNoti := resp.Messages[len(resp.Messages)-1]
		gotMsg := gotNoti.Data.Fields
		title, ok := gotMsg[tomconst.FcmKeyConversationName]
		if !ok || title.GetStringValue() != expectedTitle {
			s.ZapLogger.Error(fmt.Sprintf("err: %v", gotNoti))
			return false, fmt.Errorf("want notification title to be: %s, got %s", expectedTitle, title)
		}

		body, ok := gotMsg[tomconst.FcmKeyMessageContent]
		if !ok || body.GetStringValue() != expectContent {
			s.ZapLogger.Error(fmt.Sprintf("err: %v", gotNoti))
			return false, fmt.Errorf("want notification body to be: %s, got %s", expectContent, body)
		}
		if !silent {
			if gotNoti.Body == "" || gotNoti.Title == "" {
				return false, fmt.Errorf("non silent notification must have non-empty title and body")
			}
		} else {
			if gotNoti.Body != "" || gotNoti.Title != "" {
				return false, fmt.Errorf("silent noti must have empty body and title")
			}
		}
		return false, nil
	})
}
func (s *suite) studentReceivesNotification(ctx context.Context) (context.Context, error) {
	return s.userWithIDAndRoleReceivesNotificationFromConversation(ctx, s.studentID, cpb.UserGroup_USER_GROUP_STUDENT.String(), false, tpb.ConversationType_CONVERSATION_STUDENT.String())
}
func (s *suite) usersReceiveMatchedNewMessage(ctx context.Context, userIDs []string, matcher func(newMsg *pb.MessageResponse) error) (context.Context, error) {
	if s.chatName == "" {
		panic("forgot to set chat name in previous step")
	}
	g := new(errgroup.Group)
	for _, userID := range userIDs {
		stream, exist := s.SubV2Clients[userID]
		if !exist {
			return ctx, fmt.Errorf("stream for user %s does not exist", userID)
		}
		g.Go(func() error {
			defer func() {
				err := stream.CloseSend()
				if err != nil {
					s.ZapLogger.Error(fmt.Sprintf("error closing stream %s", err))
				}
			}()
			var newMsg *pb.MessageResponse

			for try := 0; try < 10; try++ {
				resp, err := stream.Recv()
				if err == io.EOF {
					return fmt.Errorf("received eof before try runs out")
				}
				newMsg = resp.GetEvent().GetEventNewMessage()
				if newMsg != nil {
					break
				}
			}
			if newMsg == nil {
				return fmt.Errorf("received nil after 10 tries")
			}
			if err := matcher(newMsg); err != nil {
				return err
			}
			return nil
		})
	}
	return ctx, g.Wait()
}
func (s *suite) usersReceiveSentMessage(ctx context.Context, userIDs []string) (context.Context, error) {
	sentMsg := s.sentMessages[len(s.sentMessages)-1]
	return s.usersReceiveMatchedNewMessage(ctx, userIDs, func(newMsg *pb.MessageResponse) error {
		if sentMsg.Type != newMsg.Type {
			return fmt.Errorf("expecting message type %s, %s given with content %s", sentMsg.String(), newMsg.Type.String(), newMsg.Content)
		}
		if newMsg.ConversationName != s.chatName {
			return fmt.Errorf("want message to have chat name %s, have %s", s.chatName, newMsg.ConversationName)
		}
		switch sentMsg.Type {
		case pb.MESSAGE_TYPE_TEXT:
			if newMsg.Content != sentMsg.Message {
				return fmt.Errorf("expecting text message with content %s, %v given", sentMsg.Message, newMsg)
			}

		default:
			if newMsg.UrlMedia != sentMsg.UrlMedia {
				return fmt.Errorf("expecting message with url %s, %v given", sentMsg.UrlMedia, newMsg)
			}
		}
		return nil
	})
}
func (s *suite) teachersReceiveSentMessage(ctx context.Context) (context.Context, error) {
	return s.usersReceiveSentMessage(ctx, s.teachersInConversation)
}
func (s *suite) otherTeachersReceiveSentMessage(ctx context.Context) (context.Context, error) {
	otherTeachers := []string{}
	for _, teacherID := range s.teachersInConversation {
		if teacherID != s.teacherWhoSentMessage {
			otherTeachers = append(otherTeachers, teacherID)
		}
	}
	return s.usersReceiveSentMessage(ctx, otherTeachers)
}
func (s *suite) studentSendMessage(ctx context.Context, msg string) (context.Context, error) {
	studentID := s.studentID
	sendMsgReq := &pb.SendMessageRequest{
		ConversationId: s.conversationID,
		Message:        msg,
		UrlMedia:       "",
		Type:           pb.MESSAGE_TYPE_TEXT,
		LocalMessageId: idutil.ULIDNow(),
	}

	token, err := s.generateExchangeToken(studentID, cpb.UserGroup_USER_GROUP_STUDENT.String(), applicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return ctx, err
	}

	_, err = pb.NewChatServiceClient(s.Conn).SendMessage(metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token), sendMsgReq)
	return ctx, err
}
func (s *suite) studentIsNotPresent(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
func (s *suite) teachersAreNotPresent(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
func (s *suite) teachersArePresent(ctx context.Context) (context.Context, error) {
	tok, err := s.genTeacherTokens(s.teachersInConversation)
	if err != nil {
		return ctx, err
	}
	return s.makeUsersSubscribeV2Ctx(ctx, s.teachersInConversation, tok)
}
