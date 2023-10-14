package lesson

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/app/core"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	lessondomain "github.com/manabie-com/backend/internal/tom/domain/lesson"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Logic changing lesson chat goes here
// If communication needed, use embed ChatService
type ChatModifier struct {
	DB                     database.Ext
	ChatInfra              core.ChatInfra
	ChatService            core.ChatService
	ChatReader             core.ChatReader
	Logger                 *zap.Logger
	ConversationMemberRepo interface {
		Create(ctx context.Context, db database.QueryExecer, c *domain.ConversationMembers) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversationMembers []*domain.ConversationMembers) error
		SetStatus(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.TextArray, status pgtype.Text) error
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (map[pgtype.Text]domain.ConversationMembers, error)
		FindByCIDAndUserID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text) (c *domain.ConversationMembers, err error)
	}
	ConversationRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*domain.Conversation) error
	}
	ConversationLessonRepo interface {
		FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*lessondomain.ConversationLesson, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*lessondomain.ConversationLesson) error
		FindByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray, includeSoftDeleted bool) ([]*lessondomain.ConversationLesson, error)
	}

	PrivateConversationLessonRepo interface {
		Create(ctx context.Context, db database.QueryExecer, privateConversation *lessondomain.PrivateConversationLesson) error
	}
	MessageRepo interface {
		Create(context.Context, database.QueryExecer, *domain.Message) error
	}
	UserRepo interface {
		FindByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (map[string]*domain.User, error)
	}
}

func (rcv *ChatModifier) HandleEventUpdateLesson(ctx context.Context, msg *bobproto.EvtLesson_UpdateLesson) error {
	lessonConversation, err := rcv.ConversationLessonRepo.FindByLessonID(ctx, rcv.DB, database.Text(msg.LessonId))
	if err != nil {
		return fmt.Errorf("rcv.ConversationLessonRepo.FindByLessonID: %w", err)
	}
	currentMembers, err := rcv.ConversationMemberRepo.FindByConversationID(ctx, rcv.DB, database.Text(lessonConversation.ConversationID.String))
	if err != nil {
		return fmt.Errorf("rcv.ConversationMemberRepo.FindByConversationID: %w", err)
	}
	// include new membership + soft deleted membership
	upsertMemberships := []*domain.ConversationMembers{}
	newMemberCheckList := map[string]struct{}{}
	for _, reqMember := range msg.LearnerIds {
		_, exist := currentMembers[database.Text(reqMember)]
		newMemberCheckList[reqMember] = struct{}{}
		if !exist {
			e := &domain.ConversationMembers{}
			err := multierr.Combine(
				e.ID.Set(idutil.ULIDNow()),
				e.UserID.Set(reqMember),
				e.ConversationID.Set(lessonConversation.ConversationID.String),
				e.Role.Set(cpb.UserGroup_USER_GROUP_STUDENT.String()),
				e.Status.Set(domain.ConversationStatusActive),
				e.SeenAt.Set(time.Now()),
				e.LastNotifyAt.Set(nil),
				e.CreatedAt.Set(time.Now()),
				e.UpdatedAt.Set(time.Now()),
			)
			if err != nil {
				return fmt.Errorf("domain fields set: %w", err)
			}
			upsertMemberships = append(upsertMemberships, e)
		}
	}
	for userID := range currentMembers {
		_, remain := newMemberCheckList[userID.String]
		if !remain && currentMembers[userID].Role.String == cpb.UserGroup_USER_GROUP_STUDENT.String() {
			currentMember := currentMembers[userID]
			err := currentMember.Status.Set(domain.ConversationStatusInActive)
			if err != nil {
				return fmt.Errorf("currentMember.Status.Set: %w", err)
			}
			upsertMemberships = append(upsertMemberships, &currentMember)
		}
	}

	err = database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		return rcv.ConversationMemberRepo.BulkUpsert(ctx, tx, upsertMemberships)
	})
	if err != nil {
		return fmt.Errorf("database.ExecInTx(ConversationMemberRepo.BulkUpsert): %w", err)
	}

	return nil
}

