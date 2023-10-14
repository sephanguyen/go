package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/notification/infra/metrics"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/util/validation"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpsertSystemNotificationConsumerHandler struct {
	Logger *zap.Logger
	DB     database.Ext

	SystemNotificationCommandHandler interface {
		UpsertSystemNotification(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationPayload) error
	}
	SystemNotificationRecipientCommandHandler interface {
		UpsertSystemNotificationRecipients(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationRecipientPayload) error
	}
	SystemNotificationContentCommandHandler interface {
		UpsertSystemNotificationContents(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationContentPayload) error
	}
	SoftDeleteSystemNotificationCommandHander interface {
		SoftDeleteSystemNotification(ctx context.Context, db database.QueryExecer, softDeletePayload *payloads.SoftDeleteSystemNotificationPayload) error
	}

	metrics.NotificationMetrics
}

func (s *UpsertSystemNotificationConsumerHandler) Handle(ctx context.Context, value []byte) (bool, error) {
	err := validation.ValidateKafkaContext(ctx)
	if err != nil {
		return false, fmt.Errorf("ValidateKafkaContext: %+v", err)
	}

	kafkaPayload := &payload.UpsertSystemNotification{}
	err = json.Unmarshal(value, kafkaPayload)
	if err != nil {
		return false, fmt.Errorf("failed UpsertSystemNotificationConsumerHandler Unmarshal: %+v", err)
	}

	systemNotificationDTO := systemnotification.KafkaPayloadToDTO(kafkaPayload)

	if systemNotificationDTO.IsDeleted {
		err = s.SoftDeleteSystemNotification(ctx, systemNotificationDTO)
		if err != nil {
			return false, fmt.Errorf("failed SoftDeleteSystemNotification: %+v", err)
		}
	} else {
		err = s.UpsertSystemNotification(ctx, systemNotificationDTO)
		if err != nil {
			s.NotificationMetrics.RecordSystemNotificationError(1)
			return false, fmt.Errorf("failed UpsertSystemNotification: %+v", err)
		}
		s.NotificationMetrics.RecordSystemNotificationCreated(1)
	}

	return false, nil
}

// logic to handle kafka consume messages
func (s *UpsertSystemNotificationConsumerHandler) UpsertSystemNotification(ctx context.Context, systemNotification *dto.SystemNotification) error {
	if err := validation.ValidateSystemNotificationRequiredFields(systemNotification); err != nil {
		// log error and return but still commit kafka message
		s.Logger.Error("UpsertSystemNotification message failed validation with: %+v", zap.Error(err))
		return nil
	}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		systemNotificationPayload := &payloads.UpsertSystemNotificationPayload{
			SystemNotification: systemNotification,
		}
		err := s.SystemNotificationCommandHandler.UpsertSystemNotification(ctx, tx, systemNotificationPayload)
		if err != nil {
			return err
		}

		snID := systemNotificationPayload.SystemNotification.SystemNotificationID

		err = s.SystemNotificationContentCommandHandler.UpsertSystemNotificationContents(ctx, tx, &payloads.UpsertSystemNotificationContentPayload{
			SystemNotificationID:       snID,
			SystemNotificationContents: systemNotification.Content,
		})
		if err != nil {
			return err
		}

		err = s.SystemNotificationRecipientCommandHandler.UpsertSystemNotificationRecipients(ctx, tx, &payloads.UpsertSystemNotificationRecipientPayload{
			SystemNotificationID: snID,
			Recipients:           systemNotification.Recipients,
		})

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *UpsertSystemNotificationConsumerHandler) SoftDeleteSystemNotification(ctx context.Context, systemNotification *dto.SystemNotification) error {
	if err := validation.ValidateSystemNotificationRequiredFields(systemNotification); err != nil {
		// log error and return but still commit kafka message
		s.Logger.Error("SoftDeleteSystemNotification message failed validation with: %+v", zap.Error(err))
		return nil
	}
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		softDeletePayload := &payloads.SoftDeleteSystemNotificationPayload{
			ReferenceID: systemNotification.ReferenceID,
		}
		err := s.SoftDeleteSystemNotificationCommandHander.SoftDeleteSystemNotification(ctx, tx, softDeletePayload)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
