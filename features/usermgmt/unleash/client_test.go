package unleash

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/square/go-jose/v3/json"
	"github.com/stretchr/testify/assert"
)

type MockHttpClient struct {
	mockDo func(req *http.Request) (*http.Response, error)
}

func (client *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.mockDo(req)
}

const (
	testUnleashSrvAddr          = "https://unleash.com"
	testUnleashAPIKey           = "example-api-key"
	testUnleashLocalAdminAPIKey = "example-local-admin-api-key"
)

func TestClient_ToggleUnleashFeatureWithName(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	testcases := []struct {
		name           string
		initClient     func(t *testing.T, ctx context.Context) *Client
		input          ToggleChoice
		expectedResult error
	}{
		{
			name: "call unleash to toggle successfully",
			initClient: func(t *testing.T, ctx context.Context) *Client {
				unleashClient := NewDefaultClient(testUnleashSrvAddr, testUnleashAPIKey, testUnleashLocalAdminAPIKey)

				unleashClient.httpClient = &MockHttpClient{
					func(req *http.Request) (*http.Response, error) {
						response := &http.Response{
							Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("fake data"))),
							StatusCode: http.StatusOK,
						}
						return response, nil
					},
				}
				return unleashClient
			},
			input:          ToggleChoiceEnable,
			expectedResult: nil,
		},
		{
			name: "call unleash to toggle successfully",
			initClient: func(t *testing.T, ctx context.Context) *Client {
				unleashClient := NewDefaultClient(testUnleashSrvAddr, testUnleashAPIKey, testUnleashLocalAdminAPIKey)

				unleashClient.httpClient = &MockHttpClient{
					func(req *http.Request) (*http.Response, error) {
						response := &http.Response{
							Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("fake data"))),
							StatusCode: http.StatusBadRequest,
						}
						return response, assert.AnError
					},
				}
				return unleashClient
			},
			input:          ToggleChoiceEnable,
			expectedResult: errors.Wrap(assert.AnError, "error requesting to unleash"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			unleashClient := testcase.initClient(t, ctx)
			err := unleashClient.ToggleUnleashFeatureWithName(ctx, TestFeatureName, testcase.input)
			if testcase.expectedResult == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, testcase.expectedResult.Error(), err.Error())
			}
		})
	}
}

func bodyFromUnleashFeatureEntity(t *testing.T, unleashFeatureEntity *FeatureEntity) io.ReadCloser {
	data, err := json.Marshal(unleashFeatureEntity)
	assert.NoError(t, err)
	return io.NopCloser(bytes.NewBuffer(data))
}

func TestClient_IsFeatureToggleCorrect(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	testcases := []struct {
		name           string
		initClient     func(t *testing.T, ctx context.Context) *Client
		input          ToggleChoice
		expectedResult bool
	}{
		{
			name: "check if feature toggle is enable, unleash return enabled",
			initClient: func(t *testing.T, ctx context.Context) *Client {
				unleashClient := NewDefaultClient(testUnleashSrvAddr, testUnleashAPIKey, testUnleashLocalAdminAPIKey)

				unleashClient.httpClient = &MockHttpClient{
					func(req *http.Request) (*http.Response, error) {
						dto := &FeatureEntity{
							Enabled: true,
						}
						response := &http.Response{
							Body:       bodyFromUnleashFeatureEntity(t, dto),
							StatusCode: http.StatusOK,
						}
						return response, nil
					},
				}
				return unleashClient
			},
			input:          ToggleChoiceEnable,
			expectedResult: true,
		},
		{
			name: "check if feature toggle is disable, unleash return disable",
			initClient: func(t *testing.T, ctx context.Context) *Client {
				unleashClient := NewDefaultClient(testUnleashSrvAddr, testUnleashAPIKey, testUnleashLocalAdminAPIKey)

				unleashClient.httpClient = &MockHttpClient{
					func(req *http.Request) (*http.Response, error) {
						dto := &FeatureEntity{
							Enabled: false,
						}
						response := &http.Response{
							Body:       bodyFromUnleashFeatureEntity(t, dto),
							StatusCode: http.StatusOK,
						}
						return response, nil
					},
				}
				return unleashClient
			},
			input:          ToggleChoiceDisable,
			expectedResult: true,
		},
		{
			name: "check if feature toggle is enable, unleash return disable",
			initClient: func(t *testing.T, ctx context.Context) *Client {
				unleashClient := NewDefaultClient(testUnleashSrvAddr, testUnleashAPIKey, testUnleashLocalAdminAPIKey)

				unleashClient.httpClient = &MockHttpClient{
					func(req *http.Request) (*http.Response, error) {
						dto := &FeatureEntity{
							Enabled: false,
						}
						response := &http.Response{
							Body:       bodyFromUnleashFeatureEntity(t, dto),
							StatusCode: http.StatusOK,
						}
						return response, nil
					},
				}
				return unleashClient
			},
			input:          ToggleChoiceEnable,
			expectedResult: false,
		},
		{
			name: "check if feature toggle is disable, unleash return enabled",
			initClient: func(t *testing.T, ctx context.Context) *Client {
				unleashClient := NewDefaultClient(testUnleashSrvAddr, testUnleashAPIKey, testUnleashLocalAdminAPIKey)

				unleashClient.httpClient = &MockHttpClient{
					func(req *http.Request) (*http.Response, error) {
						dto := &FeatureEntity{
							Enabled: true,
						}
						response := &http.Response{
							Body:       bodyFromUnleashFeatureEntity(t, dto),
							StatusCode: http.StatusOK,
						}
						return response, nil
					},
				}
				return unleashClient
			},
			input:          ToggleChoiceDisable,
			expectedResult: false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			unleashClient := testcase.initClient(t, ctx)
			correct, err := unleashClient.IsFeatureToggleCorrect(ctx, TestFeatureName, testcase.input)
			assert.Equal(t, testcase.expectedResult, correct)
			assert.NoError(t, err)
		})
	}

}
