package brightcove

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	clientID            = "client-id"
	secret              = "secret"
	accountID           = "account-id"
	profile             = "profile"
	policyKey           = "policy-key"
	policyKeyWithSearch = "policy-key-with-search"
	videoID             = "12345678"
)

func TestBrightcoveService_GetVideo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case fmt.Sprintf("/playback/v1/accounts/%s/videos/%s", accountID, videoID):
			if r.Header.Get("authorization") != "BCOV-Policy policy-key-with-search" {
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte(`[
  {
  	"error_code": "INVALID_POLICY_KEY",
  	"message": "Request policy key is missing or invalid"
  }
]`))
				return
			}

			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(fmt.Sprintf(`{
  "thumbnail": "https://link/to/some/image.jpg",
  "name": "video-name",
  "duration": 1234,
  "offline_enabled": true,
  "id": "%s"
}`, videoID)))
		default:
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(`[
  {
  	"error_code": "VIDEO_NOT_FOUND",
  	"message": "The designated resource was not found."
  }
]`))
			return
		}
	}))
	defer server.Close()

	s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
	s.PlaybackURL = server.URL + "/playback/v1/accounts/%s/videos/%s"

	testcases := []struct {
		testname     string
		videoID      string
		expectedResp *GetVideoResponse
		expectedErr  error
	}{
		{
			testname: "success",
			videoID:  videoID,
			expectedResp: &GetVideoResponse{
				ID:             videoID,
				Name:           "video-name",
				Thumbnail:      "https://link/to/some/image.jpg",
				Duration:       1234,
				OfflineEnabled: true,
			},
			expectedErr: nil,
		},
		{
			testname:     "invalid video ID",
			videoID:      "invalid-video-id",
			expectedResp: nil,
			expectedErr:  errors.New("expected HTTP status 200, got 404; Brightcove response: [\n  {\n  \t\"error_code\": \"VIDEO_NOT_FOUND\",\n  \t\"message\": \"The designated resource was not found.\"\n  }\n]"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testname, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			resp, err := s.GetVideo(ctx, tc.videoID)
			assert.Equal(t, tc.expectedResp, resp)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestBrightcoveService_CreateVideo(t *testing.T) {
	t.Parallel()
	t.Run("empty request data", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/accounts/account-id/videos/":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				w.WriteHeader(422)
				w.Write([]byte(`[ {
  "error_code" : "VALIDATION_ERROR",
  "message" : "name: REQUIRED_FIELD"
} ]`))
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.CreateVideoURL = server.URL + "/v1/accounts/%s/videos/"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.CreateVideo(ctx, &CreateVideoRequest{
			Name: "",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		assert.Equal(t, `r.doRequest: HTTP code: 422, data: [ {
  "error_code" : "VALIDATION_ERROR",
  "message" : "name: REQUIRED_FIELD"
} ]`, err.Error())
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/accounts/account-id/videos/":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				reqData := new(CreateVideoRequest)
				err := json.NewDecoder(req.Body).Decode(reqData)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
					return
				}

				if reqData.Name != "20200524_123945.mp4" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("wrong name"))
					return
				}

				resp := &CreateVideoResponse{
					ID: "video-id",
				}

				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.CreateVideoURL = server.URL + "/v1/accounts/%s/videos/"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.CreateVideo(ctx, &CreateVideoRequest{
			Name: "20200524_123945.mp4",
		})
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.ID)
	})
}

