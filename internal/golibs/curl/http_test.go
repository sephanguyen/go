package curl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestHTTP_Error(t *testing.T) {
	t.Parallel()
	http := &HTTP{}

	type Param struct {
		Method Method
		Url    string
		Header map[string]string
		Data   io.Reader
		Dest   interface{}
	}

	req := &Param{
		Method: 10,
		Url:    "",
		Header: map[string]string{"test": "test"},
		Data:   nil,
		Dest:   nil,
	}

	testCases := []TestCase{
		{
			name:        "error wrong method",
			req:         req,
			expectedErr: fmt.Errorf("HttpRequest invalid method"),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		params := testCase.req.(*Param)
		err := http.Request(params.Method, params.Url, params.Header, params.Data, params.Dest)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestHTTP_Request(t *testing.T) {
	t.Parallel()

	type SampleResponse struct {
		StatusCode int `json:"status_code"`
	}

	client := &HTTP{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status_code": 200}`))
		return
	}))
	defer ts.Close()

	resp := &SampleResponse{}
	err := client.Request(GET, ts.URL, nil, nil, resp)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
