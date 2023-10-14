package consumers

import "context"

type SubscriberHandler interface {
	Handle(ctx context.Context, msg []byte) (bool, error)
}
