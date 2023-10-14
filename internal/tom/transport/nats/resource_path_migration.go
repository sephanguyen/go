package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/configurations"
	"github.com/manabie-com/backend/internal/tom/infra/migration"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ResourcePathMigration struct {
	Config *configurations.Config
	JSM    nats.JetStreamManagement
	Logger *zap.Logger

	Migrator *migration.ResourcePathMigrator
}

func (rcv *ResourcePathMigration) Subscribe() error {
	durable, queuename, deliversubj := generateConsumerMetadata(constants.ChatMigrateResourcePathConsumerKey)
	otpsResourcePathSub := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamChatMigration, durable),
			nats.DeliverSubject(deliversubj),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}
	_, err := rcv.JSM.QueueSubscribe(constants.SubjectChatMigrateResourcePath, queuename, otpsResourcePathSub, rcv.handleMigrationMsges)
	if err != nil {
		return fmt.Errorf("subLesson.QueueSubscribe: %w", err)
	}
	return nil
}

// type MsgHandler func(ctx context.Context, data []byte) (bool, error)
func (rcv *ResourcePathMigration) handleMigrationMsges(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var msg = tpb.ResourcePathMigration{}
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		rcv.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}
	switch msg.MessageType.(type) {
	case *tpb.ResourcePathMigration_Users_:
		users := msg.GetUsers()
		err = rcv.Migrator.MigrateUser(ctx, users)
		if err != nil {
			rcv.Logger.Error("rcv.Migrator.MigrateUser", zap.Error(err))
			return true, err
		}
	case *tpb.ResourcePathMigration_Lessons_:
		lessons := msg.GetLessons()
		err = rcv.Migrator.MigrateLesson(ctx, lessons)
		if err != nil {
			rcv.Logger.Error("rcv.Migrator.MigrateLesson", zap.Error(err))
			return true, err
		}
	default:
		err = fmt.Errorf("invalid message type: %T", msg.MessageType)
		rcv.Logger.Error("invalid payload", zap.Error(err))
		return false, err
	}
	return false, nil
}
