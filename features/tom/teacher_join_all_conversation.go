package tom

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// nolint:gosec
func (s *suite) randomNewConversationsCreated(ctx context.Context) (context.Context, error) {
	numberOfConversation := rand.Intn(5) + 5
	if ctx, err := s.createConversation(ctx, numberOfConversation); err != nil {
		return ctx, fmt.Errorf("unable to create conversation: %w", err)
	}
	return ctx, nil
}

func (s *suite) createConversationEachLocation(ctx context.Context, conversationEach int, locationIDs []string) error {
	schoolID := s.commonState.schoolID
	studentIDs := make([]string, 0, len(locationIDs)*conversationEach)
	for _, locID := range locationIDs {
		for i := 0; i < conversationEach; i++ {
			studentID := idutil.ULIDNow()
			studentIDs = append(studentIDs, studentID)
			evt := &upb.EvtUser{
				Message: &upb.EvtUser_CreateStudent_{
					CreateStudent: &upb.EvtUser_CreateStudent{
						StudentId:   studentID,
						StudentName: "name" + studentID,
						SchoolId:    schoolID,
						LocationIds: []string{locID},
					},
				},
			}
			data, err := proto.Marshal(evt)
			if err != nil {
				return err
			}

			_, err = s.JSM.TracedPublishAsync(ctx, "nats.TracedPublishAsync", constants.SubjectUserCreated, data)
			if err != nil {
				return fmt.Errorf("s.JSM.TracedPublishAsync: %w", err)
			}
		}
	}

	return doRetry(func() (retry bool, err error) {
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		conversationMembers, err := getConversationMemberByMemberIDs(ctx2, s.DB, database.TextArray(studentIDs))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, fmt.Errorf("found no member for students")
			}
			return false, err
		}
		if conversationMembers == nil {
			return true, fmt.Errorf("unable to get conversation member")
		}
		if len(conversationMembers) != len(studentIDs) {
			return true, fmt.Errorf("not enough member in db")
		}
		return true, err
	})
}

func (s *suite) createConversation(ctx context.Context, numberOfConversation int) (context.Context, error) {
	var locID string
	var locTypeID string

	if s.schoolID == "" {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		intschool, location, locationType, err := s.aGenerateSchool(ctx2)
		if err != nil {
			return ctx, err
		}
		s.schoolID = strconv.Itoa(int(intschool))
		locID = location
		locTypeID = locationType
	} else {
		locID = getSchoolDefaultLocation(s.schoolID)
		locTypeID = getSchoolDefaultLocationType(s.schoolID)
	}

	ctx = contextWithResourcePath(ctx, s.schoolID)
	s.CommonSuite.DefaultLocationID = locID
	s.CommonSuite.DefaultLocationTypeID = locTypeID
	s.StudentIDAndParentIDMap = make(map[string]string, 0)
	evts := make([]*upb.EvtUser, 0, numberOfConversation)
	for i := 0; i < numberOfConversation; i++ {
		studentID := idutil.ULIDNow()
		// this step only create students
		s.StudentIDAndParentIDMap[studentID] = ""
		s.StudentIDs = append(s.StudentIDs, studentID)
		evts = append(evts,
			&upb.EvtUser{
				Message: &upb.EvtUser_CreateStudent_{
					CreateStudent: &upb.EvtUser_CreateStudent{
						StudentId:   studentID,
						StudentName: "name" + studentID,
						SchoolId:    s.schoolID,
						LocationIds: []string{locID},
					},
				},
			})
	}
	for _, evt := range evts {
		data, err := proto.Marshal(evt)
		if err != nil {
			return ctx, err
		}

		_, err = s.JSM.TracedPublishAsync(ctx, "nats.TracedPublishAsync", constants.SubjectUserCreated, data)
		if err != nil {
			return ctx, fmt.Errorf("s.JSM.TracedPublishAsync: %w", err)
		}
	}
	fmt.Printf("DUMP CREATED STUDENT %+v\n", s.StudentIDs)
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(2 * time.Second)
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		conversationMembers, err := getConversationMemberByMemberIDs(ctx2, s.DB, database.TextArray(s.StudentIDs))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return attempt < 5, fmt.Errorf("found no member for students")
			}
			return false, err
		}
		if conversationMembers == nil {
			return attempt < 10, fmt.Errorf("unable to get conversation member")
		}
		if len(conversationMembers) != len(s.StudentIDs) {
			return attempt < 10, fmt.Errorf("not enough member in db")
		}
		s.ConversationIDs = retrieveConversationIDsFromConversationMembers(conversationMembers)
		return true, err
	}); err != nil {
		return ctx, err
	}
	return ctx, nil
}
func retrieveConversationIDsFromConversationMembers(cms []*entities.ConversationMembers) []string {
	cIDs := make([]string, 0)
	for _, e := range cms {
		cIDs = append(cIDs, e.ConversationID.String)
	}
	return cIDs
}