func (rcv *ChatModifier) HandleEventCreateLesson(ctx context.Context, msg *bobproto.EvtLesson_CreateLessons) error {
	conversations := make([]*domain.Conversation, 0, len(msg.Lessons))
	conversationLessons := make([]*lessondomain.ConversationLesson, 0, len(msg.Lessons))
	lessonConversationMap := map[string]string{}
	for _, lesson := range msg.Lessons {
		cID := idutil.ULIDNow()
		lessonConversationMap[lesson.LessonId] = cID
		now := time.Now()

		conversation := new(domain.Conversation)
		database.AllNullEntity(conversation)
		err := multierr.Combine(
			conversation.ID.Set(cID),
			conversation.ConversationType.Set(pb.CONVERSATION_LESSON.String()),
			conversation.Name.Set(lesson.Name),
			conversation.Status.Set(pb.CONVERSATION_STATUS_NONE.String()),
			conversation.CreatedAt.Set(now),
			conversation.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("error set conversation: %v", err)
		}
		conversations = append(conversations, conversation)

		conversationLesson := new(lessondomain.ConversationLesson)
		database.AllNullEntity(conversationLesson)
		err = multierr.Combine(
			conversationLesson.ConversationID.Set(cID),
			conversationLesson.LessonID.Set(lesson.LessonId),
			conversationLesson.CreatedAt.Set(now),
			conversationLesson.UpdatedAt.Set(now),
			conversationLesson.LatestStartTime.Set(now),
		)
		if err != nil {
			return fmt.Errorf("error set conversationLesson: %w", err)
		}
		conversationLessons = append(conversationLessons, conversationLesson)
	}

	for _, lesson := range msg.GetLessons() {
		memberships := make([]*domain.ConversationMembers, 0, len(lesson.GetLearnerIds()))
		for _, student := range lesson.GetLearnerIds() {
			membership, err := domain.CreateConversationMember(student, lessonConversationMap[lesson.LessonId], cpb.UserGroup_USER_GROUP_STUDENT.String())
			if err != nil {
				return fmt.Errorf("CreateConversationMember: %w", err)
			}
			memberships = append(memberships, membership)
		}
		err := database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
			er := rcv.ConversationRepo.BulkUpsert(ctx, tx, conversations)
			if er != nil {
				return fmt.Errorf("rcv.conversationRepo.BulkUpsert: %w", er)
			}
			er = rcv.ConversationLessonRepo.BulkUpsert(ctx, tx, conversationLessons)
			if er != nil {
				return fmt.Errorf("rcv.conversationLessonRepo.BulkUpsert: %w", er)
			}
			er = rcv.ConversationMemberRepo.BulkUpsert(ctx, tx, memberships)
			if er != nil {
				return fmt.Errorf("rcv.ConversationMemberRepo.BulkUpsert: %w", er)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("database.ExecInTx: %w", err)
		}
	}

	return nil
}

func (rcv *ChatModifier) HandleEventJoinLesson(ctx context.Context, msg *bobproto.EvtLesson_JoinLesson) (bool, error) {
	if msg.UserGroup != bobproto.UserGroup(cpb.UserGroup_USER_GROUP_TEACHER) {
		// only allow teacher to join lesson on demand
		return false, nil
	}
	lessonConversation, err := rcv.ConversationLessonRepo.FindByLessonID(ctx, rcv.DB, database.Text(msg.LessonId))
	if err != nil {
		return true, errors.Wrap(err, "rcv.conversationLessonRepo.FindByLessonID")
	}

	cID := lessonConversation.ConversationID.String

	membership, err := domain.CreateConversationMember(msg.UserId, cID, msg.UserGroup.String())
	if err != nil {
		return true, fmt.Errorf("CreateConversationMember: %w", err)
	}
	// if set SeenAt to now(), previous messages of this session will be treated as seen
	err = membership.SeenAt.Set(lessonConversation.LatestStartTime.Time)
	if err != nil {
		return true, fmt.Errorf("membership.SeenAt.Set: %w", err)
	}
	err = rcv.ConversationMemberRepo.Create(ctx, rcv.DB, membership)
	if err != nil {
		return true, fmt.Errorf("rcv.ConversationMemberRepo.Create: %w", err)
	}

	_, err = rcv.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
		ConversationId: cID,
		Message:        pb.CODES_MESSAGE_TYPE_JOINED_LESSON.String(),
		UrlMedia:       "",
		TargetUser:     msg.UserId,
		Type:           pb.MESSAGE_TYPE_SYSTEM,
	}, domain.MessageToConversationOpts{Persist: true, AsUser: false})
	if err != nil {
		rcv.Logger.Warn("rcv.ChatService.SendMessageToConversation", zap.Error(err))
	}
	return false, nil
}

