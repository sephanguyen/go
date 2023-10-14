package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_RecordNotificationCreated(t *testing.T) {
	t.Parallel()
	coll := NewClientMetrics("notification")

	notificationCreatedCounter := coll.GetCollectors()[0]
	r := prometheus.NewRegistry()
	r.MustRegister(notificationCreatedCounter)
	// one metric exposed
	assert.Equal(t, 1, testutil.CollectAndCount(notificationCreatedCounter))

	assert.Equal(t, 0.0, testutil.ToFloat64(notificationCreatedCounter))

	totalNotiCreated := 15.0
	for i := 0; i < int(totalNotiCreated); i++ {
		coll.RecordNotificationCreated(1)
	}
	assert.Equal(t, totalNotiCreated, testutil.ToFloat64(notificationCreatedCounter))
}

func Test_RecordUserNotificationCreated(t *testing.T) {
	t.Parallel()
	coll := NewClientMetrics("notification")

	userNotificationCreatedCounter := coll.GetCollectors()[1]
	r := prometheus.NewRegistry()
	r.MustRegister(userNotificationCreatedCounter)
	// one metric exposed
	assert.Equal(t, 1, testutil.CollectAndCount(userNotificationCreatedCounter))

	assert.Equal(t, 0.0, testutil.ToFloat64(userNotificationCreatedCounter))

	totalUserNotiCreated := 15.0
	for i := 0; i < int(totalUserNotiCreated); i++ {
		coll.RecordUserNotificationCreated(1)
	}
	assert.Equal(t, totalUserNotiCreated, testutil.ToFloat64(userNotificationCreatedCounter))
}

func Test_RecordPushNotificationErrors(t *testing.T) {
	t.Parallel()
	coll := NewClientMetrics("notification")

	errorCounterCollector := coll.GetCollectors()[2]
	r := prometheus.NewRegistry()
	r.MustRegister(errorCounterCollector)
	totalErrors := 15.0
	// recording some error events
	for i := 0; i < int(totalErrors); i++ {
		coll.RecordPushNotificationErrors("ok", 1)
	}
	assert.Equal(t, totalErrors, testutil.ToFloat64(errorCounterCollector))
}
