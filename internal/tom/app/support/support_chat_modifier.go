package support

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/try"
	utils "github.com/manabie-com/backend/internal/tom/app"
	"github.com/manabie-com/backend/internal/tom/app/core"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	sentities "github.com/manabie-com/backend/internal/tom/domain/support"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Logic changing student/parent chat goes here
// If communication needed, use embed ChatService
type ChatModifier struct {
	DB database.Ext

	ChatService  core.ChatService
	Logger       *zap.Logger
	JSM          nats.JetStreamManagement
	LocationRepo interface {
		FindAccessPaths(ctx context.Context, db database.Ext, locationIDs []string) ([]string, error)
	}

	ConversationMemberRepo interface {
		FindByCIDAndUserID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text) (c *domain.ConversationMembers, err error)
		FindByCIDsAndUserID(ctx context.Context, db database.QueryExecer, conversationID pgtype.TextArray, userID pgtype.Text) (c []*domain.ConversationMembers, err error)
		Create(ctx context.Context, db database.QueryExecer, c *domain.ConversationMembers) error
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (map[pgtype.Text]domain.ConversationMembers, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversationMembers []*domain.ConversationMembers) error
		SetStatus(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.TextArray, status pgtype.Text) error
		SetStatusByConversationAndUserIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, userIDs pgtype.TextArray, status pgtype.Text) error
		FindUserIDConversationIDsMapByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) (map[string][]string, error)
	}
	//TODO: move to core
	ConversationLocationRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []domain.ConversationLocation) error
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (map[string][]domain.ConversationLocation, error)
	}
	ConversationStudentRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*sentities.ConversationStudent) error
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, conversationType pgtype.Text) ([]string, error)
		FindByStaffIDs(ctx context.Context, db database.QueryExecer, staffIDs pgtype.TextArray) ([]string, error)
	}
	ConversationRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*domain.Conversation) error
		FindConversationIdsBySchoolIds(ctx context.Context, db database.QueryExecer, schoolIds pgtype.TextArray) ([]*domain.Conversation, error)
		ListConversationUnjoined(ctx context.Context, db database.QueryExecer, filter *domain.ListConversationUnjoinedFilter) ([]*domain.Conversation, error)
		ListConversationUnjoinedInLocations(ctx context.Context, db database.QueryExecer, filter *domain.ListConversationUnjoinedFilter) ([]*domain.Conversation, error)
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *domain.Conversation, err error)
	}
	GrantedPermissionRepo interface {
		FindByUserIDAndPermissionName(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, permissionName pgtype.Text) (map[string][]*domain.GrantedPermission, error)
	}
	UserGroupMemberRepo interface {
		FindUserIDsByUserGroupID(ctx context.Context, db database.QueryExecer, userGroupID pgtype.Text) ([]string, error)
	}
}

func NewChatModifier(
	dbTrace database.Ext,
	chatService core.ChatService,
	zapLogger *zap.Logger,
) *ChatModifier {
	supportChatModifier := &ChatModifier{
		DB:          dbTrace,
		ChatService: chatService,
		Logger:      zapLogger,
	}
	domain.RegisterSubscriber(domain.Subscription{
		Event:   domain.MessageSentEventStr,
		Handler: supportChatModifier.HandleCoreMessageSent,
	})
	return supportChatModifier
}

func toConversationLocations(convID string, locs []string) ([]domain.ConversationLocation, error) {
	ret := make([]domain.ConversationLocation, 0, len(locs))

	now := time.Now()
	for _, loc := range locs {
		var e domain.ConversationLocation
		database.AllNullEntity(&e)
		err := multierr.Combine(
			e.ConversationID.Set(convID),
			e.LocationID.Set(loc),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return nil, err
		}
		ret = append(ret, e)
	}
	return ret, nil
}

func toConversationEn(conversationID string, studentID string, conversationType string) (*sentities.ConversationStudent, error) {
	var e sentities.ConversationStudent
	database.AllNullEntity(&e)
	now := time.Now()
	err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.ConversationID.Set(conversationID),
		e.StudentID.Set(studentID),
		e.ConversationType.Set(conversationType),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return &e, err
}

func CreateConversationMember(userID, conversationID, role string) (*domain.ConversationMembers, error) {
	e := &domain.ConversationMembers{}
	now := time.Now()
	err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.UserID.Set(userID),
		e.ConversationID.Set(conversationID),
		e.Role.Set(role),
		e.Status.Set(domain.ConversationStatusActive),
		e.SeenAt.Set(time.Now()),
		e.LastNotifyAt.Set(nil),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return e, err
}

