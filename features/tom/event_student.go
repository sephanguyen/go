package tom

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"
)

func (s *suite) aEvtUserWithMessageAndLanguageAndSchoolID(ctx context.Context, event string, language string, schoolID string, locs []string) (context.Context, error) {
	if s.studentID == "" {
		s.studentID = idutil.ULIDNow()
	}

	if s.parentID == "" {
		s.parentID = idutil.ULIDNow()
	}
	s.schoolID = schoolID

	switch event {
	case "CreateStudent":
		if s.chatName == "" {
			s.chatName = pseudoNameForStudentChat(s.studentID, language)
		}
		s.Request = &upb.EvtUser{
			Message: &upb.EvtUser_CreateStudent_{
				CreateStudent: &upb.EvtUser_CreateStudent{
					StudentId:   s.studentID,
					StudentName: s.chatName, // TODO check if this change make other BDD fails
					SchoolId:    s.schoolID,
					LocationIds: locs,
				},
			},
		}
	case "CreateParent":
		if s.chatName == "" {
			s.chatName = pseudoNameForParentChat(s.studentID, language)
		}
		s.Request = &upb.EvtUser{
			Message: &upb.EvtUser_CreateParent_{
				CreateParent: &upb.EvtUser_CreateParent{
					StudentId:   s.studentID,
					StudentName: s.chatName,
					SchoolId:    s.schoolID,
					ParentId:    s.parentID,
				},
			},
		}
	case "ParentAssignedToStudent":
		if s.chatName == "" {
			s.chatName = pseudoNameForParentChat(s.studentID, language)
		}
		s.Request = &upb.EvtUser{
			Message: &upb.EvtUser_ParentAssignedToStudent_{
				ParentAssignedToStudent: &upb.EvtUser_ParentAssignedToStudent{
					ParentId:  s.parentID,
					StudentId: s.studentID,
				},
			},
		}
	}
	return ctx, nil
}
func (s *suite) aEvtUserWithMessage(ctx context.Context, event string) (context.Context, error) {
	// for backward compatibility
	return s.aEvtUserWithMessageAndLanguageAndSchoolID(ctx, event, "english", strconv.Itoa(constants.ManabieSchool), []string{constants.ManabieOrgLocation})
}

func (s *suite) findStudentsConvIDs(ctx context.Context, ids []string, school string, locs []string) (map[string]string, error) {
	query := `SELECT cs.conversation_id,cs.student_id FROM conversation_students cs LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
	WHERE cs.student_id = ANY($1) AND owner = $2 AND c.status= 'CONVERSATION_STATUS_NONE' AND cs.conversation_type = 'CONVERSATION_STUDENT'`
	mapStuIDConvID := map[string]string{}
	err := doRetry(func() (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		rows, err := s.DB.Query(ctx2, query, database.TextArray(ids), school)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, err
			}
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var (
				convID, stuID string
			)
			err := rows.Scan(&convID, &stuID)
			if err != nil {
				return false, err
			}
			mapStuIDConvID[stuID] = convID
		}
		if len(mapStuIDConvID) != len(ids) {
			return true, fmt.Errorf("not enough student conversations created, want %d has %d", len(ids), len(mapStuIDConvID))
		}
		for _, convID := range mapStuIDConvID {
			checkLocs := "select count(*) from conversation_locations cl where cl.conversation_id=$1 and location_id=ANY($2)"
			var count pgtype.Int8
			ctx3, cancel2 := context.WithTimeout(ctx, 2*time.Second)
			defer cancel2()
			if err := s.DB.QueryRow(ctx3, checkLocs, convID, database.TextArray(locs)).Scan(&count); err != nil {
				return false, err
			}
			if int(count.Int) != len(locs) {
				return false, fmt.Errorf("conversation %s has %d locations among %v expected locations", convID, count.Int, locs)
			}
		}
		return false, nil
	})

	return mapStuIDConvID, err
}
func (s *suite) studentMustBeInConversation(ctx context.Context) (context.Context, error) {
	req := s.Request.(*upb.EvtUser)
	studentEvent := req.GetCreateStudent()
	mapStuConvIDs, err := s.findStudentsConvIDs(ctx, []string{studentEvent.StudentId}, studentEvent.SchoolId, studentEvent.GetLocationIds())
	if err != nil {
		return ctx, err
	}
	convID := mapStuConvIDs[studentEvent.GetStudentId()]
	s.ConversationIDs = []string{convID}
	s.conversationID = convID
	return ctx, nil
}

