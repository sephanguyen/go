package core

import "context"

var (
	DomainEvtBus = DomainEventPublisher{}
)

type DomainEventPublisher struct {
	subscribers []Subscription
}
type Subscription struct {
	Event   string
	Handler func(ctx context.Context, data interface{}) error
}

func (d *DomainEventPublisher) Publish(ctx context.Context, event string, data interface{}) error {
	for _, h := range d.subscribers {
		if h.Event == event {
			err := h.Handler(ctx, data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DomainEventPublisher) RegisterSubscriber(f Subscription) {
	d.subscribers = append(d.subscribers, f)
}

func RegisterSubscriber(f Subscription) {
	DomainEvtBus.RegisterSubscriber(f)
}
