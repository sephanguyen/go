package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSInternalServerError(t *testing.T) {
	expectedErr := "expected error"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
	w := httptest.NewRecorder()
	internalServerError(w, expectedErr)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestErrorRespond(t *testing.T) {
	expectedErr := "expected error"
	errStatusCode := http.StatusNotImplemented
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
	w := httptest.NewRecorder()
	errorResponse(w, expectedErr, expectedErr, errStatusCode)

	assert.Equal(t, errStatusCode, w.Code)
}
