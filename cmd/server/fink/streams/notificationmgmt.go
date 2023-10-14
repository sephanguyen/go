package streams

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	nats_org "github.com/nats-io/nats.go"
)

func GetStreamConfigNotificationmgmt() []*nats_org.StreamConfig {
	arrStreamConfig := []*nats_org.StreamConfig{
		{
			Name:      constants.StreamNotification,
			Retention: nats_org.InterestPolicy,
			Subjects:  []string{"Notification.*"},
			Replicas:  3,
			MaxBytes:  1024 * 1024 * 1024 * 4, // 4 GB
		},
	}

	return arrStreamConfig
}
