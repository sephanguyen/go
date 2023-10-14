package controller

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type VirtualClassroomReaderService struct {
	Cfg configurations.Config

	LiveLessonCommand          *commands.LiveLessonCommand
	LessonRoomStateQuery       queries.LessonRoomStateQuery
	UserInfoQuery              queries.UserInfoQuery
	VirtualClassRoomLogService *controller.VirtualClassRoomLogService
}

func (v *VirtualClassroomReaderService) RetrieveWhiteboardToken(ctx context.Context, req *vpb.RetrieveWhiteboardTokenRequest) (*vpb.RetrieveWhiteboardTokenResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}

	roomID, whiteboardToken, err := v.LiveLessonCommand.GetRoomIDAndWhiteboardToken(ctx, req.LessonId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.RetrieveWhiteboardTokenResponse{
		WhiteboardToken: whiteboardToken,
		RoomId:          roomID,
		WhiteboardAppId: v.Cfg.Whiteboard.AppID,
	}, nil
}

func (v *VirtualClassroomReaderService) GetLiveLessonState(ctx context.Context, req *vpb.GetLiveLessonStateRequest) (*vpb.GetLiveLessonStateResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}

	response, err := v.LessonRoomStateQuery.GetLiveLessonState(ctx, queries.LessonRoomStateQueryPayload{LessonID: req.LessonId})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userID := interceptors.UserIDFromContext(ctx)
	if err = v.VirtualClassRoomLogService.LogWhenGetRoomState(ctx, req.LessonId); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenGetRoomState: could not log this activity",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	return toGetLiveLessonStatePb(response), nil
}

func (v *VirtualClassroomReaderService) GetUserInformation(ctx context.Context, req *vpb.GetUserInformationRequest) (*vpb.GetUserInformationResponse, error) {
	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user IDs can't empty")
	}

	userInfos, err := v.UserInfoQuery.GetUserInfosByIDs(ctx, req.UserIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.GetUserInformationResponse{
		UserInfos: ToUserInfosPb(userInfos),
	}, nil
}

func toGetLiveLessonStatePb(lls *queries.GetLiveLessonStateResponse) *vpb.GetLiveLessonStateResponse {
	res := &vpb.GetLiveLessonStateResponse{
		LessonId:    lls.LessonID,
		CurrentTime: timestamppb.New(time.Now()),
	}

	if lls.LessonRoomState.CurrentMaterial != nil {
		res.CurrentMaterial = ToCurrentMaterialPb(
			lls.LessonRoomState.CurrentMaterial,
			lls.Media,
		)
	}

	if len(lls.UserStates.LearnersState) > 0 {
		res.UsersState = ToUsersStatePb(lls.UserStates)
	}

	if lls.LessonRoomState.CurrentPolling != nil {
		res.CurrentPolling = ToCurrentPollingPb(lls.LessonRoomState.CurrentPolling)
	}

	if lls.LessonRoomState.Recording != nil {
		res.Recording = ToRecordingPb(lls.LessonRoomState.Recording)
	}

	if lls.LessonRoomState.SpotlightedUser != "" {
		res.Spotlight = ToSpotlightedUserPb(lls.LessonRoomState.SpotlightedUser)
	}

	if lls.LessonRoomState.WhiteboardZoomState != nil {
		res.WhiteboardZoomState = ToWhiteboardZoomStatePb(lls.LessonRoomState.WhiteboardZoomState)
	}

	if lls.LessonRoomState.SessionTime != nil {
		res.SessionTime = timestamppb.New(*lls.LessonRoomState.SessionTime)
	}

	return res
}

func toMediaPb(media *media_domain.Media) *vpb.Media {
	return &vpb.Media{
		MediaId:   media.ID,
		Name:      media.Name,
		Resource:  media.Resource,
		CreatedAt: timestamppb.New(media.CreatedAt),
		UpdatedAt: timestamppb.New(media.UpdatedAt),
		Comments:  toCommentsPb(media.Comments),
		Type:      vpb.MediaType(vpb.MediaType_value[string(media.Type)]),
		Images:    toImagesPb(media.ConvertedImages),
	}
}

func toCommentsPb(comments []media_domain.Comment) []*vpb.Comment {
	commentsPb := make([]*vpb.Comment, 0, len(comments))
	for _, comment := range comments {
		commentsPb = append(commentsPb, &vpb.Comment{
			Comment:  comment.Comment,
			Duration: durationpb.New(time.Duration(comment.Duration) * time.Second),
		})
	}
	return commentsPb
}

func toImagesPb(convertedImages []media_domain.ConvertedImage) []*vpb.ConvertedImage {
	imagesPb := make([]*vpb.ConvertedImage, 0, len(convertedImages))
	for _, image := range convertedImages {
		imagesPb = append(imagesPb, &vpb.ConvertedImage{
			Width:    image.Width,
			Height:   image.Height,
			ImageUrl: image.ImageURL,
		})
	}
	return imagesPb
}