var (
	errOnlyTeacherCanLeaveLesson      = errors.New("only teacher allowed to leave lesson")
	errLessonConversationDoesNotExist = errors.New("lesson conversation does not exist")
)

func (rcv *ChatModifier) HandleEventLeaveLesson(ctx context.Context, msg *bobproto.EvtLesson_LeaveLesson) (bool, error) {
	lessonConversation, err := rcv.ConversationLessonRepo.FindByLessonID(ctx, rcv.DB, database.Text(msg.LessonId))
	if err != nil {
		return true, fmt.Errorf("rcv.ConversationLessonRepo.FindByLessonID: %w", err)
	}

	membership, err := rcv.ConversationMemberRepo.FindByCIDAndUserID(ctx, rcv.DB, lessonConversation.ConversationID, database.Text(msg.UserId))
	if err != nil {
		return true, fmt.Errorf("rcv.ConversationMemberRepo.FindByCIDAndUserID: %w", err)
	}
	// only teacher is allowed to actively leave lesson
	if membership.Role.String != cpb.UserGroup_USER_GROUP_TEACHER.String() {
		return false, errOnlyTeacherCanLeaveLesson
	}

	err = rcv.ConversationMemberRepo.SetStatus(ctx, rcv.DB, lessonConversation.ConversationID, database.TextArray([]string{msg.UserId}), database.Text(domain.ConversationStatusInActive))
	if err != nil {
		return true, fmt.Errorf("rcv.ConversationMemberRepo.SetStatus: %w", err)
	}
	_, err = rcv.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
		ConversationId: lessonConversation.ConversationID.String,
		Message:        pb.CODES_MESSAGE_TYPE_LEFT_LESSON.String(),
		UrlMedia:       "",
		TargetUser:     msg.UserId,
		Type:           pb.MESSAGE_TYPE_SYSTEM,
	}, domain.MessageToConversationOpts{Persist: true, AsUser: false})

	if err != nil {
		rcv.Logger.Warn("rcv.sendMessage", zap.Error(err))
	}
	return false, nil
}

func (rcv *ChatModifier) HandleEventEndLiveLesson(ctx context.Context, msg *bobproto.EvtLesson_EndLiveLesson) error {
	c, err := rcv.ConversationLessonRepo.FindByLessonID(ctx, rcv.DB, database.Text(msg.LessonId))
	if err != nil {
		return fmt.Errorf("rcv.conversationLessonRepo.FindByLessonID: %w", err)
	}

	_, err = rcv.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
		ConversationId: c.ConversationID.String,
		Message:        pb.CODES_MESSAGE_TYPE_END_LIVE_LESSON.String(),
		Type:           pb.MESSAGE_TYPE_SYSTEM,
	}, domain.MessageToConversationOpts{Persist: true, AsUser: false})
	if err != nil {
		rcv.Logger.Warn("rcv.sendMessage", zap.Error(err))
	}

	return nil
}
func (rcv *ChatModifier) SyncLessonConversationStudents(ctx context.Context, reqs []*npb.EventSyncUserCourse_StudentLesson) error {
	memberships := []*domain.ConversationMembers{}
	lessons := []string{}
	for _, student := range reqs {
		lessons = append(lessons, student.GetLessonIds()...)
	}
	uniqueLessons := golibs.GetUniqueElementStringArray(lessons)
	lessonConvs, err := rcv.ConversationLessonRepo.FindByLessonIDs(ctx, rcv.DB, database.TextArray(uniqueLessons), false)
	if err != nil {
		return fmt.Errorf("ConversationLessonRepo.FindByLessonIDs: %w", err)
	}
	lessonConvMap := map[string]string{}
	for _, lessonConv := range lessonConvs {
		lessonConvMap[lessonConv.LessonID.String] = lessonConv.ConversationID.String
	}

	for _, req := range reqs {
		setStatus := domain.ConversationStatusActive
		if req.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			setStatus = domain.ConversationStatusInActive
		}
		for _, lesson := range req.GetLessonIds() {
			convID, exist := lessonConvMap[lesson]
			if !exist {
				return errLessonConversationDoesNotExist
			}
			e := &domain.ConversationMembers{}
			err := multierr.Combine(
				e.ID.Set(idutil.ULIDNow()),
				e.UserID.Set(req.GetStudentId()),
				e.ConversationID.Set(convID),
				e.Role.Set(cpb.UserGroup_USER_GROUP_STUDENT.String()),
				e.Status.Set(setStatus),
				e.SeenAt.Set(time.Now()),
				e.LastNotifyAt.Set(nil),
				e.CreatedAt.Set(time.Now()),
				e.UpdatedAt.Set(time.Now()),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine(domainFields.Set): %w", err)
			}
			memberships = append(memberships, e)
		}
	}

	err = rcv.ConversationMemberRepo.BulkUpsert(ctx, rcv.DB, memberships)
	if err != nil {
		return fmt.Errorf("ConversationMember.BulkUpsert: %w", err)
	}
	return nil
}

