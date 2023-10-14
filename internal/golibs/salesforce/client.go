package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type QueryResponse[T any] struct {
	Done           bool   `json:"done"`
	NextRecordsURL string `json:"nextRecordsUrl"`
	Records        []T    `json:"records"`
	TotalSize      int    `json:"totalSize"`
}

type SuccessResponse struct {
	ID      string        `json:"id"`
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
}

type SFError struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

func (e SFError) Error() string {
	return fmt.Sprintf("salesforce message: %s, error code: %s", e.Message, e.ErrorCode)
}

type SFClient interface {
	Query(query string, queryResponse interface{}) error
	Post(object string, body interface{}) (SuccessResponse, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type SFClientImpl struct {
	HTTPClient HTTPClient
	Token      string
	Endpoint   Endpoint
}

func NewClient(_ context.Context) (SFClient, error) {
	token := "" // TODO: Should impl auth here
	httpClient := &http.Client{}
	endpoint := NewEndpoint()

	return &SFClientImpl{
		HTTPClient: httpClient,
		Endpoint:   endpoint,
		Token:      token,
	}, nil
}

func (s *SFClientImpl) HandleError(resp *http.Response) ([]byte, error) {
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var errorJSON []SFError
		if err = json.Unmarshal(payload, &errorJSON); err != nil {
			return nil, err
		}
		return nil, errorJSON[0]
	}
	return payload, nil
}

func (s *SFClientImpl) MakeHTTPRequest(method string, endPoint string, bodyReq io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, endPoint, bodyReq)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+s.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	payload, err := s.HandleError(resp)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *SFClientImpl) Query(query string, queryResponse interface{}) error {
	values := url.Values{}
	values.Set("q", query)
	endpoint := s.Endpoint.GetQueryEndPoint(values)

	payload, err := s.MakeHTTPRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, queryResponse)
}

func (s *SFClientImpl) Post(object string, body interface{}) (SuccessResponse, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return SuccessResponse{}, err
	}
	endpoint := s.Endpoint.GetObjectEndPoint(object)

	payload, err := s.MakeHTTPRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestBody))

	if err != nil {
		return SuccessResponse{}, err
	}
	var resp SuccessResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return SuccessResponse{}, err
	}
	return resp, nil
}