func TestBrightcoveService_UploadUrls(t *testing.T) {
	t.Parallel()
	t.Run("empty request data", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/accounts/account-id/videos//upload-urls/":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				w.WriteHeader(404)
				w.Write([]byte(`[{"error_code": "RESOURCE_NOT_FOUND"}]`))
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.UploadURLsURL = server.URL + "/v1/accounts/%s/videos/%s/upload-urls/%s"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.UploadUrls(ctx, &UploadUrlsRequest{
			VideoID: "",
			Name:    "",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		assert.Equal(t, `r.doRequest: HTTP code: 404, data: [{"error_code": "RESOURCE_NOT_FOUND"}]`, err.Error())
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/accounts/account-id/videos/video-id/upload-urls/" + url.QueryEscape("20200524_123945.mp4"):
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				resp := &UploadUrlsResponse{
					SignedURL:     "signed-url",
					APIRequestURL: "api-request-url",
				}

				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.UploadURLsURL = server.URL + "/v1/accounts/%s/videos/%s/upload-urls/%s"
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		uploadResp, err := s.UploadUrls(ctx, &UploadUrlsRequest{
			VideoID: "video-id",
			Name:    "20200524_123945.mp4",
		})

		assert.Nil(t, err)
		assert.NotNil(t, uploadResp)
		assert.NotEmpty(t, uploadResp.APIRequestURL)
		assert.NotEmpty(t, uploadResp.SignedURL)
	})
}

func TestBrightcoveService_SubmitDynamicIngress(t *testing.T) {
	t.Parallel()
	t.Run("empty request data", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/accounts/account-id/videos//ingest-requests":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				w.WriteHeader(404)
				w.Write([]byte(`[{"error_code": "RESOURCE_NOT_FOUND"}]`))
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.DynamicIngestURL = server.URL + "/v1/accounts/%s/videos/%s/ingest-requests"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.SubmitDynamicIngress(ctx, &SubmitDynamicIngressRequest{})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		assert.Equal(t, `r.doRequest: HTTP code: 404, data: [{"error_code": "RESOURCE_NOT_FOUND"}]`, err.Error())
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/accounts/account-id/videos/video-id/ingest-requests":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				reqData := new(SubmitDynamicIngressRequest)
				err := json.NewDecoder(req.Body).Decode(reqData)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
					return
				}

				if reqData.Master.URL != "api_request_url" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`missing api_request_url`))
					return
				}

				if reqData.Profile != "multi-platform-standard-static" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`profile: must be multi-platform-standard-static`))
					return
				}

				resp := &SubmitDynamicIngressResponse{
					JobID: "job-id",
				}

				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.DynamicIngestURL = server.URL + "/v1/accounts/%s/videos/%s/ingest-requests"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.SubmitDynamicIngress(ctx, &SubmitDynamicIngressRequest{
			Master: Master{
				URL: "api_request_url",
			},
			Profile:       "multi-platform-standard-static",
			CaptureImages: true,
			VideoID:       "video-id",
		})

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.JobID)
	})
}


func TestBrightcoveService_GetResumePosition(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/xdr/accounts/account-id/playheads/user-id/video-id":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				resp := &GetResumePositionResponse{
					Items: []ResumePositionItem{
						{
							VideoID: "video1",
							Seconds: 10,
						},
					},
				}
	
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
				return
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.ResumePositionURL = server.URL + "/v1/xdr/accounts/%s/playheads/%s/%s"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.GetResumePosition(ctx, &GetResumePositionRequest{
			UserID: "user-id",
			VideoID: "video-id",
		})

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)
	})

	t.Run("empty request data", func(t *testing.T) {
		t.Parallel()
		handler := func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.String() {
			case "/v4/access_token":
				userName, password, ok := req.BasicAuth()
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`missing basic authen`))
					return
				}

				if !(userName == clientID && password == secret) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong clientID or password`))
					return
				}

				resp := &OAuthResponse{
					AccessToken: "access-token",
					TokenType:   "bearer",
					ExpiresIn:   300,
				}
				data, _ := json.Marshal(resp)
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			case "/v1/xdr/accounts/account-id/playheads/user-id/video-id":
				if req.Header.Get("authorization") != "Bearer access-token" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`wrong access_token`))
					return
				}

				w.WriteHeader(404)
				w.Write([]byte(`[{"error_code": "RESOURCE_NOT_FOUND"}]`))
			}
		}

		// Start a local HTTP server
		server := httptest.NewServer(http.HandlerFunc(handler))
		// Close the server when test finishes
		defer server.Close()

		s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
		s.AccessTokenURL = server.URL + "/v4/access_token"
		s.ResumePositionURL = server.URL + "/v1/xdr/accounts/%s/playheads/%s/%s"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := s.GetResumePosition(ctx, &GetResumePositionRequest{
			UserID: "user-id",
			VideoID: "video-id",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		assert.Equal(t, `r.doRequest: HTTP code: 404, data: [{"error_code": "RESOURCE_NOT_FOUND"}]`, err.Error())
	})
}

func TestBrightcoveServiceImpl_GetPolicyKey(t *testing.T) {
	t.Parallel()
	s := NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile)
	assert.NotNil(t, s)
	assert.Equal(t, s.GetPolicyKey(), policyKey)
}