func (c *ChatModifier) HandleParentRemovedFromStudent(ctx context.Context, msg *upb.EvtUser_ParentRemovedFromStudent) error {
	var cID string
	err := validateParentRemovedFromStudent(msg)
	if err != nil {
		return fmt.Errorf("validateParentRemoveFromStudent: %w", err)
	}
	err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		convIDs, err := c.ConversationStudentRepo.FindByStudentIDs(ctx, tx, database.TextArray([]string{msg.GetStudentId()}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String()))
		if err != nil {
			return fmt.Errorf("c.ConversationStudentRepo.FindByStudentIDs: %w", err)
		}
		if len(convIDs) != 1 {
			return fmt.Errorf("student should have 1 conversation type parent, but has%d", len(convIDs))
		}
		cID = convIDs[0]
		_, err = c.RemoveMembership(ctx, tx, []string{cID}, msg.GetParentId())

		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}

	event := &tpb.ConversationInternal{
		Message: &tpb.ConversationInternal_MemberRemoved{
			MemberRemoved: &tpb.ConversationInternal_MemberRemovedFromConversation{
				ConversationId: cID,
				MemberId:       msg.GetParentId(),
			},
		},
		TriggeredAt: timestamppb.Now(),
	}
	bs, err := proto.Marshal(event)
	if err != nil {
		c.Logger.Warn("proto.Marshal", zap.Error(err))
		return err
	}
	_, err = c.JSM.TracedPublish(ctx, "ParentRemovedFromStudent.TracedPublish", constants.SubjectChatMembersUpdated, bs)
	// publish to conversation evt
	if err != nil {
		return err
	}

	logger := ctxzap.Extract(ctx)
	// also notify other clients of this user that he has left conversations, in case of multiple tabs/devices
	msgpb, err := c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
		ConversationId: cID,
		Message:        pb.CODES_MESSAGE_TYPE_LEAVE_CONVERSATION.String(),
		Type:           pb.MESSAGE_TYPE_SYSTEM,
		TargetUser:     msg.GetParentId(),
	}, domain.MessageToConversationOpts{Persist: true, AsUser: false})

	if err != nil {
		logger.Error("c.ChatService.SendMessageToConversation:", zap.Error(err))
	}
	// push message to clients of this user
	err = c.ChatService.SendMessageToUsers(ctx, []string{msg.GetParentId()}, &pb.Event{
		Event: &pb.Event_EventNewMessage{
			EventNewMessage: msgpb,
		},
	}, domain.MessageToUserOpts{Notification: domain.NotificationOpts{}})
	if err != nil {
		logger.Error("c.ChatService.PushMessage", zap.Error(err))
	}

	return nil
}

func (c *ChatModifier) HandleEventCreateStudentConversation(ctx context.Context, msg *upb.EvtUser_CreateStudent) error {
	conversation := new(domain.Conversation)
	database.AllNullEntity(conversation)
	cID := idutil.ULIDNow()
	now := time.Now()

	err := multierr.Combine(
		conversation.ID.Set(cID),
		conversation.ConversationType.Set(tpb.ConversationType_CONVERSATION_STUDENT.String()),
		conversation.Name.Set(msg.StudentName),
		conversation.Status.Set(pb.CONVERSATION_STATUS_NONE.String()),
		conversation.CreatedAt.Set(now),
		conversation.UpdatedAt.Set(now),
		conversation.Owner.Set(msg.SchoolId),
	)
	if err != nil {
		return fmt.Errorf("error set conversation: %v", err)
	}
	member, err := CreateConversationMember(msg.StudentId, cID, cpb.UserGroup_USER_GROUP_STUDENT.String())
	if err != nil {
		return err
	}
	var existAlready = false

	// TODO: this Tx can be split by:
	// make conversation_student has a nullable conversation_id, insert this first
	// then create conversation related entities, provide a unique constraint in conversation table: (conversation_type,process_id) for ex
	err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		studentConversation, err := toConversationEn(cID, msg.StudentId, tpb.ConversationType_CONVERSATION_STUDENT.String())
		if err != nil {
			return fmt.Errorf("toConversationEn: %w", err)
		}
		err = c.ConversationRepo.BulkUpsert(ctx, tx, []*domain.Conversation{conversation})
		if err != nil {
			return fmt.Errorf("c.conversationRepo.BulkCreate: %v", err)
		}
		locations, err := toConversationLocations(cID, msg.GetLocationIds())
		if err != nil {
			return fmt.Errorf("toConvAccessPaths %v", err)
		}

		err = c.ConversationLocationRepo.BulkUpsert(ctx, tx, locations)
		if err != nil {
			return fmt.Errorf("c.conversationAccessPath.BulkCreate: %v", err)
		}

		err = c.ConversationMemberRepo.Create(ctx, tx, member)
		if err != nil {
			return fmt.Errorf("c.ConversationMemberRepo.Create: %v", err)
		}
		err = c.ConversationStudentRepo.BulkUpsert(ctx, tx, []*sentities.ConversationStudent{studentConversation})
		if err != nil {
			pgerr, ok := errors.Unwrap(err).(*pgconn.PgError)
			if ok && pgerr.Code == pgerrcode.UniqueViolation {
				existAlready = true
				return nil
			}
			return fmt.Errorf("c.ConversationStudentRepo.BulkUpsert: %v", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	conversationID := member.ConversationID.String

	if existAlready {
		// find the already created conversation and notify elasticsearch to find this conversation
		conversationIds, err := c.ConversationStudentRepo.FindByStudentIDs(ctx, c.DB,
			database.TextArray([]string{msg.GetStudentId()}),
			database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()))
		if err != nil {
			return err
		}
		if len(conversationIds) == 0 {
			return fmt.Errorf("existAlready is true but conversation ID not found")
		}
		if len(conversationIds) > 1 {
			return fmt.Errorf("existAlready is true but found many conversation of same student")
		}
		conversationID = conversationIds[0]
	}

	// publish to conversation evt
	event := &tpb.ConversationInternal{
		TriggeredAt: timestamppb.Now(),
		Message: &tpb.ConversationInternal_ConversationCreated_{
			ConversationCreated: &tpb.ConversationInternal_ConversationCreated{
				ConversationId: conversationID,
				Type:           tpb.ConversationType_CONVERSATION_STUDENT.String(),
			},
		},
	}
	bs, err := proto.Marshal(event)
	if err != nil {
		c.Logger.Warn("proto.Marshal", zap.Error(err))
		return err
	}
	_, err = c.JSM.TracedPublish(ctx, "CreateStudentConversation.TracedPublish", constants.SubjectChatCreated, bs)
	if err != nil {
		c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
		return err
	}

	if existAlready {
		return nil
	}

	_, err = c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
		ConversationId: cID,
		Message:        pb.CODES_MESSAGE_TYPE_CREATED_CONVERSATION.String(),
		Type:           pb.MESSAGE_TYPE_SYSTEM,
	}, domain.MessageToConversationOpts{Persist: true, AsUser: false})

	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error("Error sending message:", zap.Error(err))
	}

	return nil
}

