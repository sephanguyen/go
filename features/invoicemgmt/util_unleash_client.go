package invoicemgmt

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	unleash_client_entities "github.com/manabie-com/backend/internal/golibs/unleashclient/entities"
)

const (
	unleashRetrieveFeatureEndpoint = "/unleash/api/client/features/"
	jsonContentType                = "application/json"
)

var httpTransport = &http.Transport{
	// #nosec G402
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

func isFeatureToggleEnabled(unleashSrvAddr string, apiKey string, featureName string) bool {
	// Wait for the change request to be done successfully
	httpClient := &http.Client{
		Transport: httpTransport,
	}
	featureEntity := unleash_client_entities.UnleashFeatureEntity{}
	unleashURL := unleashRetrieveFeatureEndpoint + featureName

	header := make(map[string][]string)
	header["Authorization"] = []string{apiKey}
	header["Content-Type"] = []string{jsonContentType}

	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   unleashSrvAddr,
			Path:   unleashURL,
		},
		Header: header,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	// parse response
	err = json.Unmarshal(body, &featureEntity)
	if err != nil {
		return false
	}

	return featureEntity.Enabled
}
