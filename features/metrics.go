package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cucumber/godog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type metricsService struct {
	pusher *push.Pusher
}

func (m *metricsService) suiteHook(svcName string, next func(*godog.TestSuiteContext)) func(*godog.TestSuiteContext) {
	if m == nil {
		return next
	}

	return func(sc *godog.TestSuiteContext) {
		sc.AfterSuite(func() {
			hostname, err := os.Hostname()
			if err != nil {
				log.Printf("cannot get current hostname: %+v\n", err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := m.pusher.Grouping("instance", hostname).Grouping("service_name", svcName).AddContext(ctx); err != nil {
				log.Printf("cannot push metrics to the Pushgateway: %+v\n", err)
			}
		})

		next(sc)
	}
}

func (m *metricsService) scenarioHook(next func(*godog.ScenarioContext)) func(*godog.ScenarioContext) {
	if m == nil {
		return next
	}

	var (
		scenarioCounter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "manabie_bdd_tests",
				Name:      "scenarios_total",
				Help:      "Total BDD scenarios",
			},
			[]string{"scenario_name", "scenario_result"},
		)
		// stepCounter = prometheus.NewCounterVec(
		// 	prometheus.CounterOpts{
		// 		Namespace: "manabie_bdd_tests",
		// 		Name:      "scenario_steps_total",
		// 		Help:      "Total BDD steps",
		// 	},
		// 	[]string{"step_name", "step_result"},
		// )
		scenarioHistogram = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "manabie_bdd_tests",
				Name:      "scenario_duration_milliseconds",
				Help:      "Scenario duration",
				Buckets:   []float64{1_000, 3_000, 5_000, 10_000, 30_000, 50_000, 100_000},
			},
			[]string{"scenario_name"},
		)
	)
	m.pusher.Collector(scenarioCounter).Collector(scenarioHistogram)

	return func(ctx *godog.ScenarioContext) {
		ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			ctx = context.WithValue(ctx, scenarioDurationKey{}, time.Now())
			return ctx, nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			var testResult string
			if err != nil {
				testResult = "failed"
			} else {
				if startTime, ok := ctx.Value(scenarioDurationKey{}).(time.Time); ok {
					elapsed := time.Since(startTime)
					latencyMs := float64(elapsed) / float64(time.Millisecond)
					scenarioHistogram.WithLabelValues(sc.Name).Observe(latencyMs)
				}
				testResult = "passed"
			}

			scenarioCounter.WithLabelValues(sc.Name, testResult).Inc()

			return ctx, err
		})

		// ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		// 	stepCounter.WithLabelValues(st.Text, status.String()).Inc()
		//
		// 	return ctx, err
		// })

		next(ctx)
	}
}

type scenarioDurationKey struct{}
