package bootstrap

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.uber.org/zap"
)

// NatsServicer represents a service that uses NATS Jetstream.
// Users should implements this interface with their server struct
// when they want to use NATS.
type NatsServicer[T any] interface {
	// RegisterNatsSubscribers should be implemented by users to use NATS Jetstream.
	// It should register all the necessary subscriptions to streams.
	RegisterNatsSubscribers(context.Context, T, *Resources) error
}

func initnats(_ context.Context, config interface{}, rsc *Resources) error {
	c, err := extract[configs.NatsJetStreamConfig](config, natsjsFieldName)
	if err != nil {
		return ignoreErrFieldNotFound(err)
	}

	_ = rsc.WithNATSC(c)
	return nil
}

// NATSJetstreamer handles the connection to NATS server.
type NATSJetstreamer interface {
	// NewJetStreamManagement returns a new nats.JetStreamManagement instance.
	NewJetStreamManagement(zapLogger *zap.Logger, c *configs.NatsJetStreamConfig) (nats.JetStreamManagement, error)
}

// natsJetstreamImpl implements NATSJetstreamer.
type natsJetstreamImpl struct{}

func newNATSJetstreamImpl() *natsJetstreamImpl {
	return &natsJetstreamImpl{}
}

func (n *natsJetstreamImpl) NewJetStreamManagement(zapLogger *zap.Logger, c *configs.NatsJetStreamConfig) (nats.JetStreamManagement, error) {
	return nats.NewJetStreamManagement(c.Address, c.User, c.Password, c.MaxReconnects, c.ReconnectWait, c.IsLocal, zapLogger)
}
