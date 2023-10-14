package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/brightcove"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type BrightcoveService struct {
	Env                        string
	BrightcoveExtService       brightcove.ExternalService
	MastermgmtInternalServices mpb.InternalServiceClient
	UnleashClientIns           unleashclient.ClientInstance
}

func (s *BrightcoveService) RetrieveBrightCoveProfileData(ctx context.Context, req *ypb.RetrieveBrightCoveProfileDataRequest) (*ypb.RetrieveBrightCoveProfileDataResponse, error) {
	accountID := s.BrightcoveExtService.GetAccountID()
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled("Architecture_BACKEND_RetrieveBrightCoveProfile", s.Env)
	if err != nil {
		ctxzap.Error(ctx, "error when checking unleash feature", zap.Error(err))
	}

	if isUnleashToggled {
		rsp, err := s.MastermgmtInternalServices.GetConfigurations(common.SignCtx(ctx), &mpb.GetConfigurationsRequest{
			Paging:  &cpb.Paging{},
			Keyword: constants.BrightcoveAccountID,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		for _, config := range rsp.Items {
			if config.ConfigKey == constants.BrightcoveConfigKey {
				accountID = config.ConfigValue
			}
		}
	}

	return &ypb.RetrieveBrightCoveProfileDataResponse{
		AccountId: accountID,
		PolicyKey: s.BrightcoveExtService.GetPolicyKey(),
	}, nil
}

func (s *BrightcoveService) CreateBrightCoveUploadUrl(ctx context.Context, req *ypb.CreateBrightCoveUploadUrlRequest) (*ypb.CreateBrightCoveUploadUrlResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "missing name")
	}

	createVideoResp, err := s.BrightcoveExtService.CreateVideo(ctx, &brightcove.CreateVideoRequest{
		Name: req.Name,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	uploadUrlsResp, err := s.BrightcoveExtService.UploadUrls(ctx, &brightcove.UploadUrlsRequest{
		VideoID: createVideoResp.ID,
		Name:    req.Name,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ypb.CreateBrightCoveUploadUrlResponse{
		SignedUrl:     uploadUrlsResp.SignedURL,
		ApiRequestUrl: uploadUrlsResp.APIRequestURL,
		VideoId:       createVideoResp.ID,
	}, nil
}

func (s *BrightcoveService) FinishUploadBrightCove(ctx context.Context, req *ypb.FinishUploadBrightCoveRequest) (*ypb.FinishUploadBrightCoveResponse, error) {
	if req.ApiRequestUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "missing apiRequestUrl")
	}

	if req.VideoId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing videoId")
	}

	_, err := s.BrightcoveExtService.SubmitDynamicIngress(ctx, &brightcove.SubmitDynamicIngressRequest{
		Master: brightcove.Master{
			URL: req.ApiRequestUrl,
		},
		Profile:       s.BrightcoveExtService.GetProfile(),
		CaptureImages: true,
		VideoID:       req.VideoId,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ypb.FinishUploadBrightCoveResponse{}, nil
}

func (s *BrightcoveService) GetBrightcoveVideoInfo(ctx context.Context, req *ypb.GetBrightCoveVideoInfoRequest) (*ypb.GetBrightCoveVideoInfoResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id must not be empty")
	}
	if req.VideoId == "" {
		return nil, status.Error(codes.InvalidArgument, "video_id must not be empty")
	}

	res, err := s.BrightcoveExtService.GetVideo(ctx, req.VideoId)
	if err != nil {
		if strings.Contains(err.Error(), "\"error_code\": \"VIDEO_NOT_PLAYABLE\"") && strings.Contains(err.Error(), "expected HTTP status 200, got 403;") {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("BrightcoveExtService.GetVideo: %s", err))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("BrightcoveExtService.GetVideo: %s", err))
	}

	return &ypb.GetBrightCoveVideoInfoResponse{
		Id:             res.ID,
		Name:           res.Name,
		Thumbnail:      res.Thumbnail,
		Duration:       durationpb.New(time.Duration(res.Duration * int64(time.Millisecond))),
		OfflineEnabled: res.OfflineEnabled,
	}, nil
}

func (s *BrightcoveService) GetVideoBrightcoveResumePosition(ctx context.Context, req *ypb.GetVideoBrightcoveResumePositionRequest) (*ypb.GetVideoBrightcoveResumePositionResponse, error) {
	if req.VideoId == "" {
		return nil, status.Error(codes.InvalidArgument, "video_id must not be empty")
	}

	res, err := s.BrightcoveExtService.GetResumePosition(ctx, &brightcove.GetResumePositionRequest{
		UserID:  interceptors.UserIDFromContext(ctx),
		VideoID: req.VideoId,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("BrightcoveExtService.GetResumePosition: %s", err))
	}

	if len(res.Items) == 0 {
		return &ypb.GetVideoBrightcoveResumePositionResponse{
			VideoId: req.VideoId,
			Seconds: 0,
		}, nil
	}

	return &ypb.GetVideoBrightcoveResumePositionResponse{
		VideoId: res.Items[0].VideoID,
		Seconds: int32(res.Items[0].Seconds),
	}, nil
}
