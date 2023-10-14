package enigma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func (s *suite) healthCheckEndpointCalled(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/healthcheck/status", s.EnigmaSrvURL)
	bodyBytes, err := s.makeHTTPRequest(http.MethodGet, url)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bodyBytes == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("body is nil")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) everythingIsOK(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) makeHTTPRequest(method, url string) ([]byte, error) {
	bodyRequest, err := json.Marshal(s.Request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("JPREP-Signature", s.JPREPSignature)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}
	s.Response = resp
	return body, nil
}
