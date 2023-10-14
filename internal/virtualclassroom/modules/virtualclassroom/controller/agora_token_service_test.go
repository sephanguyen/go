package controller

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/stretchr/testify/require"
)

func TestAgoraTokenService_GenerateAgoraStreamToken(t *testing.T) {
	t.Parallel()

	tokenService := &AgoraTokenService{
		AgoraCfg: configurations.AgoraConfig{},
	}

	lessonID := "lesson-id1"
	userID := "user-id1"

	testCases := []struct {
		name     string
		role     domain.AgoraRole
		setup    func(context.Context)
		hasError bool
	}{
		{
			name:     "success generate token with subscriber",
			role:     domain.RoleSubscriber,
			setup:    func(ctx context.Context) {},
			hasError: false,
		},
		{
			name:     "success generate token with publisher",
			role:     domain.RolePublisher,
			setup:    func(ctx context.Context) {},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := tokenService.GenerateAgoraStreamToken(lessonID, userID, tc.role)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, resp)
			}
		})
	}
}

func TestAgoraTokenService_BuildRTMToken(t *testing.T) {
	t.Parallel()

	tokenService := &AgoraTokenService{
		AgoraCfg: configurations.AgoraConfig{},
	}

	lessonID := "lesson-id1"
	userID := "user-id1"

	testCases := []struct {
		name     string
		setup    func(context.Context)
		hasError bool
	}{
		{
			name:     "success generate RTM token",
			setup:    func(ctx context.Context) {},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := tokenService.BuildRTMToken(lessonID, userID)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, resp)
			}
		})
	}
}

func TestAgoraTokenService_BuildRTMTokenByUserID(t *testing.T) {
	t.Parallel()

	tokenService := &AgoraTokenService{
		AgoraCfg: configurations.AgoraConfig{},
	}
	userID := "user-id1"

	testCases := []struct {
		name     string
		setup    func(context.Context)
		hasError bool
	}{
		{
			name:     "success generate RTM token",
			setup:    func(ctx context.Context) {},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := tokenService.BuildRTMTokenByUserID(userID)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, resp)
			}
		})
	}
}
