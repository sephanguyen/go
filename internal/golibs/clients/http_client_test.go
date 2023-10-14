package clients

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockHttpClient struct {
	mockDo func(req *http.Request) (*http.Response, error)
}

func (client *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.mockDo(req)
}

func TestHTTPClient_SendRequest(t *testing.T) {

	t.Run("send Request success when not add headers", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient := &MockHttpClient{
			func(req *http.Request) (*http.Response, error) {
				response := &http.Response{
					StatusCode: http.StatusOK,
				}
				return response, nil
			},
		}

		client := InitMockHTTPClient(mockHTTPClient)

		res, err := client.SendRequest(ctx, &http.Request{})
		assert.NoError(t, err)
		assert.NotNil(t, res)

	})

	t.Run("send Request success when add headers", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient := &MockHttpClient{
			func(req *http.Request) (*http.Response, error) {
				response := &http.Response{
					StatusCode: http.StatusOK,
				}
				return response, nil
			},
		}

		client := InitMockHTTPClient(mockHTTPClient)
		headers := make(Headers)
		headers["User-Agent"] = "Zoom-api-Jwt-Request"
		headers["content-type"] = "application/json"

		res, err := client.SendRequest(ctx, &http.Request{})
		assert.NoError(t, err)
		assert.NotNil(t, res)

	})

	t.Run("should throw error when send request fail", func(t *testing.T) {
		ctx := context.Background()

		mockHTTPClient := &MockHttpClient{
			func(req *http.Request) (*http.Response, error) {

				return nil, fmt.Errorf("request fail")
			},
		}

		client := InitMockHTTPClient(mockHTTPClient)

		res, err := client.SendRequest(ctx, &http.Request{})
		assert.Error(t, err)
		assert.Nil(t, res)

	})
}
