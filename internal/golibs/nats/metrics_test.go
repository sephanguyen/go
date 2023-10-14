package nats

import (
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_disconnectMetric(t *testing.T) {
	t.Parallel()
	jsm := &jetStreamManagementImpl{}
	coll := NewClientMetrics("tom", jsm)

	disconnectCollector := coll[0]
	r := prometheus.NewRegistry()
	r.MustRegister(disconnectCollector)
	// one metric exposed
	assert.Equal(t, 1, testutil.CollectAndCount(disconnectCollector))

	assert.Equal(t, 0.0, testutil.ToFloat64(disconnectCollector))

	totalDisconnections := 2.0
	// recording multiple disconnect events
	for i := 0; i < int(totalDisconnections); i++ {
		for _, h := range jsm.disconnectErrHandlers {
			h(&nats.Conn{}, fmt.Errorf("dummy"))
		}
	}
	assert.Equal(t, totalDisconnections, testutil.ToFloat64(disconnectCollector))

	// recording reconnect event
	for _, h := range jsm.reconnectHandlers {
		h(&nats.Conn{})
	}
	assert.Equal(t, 0.0, testutil.ToFloat64(disconnectCollector))
}
