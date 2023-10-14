package metrics

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	coll := NewMetricCollector()
	coll.RegisterCounterFunc(MetricOpt{Name: "a"}, func() float64 {
		return 10
	})
	coll.RegisterGaugeFunc(MetricOpt{Name: "b"}, func() float64 {
		return 11
	})

	r := prometheus.NewRegistry()
	r.MustRegister(coll.metrics...)

	assert.Equal(t, 10.0, testutil.ToFloat64(coll.metrics[0]))
	assert.Equal(t, 11.0, testutil.ToFloat64(coll.metrics[1]))
}

func TestPrometheusCollector_RegisterHistogram(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic occurred:", err)
		}
	}()

	coll := NewMetricCollector()
	upperBound := 3
	width := 4
	count := 5
	histogramCollector := coll.RegisterHistogram(MetricOpt{
		Name:       "sample_metric",
		Help:       "This is help",
		Labels:     map[string]string{},
		LabelNames: []string{"name_space", "type"},
	}, prometheus.LinearBuckets(float64(upperBound), float64(width), count))

	r := prometheus.NewRegistry()
	err := r.Register(coll.metrics[0])
	require.NoError(t, err)

	labels := [][]string{
		{"backend", "one"},
		{"backend", "two"},
		{"frontend", "two"},
	}
	for _, l := range labels {
		histogramCollector.WithLabelValues(l...).Observe(10)
		histogramCollector.WithLabelValues(l...).Observe(20)
		histogramCollector.WithLabelValues(l...).Observe(2)
	}

	assert.Equal(t, 3, testutil.CollectAndCount(coll.metrics[0], "sample_metric"))

	for _, l := range labels {
		expectedUpperBound := upperBound
		metric := &dto.Metric{}
		m, err := histogramCollector.MetricVec.GetMetricWithLabelValues(l...)
		require.NoError(t, err)
		err = m.Write(metric)
		require.NoError(t, err)
		require.NotNil(t, metric.Histogram)
		assert.Equal(t, float64(32), *metric.Histogram.SampleSum)
		assert.Equal(t, uint64(3), *metric.Histogram.SampleCount)
		assert.Len(t, metric.Histogram.Bucket, count)
		for i, bucket := range metric.Histogram.Bucket {
			if i > 1 {
				assert.Equal(t, uint64(2), *bucket.CumulativeCount)
			} else {
				assert.Equal(t, uint64(1), *bucket.CumulativeCount)
			}
			assert.Equal(t, float64(expectedUpperBound), *bucket.UpperBound)
			expectedUpperBound += width
		}
	}
}

func TestPrometheusCollector_RegisterGauge(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic occurred:", err)
		}
	}()

	coll := NewMetricCollector()
	opsQueued := coll.RegisterGauge(MetricOpt{
		Name:       "ops_queued",
		Help:       "Number of blob storage operations waiting to be processed.",
		LabelNames: []string{"a"},
	})
	prometheus.MustRegister(opsQueued)

	// 10 operations queued by the goroutine managing incoming requests.
	opsQueued.WithLabelValues("a").Add(10)
	// A worker goroutine has picked up a waiting operation.
	opsQueued.WithLabelValues("a").Dec()
	// And once more...
	opsQueued.WithLabelValues("a").Dec()

	g, err := opsQueued.GetMetricWithLabelValues("a")
	require.NoError(t, err)
	metric := &dto.Metric{}
	err = g.Write(metric)
	require.NoError(t, err)
	assert.EqualValues(t, 8, *metric.Gauge.Value)
}
