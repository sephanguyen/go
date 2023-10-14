package tom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) aStudentAndConversationCreated(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.aSignedAsATeacher,
		s.createConversation, 1, // create student and it's conversation
	)
}

func (s *suite) aParentAndConversationCreated(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.createConversationParent, // create parent and it's conversation based on students
	)
}

func (s *suite) createConversationParent(ctx context.Context) (context.Context, error) {
	var locID string
	var locTypeID string

	if len(s.StudentIDAndParentIDMap) == 0 {
		return ctx, fmt.Errorf("no student ID to create parent with")
	}

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

	evts := make([]*upb.EvtUser, 0, len(s.StudentIDAndParentIDMap))
	parentIDs := []string{}
	for studentID := range s.StudentIDAndParentIDMap {
		// this step will create parent relationship with student
		parentID := "parent-" + studentID
		s.StudentIDAndParentIDMap[studentID] = parentID
		parentIDs = append(parentIDs, parentID)
		evts = append(evts,
			&upb.EvtUser{
				Message: &upb.EvtUser_CreateParent_{
					CreateParent: &upb.EvtUser_CreateParent{
						StudentId:   studentID,
						StudentName: "name" + studentID,
						SchoolId:    s.schoolID,
						ParentId:    parentID,
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
	// find conversation parents created
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(2 * time.Second)
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		conversationMembers, err := getConversationMemberByMemberIDs(ctx2, s.DB, database.TextArray(parentIDs))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return attempt < 5, fmt.Errorf("found no member for parents")
			}
			return false, err
		}
		if conversationMembers == nil {
			return attempt < 10, fmt.Errorf("unable to get conversation member")
		}
		if len(conversationMembers) != len(parentIDs) {
			return attempt < 10, fmt.Errorf("not enough member in db")
		}
		s.ParentConversationIDs = retrieveConversationIDsFromConversationMembers(conversationMembers)
		return true, err
	}); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *suite) teacherJoinConversationOf(ctx context.Context, conversationType string) (context.Context, error) {
	// usually max time out at 15 sec, may change tho
	ctx2, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req := &tpb.JoinConversationsRequest{}
	switch conversationType {
	case "student":
		req.ConversationIds = s.ConversationIDs
	case "parent":
		req.ConversationIds = s.ParentConversationIDs
	}

	s.Request = req
	s.ResponseErr = try.Do(func(attempt int) (bool, error) {
		s.Response, s.ResponseErr = tpb.NewChatModifierServiceClient(s.Conn).
			JoinConversations(contextWithToken(ctx2, s.TeacherToken), req)

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
