package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/zegocloudtokengen/token04"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ZegoCloudService struct {
	ZegoCloudCfg configurations.ZegoCloudConfig
}

func (z *ZegoCloudService) GetAuthenticationToken(ctx context.Context, req *vpb.GetAuthenticationTokenRequest) (*vpb.GetAuthenticationTokenResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	if len(strings.TrimSpace(req.UserId)) != 0 {
		userID = req.UserId
	}
	payload := ""

	token, err := token04.GenerateToken04(uint32(z.ZegoCloudCfg.AppID), userID, z.ZegoCloudCfg.ServerSecret, int64(z.ZegoCloudCfg.TokenValidity), payload)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in ZegoCloud Generate Token: %s", err))
	}

	return &vpb.GetAuthenticationTokenResponse{
		AuthToken: token,
		AppId:     int32(z.ZegoCloudCfg.AppID),
		AppSign:   z.ZegoCloudCfg.AppSign,
	}, nil
}

func (z *ZegoCloudService) GetAuthenticationTokenV2(ctx context.Context, req *vpb.GetAuthenticationTokenV2Request) (*vpb.GetAuthenticationTokenV2Response, error) {
	userID := interceptors.UserIDFromContext(ctx)
	if len(strings.TrimSpace(req.UserId)) != 0 {
		userID = req.UserId
	}
	payload := ""

	token, err := token04.GenerateToken04(uint32(z.ZegoCloudCfg.AppID), userID, z.ZegoCloudCfg.ServerSecret, int64(z.ZegoCloudCfg.TokenValidity), payload)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in ZegoCloud Generate Token: %s", err))
	}

	return &vpb.GetAuthenticationTokenV2Response{
		AuthToken: token,
	}, nil
}

func (z *ZegoCloudService) GetChatConfig(_ context.Context, _ *vpb.GetChatConfigRequest) (*vpb.GetChatConfigResponse, error) {
	return &vpb.GetChatConfigResponse{
		AppId:   int32(z.ZegoCloudCfg.AppID),
		AppSign: z.ZegoCloudCfg.AppSign,
	}, nil
}
