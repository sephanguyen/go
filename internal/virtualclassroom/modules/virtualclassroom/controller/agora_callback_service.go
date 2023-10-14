package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	recording "github.com/manabie-com/backend/internal/golibs/recording"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	lr_queries "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	lr_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AgoraCallbackService struct {
	Cfg                  configurations.Config
	Logger               *zap.Logger
	RecordingCommand     commands.RecordingCommand
	LessonRoomStateQuery queries.LessonRoomStateQuery
	LiveRoomStateQuery   lr_queries.LiveRoomStateQuery
	OrganizationQuery    queries.OrganizationQuery
}

func (a *AgoraCallbackService) CallBack(c *gin.Context) {
	req, logger, err := a.toAgoraCallbackPayload(c)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
	logger.Info("START Callback")

	orgMap, err := a.OrganizationQuery.GetOrganizationMap(context.Background())
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resourcePathMapID := req.Payload.UID[len(req.Payload.UID)-3:] // get 3 last character
	resourcePath := orgMap[resourcePathMapID]
	if resourcePath == "" {
		logger.Error(
			"resource_path id do not match any organization in init list",
			zap.String("resourcePathMapID", resourcePathMapID),
			zap.Any("OrgMap", orgMap),
		)
		return
	}

	recordingChannel := req.Payload.ChannelName
	if recordingChannel == "" {
		logger.Error("lesson ID or channel ID from recording channel is empty")
		return
	}
	logger = logger.With(zap.String("lesson_id/channel_id", recordingChannel))

	// add initial value so if both are empty has default false value
	recordingState := &domain.CompositeRecordingState{
		IsRecording: false,
	}
	var recordingRef constant.RecordingReference
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(c, claim)
	liveLessonState, err := a.LessonRoomStateQuery.GetLessonRoomStateByLessonIDWithoutCheck(ctx, queries.LessonRoomStateQueryPayload{
		LessonID: recordingChannel,
	})
	switch {
	case err != nil && err != domain.ErrLessonRoomStateNotFound:
		logger.Error(fmt.Sprintf("error in LessonRoomStateQuery.GetLessonRoomStateByLessonID %s", err.Error()))
		return
	case err == domain.ErrLessonRoomStateNotFound:
		logger.Warn(fmt.Sprintf("lesson ID %s state is not found, checking live room", recordingChannel))

		liveRoomState, err := a.LiveRoomStateQuery.GetLiveRoomStateOnlyByChannelIDWithoutCheck(ctx, recordingChannel)
		if err != nil && err != lr_domain.ErrChannelNotFound {
			logger.Error(fmt.Sprintf("error in LiveRoomStateQuery.GetLiveRoomStateOnlyByChannelIDWithoutCheck %s", err.Error()))
			return
		}
		if err == lr_domain.ErrChannelNotFound {
			logger.Warn(fmt.Sprintf("channel ID %s state is not found, finishing callback", recordingChannel))
			return
		}
		recordingState = liveRoomState.Recording
		recordingRef = constant.LiveRoomRecordingRef
	case err == nil:
		recordingState = liveLessonState.Recording
		recordingRef = constant.LessonRecordingRef
	}

	if recordingState.IsRecording {
		newUID := fmt.Sprintf(recording.UIDFormat, recordingState.UID)
		if newUID != req.Payload.UID || recordingState.SID != req.Payload.SID {
			logger.Warn(fmt.Sprintf("request not match with current %s room state: uid:%s, sid:%s", recordingRef, newUID, recordingState.SID))
			return
		}
	}

	switch req.EventType {
	case domain.CloudRecordingServiceError:
		logger.Warn("cloud recording service error message received")
	case domain.CloudRecordingServiceExited:
		logger.Warn("cloud recording service exited log received, will clear out recording state if it is in recording state")
		if recordingState.IsRecording {
			a.clearRecordingState(ctx, logger, recordingChannel, recordingRef)
		}
	case domain.CloudRecordingServiceWarning:
		logger.Warn("warning about the cloud recording")
	case domain.CloudRecordingServiceStatusUpdate:
		logger.Warn("cloud recording status update")
	case domain.CloudRecordingServiceFileInfo, domain.CloudRecordingServiceFailover:
		break
	case domain.UploadingStarts, domain.UploadingDone, domain.UploadingBackupDone, domain.UploadingProgress:
		break
	case domain.RecordingStarts, domain.RecordingSliceStart, domain.RecordingAudioStreamStateChanged,
		domain.RecordingVideoStreamStateChanged, domain.RecordingSnapshotFile:
		break
	case domain.RecordingExits:
		logger.Warn("recorder leaving the channel has been detected, will clear out recording state if it is in recording state")
		if recordingState.IsRecording {
			a.clearRecordingState(ctx, logger, recordingChannel, recordingRef)
		}
	case domain.DownloadFailed:
		logger.Warn("recording service failed to download recorded file")
	default:
		logger.Warn("agora callback event type not supported")
	}

	logger.Info("END Callback")
}

func (a *AgoraCallbackService) toAgoraCallbackPayload(c *gin.Context) (*domain.AgoraCallbackPayload, *zap.Logger, error) {
	logger := a.Logger.With(zap.String("service", "agora callback"))
	signature := c.GetHeader(middlewares.AgoraHeaderKey)
	logger = logger.With(zap.String("signature", signature))

	payload := middlewares.PayloadFromContext(c)
	if len(payload) == 0 {
		return nil, logger, fmt.Errorf("payload body is empty")
	}

	req := &domain.AgoraCallbackPayload{}
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, logger, err
	}

	logger = logger.With(zap.String("callback_payload", string(payload)))
	return req, logger, nil
}

func (a *AgoraCallbackService) clearRecordingState(ctx context.Context, logger *zap.Logger, recordingChannel string, recordingRef constant.RecordingReference) {
	if err := a.RecordingCommand.UpsertRecordingState(ctx, &commands.UpsertRecordingStatePayload{
		RecordingRef:     recordingRef,
		RecordingChannel: recordingChannel,
		Recording:        nil,
	}); err != nil {
		logger.Error(
			fmt.Sprintf("fail when call RecordingCommand.UpsertRecordingState for %s %s", recordingRef, recordingChannel),
			zap.Error(err),
		)
	}
}
