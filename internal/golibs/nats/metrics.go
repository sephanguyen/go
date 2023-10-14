package nats

import (
	"github.com/prometheus/client_golang/prometheus"

	natsgo "github.com/nats-io/nats.go"
)

func (c *clientMetrics) recordDisconnect() {
	c.disconnect.Inc()
}
func (c *clientMetrics) recordReconnect() {
	c.disconnect.Set(0)
}

// TODO: add more metric here if needed
type clientMetrics struct {
	disconnect prometheus.Gauge
}

func NewClientMetrics(appname string, jsm JetStreamManagement) []prometheus.Collector {
	c := &clientMetrics{
		disconnect: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "manabie_app_nats_disconnect",
			Help: "Whether disconnect error has happened recently",
			ConstLabels: map[string]string{
				"app": appname, // this label is useful to route which alert goes to which channel (slack/opsgenie)
			},
		}),
	}

	jsm.RegisterDisconnectErrHandler(func(conn *natsgo.Conn, err error) { c.recordDisconnect() })
	jsm.RegisterReconnectHandler(func(conn *natsgo.Conn) { c.recordReconnect() })

	return []prometheus.Collector{c.disconnect}
}
