package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/ernesto-jimenez/httplogger"
	"go.uber.org/zap"
)

type ModuleHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
type HTTPClient struct {
	client ModuleHTTPClient
}

type httpLogger struct {
	log *zap.Logger
}

type HTTPClientInterface interface {
	SendRequest(ctx context.Context, request *http.Request) (*http.Response, error)
}

type HTTPClientConfig struct {
	TimeOut time.Duration
}

type Headers map[string]string
type Request io.Reader
type FRequest func() (*http.Response, error)

func newLogger(log *zap.Logger) *httpLogger {
	return &httpLogger{
		log: log,
	}
}

func (l *httpLogger) LogRequest(req *http.Request) {
	x, _ := httputil.DumpRequest(req, true)
	l.log.Info(fmt.Sprintf("Request: %q", x))
}

func (l *httpLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	if err != nil {
		if req != nil {
			reqByte, err := httputil.DumpRequest(req, true)
			if err != nil {
				l.log.Error(fmt.Sprintf("Request Error: %q", reqByte))
			}
		}
		if res != nil {
			resByte, err := httputil.DumpResponse(res, true)
			if err != nil {
				l.log.Error(fmt.Sprintf("Response Error: %q", resByte))
			}
		}
	} else {
		x, _ := httputil.DumpResponse(res, true)
		l.log.Info(fmt.Sprintf("Res: %q", x))
	}
}

func InitHTTPClient(config *HTTPClientConfig, zLogger *zap.Logger) *HTTPClient {
	client := &HTTPClient{
		client: &http.Client{
			Timeout:   config.TimeOut,
			Transport: httplogger.NewLoggedTransport(http.DefaultTransport, newLogger(zLogger)),
		},
	}
	return client
}

func InitMockHTTPClient(mockHTTP ModuleHTTPClient) *HTTPClient {
	client := &HTTPClient{
		client: mockHTTP,
	}
	return client
}

func (c *HTTPClient) SendRequest(ctx context.Context, request *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(request)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

type RequestInput struct {
	Ctx     context.Context
	Method  string
	URL     string
	Body    Request
	Headers *Headers
}

func createRequest(method string, url string, body Request, headers *Headers) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for key, value := range *headers {
			request.Header.Set(key, value)
		}
	}

	return request, nil
}

func HandleHTTPRequest[T any](s HTTPClientInterface, input *RequestInput) (*T, error) {
	req, err := createRequest(input.Method, input.URL, input.Body, input.Headers)
	if err != nil {
		return nil, err
	}
	resp, err := s.SendRequest(input.Ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return nil, nil
	}
	data, err := HandleHTTPResponse[T](resp)
	if err != nil {
		return nil, err
	}
	return data, err
}

func HandleHTTPResponse[T any](response *http.Response) (*T, error) {
	var data T
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func IsHeaderAllowed(s string) (string, bool) {
	var allowedHeaders = map[string]struct{}{
		"x-request-id": {},
	}
	// check if allowedHeaders contain the header
	if _, isAllowed := allowedHeaders[s]; isAllowed {
		return strings.ToUpper(s), true
	}
	// if not in the allowed header, don't send the header
	return s, false
}
