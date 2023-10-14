package mock

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/firebase"

	"firebase.google.com/go/v4/messaging"
	"github.com/gogo/protobuf/types"
)

type Notifier struct {
	pushedMulticastMessages map[string][]*messaging.MulticastMessage
}

func NewNotifier() *Notifier {
	return &Notifier{
		pushedMulticastMessages: make(map[string][]*messaging.MulticastMessage),
	}
}

// nolint:staticcheck
func (n *Notifier) SendTokens(_ context.Context, msg *messaging.MulticastMessage, deviceTokens []string) (int, int, *firebase.SendTokensError) {
	for _, deviceToken := range deviceTokens {
		n.pushedMulticastMessages[deviceToken] = append(n.pushedMulticastMessages[deviceToken], msg)
	}

	// sucessCount, failureCount, error
	return len(deviceTokens), 0, nil
}

func (n *Notifier) RetrievePushedMessages(_ context.Context, deviceToken string, _ int, _ *types.Timestamp) ([]*messaging.MulticastMessage, error) {
	return n.pushedMulticastMessages[deviceToken], nil
}
