package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
	config       configs.AppsmithAPI
}

func TestAppsmithCommandHandler_PullMetadata(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockHTTPClient := &mock_clients.MockHTTPClient{}
	command := AppsmithCommandHandler{
		nil, nil, mockHTTPClient,
	}
	organizationID := "id"
	tc := []TestCase{
		{
			name: "should pull metadata success",
			ctx:  auth.InjectFakeJwtToken(ctx, organizationID),
			config: configs.AppsmithAPI{
				ENDPOINT:      "https://appsmith.staging-green.manabie.io/api/v1",
				ApplicationID: "app_id",
				Authorization: "abc",
			},
			req:          nil,
			expectedErr:  nil,
			expectedResp: &domain.AppsmithResponse{},
			setup: func(ctx context.Context) {
				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, nil).Once()
			},
		},
		{
			name: "should pullmetadata error",
			ctx:  auth.InjectFakeJwtToken(ctx, organizationID),
			config: configs.AppsmithAPI{
				ENDPOINT:      "https://appsmith.staging-green.manabie.io/api/v1",
				ApplicationID: "app_id",
				Authorization: "abc",
			},
			req:          nil,
			expectedErr:  fmt.Errorf("internal err"),
			expectedResp: &domain.AppsmithResponse{},
			setup: func(ctx context.Context) {
				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, fmt.Errorf("internal err")).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := command.PullMetadata(testCase.ctx, "staging", testCase.config)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestAppsmithCommandHandler_DiscardMetadata(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockHTTPClient := &mock_clients.MockHTTPClient{}
	command := AppsmithCommandHandler{
		nil, nil, mockHTTPClient,
	}
	organizationID := "id"
	tc := []TestCase{
		{
			name: "should discard metadata success",
			ctx:  auth.InjectFakeJwtToken(ctx, organizationID),
			config: configs.AppsmithAPI{
				ENDPOINT:      "https://appsmith.staging-green.manabie.io/api/v1",
				ApplicationID: "app_id",
				Authorization: "abc",
			},
			req:          nil,
			expectedErr:  nil,
			expectedResp: &domain.AppsmithResponse{},
			setup: func(ctx context.Context) {
				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, nil).Once()
			},
		},
		{
			name: "should discard metadata error",
			ctx:  auth.InjectFakeJwtToken(ctx, organizationID),
			config: configs.AppsmithAPI{
				ENDPOINT:      "https://appsmith.staging-green.manabie.io/api/v1",
				ApplicationID: "app_id",
				Authorization: "abc",
			},
			req:          nil,
			expectedErr:  fmt.Errorf("internal err"),
			expectedResp: &domain.AppsmithResponse{},
			setup: func(ctx context.Context) {
				getTokenResponse := `{"access_token" : "access_token"}`
				mockHTTPClient.On("SendRequest", ctx, mock.Anything).
					Return(&http.Response{Body: io.NopCloser(bytes.NewBuffer([]byte(getTokenResponse)))}, fmt.Errorf("internal err")).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := command.DiscardChange(testCase.ctx, "staging", testCase.config)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
