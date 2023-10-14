package kafka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/infra/metrics"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/consumers"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"

	"go.uber.org/zap"
)

type ConsumerHandler interface {
	Handle(ctx context.Context, msg []byte) (bool, error)
}

type UpsertSystemNotificationConsumer struct {
	Logger    *zap.Logger
	KafkaMgmt kafka.KafkaManagement

	ConsumerHandler
}

func (s *UpsertSystemNotificationConsumer) Consume() error {
	consumerGroupID := s.KafkaMgmt.GenNewConsumerGroupID(consts.ServiceName, constants.SystemNotificationUpsertingTopic)
	opts := kafka.Option{
		SpanName: "CONSUMER." + constants.SystemNotificationUpsertingTopic,
	}
	err := s.KafkaMgmt.Consume(constants.SystemNotificationUpsertingTopic, consumerGroupID, opts, s.Handle)
	if err != nil {
		fmt.Printf("Error UpsertSystemNotificationConsumer consuming messages: %v\n", err)
	}
	return nil
}

func NewUpsertSystemNotificationConsumer(db database.Ext, kafka kafka.KafkaManagement, l *zap.Logger, metrics metrics.NotificationMetrics) *UpsertSystemNotificationConsumer {
	handler := &consumers.UpsertSystemNotificationConsumerHandler{
		Logger: l,
		DB:     db,
		SystemNotificationCommandHandler: &commands.SystemNotificationCommandHandler{
			SystemNotificationRepo: &repo.SystemNotificationRepo{},
		},
		SystemNotificationRecipientCommandHandler: &commands.SystemNotificationRecipientCommandHandler{
			SystemNotificationRecipientRepo: &repo.SystemNotificationRecipientRepo{},
		},
		SystemNotificationContentCommandHandler: &commands.SystemNotificationContentHandler{
			SystemNotificationContentRepo: &repo.SystemNotificationContentRepo{},
		},
		SoftDeleteSystemNotificationCommandHander: &commands.DeleteSystemNotificationCommandHandler{
			SystemNotificationRepo:          &repo.SystemNotificationRepo{},
			SystemNotificationRecipientRepo: &repo.SystemNotificationRecipientRepo{},
			SystemNotificationContentRepo:   &repo.SystemNotificationContentRepo{},
		},
		NotificationMetrics: metrics,
	}
	return &UpsertSystemNotificationConsumer{
		Logger:          l,
		KafkaMgmt:       kafka,
		ConsumerHandler: handler,
	}
}