func (s *suite) teacherJoinAllConversations(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = s.userJoinAllConversations(ctx, "teacher")
	if s.ResponseErr != nil {
		err = s.ResponseErr
	}
	return ctx, err
}

func (s *suite) userJoinAllConversations(ctx context.Context, role string) (context.Context, error) {
	// usually max time out at 15 sec, may change tho
	ctx2, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var token string
	switch role {
	case "teacher":
		token = s.TeacherToken
	case "school admin":
		token = s.schoolAdminToken
	case "student":
		token = s.studentToken
	case "parent":
		token = s.parentToken
	default:
		return ctx, fmt.Errorf("not handle %s role to join all conversations yet", role)
	}

	req := &tpb.JoinAllConversationRequest{}
	s.Request = req
	s.ResponseErr = try.Do(func(attempt int) (bool, error) {
		s.Response, s.ResponseErr = tpb.NewChatModifierServiceClient(s.Conn).
			JoinAllConversations(contextWithToken(ctx2, token), req)

		if s.ResponseErr != nil {
			time.Sleep(1 * time.Second)
			return attempt < 5, s.ResponseErr
		}
		return false, nil
	})

	if s.ResponseErr != nil {
		if ctx.Err() != nil {
			var count int
			err := s.DB.QueryRow(ctx2, "select count(*) from conversations where owner=$1", s.schoolID).Scan(&count)
			if err == nil {
				s.ZapLogger.Sugar().Infof("deadline when calling join all: there are %d conversations with school %s", count, s.schoolID)
			}
		}
		return ctx, nil
	}
	return ctx, nil
}

const (
	schoolBelongToTeacher = `
		SELECT count(*) FROM conversations where owner = any ($1::text[]);
	`
	getConversationByMemberID = `
	SELECT count(*) FROM conversation_members cm 
	WHERE cm.user_id = $1 
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE';
	`
	findTeacherConversationID = `
	SELECT conversation_id FROM conversation_members cm 
	WHERE cm.user_id = $1 
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE' 
	AND cm.role ='USER_GROUP_TEACHER';
	`
)

func (s *suite) teacherMustBeMemberOfAllConversationsWithSpecifySchools(ctx context.Context) (context.Context, error) {
	return s.userMustBeMemberOfAllConversationsWithSpecifySchools(ctx, "teacher")
}

func (s *suite) userMustBeMemberOfAllConversationsWithSpecifySchools(ctx context.Context, role string) (context.Context, error) {
	var userID string
	switch role {
	case "teacher":
		userID = s.teacherID
	case "school admin":
		userID = s.schoolAdminID
	case "student":
		userID = s.studentID
	default:
		return ctx, fmt.Errorf("not handle %s role to check be member of all conversations", role)
	}
	var schoolCount int64
	var conversationCount int64
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(3 * time.Second)
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := s.DB.QueryRow(ctx2, schoolBelongToTeacher, &s.SchoolIds).Scan(&schoolCount); err != nil {
			return false, err
		}
		if err := s.DB.QueryRow(ctx2, getConversationByMemberID, &userID).Scan(&conversationCount); err != nil {
			return false, err
		}
		if schoolCount == conversationCount {
			return false, nil
		}

		return attempt < 10, fmt.Errorf("user is not a member of all conversation belong to schoolIds: expected join %d conversation but just join %d", schoolCount, conversationCount)
	}); err != nil {
		return ctx, err
	}

	if schoolCount != conversationCount {
		return ctx, fmt.Errorf("user is not a member of all conversation belong to schoolIds")
	}
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var conversationIDs []string
	rows, err := s.DB.Query(ctx2, findTeacherConversationID, &userID)
	if err != nil {
		return ctx, fmt.Errorf("error when find user conversation id %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var conversationID string
		err := rows.Scan(&conversationID)
		if err != nil {
			return ctx, err
		}
		conversationIDs = append(conversationIDs, conversationID)
	}
	s.ConversationIDs = conversationIDs
	s.JoinedConversationIDs = s.ConversationIDs
	return ctx, nil
}

