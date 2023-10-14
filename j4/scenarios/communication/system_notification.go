package communication

import (
	"context"
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/j4/infras"

	j4 "github.com/manabie-com/j4/pkg/runner"
)

func SystemNotificationScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: c.SchoolID,
			UserID:       c.AdminID,
		},
	})

	// tokenGenerator := serviceutil.NewTokenGenerator(c, dep.Connections)

	runConfig, err := c.GetScenarioConfig("Notification_SystemNotificationTest")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)
	runCfg.TestFunc = func(_ context.Context) error {
		payload := makePayloadUpsertNotification()
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		err = dep.Kafka.TracedPublishContext(ctx, "Notification_SystemNotificationTest", constants.SystemNotificationUpsertingTopic, []byte(payload.ReferenceID), data)
		if err != nil {
			return err
		}
		return nil
	}

	scenario, err := j4.NewScenario("Notification_SystemNotificationTest", *runCfg)
	if err != nil {
		return nil, err
	}

	return []*j4.Scenario{
		scenario,
	}, nil
}

func makePayloadUpsertNotification() *payload.UpsertSystemNotification {
	payload := &payload.UpsertSystemNotification{
		ReferenceID: idutil.ULIDNow(),
		Content:     makeContent(),
		URL:         "http://random",
		ValidFrom:   time.Now(),
		Recipients:  makeRecipients(),
		IsDeleted:   false,
	}

	return payload
}

func makeContent() []payload.SystemNotificationContent {
	content := payload.SystemNotificationContent{
		Language: "en",
		Text:     "Donec ultrices pretium enim, eget malesuada nisl ornare a. Nunc consequat magna ac elit volutpat tincidunt. Phasellus suscipit convallis facilisis. Donec dictum lacus arcu, id elementum justo aliquet eu. Nulla odio nibh, laoreet quis magna non, volutpat pretium massa. Aliquam sed velit pretium, scelerisque purus ut, tempus mi. Proin aliquam pretium lectus eu suscipit. Phasellus convallis porta ipsum, eu iaculis mauris euismod eu. Sed aliquet, sapien in consequat gravida, mi ex pulvinar felis, id sollicitudin leo dolor feugiat orci. Nullam ut metus at nisi placerat gravida. Etiam non elit maximus, pharetra dolor nec, commodo ex. Quisque sagittis metus vel pellentesque consectetur. Nam at aliquam est, at congue justo. Suspendisse dapibus nisl ut lacus fringilla ornare.",
	}
	return []payload.SystemNotificationContent{content}
}

func makeRecipients() []payload.SystemNotificationRecipient {
	recipients := []payload.SystemNotificationRecipient{}
	for i := 0; i < 10; i++ {
		recipients = append(recipients, payload.SystemNotificationRecipient{
			UserID: idutil.ULIDNow(),
		})
	}

	return recipients
}
