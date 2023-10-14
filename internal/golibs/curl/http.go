package curl

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/try"
)

type Method int

const (
	GET Method = iota
	POST
	PUT
	DELETE
)

type IHTTP interface {
	Request(method Method, url string, header map[string]string, data io.Reader, dest interface{}) error
}

type HTTP struct {
	InsecureSkipVerify bool
}

func (h *HTTP) Request(method Method, url string, header map[string]string, data io.Reader, dest interface{}) error {
	methodStr := ""
	switch method {
	case GET:
		methodStr = "GET"
	case POST:
		methodStr = "POST"
	case PUT:
		methodStr = "PUT"
	case DELETE:
		methodStr = "DELETE"
	default:
		return fmt.Errorf("HttpRequest invalid method")
	}

	req, err := http.NewRequest(methodStr, url, data)
	if err != nil {
		return fmt.Errorf("HttpRequest.NewRequest: %w", err)
	}

	for key, element := range header {
		req.Header.Set(key, element)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: h.InsecureSkipVerify, //nolint:gosec
			},
		},
	}

	if err := try.Do(func(attempt int) (bool, error) {
		isRetryable := attempt < 5

		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Errorf("HttpRequest.Do: %w", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("HttpRequest.ReadAll: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			if dest != nil {
				if err := json.Unmarshal(body, dest); err != nil {
					return false, fmt.Errorf("HttpRequest.Unmarshal: %w", err)
				}
			}
			return false, nil
		}

		return isRetryable, fmt.Errorf("error: %w", errors.New(string(body)))
	}); err != nil {
		return err
	}

	return nil
}
