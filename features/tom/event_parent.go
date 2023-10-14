package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"

	"go.uber.org/multierr"
)

func genereateConversationMember(conversationID string, teacherID string) (*entities.ConversationMembers, error) {
	e := &entities.ConversationMembers{}
	now := time.Now()
	database.AllNullEntity(e)
	err := multierr.Combine(e.ID.Set(idutil.ULIDNow()), e.UserID.Set(teacherID), e.ConversationID.Set(conversationID), e.Role.Set(entities.ConversationRoleTeacher), e.Status.Set(entities.ConversationStatusActive), e.SeenAt.Set(nil), e.LastNotifyAt.Set(nil), e.UpdatedAt.Set(now), e.CreatedAt.Set(now))
	return e, err
}

func (s *suite) teachersReceiveMessageWithContent(ctx context.Context, msgtype string, msgContent string) (context.Context, error) {
	for _, id := range s.teacherIDs {
		stream, ok := s.SubV2Clients[id]
		if !ok {
			return ctx, fmt.Errorf("stream for parent %s not found", id)
		}
		ctx, err := s.receiveMsgWithTypeContentFromStream(ctx, stream, msgtype, msgContent)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) currentParentsReceiveMessageWithContent(ctx context.Context, msgtype string, msgContent string) (context.Context, error) {
	all := s.parentIDs
	take := make([]string, 0, len(all))
	for _, id := range all {
		if id == s.additionalParentID {
			continue
		}
		take = append(take, id)
	}
	for _, id := range take {
		stream, ok := s.SubV2Clients[id]
		if !ok {
			return ctx, fmt.Errorf("stream for parent %s not found", id)
		}
		ctx, err := s.receiveMsgWithTypeContentFromStream(ctx, stream, msgtype, msgContent)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) teacherJoinstudentConversation(ctx context.Context, teacherNum int) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `SELECT conversation_id FROM conversation_students WHERE student_id = $1 AND conversation_type = 'CONVERSATION_STUDENT'`
	var conversationID string
	err := s.DB.QueryRow(ctx2, query, &s.studentID).Scan(&conversationID)
	if err != nil {
		return ctx, err
	}
	s.conversationID = conversationID

	s.teacherIDs = []string{}
	for i := 0; i < teacherNum; i++ {
		s.teacherIDs = append(s.teacherIDs, idutil.ULIDNow())
		conversationMember, err := genereateConversationMember(conversationID, s.teacherIDs[i])
		if err != nil {
			return ctx, err
		}
		cmdTag, err := database.Insert(ctx2, conversationMember, s.DB.Exec)
		if err != nil {
			return ctx, err
		}
		if cmdTag.RowsAffected() != 1 {
			return ctx, fmt.Errorf("error ")
		}
	}
	return ctx, nil
}

func (s *suite) aStudentConversationWithTeacher(ctx context.Context, teacherNum int) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.createStudentConversation,
		s.teacherJoinstudentConversation, teacherNum,
	)
}

func (s *suite) tomMustCreateConversationForParent(ctx context.Context) (context.Context, error) {
	studentID, chatName := s.studentID, s.chatName

	var convID string
	if err := doRetry(func() (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		query := `SELECT cs.conversation_id FROM conversation_students cs LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
    WHERE cs.student_id = $1 AND c.name = $2 AND owner = $3  AND c.status= 'CONVERSATION_STATUS_NONE' AND cs.conversation_type = 'CONVERSATION_PARENT'`
		rows, err := s.DB.Query(ctx2, query, studentID, chatName, resourcePathFromCtx(ctx))
		if err != nil {
			return false, err
		}
		defer rows.Close()
		var count int
		for rows.Next() {
			err := rows.Scan(&convID)
			if err != nil {
				return false, err
			}
			count++
			if count > 1 {
				return false, fmt.Errorf("only expect to create 1 parent conversation, check student %s", studentID)
			}
		}
		if count == 0 {
			return true, fmt.Errorf("conversation not inserted")
		}
		return false, nil
	}); err != nil {
		return ctx, err
	}
	s.ConversationIDs = append(s.ConversationIDs, convID)
	return ctx, nil
}
func findStudentConversation(ctx context.Context, studentID, conversationName string, db database.QueryExecer) (string, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	query := `SELECT cs.conversation_id FROM conversation_students cs LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
    WHERE cs.student_id = $1 AND c.name = $2 AND c.status= 'CONVERSATION_STATUS_NONE' AND cs.conversation_type = 'CONVERSATION_PARENT'`
	var conversationID string
	err := db.QueryRow(ctx2, query, studentID, conversationName).Scan(&conversationID)
	if err != nil {
		return "", err
	}
	return conversationID, nil
}

func (s *suite) allTeacherInStudentConversationMustBeInParentConversation(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var (
		count int
	)
	studentID, chatName := s.studentID, s.chatName

	conversationID, err := findStudentConversation(ctx2, studentID, chatName, s.DB)
	if err != nil {
		return ctx, err
	}

	checkConversationMember := `SELECT count(*) FROM conversation_members cm WHERE cm.conversation_id =$1 AND cm.user_id =ANY($2)`
	err = s.DB.QueryRow(ctx, checkConversationMember, conversationID, s.teacherIDs).Scan(&count)
	if err != nil {
		return ctx, err
	}

	if count != len(s.teacherIDs) {
		return ctx, fmt.Errorf("not all teacher from student's conversation is in parent conversation: expected %d got %d", len(s.teacherIDs), count)
	}
	return ctx, nil
}
