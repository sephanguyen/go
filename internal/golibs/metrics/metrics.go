package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MetricCollector interface {
	RegisterGaugeFunc(MetricOpt, func() float64)
	RegisterGauge(opts MetricOpt) *prometheus.GaugeVec
	RegisterCounterFunc(MetricOpt, func() float64)
	RegisterHistogram(opts MetricOpt, buckets []float64) *prometheus.HistogramVec
}
type PrometheusCollector struct {
	metrics []prometheus.Collector
}

func NewMetricCollector() *PrometheusCollector {
	return &PrometheusCollector{}
}

func (coll *PrometheusCollector) Collectors() []prometheus.Collector {
	return coll.metrics
}

func (coll *PrometheusCollector) RegisterCounterFunc(opts MetricOpt, recorder func() float64) {
	newMetric := prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name:        opts.Name,
		ConstLabels: opts.Labels,
		Help:        opts.Help,
	}, recorder)
	coll.metrics = append(coll.metrics, newMetric)
}

func (coll *PrometheusCollector) RegisterGaugeFunc(opts MetricOpt, recorder func() float64) {
	newMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        opts.Name,
		ConstLabels: opts.Labels,
		Help:        opts.Help,
	}, recorder)
	coll.metrics = append(coll.metrics, newMetric)
}

func (coll *PrometheusCollector) RegisterGauge(opts MetricOpt) *prometheus.GaugeVec {
	newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        opts.Name,
		ConstLabels: opts.Labels,
		Help:        opts.Help,
	}, opts.LabelNames)
	coll.metrics = append(coll.metrics, newMetric)

	return newMetric
}

func (coll *PrometheusCollector) RegisterHistogram(opts MetricOpt, buckets []float64) *prometheus.HistogramVec {
	newMetric := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        opts.Name,
		Help:        opts.Help,
		ConstLabels: opts.Labels,
		Buckets:     buckets,
	}, opts.LabelNames)
	coll.metrics = append(coll.metrics, newMetric)

	return newMetric
}

type MetricOpt struct {
	Name       string
	Help       string
	Labels     map[string]string // constant labels
	LabelNames []string
}
