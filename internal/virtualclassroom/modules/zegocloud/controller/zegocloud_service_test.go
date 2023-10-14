package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/stretchr/testify/require"
)

func TestZegoCloudService_GetAuthenticationToken(t *testing.T) {
	t.Parallel()

	request := &vpb.GetAuthenticationTokenRequest{
		UserId: "user_id",
	}
	config := configurations.ZegoCloudConfig{
		AppID:         123456,
		ServerSecret:  "ABCDEFGEDSAFAIURGGFDGNDFUIFDGDFA",
		AppSign:       "app_sign",
		TokenValidity: 3600,
	}
	teacherID := "teacher_id1"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetAuthenticationTokenRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "user gets zegocloud auth token",
			reqUserID: teacherID,
			req:       request,
			setup:     func(ctx context.Context) {},
			hasError:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &ZegoCloudService{
				ZegoCloudCfg: config,
			}

			response, err := service.GetAuthenticationToken(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response.AuthToken)
				require.Equal(t, response.AppId, int32(config.AppID))
				require.Equal(t, response.AppSign, config.AppSign)
			}
		})
	}
}

func TestZegoCloudService_GetAuthenticationTokenV2(t *testing.T) {
	t.Parallel()

	request := &vpb.GetAuthenticationTokenV2Request{
		UserId: "user_id",
	}
	config := configurations.ZegoCloudConfig{
		AppID:         123456,
		ServerSecret:  "ABCDEFGEDSAFAIURGGFDGNDFUIFDGDFA",
		AppSign:       "app_sign",
		TokenValidity: 3600,
	}
	teacherID := "teacher_id1"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetAuthenticationTokenV2Request
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "user gets zegocloud auth token",
			reqUserID: teacherID,
			req:       request,
			setup:     func(ctx context.Context) {},
			hasError:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &ZegoCloudService{
				ZegoCloudCfg: config,
			}

			response, err := service.GetAuthenticationTokenV2(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response.AuthToken)
			}
		})
	}
}

func TestZegoCloudService_GetChatConfig(t *testing.T) {
	t.Parallel()

	request := &vpb.GetChatConfigRequest{}
	config := configurations.ZegoCloudConfig{
		AppID:         123456,
		ServerSecret:  "ABCDEFGEDSAFAIURGGFDGNDFUIFDGDFA",
		AppSign:       "app_sign",
		TokenValidity: 3600,
	}
	teacherID := "teacher_id1"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetChatConfigRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "user gets zegocloud chat config",
			reqUserID: teacherID,
			req:       request,
			setup:     func(ctx context.Context) {},
			hasError:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &ZegoCloudService{
				ZegoCloudCfg: config,
			}

			response, err := service.GetChatConfig(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, response.AppId, int32(config.AppID))
				require.Equal(t, response.AppSign, config.AppSign)
			}
		})
	}
}
