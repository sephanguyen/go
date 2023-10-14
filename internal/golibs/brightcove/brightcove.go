package brightcove

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

const (
	AccessTokenURL    = "https://oauth.brightcove.com/v4/access_token"                               //nolint:gosec
	CreateVideoURL    = "https://cms.api.brightcove.com/v1/accounts/%s/videos/"                      // accountID
	UploadURLsURL     = "https://cms.api.brightcove.com/v1/accounts/%s/videos/%s/upload-urls/%s"     // accountID, videoID, sourceName
	DynamicIngestURL  = "https://ingest.api.brightcove.com/v1/accounts/%s/videos/%s/ingest-requests" // accountID, videoID
	PlaybackURL       = "https://edge.api.brightcove.com/playback/v1/accounts/%s/videos/%s"          // accountID, videoID
	ResumePositionURL = "https://data.brightcove.com/v1/xdr/accounts/%s/playheads/%s/%s"             // accountID, userID, videoID
)

type ExternalService interface {
	GetAccountID() string
	GetPolicyKey() string
	GetProfile() string
	GetVideo(ctx context.Context, videoID string) (*GetVideoResponse, error)
	CreateVideo(ctx context.Context, req *CreateVideoRequest) (*CreateVideoResponse, error)
	UploadUrls(ctx context.Context, req *UploadUrlsRequest) (*UploadUrlsResponse, error)
	SubmitDynamicIngress(ctx context.Context, req *SubmitDynamicIngressRequest) (*SubmitDynamicIngressResponse, error)
	GetResumePosition(ctx context.Context, req *GetResumePositionRequest) (*GetResumePositionResponse, error)
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type CreateVideoRequest struct {
	Name string `json:"name"`
}

type CreateVideoResponse struct {
	ID string `json:"id"`
}

type UploadUrlsRequest struct {
	VideoID string `json:"-"`
	Name    string `json:"-"`
}

type UploadUrlsResponse struct {
	SignedURL     string `json:"signed_url"`
	APIRequestURL string `json:"api_request_url"`
}

type Master struct {
	URL string `json:"url"`
}

type SubmitDynamicIngressRequest struct {
	Master        Master `json:"master"`
	Profile       string `json:"profile"`
	CaptureImages bool   `json:"capture-images"`
	VideoID       string `json:"-"`
}

type SubmitDynamicIngressResponse struct {
	JobID string `json:"id"`
}

type GetResumePositionRequest struct {
	UserID  string `json:"-"`
	VideoID string `json:"-"`
}

type GetResumePositionResponse struct {
	Items []ResumePositionItem `json:"items"`
}

type ResumePositionItem struct {
	VideoID string `json:"video_id"`
	Seconds int    `json:"playhead_seconds"`
}

type ServiceImpl struct {
	AccessTokenURL      string
	CreateVideoURL      string
	UploadURLsURL       string
	DynamicIngestURL    string
	PlaybackURL         string
	ResumePositionURL   string
	ClientID            string
	Secret              string
	PolicyKey           string
	PolicyKeyWithSearch string
	Profile             string
	AccountID           string

	httpClient     *http.Client
	requestGroup   singleflight.Group
	accessToken    string
	tokenExpiredAt time.Time
}

func NewBrightcoveService(clientID, secret, accountID, policyKey, policyKeyWithSearch, profile string) *ServiceImpl {
	return &ServiceImpl{
		AccessTokenURL:      AccessTokenURL,
		CreateVideoURL:      CreateVideoURL,
		UploadURLsURL:       UploadURLsURL,
		DynamicIngestURL:    DynamicIngestURL,
		PlaybackURL:         PlaybackURL,
		ResumePositionURL:   ResumePositionURL,
		ClientID:            clientID,
		Secret:              secret,
		PolicyKey:           policyKey,
		PolicyKeyWithSearch: policyKeyWithSearch,
		Profile:             profile,
		AccountID:           accountID,

		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		requestGroup: singleflight.Group{},
	}
}

func (r *ServiceImpl) GetAccountID() string {
	return r.AccountID
}

func (r *ServiceImpl) GetPolicyKey() string {
	return r.PolicyKey
}

func (r *ServiceImpl) GetProfile() string {
	return r.Profile
}

type GetVideoResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Thumbnail      string `json:"thumbnail"`
	Duration       int64  `json:"duration"` // in milliseconds, ref: https://apis.support.brightcove.com/cms/references/cms-api-video-fields-reference.html
	OfflineEnabled bool   `json:"offline_enabled"`
}

