package bob

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	unleash_client_entities "github.com/manabie-com/backend/internal/golibs/unleashclient/entities"
)

const (
	ToggleStatusEnable             = "enable"
	ToggleStatusDisable            = "disable"
	jsonContentType                = "application/json"
	unleashAdminEndpoint           = "/unleash/api/admin/projects/default/features/"
	unleashRetrieveFeatureEndpoint = "/unleash/api/client/features/"
)

var httpTransport = &http.Transport{
	// #nosec G402
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

func (s *suite) ToggleUnleashFeatureWithName(ctx context.Context, toggleChoice string, featureName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	httpClient := &http.Client{
		Transport: httpTransport,
	}
	var unleashPath string
	// Login to get credentials
	switch toggleChoice {
	case ToggleStatusEnable:
		{
			unleashPath = unleashAdminEndpoint + featureName + "/environments/default/on"
		}
	case ToggleStatusDisable:
		{
			unleashPath = unleashAdminEndpoint + featureName + "/environments/default/off"
		}
	}

	header := make(map[string][]string)
	header["Authorization"] = []string{s.UnleashLocalAdminAPIKey}
	header["Content-Type"] = []string{jsonContentType}

	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   s.UnleashSrvAddr,
			Path:   unleashPath,
		},
		Header: header,
	}
	// Request Unleash to toggle, loop request until the toggle is activated
	isToggleChoiceCorrect := false
	if err := try.Do(func(attempt int) (bool, error) {
		resp, err := httpClient.Do(req)
		if err != nil {
			return false, fmt.Errorf("error requesting to unleash:%w", err)
		}
		defer resp.Body.Close()
		isToggleChoiceCorrect = isFeatureToggleCorrect(s.UnleashSrvAddr, s.UnleashLocalAdminAPIKey, featureName, toggleChoice)

		if !isToggleChoiceCorrect {
			time.Sleep(1 * time.Second)
			return attempt < 5, fmt.Errorf("can't toggle unleash %s", featureName)
		}

		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

// This function is used for retrying to toggle Unleash if the toggle is not changed yet
func isFeatureToggleCorrect(unleashSrvAddr string, apiKey string, featureName string, toggleSelect string) bool {
	// Wait for the change request to be done successfully
	time.Sleep(time.Second * 5)
	httpClient := &http.Client{
		Transport: httpTransport,
	}
	featureEntity := unleash_client_entities.UnleashFeatureEntity{}
	unleashUrl := unleashRetrieveFeatureEndpoint + featureName

	header := make(map[string][]string)
	header["Authorization"] = []string{apiKey}
	header["Content-Type"] = []string{jsonContentType}

	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   unleashSrvAddr,
			Path:   unleashUrl,
		},
		Header: header,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	// parse response
	err = json.Unmarshal(body, &featureEntity)
	if err != nil {
		return false
	}
	if (toggleSelect == ToggleStatusEnable && featureEntity.Enabled) || (toggleSelect == ToggleStatusDisable && !featureEntity.Enabled) {
		return true
	}
	return false
}
