package salesforce

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock HTTP Client for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestSFClientImpl_Query(t *testing.T) {
	t.Run("happy case: query success", func(t *testing.T) {

		httpClient := new(MockHTTPClient)
		client := &SFClientImpl{
			HTTPClient: httpClient,
			Token:      "testToken",
			Endpoint:   NewEndpoint(),
		}

		query := "SELECT Id, Name FROM Account"
		expectedQueryResponse := QueryResponse[map[string]interface{}]{
			Done:           true,
			TotalSize:      1,
			NextRecordsURL: "link",
			Records: []map[string]interface{}{
				{
					"Id":   "001",
					"Name": "Test Account",
				},
			},
		}
		responseBody := []byte(`{
			"done": true,
			"totalSize": 1,
			"records": [
				{
					"Id": "001",
					"Name": "Test Account"
				}
			],
			"nextRecordsUrl": "link"
		}`)
		response := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		httpClient.On("Do", mock.Anything).Return(response, nil)

		var queryResponse QueryResponse[map[string]interface{}]
		err := client.Query(query, &queryResponse)

		assert.NoError(t, err)
		assert.Equal(t, expectedQueryResponse, queryResponse)
		httpClient.AssertExpectations(t)

	})

	t.Run("error case: query error", func(t *testing.T) {
		httpClient := new(MockHTTPClient)
		client := &SFClientImpl{
			HTTPClient: httpClient,
			Token:      "testToken",
			Endpoint:   NewEndpoint(),
		}

		query := "SELECT Id, Name FROM Account"
		errorResponse := SFError{Message: "Invalid query", ErrorCode: "MALFORMED_QUERY"}
		responseBody, _ := json.Marshal([]SFError{errorResponse})
		response := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		httpClient.On("Do", mock.Anything).Return(response, nil)

		var queryResponse QueryResponse[map[string]interface{}]
		err := client.Query(query, &queryResponse)

		assert.Error(t, err)
		assert.Equal(t, errorResponse.Error(), err.Error())
		httpClient.AssertExpectations(t)
	})
}

func TestSFClientImpl_Post(t *testing.T) {
	t.Run("happy case: post success", func(t *testing.T) {
		httpClient := new(MockHTTPClient)
		client := &SFClientImpl{
			HTTPClient: httpClient,
			Token:      "testToken",
			Endpoint:   NewEndpoint(),
		}

		object := "Account"
		requestBody := map[string]interface{}{"Name": "Test Account"}
		expectedResponse := SuccessResponse{ID: "001", Success: true}
		responseBody, _ := json.Marshal(expectedResponse)
		response := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		httpClient.On("Do", mock.Anything).Return(response, nil)

		resp, err := client.Post(object, requestBody)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		httpClient.AssertExpectations(t)
	})
	t.Run("error case: post error", func(t *testing.T) {
		httpClient := new(MockHTTPClient)
		client := &SFClientImpl{
			HTTPClient: httpClient,
			Token:      "testToken",
			Endpoint:   NewEndpoint(),
		}

		object := "Account"
		requestBody := map[string]interface{}{"Name": "Test Account"}
		errorResponse := SFError{Message: "Invalid input", ErrorCode: "INVALID_INPUT"}
		responseBody, _ := json.Marshal([]SFError{errorResponse})
		response := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		httpClient.On("Do", mock.Anything).Return(response, nil)

		_, err := client.Post(object, requestBody)

		assert.Error(t, err)
		assert.Equal(t, errorResponse.Error(), err.Error())
		httpClient.AssertExpectations(t)
	})
}

func TestSFClientImpl_Post_MarshalError(t *testing.T) {
	httpClient := new(MockHTTPClient)
	client := &SFClientImpl{
		HTTPClient: httpClient,
		Token:      "testToken",
		Endpoint:   NewEndpoint(),
	}

	object := "Account"
	requestBody := make(chan int) // Invalid type for JSON marshaling

	_, err := client.Post(object, requestBody)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json: unsupported type: chan int")
	httpClient.AssertExpectations(t)
}