func (s *suite) yasuoSendEventEvtUser(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	subj := constants.SubjectUserCreated
	var (
		data    []byte
		err     error
		msgType string
	)
	switch msg := s.Request.(type) {
	case *bpb.EvtUser:
		msgType = fmt.Sprintf("%T", msg.GetMessage())
		data, err = proto.Marshal(msg)
	case *upb.EvtUser:
		msgType = fmt.Sprintf("%T", msg.GetMessage())
		data, err = proto.Marshal(msg)
		switch msg.GetMessage().(type) {
		case *upb.EvtUser_ParentAssignedToStudent_:
			subj = constants.SubjectUserUpdated
		case *upb.EvtUser_ParentRemovedFromStudent_:
			subj = constants.SubjectUserUpdated
		case *upb.EvtUser_UpdateStudent_:
			subj = constants.SubjectUserUpdated
		}
	default:
		return ctx, fmt.Errorf("invalid req type %T", s.Request)
	}
	if err != nil {
		return ctx, err
	}
	s.RequestAt = time.Now()

	_, err = s.JSM.TracedPublish(ctx2, fmt.Sprintf("yasuoSendEventEvtUser(%s)", msgType), subj, data)
	return ctx, err
}

func (s *suite) createStudentConversation(ctx context.Context) (context.Context, error) {
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32(constants.ManabieSchool))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	stu, err := s.CommonSuite.CreateStudent(ctx, []string{constants.ManabieOrgLocation}, nil)
	if err != nil {
		return ctx, err
	}
	s.studentID = stu.UserProfile.UserId

	token, err := s.genStudentToken(s.studentID)
	if err != nil {
		return ctx, err
	}

	s.studentToken = token
	s.chatName = stu.UserProfile.Name
	schoolText := strconv.Itoa(constants.ManabieSchool)
	mapStuConvIDs, err := s.findStudentsConvIDs(ctx, []string{s.studentID}, schoolText, []string{constants.ManabieOrgLocation})
	if err != nil {
		return ctx, err
	}
	convID := mapStuConvIDs[s.studentID]
	s.ConversationIDs = []string{convID}
	s.conversationID = convID
	return ctx, nil
}

func (s *suite) createStudentConversationWithLocations(ctx context.Context) (context.Context, error) {
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32(constants.ManabieSchool))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	stu, err := s.CommonSuite.CreateStudent(ctx, s.CommonSuite.LocationIDs, nil)
	if err != nil {
		return ctx, err
	}
	s.studentID = stu.UserProfile.UserId

	token, err := s.genStudentToken(s.studentID)
	if err != nil {
		return ctx, err
	}

	s.studentToken = token
	s.chatName = stu.UserProfile.Name
	schoolText := strconv.Itoa(constants.ManabieSchool)
	mapStuConvIDs, err := s.findStudentsConvIDs(ctx, []string{s.studentID}, schoolText, s.CommonSuite.LocationIDs)
	if err != nil {
		return ctx, err
	}
	convID := mapStuConvIDs[s.studentID]
	s.ConversationIDs = []string{convID}
	s.conversationID = convID
	return ctx, nil
}

func (s *suite) returnConversationListMustHave(ctx context.Context, criteriaList string) (context.Context, error) {
	item := strings.Split(criteriaList, ",")

	rsp := s.Response.(*pb.ConversationListResponse)
	for _, conversation := range rsp.Conversations {
		for _, criteria := range item {
			parts := strings.Split(criteria, " ")
			field, value := parts[0], parts[1]
			switch field {
			case "type":
				if conversation.ConversationType.String() != value {
					return ctx, fmt.Errorf("expect conversation %s has type %s but got %s", conversation.ConversationId, value, conversation.ConversationType.String())
				}
			case "latest_message":
				if conversation.GetLastMessage().Content != value {
					return ctx, fmt.Errorf("expect latest msg to be %s but have %s", value, conversation.LastMessage.Content)
				}
			default:
				panic(fmt.Sprintf("invalid field %s", field))
			}
		}
	}
	return ctx, nil
}
