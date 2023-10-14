package helper

import "context"

func (h *CommunicationHelper) PublishToNats(ctx context.Context, topic string, data []byte) error {
	_, err := h.jsm.PublishContext(ctx, topic, data)
	return err
}