func (s *suite) theTeacherJoinSomeConversations(ctx context.Context) (context.Context, error) {
	// nolint:gosec
	numOfJoinConversation := rand.Intn(len(s.StudentIDs)-3) + 3

	err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)

		_, err := tpb.NewChatModifierServiceClient(conn).JoinConversations(metadata.AppendToOutgoingContext(context.Background(), "token", s.TeacherToken), &tpb.JoinConversationsRequest{
			ConversationIds: s.ConversationIDs[:numOfJoinConversation]})

		if err != nil {
			return attempt < 5, fmt.Errorf("teacher unable to join some conversations: %w", err)
		}

		mapConversation, err := getConversationMap(ctx, s.DB, s.ConversationIDs[:numOfJoinConversation])
		if err != nil {
			return true, err
		}
		s.OldConversations = mapConversation
		joinedCounter := 0
		for _, c := range mapConversation {
			if c.MessageType == tpb.MessageType_MESSAGE_TYPE_SYSTEM.String() && c.Content == tpb.CodesMessageType_CODES_MESSAGE_TYPE_JOINED_CONVERSATION.String() {
				s.JoinedConversationIDs = append(s.JoinedConversationIDs, c.ConversationID)
				joinedCounter++
			}
		}
		if joinedCounter != numOfJoinConversation {
			return attempt < 10, fmt.Errorf("unexpected number of conversations that teacher joined, want: %d, actual: %d", numOfJoinConversation, joinedCounter)
		}
		return false, nil
	})

	if err != nil {
		return ctx, fmt.Errorf("teacher unable to join some conversations: %w", err)
	}

	return ctx, nil
}
func (s *suite) systemMustOnlySendConversationMessageWhichUnjoinedBefore(ctx context.Context) (context.Context, error) {
	closure := func() error {
		ctx2, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		mapMessage, err := getConversationMap(ctx2, s.DB, s.JoinedConversationIDs)
		if err != nil {
			return fmt.Errorf("getConversationMap: %w", err)
		}

		if len(s.JoinedConversationIDs) != len(mapMessage) {
			return fmt.Errorf("wrong length of conversations in school")
		}
		for _, cID := range s.JoinedConversationIDs {
			if mapMessage[cID].MessageType != tpb.MessageType_MESSAGE_TYPE_SYSTEM.String() {
				return fmt.Errorf("wrong type of conversation last message")
			}
			if mapMessage[cID].Content != tpb.CodesMessageType_CODES_MESSAGE_TYPE_JOINED_CONVERSATION.String() {
				return fmt.Errorf("wrong content message of conversation last message")
			}
		}
		return nil
	}
	return ctx, try.Do(func(attempt int) (bool, error) {
		err := closure()
		if err != nil {
			time.Sleep(1 * time.Second)
			return attempt < 5, err
		}
		return false, nil
	})
}

func (s *suite) aSignedAsATeacher(ctx context.Context) (context.Context, error) {
	schoolID := s.getSchool()
	fmt.Printf("SCHOOL %+v\n", schoolID)
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32(schoolID))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	profile, tok, err := s.CommonSuite.CreateTeacher(ctx)
	if err != nil {
		return ctx, err
	}
	s.teacherID = profile.StaffId
	s.TeacherToken = tok
	s.SchoolIds = []string{strconv.Itoa(int(schoolID))}
	return ctx, err
}

func (s *suite) aSignedAsATeacherWithUserGroups(ctx context.Context, userGroupIDs []string) (context.Context, error) {
	schoolID := s.getSchool()
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32(schoolID))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	profile, tok, err := s.CommonSuite.CreateTeacherWithUserGroups(ctx, userGroupIDs)
	if err != nil {
		return ctx, err
	}
	s.teacherID = profile.StaffId
	s.TeacherToken = tok
	s.SchoolIds = []string{strconv.Itoa(int(schoolID))}
	return ctx, err
}

