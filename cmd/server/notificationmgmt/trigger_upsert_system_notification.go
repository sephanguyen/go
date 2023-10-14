package notificationmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/notification/config"

	"github.com/jackc/pgx/v4/pgxpool"
)

// this job is to trigger a fake UpsertSystemNotification message sent to kafka -> notification
// it receives list of UserIDs as Recipients, Content, URL, ValidFrom, ValidTo
// ResourcePath: required
// ReferenceID: required
// Content: required
// URL: optional
// ValidFrom: required
// UserIDs: required
var (
	pResourcePath string
	pReferenceID  string
	pContent      []byte
	pURL          string
	pValidFrom    string // format: "2023-05-09 14:42:00 +0700"
	pIsDeleted    bool
	pUserIDs      string // format: "user_id1;user_id2;user_id3;..."

	timeLayout = "2006-01-02 15:04:05 -0700"
)

func init() {
	bootstrap.RegisterJob("notificationmgmt_trigger_upsert_system_notification", RunTriggerUpsertSystemNotification).
		StringVar(&pReferenceID, "referenceID", "", "reference ID of the system notification").
		BytesBase64Var(&pContent, "content", []byte(`[]`), "content of the system notification in JSON").
		StringVar(&pURL, "url", "", "url of system notification").
		StringVar(&pValidFrom, "validFrom", "", "the time of the system notification will be enabled").
		StringVar(&pUserIDs, "userIDs", "", "the recipient ids of system notification").
		BoolVar(&pIsDeleted, "isDeleted", false, "the flag to delete the system notification").
		StringVar(&pResourcePath, "resourcePath", "", "resource path of the message")
}

func RunTriggerUpsertSystemNotification(ctx context.Context, _ config.Config, rsc *bootstrap.Resources) error {
	err := ValidateJobInput()
	if err != nil {
		return err
	}

	// get UserID from NotificationScheduledJob correspond with the resource path
	bobDB := rsc.DBWith("bob")
	tenantAndUserCtx, err := makeTenantWithUserCtx(ctx, bobDB.DB.(*pgxpool.Pool), pResourcePath)
	if err != nil {
		return err
	}

	kafkaPayload, err := ToKafkaPayload()
	if err != nil {
		return fmt.Errorf("failed ToKafkaPayload: %+v", err)
	}
	data, err := json.Marshal(kafkaPayload)
	if err != nil {
		return fmt.Errorf("failed marshal kafkaPayload: %+v", err)
	}

	msgKey, err := json.Marshal(pReferenceID)
	if err != nil {
		return fmt.Errorf("failed marshal Message Key: %+v", err)
	}

	kafka := rsc.Kafka()
	err = kafka.PublishContext(tenantAndUserCtx, constants.SystemNotificationUpsertingTopic, msgKey, data)
	if err != nil {
		return fmt.Errorf("failed kafka PublishContext: %+v", err)
	}

	return nil
}

func ToKafkaPayload() (*payload.UpsertSystemNotification, error) {
	snContent := []payload.SystemNotificationContent{}

	err := json.Unmarshal(pContent, &snContent)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal: %+v", err)
	}

	kafkaPayload := &payload.UpsertSystemNotification{
		ReferenceID: pReferenceID,
		Content:     snContent,
		URL:         pURL,
		IsDeleted:   pIsDeleted,
		Status:      payload.SystemNotificationStatusNew,
	}

	userIDs := strings.Split(pUserIDs, ";")

	recipients := []payload.SystemNotificationRecipient{}
	for _, u := range userIDs {
		recipients = append(recipients, payload.SystemNotificationRecipient{
			UserID: u,
		})
	}
	kafkaPayload.Recipients = recipients

	validFrom, err := time.Parse(timeLayout, pValidFrom)
	if err != nil {
		return nil, fmt.Errorf("failed parsing ValidFrom: %+v", err)
	}
	kafkaPayload.ValidFrom = validFrom

	return kafkaPayload, nil
}

func ValidateJobInput() error {
	if pResourcePath == "" {
		return fmt.Errorf("resource path is required")
	}

	if pReferenceID == "" {
		return fmt.Errorf("reference ID is required")
	}

	if pValidFrom == "" {
		return fmt.Errorf("valid from is required")
	}

	if pUserIDs == "" {
		return fmt.Errorf("user ids is required")
	}

	if len(pContent) == 0 {
		return fmt.Errorf("content is required")
	}

	return nil
}
