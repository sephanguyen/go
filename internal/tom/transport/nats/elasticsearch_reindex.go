package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/app/support"
	"github.com/manabie-com/backend/internal/tom/configurations"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ElasticsearchReindexSubscription struct {
	Config  *configurations.Config
	Logger  *zap.Logger
	Indexer *support.SearchIndexer
	JSM     nats.JetStreamManagement
}

// spliting out to multiple subjects so that stampede on update does not block on new index
func (j *ElasticsearchReindexSubscription) SubscribeConversationInternal() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamChat, constants.DurableElasticChatCreated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverElasticChatCreated),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "ElasticChatCreated",
	}

	_, err := j.JSM.QueueSubscribe(constants.SubjectChatCreated, constants.QueueElasticChatCreated, opts,
		j.handleConversationInternal)
	if err != nil {
		return fmt.Errorf("QueueSubscribe %s %w", constants.QueueElasticChatCreated, err)
	}
	opts = nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamChat, constants.DurableElasticChatUpdated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverElasticChatUpdated),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "ConsumerElasticChatUpdated",
	}

	_, err = j.JSM.QueueSubscribe(constants.SubjectChatUpdated, constants.QueueElasticChatUpdated, opts,
		j.handleConversationInternal)
	if err != nil {
		return fmt.Errorf("QueueSubscribe %s %w", constants.QueueElasticChatUpdated, err)
	}

	opts = nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamChat, constants.DurableElasticChatMembersUpdated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverElasticChatMembersUpdated),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "ElasticChatMembersUpdated",
	}

	_, err = j.JSM.QueueSubscribe(constants.SubjectChatMembersUpdated, constants.QueueElasticChatMembersUpdated, opts,
		j.handleConversationInternal)
	if err != nil {
		return fmt.Errorf("ElasticsearchReindexSubscription.JSM.QueueSubscribe: %v", err)
	}

	opts = nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamChat, constants.DurableElasticChatMessageCreated),
			nats.DeliverSubject(constants.DeliverElasticChatMessageCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "ElasticChatMessageCreated",
	}

	_, err = j.JSM.QueueSubscribe(constants.SubjectChatMessageCreated, constants.QueueElasticChatMessageCreated, opts,
		j.handleConversationInternal)
	if err != nil {
		return fmt.Errorf("ElasticsearchReindexSubscription.JSM.QueueSubscribe: %v", err)
	}

	return nil
}

// subscribe to student course to get which course belong to the student (students are a part of member of conversation)
func (j *ElasticsearchReindexSubscription) SubscribeCourseStudent() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.Bind(constants.StreamESConversation, constants.DurableCourseStudentEventNats),
			nats.DeliverSubject(constants.DeliverCourseStudentEventNats),
		},
		SpanName: "SubscribeCourseStudent.handleCourseStudentEvt",
	}

	_, err := j.JSM.QueueSubscribe(constants.SubjectCourseStudentEventNats, constants.QueueCourseStudentEventNats,
		option, j.handleCourseStudentEvt)
	if err != nil {
		return fmt.Errorf("ElasticsearchReindexSubscription.SubscribeCourseStudent: %w", err)
	}
	return nil
}

func (j *ElasticsearchReindexSubscription) handleConversationInternal(ctx context.Context, rawMsg []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var (
		req tpb.ConversationInternal
		err error
	)
	err = proto.Unmarshal(rawMsg, &req)
	if err != nil {
		j.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}
	switch req.Message.(type) {
	case *tpb.ConversationInternal_ConversationCreated_:
		return true, j.Indexer.ReindexConversationDocument(ctx, req.GetTriggeredAt().AsTime(), []string{req.GetConversationCreated().GetConversationId()})
	case *tpb.ConversationInternal_ConversationUpdated_:
		return true, j.Indexer.ReindexConversationDocument(ctx, req.GetTriggeredAt().AsTime(), []string{req.GetConversationUpdated().GetConversationId()})
	case *tpb.ConversationInternal_MemberRemoved:
		return true, j.Indexer.ReindexConversationDocument(ctx, req.GetTriggeredAt().AsTime(), []string{req.GetMemberRemoved().GetConversationId()})
	case *tpb.ConversationInternal_MessageSent:
		return true, j.Indexer.ReindexConversationDocument(ctx, req.GetTriggeredAt().AsTime(), []string{req.GetMessageSent().GetConversationId()})
	case *tpb.ConversationInternal_MemberAdded:
		return true, j.Indexer.ReindexConversationDocument(ctx, req.GetTriggeredAt().AsTime(), []string{req.GetMemberAdded().GetConversationId()})
	case *tpb.ConversationInternal_ConversationsUpdated_:
		return true, j.Indexer.ReindexConversationDocument(ctx, req.TriggeredAt.AsTime(), req.GetConversationsUpdated().GetConversationIds())
	default:
		return false, fmt.Errorf("invalid type for conversation internal message: %T", req.Message)
	}
}

func (j *ElasticsearchReindexSubscription) handleCourseStudentEvt(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var (
		req npb.EventCourseStudent
		err error
	)
	if err = proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("handleCourseStudentEvt proto.Unmarshal: %w", err)
	}

	if len(req.StudentIds) == 0 {
		return false, fmt.Errorf("handleCourseStudentEvt length of StudentIds = 0")
	}

	err = nats.ChunkHandler(len(req.GetStudentIds()), constants.MaxRecordProcessPertime, func(start, end int) error {
		err := j.Indexer.HandleStudentCourseUpdated(ctx, &npb.EventCourseStudent{
			StudentIds: req.GetStudentIds()[start:end],
		})
		return err
	})

	if err != nil {
		return true, fmt.Errorf("handleCourseStudentEvt err HandleStudentCourseUpdated: %w", err)
	}

	return false, nil
}

// Deprecated
func (j *ElasticsearchReindexSubscription) handleConversationEvt(ctx context.Context, rawMsg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var (
		req npb.EventConversation
		err error
	)
	err = proto.Unmarshal(rawMsg, &req)
	if err != nil {
		j.Logger.Error("proto.Unmarshal", zap.Error(err))
		return err
	}

	if len(req.ConversationIds) == 0 && len(req.GetUserIds()) == 0 {
		return nil
	}

	if len(req.GetConversationIds()) != 0 {
		err = nats.ChunkHandler(len(req.GetConversationIds()), constants.MaxRecordProcessPertime, func(start, end int) error {
			_, err := j.Indexer.BuildConversationDocument(ctx, &tpb.BuildConversationDocumentRequest{
				ConversationIds: req.GetConversationIds()[start:end],
			})
			return err
		})
	} else if len(req.GetUserIds()) != 0 {
		err = nats.ChunkHandler(len(req.GetUserIds()), constants.MaxRecordProcessPertime, func(start, end int) error {
			_, err = j.Indexer.BuildConversationDocument(ctx, &tpb.BuildConversationDocumentRequest{
				UserIds: req.GetUserIds()[start:end],
			})
			return err
		})
	}
	if err != nil {
		j.Logger.Error("err handleConversationEvt", zap.Error(err))
		return err
	}
	return nil
}
