package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	lr_infrastructure "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type RecordingCommand struct {
	LessonmgmtDB        database.Ext
	WrapperDBConnection *support.WrapperDBConnection
	LessonRoomStateRepo interface {
		UpsertRecordingState(ctx context.Context, db database.QueryExecer, lessonID string, recording *domain.CompositeRecordingState) error
	}
	RecordedVideoRepo          infrastructure.RecordedVideoRepo
	MediaModulePort            infrastructure.MediaModulePort
	LiveRoomStateRepo          lr_infrastructure.LiveRoomStateRepo
	LiveRoomRecordedVideosRepo lr_infrastructure.LiveRoomRecordedVideos
}

func (r *RecordingCommand) UpsertRecordingState(ctx context.Context, payload *UpsertRecordingStatePayload) error {
	conn, err := r.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	switch payload.RecordingRef {
	case constant.LessonRecordingRef:
		err := r.LessonRoomStateRepo.UpsertRecordingState(ctx, conn, payload.RecordingChannel, payload.Recording)
		if err != nil {
			return fmt.Errorf("error in LessonRoomStateRepo.UpsertRecordingState, lesson %s: %w", payload.RecordingChannel, err)
		}
	case constant.LiveRoomRecordingRef:
		err := r.LiveRoomStateRepo.UpsertRecordingState(ctx, r.LessonmgmtDB, payload.RecordingChannel, payload.Recording)
		if err != nil {
			return fmt.Errorf("error in LiveRoomStateRepo.UpsertRecordingState, channel %s: %w", payload.RecordingChannel, err)
		}
	}

	return nil
}

func (r *RecordingCommand) NewRecordedVideos(ctx context.Context, payload *NewRecordingVideoPayload) error {
	if payload.RecordingChannel == "" {
		return fmt.Errorf("recording channel cannot be empty")
	}
	if payload.RecordingRef == "" {
		return fmt.Errorf("recording reference cannot be empty")
	}
	if len(payload.RecordedVideos) == 0 {
		return fmt.Errorf("recordedVideos must have at least 1 video")
	}
	conn, err := r.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	rv := make([]*domain.RecordedVideo, 0, len(payload.RecordedVideos))
	for _, v := range payload.RecordedVideos {
		if v == nil {
			return fmt.Errorf("could not new a null recorded video")
		}
		v.Media.PreCreate()
		if err := v.Media.IsValid(); err != nil {
			return fmt.Errorf("invalid recorded video media: %w", err)
		}
		if err := r.MediaModulePort.CreateMedia(ctx, v.Media); err != nil {
			return fmt.Errorf("err when create media: %w", err)
		}
		v.PreInsert()
		if err := v.IsValid(ctx); err != nil {
			return fmt.Errorf("invalid recorded video: %w", err)
		}
		rv = append(rv, v)
	}

	switch payload.RecordingRef {
	case constant.LessonRecordingRef:
		err := r.RecordedVideoRepo.InsertRecordedVideos(ctx, conn, rv)
		if err != nil {
			return fmt.Errorf("could not insert recorded videos for lesson %s : RecordedVideoRepo.InsertRecordedVideos: %w", payload.RecordingChannel, err)
		}
	case constant.LiveRoomRecordingRef:
		err := r.LiveRoomRecordedVideosRepo.InsertRecordedVideos(ctx, r.LessonmgmtDB, rv)
		if err != nil {
			return fmt.Errorf("could not insert recorded videos for live room %s : LiveRoomRecordedVideosRepo.InsertRecordedVideos: %w", payload.RecordingChannel, err)
		}
	}

	return nil
}