func (rcv *ChatModifier) CreateLiveLessonPrivateConversation(ctx context.Context, req *tpb.CreateLiveLessonPrivateConversationRequest) (*tpb.CreateLiveLessonPrivateConversationResponse, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)

	var userIDs []string
	userIDs = append(userIDs, req.UserIds...)
	// if the request does not contain sender Id, add sender Id to the slice
	if !slices.Contains(userIDs, currentUserID) {
		userIDs = append(userIDs, currentUserID)
	}

	conversationID := idutil.ULIDNow()
	now := time.Now()

	conversation := new(domain.Conversation)
	database.AllNullEntity(conversation)
	err := multierr.Combine(
		conversation.ID.Set(conversationID),
		conversation.ConversationType.Set(tpb.ConversationType_CONVERSATION_LESSON_PRIVATE.String()),
		conversation.Status.Set(pb.CONVERSATION_STATUS_NONE.String()),
		conversation.CreatedAt.Set(now),
		conversation.UpdatedAt.Set(now),
	)
	if err != nil {
		return nil, fmt.Errorf("error set Conversation: %v", err)
	}

	memberships := make([]*domain.ConversationMembers, 0, len(userIDs))
	ids := make([]string, 0, len(userIDs))

	usersMap, err := rcv.UserRepo.FindByUserIDs(ctx, rcv.DB, userIDs)
	if err != nil {
		return nil, fmt.Errorf("rcv.UserRepo.FindByUserIDs %w", err)
	}
	convUsers := make([]*tpb.Conversation_User, 0, len(userIDs))
	for _, userID := range userIDs {
		user, userExists := usersMap[userID]
		if !userExists {
			return nil, fmt.Errorf("error not found user")
		}
		member, err := domain.CreateConversationMember(userID, conversationID, user.UserGroup.String)
		if err != nil {
			return nil, fmt.Errorf("cannot create conversation member :%w", err)
		}
		ids = append(ids, member.UserID.String)
		memberships = append(memberships, member)

		convUsers = append(convUsers, &tpb.Conversation_User{
			Id:        member.UserID.String,
			Group:     cpb.UserGroup(cpb.UserGroup_value[member.Role.String]),
			IsPresent: member.Status.String == domain.ConversationStatusActive,
			SeenAt:    timestamppb.New(member.SeenAt.Time),
		})
	}

	if len(memberships) != 2 {
		return nil, fmt.Errorf("error invalid memberships")
	}
	if len(ids) != 2 {
		return nil, fmt.Errorf("error invalid user ids")
	}

	flattenUserIds := getFlattenUserIdsByAscending(ids)
	pcLesson := new(lessondomain.PrivateConversationLesson)
	database.AllNullEntity(pcLesson)
	err = multierr.Combine(
		pcLesson.ConversationID.Set(conversationID),
		pcLesson.LessonID.Set(req.LessonId),
		pcLesson.FlattenUserIds.Set(flattenUserIds),
		pcLesson.CreatedAt.Set(now),
		pcLesson.UpdatedAt.Set(now),
		pcLesson.LatestStartTime.Set(now),
	)
	if err != nil {
		return nil, fmt.Errorf("error set PrivateConversationLesson: %w", err)
	}

	err = database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		conversations := make([]*domain.Conversation, 0, 1)
		conversations = append(conversations, conversation)
		er := rcv.ConversationRepo.BulkUpsert(ctx, tx, conversations)
		if er != nil {
			return fmt.Errorf("rcv.conversationRepo.BulkUpsert: %w", er)
		}
		er = rcv.PrivateConversationLessonRepo.Create(ctx, tx, pcLesson)
		if er != nil {
			return fmt.Errorf("rcv.PrivateConversationLessonRepo.Create: %w", er)
		}
		er = rcv.ConversationMemberRepo.BulkUpsert(ctx, tx, memberships)
		if er != nil {
			return fmt.Errorf("rcv.ConversationMemberRepo.BulkUpsert: %w", er)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("database.ExecInTx: %w", err)
	}

	msg, err := rcv.persistCreatedMessagePrivateConversation(ctx, conversationID, currentUserID)
	if err != nil {
		rcv.Logger.Error("persistCreatedMessagePrivateConversation", zap.Error(err))
	}

	rcv.broadcastPrivateConversationLessonCreated(ctx, currentUserID, req.UserIds, conversationID, msg)

	resp := new(tpb.CreateLiveLessonPrivateConversationResponse)
	resp.Conversation = &tpb.Conversation{
		ConversationId:   conversation.ID.String,
		Seen:             false,
		Users:            convUsers,
		Status:           tpb.ConversationStatus(tpb.ConversationStatus_value[conversation.Status.String]),
		ConversationType: tpb.ConversationType(tpb.ConversationType_value[conversation.ConversationType.String]),
		ConversationName: conversation.Name.String,
	}

	return resp, nil
}

