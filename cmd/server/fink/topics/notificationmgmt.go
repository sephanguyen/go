package topics

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	"github.com/segmentio/kafka-go"
)

func GetTopicConfigNotificationmgmt() []*kafka.TopicConfig {
	var arrTopicConfig = []*kafka.TopicConfig{
		{
			Topic:             constants.SystemNotificationUpsertingTopic,
			NumPartitions:     10,
			ReplicationFactor: 3,
		},
	}

	return arrTopicConfig
}
