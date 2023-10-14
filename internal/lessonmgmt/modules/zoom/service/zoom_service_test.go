package service

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mock_service "github.com/manabie-com/backend/mock/lessonmgmt/zoom/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestZoomService_GenerateZoomLink(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}

	s := InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	organizationID := "id"
	tc := []TestCase{
		{
			name:        "should generate zoom link success",
			ctx:         auth.InjectFakeJwtToken(ctx, organizationID),
			req:         nil,
			expectedErr: nil,
			expectedResp: &domain.GenerateZoomLinkResponse{
				URL: "URL", ZoomID: 1,
			},
			setup: func(ctx context.Context) {

				mockExternalConfigService.On("GetConfigByResource", ctx).
					Return(&domain.ZoomConfig{AccountID: "AccountID", ClientID: "ClientID", ClientSecret: "ClientSecret"}, nil).Once()

				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, nil).Once()

				getGenerateLinkResponse := `{"uuid": "uuid", "id": 1, "join_url" : "URL"}`

				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(getGenerateLinkResponse)))}, nil).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.RetryGenerateZoomLink(testCase.ctx, "AccountID", &domain.ZoomGenerateMeetingRequest{})
			if testCase.expectedErr != nil {
				fmt.Println("err:", err)

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestZoomService_GenerateMultiZoomLink(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}

	s := InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	organizationID := "id"
	tc := []TestCase{
		{
			name: "should generate zoom link success",
			ctx:  auth.InjectFakeJwtToken(ctx, organizationID),
			req: []*domain.ZoomGenerateMeetingRequest{
				{
					Topic: "",
				},
			},
			expectedErr: nil,
			expectedResp: []*domain.GenerateZoomLinkResponse{
				{
					URL:    "URL",
					ZoomID: 1,
				},
			},
			setup: func(ctx context.Context) {

				mockExternalConfigService.On("GetConfigByResource", ctx).
					Return(&domain.ZoomConfig{AccountID: "AccountID", ClientID: "ClientID", ClientSecret: "ClientSecret"}, nil).Once()

				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, nil).Once()

				getGenerateLinkResponse := `{"uuid": "uuid", "id": 1, "join_url" : "URL"}`

				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(getGenerateLinkResponse)))}, nil).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.([]*domain.ZoomGenerateMeetingRequest)
			resp, err := s.GenerateMultiZoomLink(testCase.ctx, "AccountID", req)
			if testCase.expectedErr != nil {
				fmt.Println("err:", err)

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestZoomService_RetryDeleteZoomLink(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockExternalConfigService := &mock_service.MockExternalConfigService{}
	mockHTTPClient := &mock_clients.MockHTTPClient{}
	zcf := &configs.ZoomConfig{}

	s := InitZoomService(zcf, mockExternalConfigService, mockHTTPClient)
	organizationID := "id"
	tc := []TestCase{
		{
			name:         "should delete zoom link success",
			ctx:          auth.InjectFakeJwtToken(ctx, organizationID),
			req:          "zoomID",
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {

				mockExternalConfigService.On("GetConfigByResource", ctx).
					Return(&domain.ZoomConfig{AccountID: "AccountID", ClientID: "ClientID", ClientSecret: "ClientSecret"}, nil).Once()

				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, nil).Once()

				deleteZoomLinkResponse := `{"uuid": "uuid", "id": 1, "join_url" : "URL"}`

				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(deleteZoomLinkResponse)))}, nil).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			zoomID := testCase.req.(string)
			resp, err := s.RetryDeleteZoomLink(testCase.ctx, zoomID)
			if testCase.expectedErr != nil {
				fmt.Println("err:", err)

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