// get all conversation associated with latest message, but wait until latest message is joined conversation type
func getConversationMap(ctx context.Context, db database.QueryExecer, ids []string) (map[string]Message, error) {
	trial := func() (map[string]Message, error) {
		mapConversation := make(map[string]Message)

		messageRepo := &repositories.MessageRepo{}
		messages, err := messageRepo.GetLastMessageByConversationIDs(ctx, db, database.TextArray(ids), uint(len(ids)), database.Timestamptz(time.Now()), true)
		if err != nil {
			return nil, fmt.Errorf("messageRepo.GetLastMessageByConversationIDs: %w", err)
		}
		if len(messages) != len(ids) {
			return nil, fmt.Errorf("expect all conversations provided are returned")
		}
		for _, m := range messages {
			mapConversation[m.ConversationID.String] = Message{
				ConversationID: m.ConversationID.String,
				MessageID:      m.ID.String,
				MessageType:    m.Type.String,
				Content:        m.Message.String,
			}
		}
		return mapConversation, nil
	}
	var ret map[string]Message
	err := try.Do(func(attempt int) (bool, error) {
		msgs, err := trial()
		if err != nil {
			time.Sleep(1 * time.Second)
			return attempt < 5, err
		}
		ret = msgs
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func getConversationMemberByMemberIDs(ctx context.Context, db database.Ext, memberIDs pgtype.TextArray) ([]*entities.ConversationMembers, error) {
	results := make([]*entities.ConversationMembers, 0)
	stmt := `SELECT %s FROM conversation_members WHERE user_id=ANY($1::_TEXT)`
	c := &entities.ConversationMembers{}
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf(stmt, strings.Join(fields, ","))

	rows, err := db.Query(ctx, selectStmt, &memberIDs)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve conversation by user_ids: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		c := &entities.ConversationMembers{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan %w", err)
		}
		results = append(results, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

type Message struct {
	ConversationID string
	MessageID      string
	MessageType    string
	Content        string
}

func initJoinAllStepSuite(previousMap map[string]interface{}, ctx *godog.ScenarioContext, s *suite) {
	s.commonState = commonState{
		coursePool:      map[string]string{},
		locationPool:    map[string]string{},
		studentChatPool: map[string]chatInfo{},
		parentChatPool:  map[string]chatInfo{},
	}
	mergedMap := map[string]interface{}{
		`^"([^"]*)" new conversations created for each locations "([^"]*)"$`:              s.newConversationsCreatedForEachLocations,
		`^teacher "([^"]*)" be member of "([^"]*)" conversations in locations "([^"]*)"$`: s.teacherBeMemberOfConversationsInLocations,
		`^teacher joins all conversations in locations "([^"]*)"$`:                        s.teacherJoinsAllConversationsInLocations,
	}

	applyMergedSteps(ctx, previousMap, mergedMap)
}

func (s *suite) newConversationsCreatedForEachLocations(ctx context.Context, newConversationEach int, locs string) (context.Context, error) {
	locIDs := s.getLocations(locs)
	err := s.createConversationEachLocation(ctx, newConversationEach, locIDs)
	return ctx, err
}

// condition: must,mustnot
func (s *suite) teacherBeMemberOfConversationsInLocations(ctx context.Context, condition string, totalMembership int, locs string) (context.Context, error) {
	locIDs := s.getLocations(locs)
	query := `
	select count(*) from conversation_members cm left join conversations c using(conversation_id) left join conversation_locations cl using(conversation_id)
	where cm.user_id=$1 and cl.location_id=ANY($2)
	`

	return ctx, doRetry(func() (bool, error) {
		var count pgtype.Int8
		err := s.DB.QueryRow(ctx, query, s.teacherID, database.TextArray(locIDs)).Scan(&count)
		if err != nil {
			return false, err
		}
		switch condition {
		case "must not":
			if count.Int != 0 {
				return false, fmt.Errorf("teacher %s is member of %d conversation in locations %v", s.teacherID, count, locIDs)
			}
		case "must":
			if count.Int != int64(totalMembership) {
				return true, fmt.Errorf("want %d membership, has %d", totalMembership, count)
			}
		default:
			panic(fmt.Sprintf("uknown condition %s", condition))
		}
		return false, nil
	})
}

func (s *suite) teacherJoinsAllConversationsInLocations(ctx context.Context, locs string) (context.Context, error) {
	// usually max time out at 15 sec, may change tho
	ctx2, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	req := &tpb.JoinAllConversationRequest{
		LocationIds: s.getLocations(locs),
	}

	s.Request = req

	_, err := tpb.NewChatModifierServiceClient(s.Conn).
		JoinAllConversationsWithLocations(contextWithToken(ctx2, s.TeacherToken), req)
	return ctx, err
}
