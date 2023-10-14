package unleash

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var httpTransport = &http.Transport{
	// #nosec G402
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

const (
	appName               = "manabie-backend-usermgmt-unleash-client-integration-test"
	jsonContentType       = "application/json"
	prefixHTTP            = "http://"
	unleashAdminEndpoint  = "/unleash/api/admin/projects/default/features/"
	unleashClientEndpoint = "/unleash/api/"
)

type Client struct {
	httpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	unleashSrvAddr          string
	unleashAPIKey           string
	unleashLocalAdminAPIKey string

	unleashClientSDK unleashclient.ClientInstance
}

func NewDefaultClient(unleashSrvAddr, unleashAPIKey, unleashLocalAdminAPIKey string) *Client {
	client := &Client{
		httpClient:              &http.Client{Transport: httpTransport},
		unleashSrvAddr:          unleashSrvAddr,
		unleashAPIKey:           unleashAPIKey,
		unleashLocalAdminAPIKey: unleashLocalAdminAPIKey,
	}

	return client
}
func (client *Client) connectUnleashSDK() error {
	clientUnleashSDK, err := unleashclient.NewUnleashClientInstance(
		prefixHTTP+client.unleashSrvAddr+unleashClientEndpoint,
		appName,
		client.unleashAPIKey,
		zap.NewNop(),
	)

	if err != nil {
		return err
	}
	if err := clientUnleashSDK.ConnectToUnleashClient(); err != nil {
		return err
	}
	clientUnleashSDK.WaitForUnleashReady()
	client.unleashClientSDK = clientUnleashSDK

	return nil
}

func (client *Client) ToggleUnleashFeatureWithName(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
	if err := validToggleOption(toggleChoice); err != nil {
		return err
	}

	unleashPath := unleashAdminEndpoint + featureName
	// Login to get credentials
	switch toggleChoice {
	case ToggleChoiceEnable:
		unleashPath += "/environments/default/on"
	case ToggleChoiceDisable:
		unleashPath += "/environments/default/off"
	}

	header := make(map[string][]string)
	header["Authorization"] = []string{client.unleashLocalAdminAPIKey}
	header["Content-Type"] = []string{jsonContentType}

	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   client.unleashSrvAddr,
			Path:   unleashPath,
		},
		Header: header,
	}
	req = req.WithContext(ctx)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error requesting to unleash")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrFeatureFlagNotFound{FeatureFlagName: featureName}
	default:
		body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		if err != nil {
			return err
		}
		return errors.New("encounter error while interacting with unleash: " + string(body))
	}
}

type FeatureEntity struct {
	Strategies []struct {
		Name        string        `json:"name"`
		Constraints []interface{} `json:"constraints"`
		Parameters  struct {
			Environments string `json:"environments"`
		} `json:"parameters"`
	} `json:"strategies"`
	ImpressionData bool          `json:"impressionData"`
	Enabled        bool          `json:"enabled"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Project        string        `json:"project"`
	Stale          bool          `json:"stale"`
	Type           string        `json:"type"`
	Variants       []interface{} `json:"variants"`
}

// GetFeatureToggleDetails returns the feature toggle details from unleash
func (client *Client) GetFeatureToggleDetails(ctx context.Context, featureName string) (*FeatureEntity, error) {
	if client.unleashClientSDK == nil {
		if err := client.connectUnleashSDK(); err != nil {
			return nil, err
		}
	}
	enabled, err := client.unleashClientSDK.IsFeatureEnabledOnOrganization(featureName, vr.EnvStaging.String(), golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return &FeatureEntity{Name: featureName, Enabled: enabled}, nil
}

// IsFeatureToggleCorrect is used for retrying to toggle Unleash if the toggle is not changed yet
func (client *Client) IsFeatureToggleCorrect(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
	if err := validToggleOption(toggleSelect); err != nil {
		return false, err
	}

	featureEntity, err := client.GetFeatureToggleDetails(ctx, featureName)
	if err != nil {
		return false, err
	}

	tryEnablingAndFeatureIsEnabled := toggleSelect == ToggleChoiceEnable && featureEntity.Enabled
	tryDisablingAndFeatureIsDisabled := toggleSelect == ToggleChoiceDisable && !featureEntity.Enabled
	if tryEnablingAndFeatureIsEnabled || tryDisablingAndFeatureIsDisabled {
		return true, nil
	}

	return false, nil
}