func (rcv *ChatModifier) persistCreatedMessagePrivateConversation(ctx context.Context, conversationID string, userID string) (*domain.Message, error) {
	message := new(domain.Message)
	database.AllNullEntity(message)
	err := multierr.Combine(
		message.UserID.Set(userID),
		message.ID.Set(idutil.ULIDNow()),
		message.ConversationID.Set(conversationID),
		message.Message.Set(tpb.CodesMessageType_CODES_MESSAGE_TYPE_CREATED_PRIVATE_CONVERSATION_LESSON.String()),
		message.Type.Set(pb.MESSAGE_TYPE_SYSTEM),
	)
	if err != nil {
		return nil, err
	}

	err = rcv.MessageRepo.Create(ctx, rcv.DB, message)
	if err != nil {
		return nil, fmt.Errorf("rcv.MessageRepo.Create: %w", err)
	}
	return message, nil
}

func (rcv *ChatModifier) broadcastPrivateConversationLessonCreated(ctx context.Context, senderID string, targetIDs []string, conversationID string, message *domain.Message) {
	createdPrivateConversationMessageEvent := &pb.Event_EventNewMessage{
		EventNewMessage: &pb.MessageResponse{
			MessageId:      message.ID.String,
			ConversationId: conversationID,
			Content:        tpb.CodesMessageType_CODES_MESSAGE_TYPE_CREATED_PRIVATE_CONVERSATION_LESSON.String(),
			Type:           pb.MESSAGE_TYPE_SYSTEM,
			UserId:         senderID,
		},
	}

	err := rcv.ChatInfra.PushMessage(ctx, targetIDs, &pb.Event{
		Event: createdPrivateConversationMessageEvent,
	}, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			Enabled:      true,
			IgnoredUsers: []string{senderID},
			Silence:      true,
		},
	})
	if err != nil {
		rcv.Logger.Error("chatInfra.PushMessage", zap.Error(err))
	}
}
