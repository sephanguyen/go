package helpers

import "context"

func (helper *ConversationMgmtHelper) PublishToKafka(ctx context.Context, topic string, data []byte) error {
	err := helper.Kafka.PublishContext(ctx, topic, nil, data)
	return err
}
