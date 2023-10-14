package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	mock_bootstrap "github.com/manabie-com/backend/mock/golibs/bootstrap"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitnatsf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := zap.NewNop()

	type testConfig struct {
		NatsJS configs.NatsJetStreamConfig
	}
	c := &testConfig{
		NatsJS: configs.NatsJetStreamConfig{Address: "address", User: "user", Password: "password"},
	}
	t.Run("with nats config", func(t *testing.T) {
		natsjs := new(mock_nats.JetStreamManagement)
		natsjs.On("ConnectToJS").Once()
		natsjser := mock_bootstrap.NewNATSJetstreamer(t)
		natsjser.On("NewJetStreamManagement", l, &c.NatsJS).Once().Return(natsjs, nil)
		rsc := NewResources().WithLogger(l)
		rsc.natsjser = natsjser
		err := initnats(ctx, c, rsc)
		require.NoError(t, err)
		require.Equal(t, natsjs, rsc.NATS())
	})

	t.Run("with invalid nats config", func(t *testing.T) {
		natsjser := mock_bootstrap.NewNATSJetstreamer(t)
		natsjser.On("NewJetStreamManagement", l, &c.NatsJS).Once().Return(nil, errors.New("some errors"))
		rsc := NewResources().WithLogger(l)
		rsc.natsjser = natsjser
		err := initnats(ctx, c, rsc)
		require.NoError(t, err)
		require.PanicsWithError(t, "some errors", func() { _ = rsc.NATS() })
	})

	t.Run("without nats config", func(t *testing.T) {
		type testConfig struct{}
		rsc := NewResources().WithLogger(l)
		err := initnats(ctx, testConfig{}, rsc)
		require.NoError(t, err)
		require.Nil(t, rsc.natsjs)
		require.PanicsWithValue(t, "unable to init NAT Jetstream client: NATS config is not provided", func() { _ = rsc.NATS() })
	})
}