// TODO: split this function
func (c *ChatModifier) HandleEventCreateParentConversation(ctx context.Context, msg *upb.EvtUser_ParentAssignedToStudent) error {
	return try.Do(func(attempt int) (bool, error) {
		// first, check the conversation in parent is exist or not
		conversationIds, err := c.ConversationStudentRepo.FindByStudentIDs(ctx, c.DB, database.TextArray([]string{msg.GetStudentId()}),
			database.Text(tpb.ConversationType_CONVERSATION_PARENT.String()))
		if err != nil {
			return false, err
		}
		if len(conversationIds) > 1 {
			return false, fmt.Errorf("student has more than one parent's conversation")
		}
		// Parent conversation already created, just add this parent to conversation
		if len(conversationIds) == 1 {
			var existAlready = false
			err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
				// with spec, a student will have 2 conversation: one for teacher & student && two for their parents and teacher
				// to avoid violates key when more than one person related to the student in the conversation
				member, err := CreateConversationMember(msg.ParentId, conversationIds[0], cpb.UserGroup_USER_GROUP_PARENT.String())
				if err != nil {
					return err
				}
				err = c.ConversationMemberRepo.BulkUpsert(ctx, tx, []*domain.ConversationMembers{member})
				if err != nil {
					pgerr, ok := errors.Unwrap(err).(*pgconn.PgError)
					if ok && pgerr.Code == pgerrcode.UniqueViolation {
						existAlready = true
						return nil
					}
					return fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %v", err)
				}
				event := &tpb.ConversationInternal{
					Message: &tpb.ConversationInternal_MemberAdded{
						MemberAdded: &tpb.ConversationInternal_MemberAddedToConversation{
							ConversationId: member.ConversationID.String,
							MemberId:       member.UserID.String,
						},
					},
					TriggeredAt: timestamppb.Now(),
				}
				bs, err := proto.Marshal(event)
				if err != nil {
					c.Logger.Warn("proto.Marshal", zap.Error(err))
					return err
				}
				_, err = c.JSM.TracedPublish(ctx, "CreateParentConversation.TracedPublish", constants.SubjectChatMembersUpdated, bs)
				if err != nil {
					c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
					return err
				}
				return nil
			})
			if err != nil {
				return false, err
			}
			if existAlready {
				return false, nil
			}
			_, err = c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
				ConversationId: conversationIds[0],
				Message:        pb.CODES_MESSAGE_TYPE_USER_ADDED_TO_CONVERSATION.String(),
				Type:           pb.MESSAGE_TYPE_SYSTEM,
			}, domain.MessageToConversationOpts{Persist: true, AsUser: false})
			if err != nil {
				logger := ctxzap.Extract(ctx)
				logger.Error("Error sending message:", zap.Error(err))
			}
			return false, nil
		}

		// Creating new parent conversation with current parent is the first member

		// Find student's conversation
		studentConversationIDs, err := c.ConversationStudentRepo.FindByStudentIDs(ctx, c.DB, database.TextArray([]string{msg.StudentId}),
			database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()))
		if err != nil {
			return false, err
		}
		if len(studentConversationIDs) > 1 {
			return false, fmt.Errorf("student has more than one student's conversation")
		}
		if len(studentConversationIDs) == 0 {
			return false, fmt.Errorf("student has no conversation")
		}
		studentConvID := studentConversationIDs[0]

		conversationMembers, err := c.ConversationMemberRepo.FindByConversationID(ctx, c.DB, database.Text(studentConvID))
		if err != nil {
			return false, fmt.Errorf("conversationMember: %w", err)
		}
		// if the student parent's conversation doesn't existed
		conversation := new(domain.Conversation)
		database.AllNullEntity(conversation)
		cID := idutil.ULIDNow()
		now := time.Now()
		conv, err := c.ConversationRepo.FindByID(ctx, c.DB, database.Text(studentConvID))
		if err != nil {
			return false, fmt.Errorf("c.ConversationRepo.FindByID: %w", err)
		}
		err = multierr.Combine(
			conversation.ID.Set(cID),
			conversation.ConversationType.Set(tpb.ConversationType_CONVERSATION_PARENT.String()),
			conversation.Name.Set(conv.Name.String),
			conversation.Status.Set(pb.CONVERSATION_STATUS_NONE.String()),
			conversation.CreatedAt.Set(now),
			conversation.UpdatedAt.Set(now),
			conversation.Owner.Set(conv.Owner.String),
		)
		if err != nil {
			return false, fmt.Errorf("error set conversation: %v", err)
		}

		var locations []string
		studentConvLocations, err := c.ConversationLocationRepo.FindByConversationIDs(ctx, c.DB, database.TextArray([]string{studentConvID}))
		if err != nil {
			return false, fmt.Errorf("ConversationLocationRepo.FindByConversationIDs %w", err)
		}

		if locs, exist := studentConvLocations[studentConvID]; exist {
			for _, ent := range locs {
				locations = append(locations, ent.LocationID.String)
			}
		}

		// Add all teacher from student's conversation into parent's conversation
		members := make([]*domain.ConversationMembers, 0, len(conversationMembers)+1)
		for _, conversationMember := range conversationMembers {
			if conversationMember.Role.String == cpb.UserGroup_USER_GROUP_TEACHER.String() {
				member, err := CreateConversationMember(conversationMember.UserID.String, cID, cpb.UserGroup_USER_GROUP_TEACHER.String())
				if err != nil {
					return false, fmt.Errorf("cannot create conversation member :%w", err)
				}
				members = append(members, member)
			}
		}

		member, err := CreateConversationMember(msg.ParentId, cID, cpb.UserGroup_USER_GROUP_PARENT.String())
		if err != nil {
			return false, err
		}

		members = append(members, member)
		var retry bool
		err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
			err = c.ConversationRepo.BulkUpsert(ctx, tx, []*domain.Conversation{conversation})
			if err != nil {
				return fmt.Errorf("c.conversationRepo.BulkUpsert: %v", err)
			}
			if len(locations) > 0 {
				locations, err := toConversationLocations(cID, locations)
				if err != nil {
					return fmt.Errorf("toConvAccessPaths %v", err)
				}

				err = c.ConversationLocationRepo.BulkUpsert(ctx, tx, locations)
				if err != nil {
					return fmt.Errorf("c.conversationAccessPath.BulkCreate: %v", err)
				}
			}

			studentConversation, err := toConversationEn(cID, msg.StudentId, tpb.ConversationType_CONVERSATION_PARENT.String())
			if err != nil {
				return fmt.Errorf("toConversationEn: %w", err)
			}
			err = c.ConversationStudentRepo.BulkUpsert(ctx, tx, []*sentities.ConversationStudent{studentConversation})
			if err != nil {
				pgerr, ok := errors.Unwrap(err).(*pgconn.PgError)
				if ok && pgerr.Code == "23505" {
					retry = true
					return err
				}
				return fmt.Errorf("c.ConversationStudentRepo.BulkUpsert: %v", err)
			}

			err = c.ConversationMemberRepo.BulkUpsert(ctx, tx, members)
			if err != nil {
				return fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %v", err)
			}
			return nil
		})

		if err != nil {
			if retry && attempt < 2 {
				time.Sleep(10 * time.Millisecond)
				return true, err
			}
			return false, err
		}

		event := &tpb.ConversationInternal{
			TriggeredAt: timestamppb.Now(),
			Message: &tpb.ConversationInternal_ConversationCreated_{
				ConversationCreated: &tpb.ConversationInternal_ConversationCreated{
					ConversationId: member.ConversationID.String,
					Type:           tpb.ConversationType_CONVERSATION_PARENT.String(),
				},
			},
		}

		bs, err := proto.Marshal(event)
		if err != nil {
			c.Logger.Warn("proto.Marshal", zap.Error(err))
			return false, err
		}
		_, err = c.JSM.TracedPublish(ctx, "CreateParentConversation.TracedPublish", constants.SubjectChatCreated, bs)
		if err != nil {
			c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
			return false, err
		}

		_, err = c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
			ConversationId: cID,
			Message:        pb.CODES_MESSAGE_TYPE_CREATED_CONVERSATION.String(),
			Type:           pb.MESSAGE_TYPE_SYSTEM,
		}, domain.MessageToConversationOpts{Persist: true, AsUser: false})
		if err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error("Error sending message:", zap.Error(err))
		}

		return false, nil
	})
}

