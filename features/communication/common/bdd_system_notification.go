package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func createSystemNotificationKafkaPayload(ctx context.Context) *payload.UpsertSystemNotification {
	stepState := StepStateFromContext(ctx)

	payloadRecipients := []payload.SystemNotificationRecipient{}
	payloadRecipients = append(payloadRecipients, payload.SystemNotificationRecipient{
		UserID: stepState.CurrentStaff.ID,
	})

	payloadContents := []payload.SystemNotificationContent{}
	payloadContents = append(payloadContents, payload.SystemNotificationContent{
		Language: "en",
		Text:     idutil.ULIDNow(),
	})
	payloadContents = append(payloadContents, payload.SystemNotificationContent{
		Language: "jp",
		Text:     idutil.ULIDNow(),
	})

	referenceID := idutil.ULIDNow()
	kafkaPayload := &payload.UpsertSystemNotification{
		ReferenceID: referenceID,
		Content:     payloadContents,
		URL:         referenceID,
		ValidFrom:   time.Now().Add(time.Hour * time.Duration(-1)),
		Recipients:  payloadRecipients,
		Status:      payload.SystemNotificationStatusNew,
	}
	return kafkaPayload
}

func (s *NotificationSuite) CreateNumberOfSystemNotificationWithSomeStatus(ctx context.Context, numNew, numDone, numUnenabled string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numSystemNotificationNew := StrToInt(numNew)
	numSystemNotificationDone := StrToInt(numDone)
	numSystemNotificationUnenabled := StrToInt(numUnenabled)

	for i := 0; i < numSystemNotificationUnenabled; i++ {
		kafkaPayload := createSystemNotificationKafkaPayload(ctx)
		kafkaPayload.ValidFrom = time.Now().Add(time.Hour * 1)

		data, err := json.Marshal(kafkaPayload)
		if err != nil {
			return ctx, fmt.Errorf("failed to marshal kafkaPayload: %+v", err)
		}
		err = s.PublishToKafka(ctx, constants.SystemNotificationUpsertingTopic, data)
		if err != nil {
			return ctx, fmt.Errorf("failed to PublishKafka UpsertSystemNotification: %+v", err)
		}
	}

	payloadSystemNotifications := []*payload.UpsertSystemNotification{}
	for i := 0; i < numSystemNotificationNew; i++ {
		kafkaPayload := createSystemNotificationKafkaPayload(ctx)
		kafkaPayload.Status = payload.SystemNotificationStatusNew

		payloadSystemNotifications = append(payloadSystemNotifications, kafkaPayload)
		data, err := json.Marshal(kafkaPayload)
		if err != nil {
			return ctx, fmt.Errorf("failed to marshal kafkaPayload: %+v", err)
		}
		err = s.PublishToKafka(ctx, constants.SystemNotificationUpsertingTopic, data)
		if err != nil {
			return ctx, fmt.Errorf("failed to PublishKafka UpsertSystemNotification: %+v", err)
		}
	}
	for i := 0; i < numSystemNotificationDone; i++ {
		kafkaPayload := createSystemNotificationKafkaPayload(ctx)
		kafkaPayload.Status = payload.SystemNotificationStatusDone

		payloadSystemNotifications = append(payloadSystemNotifications, kafkaPayload)
		data, err := json.Marshal(kafkaPayload)
		if err != nil {
			return ctx, fmt.Errorf("failed to marshal kafkaPayload: %+v", err)
		}
		err = s.PublishToKafka(ctx, constants.SystemNotificationUpsertingTopic, data)
		if err != nil {
			return ctx, fmt.Errorf("failed to PublishKafka UpsertSystemNotification: %+v", err)
		}
	}
	stepState.PayloadSystemNotifications = payloadSystemNotifications
	stepState.TokenOfSentRecipient = stepState.CurrentStaff.Token
	return StepStateToContext(ctx, stepState), nil
}

// for now we only return 1 ID each call
func getRandomlyRecipientIDsFromStaffs(staff []*entities.Staff) (string, string) {
	i := RandRangeIn(1, 3)
	return staff[i].ID, staff[i].Token
}

func (s *NotificationSuite) GetSystemNotificationByUser(userToken string) ([]*npb.RetrieveSystemNotificationsResponse_SystemNotification, error) {
	ctx, cancel := ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()
	res, err := npb.NewSystemNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveSystemNotifications(ctx,
		&npb.RetrieveSystemNotificationsRequest{
			Paging: &cpb.Paging{
				Limit: 1000,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed RetrieveSystemNotifications: %+v", err)
	}

	return res.GetSystemNotifications(), nil
}
