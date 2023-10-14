package notificationmgmt

import (
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/notification/services"
	"github.com/manabie-com/backend/internal/notification/subscribers"
	noti_nats "github.com/manabie-com/backend/internal/notification/transports/nats"

	"go.uber.org/zap"
)

func initEventPushNotification(jsm nats.JetStreamManagement, logger *zap.Logger, notiSubscriber *subscribers.NotificationSubscriber) {
	pushNotiEvent := noti_nats.NewPushNotificationEvent(jsm, logger, notiSubscriber)
	err := pushNotiEvent.StartSubscribe()
	if err != nil {
		logger.Panic("pushNotiEvent.StartSubscribe", zap.Error(err))
	}
}

func initEventSyncStudentPackage(jsm nats.JetStreamManagement, logger *zap.Logger, notiService *services.NotificationModifierService) {
	notificationSyncStudentPackage := &noti_nats.NotificationSyncStudentPackage{
		JSM:                         jsm,
		Logger:                      logger,
		NotificationModifierService: notiService,
	}
	err := notificationSyncStudentPackage.StartToSubscribe()

	if err != nil {
		logger.Panic("subscriptions.NotificationSyncStudentPackage.Subscribe: %v", zap.Error(err))
	}
}

func initEventSyncJprepStudentPackage(jsm nats.JetStreamManagement, logger *zap.Logger, notiService *services.NotificationModifierService) {
	notificationSyncJprepStudentPackage := &noti_nats.JprepSyncStudentPackage{
		JSM:                         jsm,
		Logger:                      logger,
		NotificationModifierService: notiService,
	}

	err := notificationSyncJprepStudentPackage.StartToSubscribe()
	if err != nil {
		logger.Panic("subscriptions.NotificationSyncJprepStudentCourse.Subscribe: %v", zap.Error(err))
	}
}

func initEventSyncJprepStudentClass(jsm nats.JetStreamManagement, logger *zap.Logger, notiService *services.NotificationModifierService) {
	notificationSyncJprepStudentClass := &noti_nats.JprepSyncClassMember{
		JSM:                         jsm,
		Logger:                      logger,
		NotificationModifierService: notiService,
	}

	err := notificationSyncJprepStudentClass.StartToSubscribe()
	if err != nil {
		logger.Panic("subscriptions.initEventSyncJprepStudentClass.Subscribe: %v", zap.Error(err))
	}
}