func (c *ChatModifier) JoinConversations(ctx context.Context, req *tpb.JoinConversationsRequest) (*tpb.JoinConversationsResponse, error) {
	userGroup, userID, _ := interceptors.GetUserInfoFromContext(ctx)
	if !utils.IsStaff(userGroup) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("user do not have permission with role %s to join conversations", userGroup))
	}

	members := make([]*domain.ConversationMembers, len(req.ConversationIds))
	for i, conversationID := range req.ConversationIds {
		member, err := CreateConversationMember(userID, conversationID, userGroup)
		if err != nil {
			return nil, fmt.Errorf("error create conversation member")
		}
		members[i] = member
	}
	err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.ConversationMemberRepo.BulkUpsert(ctx, tx, members)
		if err != nil {
			return fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %w", err)
	}
	// publish to conversation evt
	for _, id := range req.ConversationIds {
		evt := &tpb.ConversationInternal{
			TriggeredAt: timestamppb.Now(),
			Message: &tpb.ConversationInternal_MemberAdded{
				MemberAdded: &tpb.ConversationInternal_MemberAddedToConversation{
					ConversationId: id,
					MemberId:       userID,
				},
			},
		}
		bs, err := proto.Marshal(evt)
		if err != nil {
			c.Logger.Warn("proto.Marshal", zap.Error(err))
			return nil, err
		}
		_, err = c.JSM.TracedPublish(ctx, "JoinConversations.TracedPublish", constants.SubjectChatMembersUpdated, bs)
		if err != nil {
			c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
			return nil, err
		}
	}

	logger := ctxzap.Extract(ctx)
	for _, cID := range req.ConversationIds {
		_, err = c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
			ConversationId: cID,
			Message:        pb.CODES_MESSAGE_TYPE_JOINED_CONVERSATION.String(),
			Type:           pb.MESSAGE_TYPE_SYSTEM,
		}, domain.MessageToConversationOpts{Persist: true, AsUser: false})
		if err != nil {
			logger.Error("JoinConversations:", zap.Error(err))
		}
	}
	return &tpb.JoinConversationsResponse{}, nil
}

