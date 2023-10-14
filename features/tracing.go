package main

import (
	"context"
	"strings"

	"github.com/cucumber/godog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func scenarioTracingInitializer(tp *tracesdk.TracerProvider, next func(*godog.ScenarioContext)) func(*godog.ScenarioContext) {
	if tp == nil {
		return next
	}

	return func(ctx *godog.ScenarioContext) {
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx, span := tp.Tracer(sc.Id).Start(ctx, sc.Uri, trace.WithAttributes(
				attribute.String("scenario_name", sc.Name),
			))

			// children steps need this parent span
			ctx = context.WithValue(ctx, traceScenarioSpanKey{}, span)
			return ctx, nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			if span, ok := ctx.Value(traceScenarioSpanKey{}).(trace.Span); ok {
				if err != nil {
					span.SetAttributes(attribute.String("x-has-error", "true"))
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
				}
				span.End()
			}

			return ctx, err
		})

		ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			// Get parent scenario span, so that next span created is a children of this span
			// instead of the current span in the context, which should be its cousin
			if span, ok := ctx.Value(traceScenarioSpanKey{}).(trace.Span); ok {
				ctx = trace.ContextWithSpan(ctx, span)
			}

			ctx, span := tp.Tracer(st.Id).Start(
				ctx,
				strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(st.Text), " ", "-"), "\"", ""),
				trace.WithAttributes(
					attribute.String("step_name", st.Text),
				),
			)
			ctx = context.WithValue(ctx, traceStepSpanKey(st.Id), span) //nolint:revive,staticcheck
			return ctx, nil
		})

		ctx.StepContext().After(func(ctx context.Context, st *godog.Step, stat godog.StepResultStatus, err error) (context.Context, error) {
			if span, ok := ctx.Value(traceStepSpanKey(st.Id)).(trace.Span); ok {
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
				}
				span.End()
			}

			return ctx, err
		})

		next(ctx)
	}
}

type traceScenarioSpanKey struct{}

func traceStepSpanKey(id string) string {
	return "x-trace-step-key" + id
}
