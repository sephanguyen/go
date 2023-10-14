package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultCountMetricValue = 0
)

type EmailMetrics interface {
	RecordEmailEvents(event EmailEventMetricType, num float64)
	GetCollectors() []prometheus.Collector

	InitCounterValue()
}

// nolint
func NewClientMetrics(appname string) *clientMetrics {
	c := &clientMetrics{
		emailEventCounter: *prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "manabie_email_event_counter",
			Help: "Counting email event (queued, processed, delivered, bounce,...)",
			ConstLabels: map[string]string{
				"app": appname,
			},
		}, []string{"event"}),
	}

	return c
}

type clientMetrics struct {
	emailEventCounter prometheus.CounterVec
}

func (c *clientMetrics) RecordEmailEvents(event EmailEventMetricType, num float64) {
	c.emailEventCounter.With(
		prometheus.Labels{"event": string(event)},
	).Add(num)
}

func (c *clientMetrics) GetCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		c.emailEventCounter,
	}
}

func (c *clientMetrics) InitCounterValue() {
	c.RecordEmailEvents(EmailQueued, defaultCountMetricValue)
	c.RecordEmailEvents(EmailDropped, defaultCountMetricValue)
	c.RecordEmailEvents(EmailBounce, defaultCountMetricValue)
	c.RecordEmailEvents(EmailProcessed, defaultCountMetricValue)
}