func (c *ChatModifier) JoinAllConversationsWithLocations(ctx context.Context, req *tpb.JoinAllConversationRequest) (*tpb.JoinAllConversationResponse, error) {
	userGroup, userID, schoolIDs := interceptors.GetUserInfoFromContext(ctx)
	if len(schoolIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no school ids")
	}
	if !utils.IsStaff(userGroup) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("ChatModifierService.JoinAllConversationsWithLocations: user do not have permission with role %s", userGroup))
	}
	pgIDs := database.TextArray(schoolIDs)

	// joinAllConversation in school
	if len(req.GetLocationIds()) == 0 {
		conversations, err := c.ConversationRepo.ListConversationUnjoined(ctx, c.DB, &domain.ListConversationUnjoinedFilter{
			UserID: database.Text(userID), OwnerIDs: pgIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("ConversationRepo.ListConversationUnjoined: %w", err)
		}
		return c.makeUserJoinConversations(ctx, conversations, userID, userGroup)
	}
	accPathInDB, err := c.LocationRepo.FindAccessPaths(ctx, c.DB, req.GetLocationIds())
	if err != nil {
		return nil, fmt.Errorf("LocationRepo.FindAccessPaths %w", err)
	}

	conversations, err := c.ConversationRepo.ListConversationUnjoinedInLocations(ctx, c.DB, &domain.ListConversationUnjoinedFilter{
		UserID: database.Text(userID), OwnerIDs: pgIDs, AccessPaths: database.TextArray(accPathInDB),
	})
	if err != nil {
		return nil, fmt.Errorf("ConversationRepo.ListConversationUnjoinedInLocations: %w", err)
	}
	return c.makeUserJoinConversations(ctx, conversations, userID, userGroup)
}

func (c *ChatModifier) makeUserJoinConversations(ctx context.Context, conversations domain.Conversations, userID, userGroup string) (*tpb.JoinAllConversationResponse, error) {
	cMemberRequest := make([]*domain.ConversationMembers, 0)

	for _, conversation := range conversations {
		cMember, err := CreateConversationMember(userID, conversation.ID.String, userGroup)
		if err != nil {
			return nil, fmt.Errorf("error create conversation member")
		}

		cMemberRequest = append(cMemberRequest, cMember)
	}

	if err := c.ConversationMemberRepo.BulkUpsert(ctx, c.DB, cMemberRequest); err != nil {
		return nil, err
	}

	// publish to conversation evt

	for _, conv := range conversations {
		evt := tpb.ConversationInternal{
			TriggeredAt: timestamppb.Now(),
			Message: &tpb.ConversationInternal_MemberAdded{
				MemberAdded: &tpb.ConversationInternal_MemberAddedToConversation{
					ConversationId: conv.ID.String,
					MemberId:       userID,
				},
			},
		}
		bs, err := proto.Marshal(&evt)
		if err != nil {
			c.Logger.Warn("proto.Marshal", zap.Error(err))
			return nil, err
		}
		_, err = c.JSM.TracedPublish(ctx, "JoinAllConversations.TracedPublish", constants.SubjectChatMembersUpdated, bs)
		if err != nil {
			c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
		}
	}

	sendMsgReqs := make([]*pb.SendMessageRequest, 0, len(conversations))
	for _, conversation := range conversations {
		sendMsgReqs = append(sendMsgReqs, &pb.SendMessageRequest{
			ConversationId: conversation.ID.String,
			Message:        pb.CODES_MESSAGE_TYPE_JOINED_CONVERSATION.String(),
			Type:           pb.MESSAGE_TYPE_SYSTEM,
		})
	}
	err := c.ChatService.SendMessageToConversations(ctx, sendMsgReqs, domain.MessageToConversationOpts{Persist: true, AsUser: false})
	if err != nil {
		c.Logger.Warn("SendMessageToConvresations error", zap.Error(err))
	}

	return &tpb.JoinAllConversationResponse{}, nil
}

