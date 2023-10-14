package topics

import (
	"github.com/manabie-com/backend/internal/golibs/constants"

	"github.com/segmentio/kafka-go"
)

func GetTopicConfigSpike() []*kafka.TopicConfig {
	var arrTopicConfig = []*kafka.TopicConfig{
		{
			Topic:             constants.EmailSendingTopic,
			NumPartitions:     10,
			ReplicationFactor: 3,
		},
	}

	return arrTopicConfig
}
