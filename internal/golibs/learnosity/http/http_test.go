package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/learnosity"

	"github.com/stretchr/testify/assert"
)

func TestClient_RequestSuccess(t *testing.T) {
	t.Parallel()

	client := &Client{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status_code": 200}`))
		return
	}))
	defer mockServer.Close()

	type SampleResponse struct {
		StatusCode int `json:"status_code"`
	}

	holder := &SampleResponse{}
	_ = client.Request(context.Background(), learnosity.MethodGet, mockServer.URL, nil, nil, holder)
	assert.Equal(t, 200, holder.StatusCode)
}

func TestClient_Request(t *testing.T) {
	t.Parallel()

	client := &Client{}

	testCases := []struct {
		Name        string
		Ctx         context.Context
		Method      learnosity.Method
		URL         string
		Header      map[string]string
		Body        io.Reader
		Holder      any
		ExpectedErr error
	}{
		{
			Name:   "error: 405 method not allowed",
			Ctx:    context.Background(),
			Method: learnosity.Method("method"),
			URL:    "https://www.google.com",
			Header: map[string]string{
				"Content-Type": "application/json",
			},
			Body:        nil,
			Holder:      nil,
			ExpectedErr: fmt.Errorf("try.Do: %w", fmt.Errorf("http request failed with status code %d", 405)),
		},
		{
			Name:   "error: json.Decode",
			Ctx:    context.Background(),
			Method: learnosity.MethodGet,
			URL:    "https://www.google.com",
			Header: map[string]string{
				"Content-Type": "application/json",
			},
			Body:        nil,
			Holder:      &struct{}{},
			ExpectedErr: fmt.Errorf("try.Do: %w", fmt.Errorf("json.Decode: invalid character")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := client.Request(tc.Ctx, tc.Method, tc.URL, tc.Header, tc.Body, tc.Holder)
			var tt *json.SyntaxError
			if errors.As(err, &tt) {
				assert.Contains(t, err.Error(), tc.ExpectedErr.Error())
			} else if err != nil {
				assert.Equal(t, tc.ExpectedErr, err)
			}
		})
	}
}
