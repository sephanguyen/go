package tom

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	golibs_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (s *suite) aEvtClassRoomWithMessage(ctx context.Context, arg1 string) (context.Context, error) {
	// nolint:gosec
	switch arg1 {
	case "CreateClass":
		s.teacherID = idutil.ULIDNow()
		s.classID = int32(rand.Intn(100000))
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_CreateClass_{
				CreateClass: &pb.EvtClassRoom_CreateClass{
					ClassId:   s.classID,
					TeacherId: s.teacherID,
				},
			},
		}
	case "CreateClassWithMultipleTeachers":
		s.teacherIDs = []string{idutil.ULIDNow(), idutil.ULIDNow()}
		s.classID = int32(rand.Intn(100000))
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_CreateClass_{
				CreateClass: &pb.EvtClassRoom_CreateClass{
					ClassId:    s.classID,
					TeacherIds: s.teacherIDs,
				},
			},
		}
	case "ActiveConversation":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_ActiveConversation_{
				ActiveConversation: &pb.EvtClassRoom_ActiveConversation{
					ClassId: s.Request.(*pb.EvtClassRoom).GetCreateClass().ClassId,
					Active:  true,
				},
			},
		}
	case "JoinClass":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_JoinClass_{
				JoinClass: &pb.EvtClassRoom_JoinClass{
					ClassId: s.classID,
				},
			},
		}
	case "JoinClass with above teacher":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_JoinClass_{
				JoinClass: &pb.EvtClassRoom_JoinClass{
					ClassId:   s.classID,
					UserId:    s.teacherIDs[0],
					UserGroup: pb.USER_GROUP_TEACHER,
				},
			},
		}
	case "LeaveClass with above userId with isKicked = false":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_LeaveClass_{
				LeaveClass: &pb.EvtClassRoom_LeaveClass{
					ClassId:  s.Request.(*pb.EvtClassRoom).GetJoinClass().ClassId,
					UserIds:  []string{s.Request.(*pb.EvtClassRoom).GetJoinClass().UserId},
					IsKicked: false,
				},
			},
		}
	case "LeaveClass with above userId with isKicked = true":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_LeaveClass_{
				LeaveClass: &pb.EvtClassRoom_LeaveClass{
					ClassId:  s.Request.(*pb.EvtClassRoom).GetJoinClass().ClassId,
					UserIds:  []string{s.Request.(*pb.EvtClassRoom).GetJoinClass().UserId},
					IsKicked: true,
				},
			},
		}
	case "LeaveClass with multiple teacher with isKicked = true":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_LeaveClass_{
				LeaveClass: &pb.EvtClassRoom_LeaveClass{
					ClassId:  s.classID,
					UserIds:  s.teacherIDs,
					IsKicked: true,
				},
			},
		}
	case "LeaveClass with invalid teacher with isKicked = true":
		s.Request = &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_LeaveClass_{
				LeaveClass: &pb.EvtClassRoom_LeaveClass{
					ClassId:  s.classID,
					UserIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					IsKicked: true,
				},
			},
		}
	}

	return ctx, nil
}
func (s *suite) bobSendEventEvtClassRoom(ctx context.Context) (context.Context, error) {
	data, err := s.Request.(*pb.EvtClassRoom).Marshal()
	if err != nil {
		return ctx, err
	}
	s.RequestAt = time.Now()
	subject := golibs_constants.SubjectClassUpserted
	_, err = s.JSM.PublishContext(ctx, subject, data)
	if err != nil {
		return ctx, fmt.Errorf("s.JSM.PublishContext: %w", err)
	}
	return ctx, nil
}
func (s *suite) tomStoreMessageInThisConversation(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var (
		cID pgtype.Text
		now pgtype.Timestamptz
	)
	_ = cID.Set(s.conversationID)
	_ = now.Set(time.Now().Add(time.Second))

	mRepo := &repositories.MessageRepo{}
	messages, err := mRepo.FindAllMessageByConversation(ctx2, s.DB, cID, 10, now)
	if err != nil {
		return ctx, err
	}

	found := false
	for _, m := range messages {
		if m.Message.String == arg2 {
			found = true
			break
		}
	}

	// nolint: goconst
	if !found && arg1 == "must" {
		return ctx, errors.New("not found msg " + arg2)
	}
	// nolint: goconst
	if found && arg1 == "do not" {
		return ctx, errors.New("found msg " + arg2)
	}

	return ctx, nil
}
func (s *suite) aValidIDInJoinClass(ctx context.Context, arg1 string) (context.Context, error) {
	userID := idutil.ULIDNow()
	// nolint: gocritic
	switch arg1 {
	case pb.USER_GROUP_STUDENT.String():
		s.studentID = userID
	}

	s.Request.(*pb.EvtClassRoom).GetJoinClass().UserId = userID
	s.Request.(*pb.EvtClassRoom).GetJoinClass().UserGroup = pb.UserGroup(pb.UserGroup_value[arg1])

	return ctx, nil
}
func (s *suite) tomMustAddAboveUserToThisConversation(ctx context.Context) (context.Context, error) {
	cStatusRepo := &repositories.ConversationMemberRepo{}

	var (
		cID       pgtype.Text
		userID    = s.Request.(*pb.EvtClassRoom).GetJoinClass().UserId
		userGroup = s.Request.(*pb.EvtClassRoom).GetJoinClass().UserGroup.String()
	)
	_ = cID.Set(s.conversationID)

	notFoundErr := errors.New("not found " + userGroup + " in conversation")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(2 * time.Second)

		ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		conversationMembers, err := cStatusRepo.FindByConversationID(ctx2, s.DB, cID)
		if err != nil {
			return false, err
		}

		for _, m := range conversationMembers {
			if m.UserID.String == userID && m.Role.String == userGroup {
				return false, nil
			}
		}

		return attempt < 10, notFoundErr
	}); err != nil {
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) tomRemoveAboveUserFromThisConversation(ctx context.Context, arg string) (context.Context, error) {
	time.Sleep(300 * time.Millisecond)
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	cStatusRepo := &repositories.ConversationMemberRepo{}

	var (
		cID pgtype.Text
	)
	_ = cID.Set(s.conversationID)
	for _, userID := range s.Request.(*pb.EvtClassRoom).GetLeaveClass().UserIds {
		conversationMembers, err := cStatusRepo.FindByCIDAndUserID(ctx2, s.DB, cID, pgtype.Text{String: userID, Status: 2})
		if err != nil {
			if errors.Is(pgx.ErrNoRows, err) && arg == "do not" {
				return ctx, nil
			}
			return ctx, err
		}

		if arg == "must" && conversationMembers.Status.String == entities.ConversationStatusActive {
			return ctx, errors.New("user is not remove from this conversation")
		}

		if arg == "do not" && conversationMembers.Status.String == entities.ConversationStatusInActive {
			return ctx, errors.New("user is remove from this conversation")
		}
	}

	return ctx, nil
}
func (s *suite) tomMustAddAboveTeachersToThisConversation(ctx context.Context) (context.Context, error) {
	time.Sleep(100 * time.Millisecond)
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	cStatusRepo := &repositories.ConversationMemberRepo{}

	var (
		cID pgtype.Text
	)
	_ = cID.Set(s.conversationID)

	conversationMembers, err := cStatusRepo.FindByConversationID(ctx2, s.DB, cID)
	if err != nil {
		return ctx, err
	}

	for _, userID := range s.teacherIDs {
		c := conversationMembers[pgtype.Text{String: userID, Status: 2}]
		if c.Status.String != entities.ConversationStatusActive {
			return ctx, errors.New("ser " + userID + " do not active in conversation status")
		}
	}

	return ctx, nil
}
