package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func NewMockRequest(method string, bodyContent interface{}, headers map[string][]string) (*http.Request, io.ReadCloser) {
	byteData, _ := json.Marshal(bodyContent)
	header := http.Header{}
	for k, vs := range headers {
		for _, v := range vs {
			header.Add(k, v)
		}
	}
	expectedBody := io.NopCloser(bytes.NewBuffer(byteData))
	r := &http.Request{
		Method: method,
		Body:   io.NopCloser(bytes.NewBuffer(byteData)),
		Header: header,
	}
	return r, expectedBody
}