func (c *ChatModifier) JoinAllConversations(ctx context.Context, req *tpb.JoinAllConversationRequest) (*tpb.JoinAllConversationResponse, error) {
	userGroup, userID, schoolIDs := interceptors.GetUserInfoFromContext(ctx)
	if len(schoolIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no school ids")
	}
	if !utils.IsStaff(userGroup) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("ChatModifierService.JoinAllConversations: user do not have permission with role %s", userGroup))
	}
	pgIds := database.TextArray(schoolIDs)

	conversations, err := c.ConversationRepo.ListConversationUnjoined(ctx, c.DB, &domain.ListConversationUnjoinedFilter{UserID: database.Text(userID), OwnerIDs: pgIds})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationRepo.ListConversationUnjoined: %w", err).Error())
	}

	res, err := c.makeUserJoinConversations(ctx, conversations, userID, userGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("makeUserJoinConversationsd: %w", err).Error())
	}

	return res, nil
}

// TODO: move to core
func (c *ChatModifier) RemoveMembership(ctx context.Context, db database.QueryExecer, convIDs []string, userID string) (affectedConversations []string, err error) {
	existMembership, err := c.ConversationMemberRepo.FindByCIDsAndUserID(ctx, db, database.TextArray(convIDs), database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("c.ConversationMemberRepo.FindByCIDsAndUserID: %w", err)
	}
	deactivatedMembership := make([]*domain.ConversationMembers, 0, len(convIDs))
	affectedConversations = make([]string, 0, len(convIDs))
	for idx := range existMembership {
		err := existMembership[idx].Status.Set(domain.ConversationStatusInActive)
		if err != nil {
			return nil, fmt.Errorf("existingMembership[idx].Status.Set: %w", err)
		}
		deactivatedMembership = append(deactivatedMembership, existMembership[idx])
		affectedConversations = append(affectedConversations, existMembership[idx].ConversationID.String)
	}
	err = c.ConversationMemberRepo.BulkUpsert(ctx, db, deactivatedMembership)
	if err != nil {
		return nil, fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %w", err)
	}
	return
}

func (c *ChatModifier) LeaveConversations(ctx context.Context, req *tpb.LeaveConversationsRequest) (*tpb.LeaveConversationsResponse, error) {
	if len(req.ConversationIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty conversation ids")
	}

	userGroup, userID, _ := interceptors.GetUserInfoFromContext(ctx)
	if !utils.IsStaff(userGroup) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("user do not have permission with role %s to leave conversations", userGroup))
	}

	var notifiedConversations []string
	err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		affectedConversations, txerr := c.RemoveMembership(ctx, tx, req.GetConversationIds(), userID)
		if txerr != nil {
			return txerr
		}
		notifiedConversations = affectedConversations
		return nil
	})
	if err != nil {
		return &tpb.LeaveConversationsResponse{}, fmt.Errorf("ExecInTx: %w", err)
	}
	for _, id := range notifiedConversations {
		event := &tpb.ConversationInternal{
			TriggeredAt: timestamppb.Now(),
			Message: &tpb.ConversationInternal_MemberRemoved{
				MemberRemoved: &tpb.ConversationInternal_MemberRemovedFromConversation{
					ConversationId: id,
					MemberId:       userID,
				},
			},
		}
		bs, err := proto.Marshal(event)
		if err != nil {
			c.Logger.Warn("proto.Marshal", zap.Error(err))
			return nil, err
		}
		_, err = c.JSM.TracedPublish(ctx, "LeaveConversation.TracedPublish", constants.SubjectChatMembersUpdated, bs)
		if err != nil {
			c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
			return nil, err
		}
	}

	logger := ctxzap.Extract(ctx)
	for _, cID := range notifiedConversations {
		// also notify other clients of this user that he has left conversations, in case of multiple tabs/devices
		msgpb, err := c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
			ConversationId: cID,
			Message:        pb.CODES_MESSAGE_TYPE_LEAVE_CONVERSATION.String(),
			Type:           pb.MESSAGE_TYPE_SYSTEM,
			TargetUser:     userID,
		}, domain.MessageToConversationOpts{Persist: true, AsUser: false})

		if err != nil {
			logger.Error("c.ChatService.SendMessageToConversation:", zap.Error(err))
		}
		// push message to clients of this user
		err = c.ChatService.SendMessageToUsers(ctx, []string{userID}, &pb.Event{
			Event: &pb.Event_EventNewMessage{
				EventNewMessage: msgpb,
			},
		}, domain.MessageToUserOpts{Notification: domain.NotificationOpts{}})
		if err != nil {
			logger.Error("c.ChatService.PushMessage", zap.Error(err))
		}
	}
	return &tpb.LeaveConversationsResponse{}, nil
}

