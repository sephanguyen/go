package unleash

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (s *suite) theRequestToCheckUnleashHealth(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   s.UnleashSrvAddr,
			Path:   "/unleash/health",
		},
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) sendRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = http.DefaultClient.Do(stepState.Request.(*http.Request))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) unleashMustReturnHealthStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("failed to get unleash health status: %v", stepState.ResponseErr)
	}
	resp := stepState.Response.(*http.Response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("expected status is 200, but the fact is: %d", resp.StatusCode)
	}
	var body HealthCheckResponse
	err := json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return ctx, fmt.Errorf("failed to convert response body to struct HealthCheckResponse: %v", err)
	}
	if body.Health != "GOOD" {
		return ctx, fmt.Errorf("expected status in body is GOOD, but the fact is %v", body.Health)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theRequestToCheckUnleashProxyHealth(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "unleash-proxy.local-manabie-unleash.svc.cluster.local:4243",
			Path:   "/proxy/health",
		},
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) unleashProxyMustReturnHealthStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("failed to get unleash health status: %v", stepState.ResponseErr)
	}
	resp := stepState.Response.(*http.Response)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("expected status is 200, but the fact is: %d", resp.StatusCode)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to convert resp.Body to []byte: %v", err)
	}
	if string(bytes) != "ok" {
		return ctx, fmt.Errorf("expected body is ok, but the fact is: %s", string(bytes))
	}
	return StepStateToContext(ctx, stepState), nil
}
