package chat

import "github.com/manabie-com/backend/internal/golibs/metrics"

func (s *Server) RegisterMetric(collector metrics.MetricCollector) {
	collector.RegisterGaugeFunc(metrics.MetricOpt{
		Name: "manabie_app_tom_total_conn",
		Help: "total connections of server",
		Labels: map[string]string{
			"app": "tom",
		},
	}, s.TotalConn)
	collector.RegisterCounterFunc(metrics.MetricOpt{
		Name: "manabie_app_tom_total_server_disconnect",
		Help: "total server side disconnection",
		Labels: map[string]string{
			"app": "tom",
		},
	}, s.serverDisconnections)
	collector.RegisterCounterFunc(metrics.MetricOpt{
		Name: "manabie_app_tom_total_client_disconnect",
		Help: "total client side disconnection",
		Labels: map[string]string{
			"app": "tom",
		},
	}, s.clientDisconnections)
}