func (c *ChatModifier) HandleUpsertStaff(ctx context.Context, msg *upb.EvtUpsertStaff) (bool, error) {
	staffID := msg.StaffId

	if staffID == "" {
		return false, errors.New("no staffID found")
	}

	conversations, err := c.ConversationStudentRepo.FindByStaffIDs(ctx, c.DB, database.TextArray([]string{staffID}))
	if err != nil {
		c.Logger.Error("c.ConversationStudentRepo.FindByStaffIDs", zap.Error(err))
		return true, err
	}

	if len(conversations) == 0 {
		c.Logger.Info("no conversations found")

		// do not retry here
		// because user has not join any conversation
		return false, nil
	}

	conversationLocationsMap, err := c.ConversationLocationRepo.FindByConversationIDs(ctx, c.DB, database.TextArray(conversations))
	if err != nil {
		c.Logger.Error("ConversationLocationRepo.FindByConversationIDs", zap.Error(err))
		return true, err
	}

	conversationLocationIDsMap := extractConversationLocationIDsMap(conversationLocationsMap)

	grantedLocationsMap, err := c.GrantedPermissionRepo.FindByUserIDAndPermissionName(ctx, c.DB, database.TextArray([]string{staffID}), database.Text("master.location.read"))
	if err != nil {
		c.Logger.Error("GrantedPermissionRepo.FindByUserIDAndPermissionName", zap.Error(err))
		// allow retry as we need to wait for granted permission data to finish sync into our db
		return true, err
	}

	grantedLocationIDsMap := extractGrantedLocationIDsMap(grantedLocationsMap)

	inactivatingStaffConversationIDs := getInactivatingConversations(grantedLocationIDsMap, conversationLocationIDsMap, conversations, staffID)
	if len(inactivatingStaffConversationIDs) == 0 {
		return true, errors.New("no inactivating conversations found")
	}

	err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.ConversationMemberRepo.SetStatusByConversationAndUserIDs(ctx, tx, database.TextArray(inactivatingStaffConversationIDs), database.TextArray([]string{staffID}), database.Text(domain.ConversationStatusInActive))
		if err != nil {
			return fmt.Errorf("ConversationMemberRepo.SetStatusByConversationAndUserIDs: %w", err)
		}

		return nil
	})
	if err != nil {
		return true, fmt.Errorf("ExecInTx: %w", err)
	}

	event := &tpb.ConversationInternal{
		TriggeredAt: timestamppb.Now(),
		Message: &tpb.ConversationInternal_ConversationsUpdated_{
			ConversationsUpdated: &tpb.ConversationInternal_ConversationsUpdated{
				ConversationIds: inactivatingStaffConversationIDs,
			},
		},
	}

	bs, err := proto.Marshal(event)
	if err != nil {
		c.Logger.Warn("proto.Marshal", zap.Error(err))
		return false, err
	}
	_, err = c.JSM.TracedPublish(ctx, "HandleUpsertStaff.TracedPublish", constants.SubjectChatMembersUpdated, bs)
	if err != nil {
		c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
		return false, err
	}

	for _, conversationID := range inactivatingStaffConversationIDs {
		// also notify other clients of this user that he has left conversations, in case of multiple tabs/devices
		msgpb, err := c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
			ConversationId: conversationID,
			Message:        pb.CODES_MESSAGE_TYPE_LEAVE_CONVERSATION.String(),
			Type:           pb.MESSAGE_TYPE_SYSTEM,
			TargetUser:     staffID,
		}, domain.MessageToConversationOpts{Persist: true, AsUser: false})

		if err != nil {
			c.Logger.Error("c.ChatService.SendMessageToConversation:", zap.Error(err))
		}
		// push message to clients of this user
		err = c.ChatService.SendMessageToUsers(ctx, []string{staffID}, &pb.Event{
			Event: &pb.Event_EventNewMessage{
				EventNewMessage: msgpb,
			},
		}, domain.MessageToUserOpts{Notification: domain.NotificationOpts{}})
		if err != nil {
			c.Logger.Error("c.ChatService.PushMessage", zap.Error(err))
		}
	}
	return false, nil
}

