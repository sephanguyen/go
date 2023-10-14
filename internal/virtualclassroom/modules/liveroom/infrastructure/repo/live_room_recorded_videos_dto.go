package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveRoomRecordedVideo struct {
	RecordedVideoID  pgtype.Text
	ChannelID        pgtype.Text
	MediaID          pgtype.Text
	Description      pgtype.Text
	DateTimeRecorded pgtype.Timestamptz
	Creator          pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (l *LiveRoomRecordedVideo) FieldMap() ([]string, []interface{}) {
	return []string{
			"recorded_video_id",
			"channel_id",
			"media_id",
			"description",
			"date_time_recorded",
			"creator",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.RecordedVideoID,
			&l.ChannelID,
			&l.MediaID,
			&l.Description,
			&l.DateTimeRecorded,
			&l.Creator,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *LiveRoomRecordedVideo) TableName() string {
	return "live_room_recorded_videos"
}

func (l *LiveRoomRecordedVideo) PreInsert() error {
	now := time.Now()
	err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	)
	return err
}

func NewLiveRoomRecordedVideoFromEntity(r *vc_domain.RecordedVideo) (*LiveRoomRecordedVideo, error) {
	dto := &LiveRoomRecordedVideo{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.RecordedVideoID.Set(r.ID),
		dto.ChannelID.Set(r.RecordingChannelID),
		dto.MediaID.Set(r.GetMediaID()),
		dto.Description.Set(r.Description),
		dto.DateTimeRecorded.Set(r.DateTimeRecorded),
		dto.Creator.Set(r.Creator),
		dto.CreatedAt.Set(r.CreatedAt),
		dto.UpdatedAt.Set(r.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from recorded video entity to live room recorded video dto: %w", err)
	}

	return dto, nil
}

type LiveRoomRecordedVideos []*LiveRoomRecordedVideo

func (l *LiveRoomRecordedVideos) Add() database.Entity {
	e := &LiveRoomRecordedVideo{}
	*l = append(*l, e)

	return e
}

func (l LiveRoomRecordedVideos) ToRecordedVideosEntity() vc_domain.RecordedVideos {
	res := make(vc_domain.RecordedVideos, 0, len(l))
	for _, video := range l {
		res = append(res, &vc_domain.RecordedVideo{
			ID:                 video.RecordedVideoID.String,
			RecordingChannelID: video.ChannelID.String,
			Description:        video.Description.String,
			DateTimeRecorded:   video.DateTimeRecorded.Time,
			Creator:            video.Creator.String,
			CreatedAt:          video.CreatedAt.Time,
			UpdatedAt:          video.UpdatedAt.Time,
			Media: &media_domain.Media{
				ID: video.MediaID.String,
			},
		})
	}

	return res
}
