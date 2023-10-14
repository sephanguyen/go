package consumers

import "context"

type ConsumerHandler interface {
	Handle(ctx context.Context, msg []byte) (bool, error)
}
