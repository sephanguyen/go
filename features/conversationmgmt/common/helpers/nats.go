package helpers

import "context"

func (helper *ConversationMgmtHelper) PublishToNats(ctx context.Context, topic string, data []byte) error {
	_, err := helper.JSM.PublishContext(ctx, topic, data)
	return err
}