func (c *ChatModifier) HandleUpsertUserGroup(ctx context.Context, msg *upb.EvtUpsertUserGroup) (bool, error) {
	userGroupID := msg.GetUserGroupId()

	if userGroupID == "" {
		return false, errors.New("no userGroupID found")
	}

	userGroupMemberIDs, err := c.UserGroupMemberRepo.FindUserIDsByUserGroupID(ctx, c.DB, database.Text(userGroupID))
	if err != nil {
		c.Logger.Error("c.UserGroupMemberRepo.FindUserIDsByUserGroupID", zap.Error(err))
		return true, err
	}

	if len(userGroupMemberIDs) == 0 {
		c.Logger.Info(fmt.Sprintf("no user group members found with ID %s", userGroupID))
		// retry because we need to wait for sync data into our db
		return true, fmt.Errorf("no user group members found with ID %s", userGroupID)
	}

	conversationMembersMap, err := c.ConversationMemberRepo.FindUserIDConversationIDsMapByUserIDs(ctx, c.DB, database.TextArray(userGroupMemberIDs))
	if err != nil {
		c.Logger.Error("c.ConversationMemberRepo.FindUserIDConversationIDsMapByUserIDs", zap.Error(err))
		return true, err
	}

	conversationIDs := make([]string, 0)
	for _, convIDs := range conversationMembersMap {
		conversationIDs = append(conversationIDs, convIDs...)
	}

	conversationIDs = golibs.GetUniqueElementStringArray(conversationIDs)

	conversationLocationsMap, err := c.ConversationLocationRepo.FindByConversationIDs(ctx, c.DB, database.TextArray(conversationIDs))
	if err != nil {
		c.Logger.Error("c.ConversationLocationRepo.FindByConversationIDs", zap.Error(err))
		return true, err
	}

	conversationLocationIDsMap := extractConversationLocationIDsMap(conversationLocationsMap)

	grantedPermissionsMap, err := c.GrantedPermissionRepo.FindByUserIDAndPermissionName(ctx, c.DB, database.TextArray(userGroupMemberIDs), database.Text("master.location.read"))
	if err != nil {
		c.Logger.Error("c.GrantedPermissionRepo.FindByUserIDAndPermissionName", zap.Error(err))
		return true, err
	}

	grantedLocationIDsMap := extractGrantedLocationIDsMap(grantedPermissionsMap)

	inactiveConversationMembersMap := make(map[string][]string, 0)
	err = database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		for userID, convIDs := range conversationMembersMap {
			inactivatingStaffConversationIDs := getInactivatingConversations(grantedLocationIDsMap, conversationLocationIDsMap, convIDs, userID)

			if len(inactivatingStaffConversationIDs) == 0 {
				return fmt.Errorf("retrying to find InactiveConversations")
			}

			err := c.ConversationMemberRepo.SetStatusByConversationAndUserIDs(ctx, tx, database.TextArray(inactivatingStaffConversationIDs), database.TextArray([]string{userID}), database.Text(domain.ConversationStatusInActive))
			if err != nil {
				return fmt.Errorf("ConversationMemberRepo.SetStatusByConversationAndUserIDs: %w", err)
			}

			inactiveConversationMembersMap[userID] = inactivatingStaffConversationIDs
		}
		return nil
	})
	if err != nil {
		return true, fmt.Errorf("ExecInTx: %w", err)
	}

	for userID, inactiveConversationIDs := range inactiveConversationMembersMap {
		event := &tpb.ConversationInternal{
			TriggeredAt: timestamppb.Now(),
			Message: &tpb.ConversationInternal_ConversationsUpdated_{
				ConversationsUpdated: &tpb.ConversationInternal_ConversationsUpdated{
					ConversationIds: inactiveConversationIDs,
				},
			},
		}

		bs, err := proto.Marshal(event)
		if err != nil {
			c.Logger.Warn("proto.Marshal", zap.Error(err))
			return false, err
		}
		_, err = c.JSM.TracedPublish(ctx, "HandleUpsertUserGroup.TracedPublish", constants.SubjectChatMembersUpdated, bs)
		if err != nil {
			c.Logger.Warn("c.JSM.TracedPublish", zap.Error(err))
			return false, err
		}

		for _, conversationID := range inactiveConversationIDs {
			// also notify other clients of this user that he has left conversations, in case of multiple tabs/devices
			msgpb, err := c.ChatService.SendMessageToConversation(ctx, &pb.SendMessageRequest{
				ConversationId: conversationID,
				Message:        pb.CODES_MESSAGE_TYPE_LEAVE_CONVERSATION.String(),
				Type:           pb.MESSAGE_TYPE_SYSTEM,
				TargetUser:     userID,
			}, domain.MessageToConversationOpts{Persist: true, AsUser: false})

			if err != nil {
				c.Logger.Error("c.ChatService.SendMessageToConversation:", zap.Error(err))
			}
			// push message to clients of this user
			err = c.ChatService.SendMessageToUsers(ctx, []string{userID}, &pb.Event{
				Event: &pb.Event_EventNewMessage{
					EventNewMessage: msgpb,
				},
			}, domain.MessageToUserOpts{Notification: domain.NotificationOpts{}})
			if err != nil {
				c.Logger.Error("c.ChatService.PushMessage", zap.Error(err))
			}
		}
	}

	return false, nil
}

func extractConversationLocationIDsMap(conversationLocationsMap map[string][]domain.ConversationLocation) map[string][]string {
	conversationLocationIDsMap := make(map[string][]string, 0)
	for conversationID, conversationLocationEnts := range conversationLocationsMap {
		for _, locationEnt := range conversationLocationEnts {
			conversationLocationIDsMap[conversationID] = append(conversationLocationIDsMap[conversationID], locationEnt.LocationID.String)
		}
	}
	return conversationLocationIDsMap
}

func extractGrantedLocationIDsMap(grantedPermissionsMap map[string][]*domain.GrantedPermission) map[string][]string {
	grantedLocationIDsMap := make(map[string][]string, 0)
	for userID, grantedPermissions := range grantedPermissionsMap {
		for _, permission := range grantedPermissions {
			grantedLocationIDsMap[userID] = append(grantedLocationIDsMap[userID], permission.LocationID.String)
		}
	}
	return grantedLocationIDsMap
}

func getInactivatingConversations(grantedLocationIDsMap map[string][]string, conversationLocationIDsMap map[string][]string, conversationIDs []string, userID string) []string {
	inactivatingStaffConversationIDs := make([]string, 0)

	grantedLocations, existed := grantedLocationIDsMap[userID]
	if !existed {
		inactivatingStaffConversationIDs = append(inactivatingStaffConversationIDs, conversationIDs...)
	} else {
		for _, convID := range conversationIDs {
			convLocationIDs, existed := conversationLocationIDsMap[convID]
			if !existed {
				continue
			}

			matchedLocations := sliceutils.Intersect(convLocationIDs, grantedLocations)
			if len(matchedLocations) == 0 {
				inactivatingStaffConversationIDs = append(inactivatingStaffConversationIDs, convID)
			}
		}
	}

	return inactivatingStaffConversationIDs
}