func ToCurrentMaterialPb(material *domain.CurrentMaterial, media *media_domain.Media) *vpb.VirtualClassroomState_CurrentMaterial {
	currentMaterial := &vpb.VirtualClassroomState_CurrentMaterial{
		MediaId:   material.MediaID,
		UpdatedAt: timestamppb.New(material.UpdatedAt),
	}

	if media != nil {
		currentMaterial.Data = toMediaPb(media)
	}

	if material.VideoState != nil {
		currentMaterial.State = &vpb.VirtualClassroomState_CurrentMaterial_VideoState_{
			VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
				CurrentTime: durationpb.New(material.VideoState.CurrentTime.Duration()),
				PlayerState: vpb.PlayerState(vpb.PlayerState_value[string(material.VideoState.PlayerState)]),
			},
		}
	} else if material.AudioState != nil {
		currentMaterial.State = &vpb.VirtualClassroomState_CurrentMaterial_AudioState_{
			AudioState: &vpb.VirtualClassroomState_CurrentMaterial_AudioState{
				CurrentTime: durationpb.New(material.AudioState.CurrentTime.Duration()),
				PlayerState: vpb.PlayerState(vpb.PlayerState_value[string(material.AudioState.PlayerState)]),
			},
		}
	}

	return currentMaterial
}

func ToUsersStatePb(userStates *domain.UserStates) *vpb.GetLiveLessonStateResponse_UsersState {
	learnersStates := make([]*vpb.GetLiveLessonStateResponse_UsersState_LearnerState, 0, len(userStates.LearnersState))

	for _, state := range userStates.LearnersState {
		learnersStates = append(
			learnersStates,
			&vpb.GetLiveLessonStateResponse_UsersState_LearnerState{
				UserId: state.UserID,
				HandsUp: &vpb.VirtualClassroomState_HandsUp{
					Value:     state.HandsUp.Value,
					UpdatedAt: timestamppb.New(state.HandsUp.UpdatedAt),
				},
				Annotation: &vpb.VirtualClassroomState_Annotation{
					Value:     state.Annotation.Value,
					UpdatedAt: timestamppb.New(state.Annotation.UpdatedAt),
				},
				PollingAnswer: &vpb.VirtualClassroomState_PollingAnswer{
					StringArrayValue: state.PollingAnswer.StringArrayValue,
					UpdatedAt:        timestamppb.New(state.PollingAnswer.UpdatedAt),
				},
				Chat: &vpb.VirtualClassroomState_Chat{
					Value:     state.Chat.Value,
					UpdatedAt: timestamppb.New(state.Chat.UpdatedAt),
				},
			},
		)
	}

	return &vpb.GetLiveLessonStateResponse_UsersState{
		Learners: learnersStates,
	}
}

func ToCurrentPollingPb(polling *domain.CurrentPolling) *vpb.VirtualClassroomState_CurrentPolling {
	options := make([]*vpb.VirtualClassroomState_PollingOption, 0, len(polling.Options))
	for _, option := range polling.Options {
		options = append(options, &vpb.VirtualClassroomState_PollingOption{
			Answer:    option.Answer,
			IsCorrect: option.IsCorrect,
			Content:   option.Content,
		})
	}

	currentPolling := &vpb.VirtualClassroomState_CurrentPolling{
		Options:   options,
		Status:    vpb.PollingState(vpb.PollingState_value[string(polling.Status)]),
		CreatedAt: timestamppb.New(polling.CreatedAt),
		IsShared: &vpb.VirtualClassroomState_PollingSharing{
			IsShared: polling.IsShared,
		},
		Question: polling.Question,
	}

	if polling.StoppedAt != nil {
		currentPolling.StoppedAt = timestamppb.New(*polling.StoppedAt)
	}

	return currentPolling
}

func ToRecordingPb(recording *domain.CompositeRecordingState) *vpb.VirtualClassroomState_Recording {
	return &vpb.VirtualClassroomState_Recording{
		IsRecording: recording.IsRecording,
		Creator:     recording.Creator,
	}
}

func ToSpotlightedUserPb(spotlightedUser string) *vpb.VirtualClassroomState_Spotlight {
	return &vpb.VirtualClassroomState_Spotlight{
		IsSpotlight: true,
		UserId:      spotlightedUser,
	}
}

func ToWhiteboardZoomStatePb(wbZoomState *domain.WhiteboardZoomState) *vpb.VirtualClassroomState_WhiteboardZoomState {
	return &vpb.VirtualClassroomState_WhiteboardZoomState{
		PdfScaleRatio: wbZoomState.PdfScaleRatio,
		CenterX:       wbZoomState.CenterX,
		CenterY:       wbZoomState.CenterY,
		PdfWidth:      wbZoomState.PdfWidth,
		PdfHeight:     wbZoomState.PdfHeight,
	}
}

func ToUserInfosPb(userInfos []domain.UserBasicInfo) []*vpb.UserInfo {
	userInfosPb := make([]*vpb.UserInfo, 0, len(userInfos))

	for _, userInfo := range userInfos {
		userInfosPb = append(userInfosPb, &vpb.UserInfo{
			UserId:            userInfo.UserID,
			Name:              userInfo.Name,
			FirstName:         userInfo.FirstName,
			LastName:          userInfo.LastName,
			FullNamePhonetic:  userInfo.FullNamePhonetic,
			FirstNamePhonetic: userInfo.FirstNamePhonetic,
			LastNamePhonetic:  userInfo.LastNamePhonetic,
		})
	}

	return userInfosPb
}