func (r *ServiceImpl) GetVideo(ctx context.Context, videoID string) (*GetVideoResponse, error) {
	url := fmt.Sprintf(r.PlaybackURL, r.AccountID, videoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("BCOV-Policy %s", r.PolicyKeyWithSearch))

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("expected HTTP status 200, got %d; failed to extract Brightcove response: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("expected HTTP status 200, got %d; Brightcove response: %s", resp.StatusCode, string(body))
	}

	getVideoResp := &GetVideoResponse{}
	err = json.NewDecoder(resp.Body).Decode(getVideoResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return getVideoResp, nil
}

func (r *ServiceImpl) getOAuthToken(ctx context.Context) (string, error) {
	if r.accessToken != "" && time.Now().Before(r.tokenExpiredAt.Add(-10*time.Second)) {
		return r.accessToken, nil
	}

	body := url.Values{}
	body.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, r.AccessTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "Cannot connect to brightcoveService")
	}

	req = req.WithContext(ctx)
	req.SetBasicAuth(r.ClientID, r.Secret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "Cannot grant brightcoveService access")
	}
	defer resp.Body.Close()

	oauthResp := &OAuthResponse{}

	err = json.NewDecoder(resp.Body).Decode(oauthResp)
	if err != nil {
		return "", errors.Wrap(err, "brightcoveService.DecodeJson")
	}

	r.accessToken = oauthResp.AccessToken
	r.tokenExpiredAt = time.Now().Add(time.Duration(oauthResp.ExpiresIn) * time.Second)

	return r.accessToken, nil
}

func (r *ServiceImpl) CreateVideo(ctx context.Context, req *CreateVideoRequest) (*CreateVideoResponse, error) {
	url := fmt.Sprintf(r.CreateVideoURL, r.AccountID)
	data, err := r.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, errors.Wrap(err, "r.doRequest")
	}

	resp := new(CreateVideoResponse)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (r *ServiceImpl) UploadUrls(ctx context.Context, req *UploadUrlsRequest) (*UploadUrlsResponse, error) {
	req.Name = url.QueryEscape(req.Name)

	url := fmt.Sprintf(r.UploadURLsURL, r.AccountID, req.VideoID, req.Name)
	data, err := r.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "r.doRequest")
	}

	resp := new(UploadUrlsResponse)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (r *ServiceImpl) SubmitDynamicIngress(ctx context.Context, req *SubmitDynamicIngressRequest) (*SubmitDynamicIngressResponse, error) {
	url := fmt.Sprintf(r.DynamicIngestURL, r.AccountID, req.VideoID)
	data, err := r.doRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, errors.Wrap(err, "r.doRequest")
	}

	resp := new(SubmitDynamicIngressResponse)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (r *ServiceImpl) GetResumePosition(ctx context.Context, req *GetResumePositionRequest) (*GetResumePositionResponse, error) {
	url := fmt.Sprintf(r.ResumePositionURL, r.AccountID, req.UserID, req.VideoID)
	data, err := r.doRequest(ctx, http.MethodGet, url, req)
	if err != nil {
		return nil, errors.Wrap(err, "r.doRequest")
	}

	resp := new(GetResumePositionResponse)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return resp, nil
}

func (r *ServiceImpl) doRequest(ctx context.Context, method, url string, req interface{}) ([]byte, error) {
	_, err, _ := r.requestGroup.Do("get-oauth-token", func() (interface{}, error) {
		return r.getOAuthToken(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("get OAuth token: %w", err)
	}

	reqBody := new(bytes.Buffer)
	if req != nil {
		err = json.NewEncoder(reqBody).Encode(req)
		if err != nil {
			return nil, errors.Wrap(err, "json Encode")
		}
	}

	request, err := http.NewRequest(
		method,
		url,
		reqBody,
	)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest")
	}

	request = request.WithContext(ctx)
	request.Header.Add("authorization", fmt.Sprintf("Bearer %s", r.accessToken))
	request.Header.Add("content-type", "application/json")

	resp, err := r.httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "client.Do")
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "json.Decode")
	}

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		return nil, errors.Errorf("HTTP code: %d, data: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, err
}
