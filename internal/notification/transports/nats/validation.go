package nats

import (
	"fmt"

	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func ValidateNatsMessage(data *ypb.NatsCreateNotificationRequest) error {
	acceptClient := false

	for _, clientID := range ClientIDsAccepted {
		if clientID == data.ClientId {
			acceptClient = true
		}
	}

	if !acceptClient {
		return fmt.Errorf("prevent client_id: %s", data.ClientId)
	}

	return nil
}
