package mastermgmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/middlewares"

	"github.com/gin-gonic/gin"
)

func (s *suite) withAppsmithTrackRequest(ctx context.Context, headerType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch headerType {
	case "wrong key":
		stepState.AuthKey = "Wrong-key"
		stepState.AuthValue = ""

	case "wrong value":
		stepState.AuthKey = middlewares.MasterHeaderKey
		stepState.AuthValue = "Wrong auth value"

	case "valid":
		stepState.AuthKey = middlewares.MasterHeaderKey
		stepState.AuthValue = middlewares.MasterAuthValue
	}
	_ = `{
		"context": {
		  "ip": "203.192.213.46",
		  "library": {
			"name": "unknown",
			"version": "unknown"
		  }
		},
		"event": "Instance Active",
		"integrations": {},
		"messageId": "api-1jokIBOkNv8nEmu2fGeNb01G1RC",
		"properties": {
		  "instanceId": "<uuid>"
		},
		"receivedAt": "2020-11-04T08:15:49.537Z",
		"timestamp": "2020-11-04T08:15:49.537Z",
		"type": "track",
		"userId": "203.192.213.46"
	  }
	  `
	logStr := domain.EventLog{
		"context": map[string]interface{}{
			"ip": "203.192.213.46",
		},
		"userId": "203.192.213.46",
		"type":   "track",
	}
	s.Request = logStr
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) trackEndpointIsCalled(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	url := fmt.Sprintf("%s/mastermgmt/api/v1/appsmith/track", "http://"+s.Cfg.MasterMgmtHTTPSrvAddr)
	res, err := s.makeHTTPRequest(http.MethodPost, url)
	stepState.RestResponseBody = res
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnTrackResponse(ctx context.Context, statusCode string, responseType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := s.RestResponse.(*http.Response)
	if fmt.Sprintf("%d", resp.StatusCode) != statusCode {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect status code %s, got %d\nauthKey: %s, authValue:%s\nrequest: %v", statusCode, resp.StatusCode, stepState.AuthKey, stepState.AuthValue, s.Request)
	}
	var expectedBody []byte

	switch responseType {
	case "error":
		expectedBody, _ = json.Marshal(gin.H{
			"error": "signature is not match",
		})
	case "success":
		expectedBody, _ = json.Marshal(gin.H{
			"success": true,
		})
	}

	resBody := string(stepState.RestResponseBody)
	if resBody != string(expectedBody) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect body %s, got %s\nauthKey: %s, authValue:%s\nrequest: %v", expectedBody, resBody, stepState.AuthKey, stepState.AuthValue, s.Request)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) makeHTTPRequest(method, url string) ([]byte, error) {
	bodyRequest, err := json.Marshal(s.Request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Close = true
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set(s.AuthKey, s.AuthValue)
	if err != nil {
		return nil, err
	}

	client := http.Client{Transport: &http.Transport{}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can not read response body: %s", err.Error())
	}
	s.StepState.RestResponse = resp
	return body, nil
}
