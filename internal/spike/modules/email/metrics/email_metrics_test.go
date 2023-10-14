package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"gotest.tools/assert"
)

func Test_RecordEmailEvents(t *testing.T) {
	t.Parallel()
	coll := NewClientMetrics("spike")

	emailEventCounter := coll.GetCollectors()[0]
	r := prometheus.NewRegistry()
	r.MustRegister(emailEventCounter)
	totalEvents := 15.0
	// recording some error events
	for i := 0; i < int(totalEvents); i++ {
		coll.RecordEmailEvents("processed", 1)
	}
	assert.Equal(t, totalEvents, testutil.ToFloat64(emailEventCounter))
}
