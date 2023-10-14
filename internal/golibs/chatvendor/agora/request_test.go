package agora

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/chatvendor/agora/dto"
	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/stretchr/testify/assert"
)

func Test_doRequest(t *testing.T) {
	type TestResOK struct {
		Res string `json:"res"`
	}

	testEndpoint := "/tests"

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get(AuthHeaderKey))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestResOK{Res: "OK"})
		}))
		defer ts.Close()

		agoraClient := &agoraClientImpl{
			AgoraConfig: configs.AgoraConfig{
				AppID:              DataMockAppID,
				PrimaryCertificate: DataMockAppCertificate,
				AppName:            DataMockAppName,
				OrgName:            DataMockOrgName,
				RestAPI:            ts.URL,
			},
			httpClient: &http.Client{},
		}

		appToken, _ := agoraClient.GetAppToken()
		header := map[string]string{
			AuthHeaderKey: appToken,
		}

		res := &TestResOK{}
		err := agoraClient.doRequest(context.Background(), MethodGet, testEndpoint, header, nil, res)
		assert.Equal(t, "OK", res.Res)
		assert.Equal(t, nil, err)
	})

	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// checking auto fill token if empty
			assert.NotEmpty(t, r.Header.Get(AuthHeaderKey))

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestResOK{Res: "OK"})
		}))
		defer ts.Close()

		agoraClient := &agoraClientImpl{
			AgoraConfig: configs.AgoraConfig{
				AppID:              DataMockAppID,
				PrimaryCertificate: DataMockAppCertificate,
				AppName:            DataMockAppName,
				OrgName:            DataMockOrgName,
				RestAPI:            ts.URL,
			},
			httpClient: &http.Client{},
		}

		res := &TestResOK{}
		err := agoraClient.doRequest(context.Background(), MethodGet, testEndpoint, make(map[string]string), nil, res)
		assert.Equal(t, "OK", res.Res)
		assert.Equal(t, nil, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get(AuthHeaderKey))
			if r.Header.Get(AuthHeaderKey) == "Bearer unauthorized_token" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(dto.ErrorResponse{
					Exception:        "Exception",
					Error:            "unauthorized",
					ErrorDescription: "unauthorized",
					Duration:         1,
					Timestamp:        1,
				})
			} else {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(TestResOK{Res: "OK"})
			}
		}))
		defer ts.Close()

		agoraClient := &agoraClientImpl{
			AgoraConfig: configs.AgoraConfig{
				AppID:              DataMockAppID,
				PrimaryCertificate: DataMockAppCertificate,
				AppName:            DataMockAppName,
				OrgName:            DataMockOrgName,
				RestAPI:            ts.URL,
			},
			httpClient: &http.Client{},
		}

		// make unauthorized_token
		header := map[string]string{
			AuthHeaderKey: "Bearer unauthorized_token",
		}

		res := &TestResOK{}
		err := agoraClient.doRequest(context.Background(), MethodGet, testEndpoint, header, nil, res)
		assert.Equal(t, "OK", res.Res)
		assert.Equal(t, nil, err)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get(AuthHeaderKey))
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(dto.ErrorResponse{
				Exception:        "Exception",
				Error:            "not found",
				ErrorDescription: "not found",
				Duration:         1,
				Timestamp:        1,
			})
		}))
		defer ts.Close()

		agoraClient := &agoraClientImpl{
			AgoraConfig: configs.AgoraConfig{
				AppID:              DataMockAppID,
				PrimaryCertificate: DataMockAppCertificate,
				AppName:            DataMockAppName,
				OrgName:            DataMockOrgName,
				RestAPI:            ts.URL,
			},
			httpClient: &http.Client{},
		}

		res := &TestResOK{}
		err := agoraClient.doRequest(context.Background(), MethodGet, testEndpoint, nil, nil, res)
		assert.Empty(t, res.Res)
		assert.Equal(t, nil, err)
	})
}
