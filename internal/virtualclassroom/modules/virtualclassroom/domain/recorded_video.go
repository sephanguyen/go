package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	recording "github.com/manabie-com/backend/internal/golibs/recording"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type RecordedVideo struct {
	ID                 string
	RecordingChannelID string // can be channel ID or lesson ID
	Description        string
	Creator            string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DateTimeRecorded   time.Time
	Media              *media_domain.Media
}

func (r *RecordedVideo) IsValid(ctx context.Context) error {
	if len(r.ID) == 0 {
		return fmt.Errorf("ID cannot be empty")
	}
	if r.DateTimeRecorded.IsZero() {
		return fmt.Errorf("datetime recorded cannot be empty")
	}
	if len(r.Creator) == 0 {
		return fmt.Errorf("creator cannot be empty")
	}
	if r.Media == nil {
		return fmt.Errorf("media cannot nil")
	}

	if len(r.Media.Resource) == 0 {
		return fmt.Errorf("media's resource can not be empty")
	}
	if r.Media.FileSizeBytes == 0 {
		return fmt.Errorf("media's file_size_bytes can not be 0")
	}
	if r.Media.Duration.Seconds() == 0 {
		return fmt.Errorf("media's duration can not be 0")
	}
	if r.Media.Type != media_domain.MediaTypeRecordingVideo {
		return fmt.Errorf("media's type must be MediaTypeRecordingVideo")
	}

	return nil
}

func (r *RecordedVideo) PreInsert() {
	r.ID = idutil.ULIDNow()
}

func (r *RecordedVideo) GetMediaID() string {
	if r.Media != nil {
		return r.Media.ID
	}
	return ""
}

func (r *RecordedVideo) AssignByMediaFile(media *media_domain.Media) error {
	if media == nil {
		return fmt.Errorf("could not use null media to assign recorded video")
	}
	if err := media.IsValid(); err != nil {
		return fmt.Errorf("could not use an invalid media to assign recorded video: %v", err)
	}
	if media.Type != media_domain.MediaTypeRecordingVideo {
		return fmt.Errorf("only use video media to assign for recorded video")
	}
	r.Media = media

	return nil
}

type RecordedVideos []*RecordedVideo

func (rs RecordedVideos) GetRecordIDs() []string {
	ids := make([]string, 0, len(rs))
	for _, v := range rs {
		ids = append(ids, v.ID)
	}
	return ids
}

func (rs RecordedVideos) GetResources() []string {
	resources := make([]string, 0, len(rs))
	for _, v := range rs {
		resources = append(resources, v.Media.Resource)
	}
	return resources
}

func (rs RecordedVideos) GetMediaIDs() []string {
	medias := make([]string, 0, len(rs))
	for _, v := range rs {
		medias = append(medias, v.Media.ID)
	}
	return medias
}
func (rs RecordedVideos) WithMedias(medias media_domain.Medias) error {
	if len(medias) == 0 {
		return fmt.Errorf("medias don't have any element")
	}
	videoIDByMediaID := make(map[string]*media_domain.Media)
	for _, media := range medias {
		videoIDByMediaID[media.ID] = media
	}
	var isExisted bool
	for _, video := range rs {
		video.Media, isExisted = videoIDByMediaID[video.Media.ID]
		if !isExisted {
			return fmt.Errorf("the recorded video don't have any match media")
		}
	}
	return nil
}

type RecordedVideoBuilder struct {
	recordedVideo *RecordedVideo
}

func NewRecordedVideoBuilder() *RecordedVideoBuilder {
	return &RecordedVideoBuilder{
		recordedVideo: &RecordedVideo{},
	}
}

func (r *RecordedVideoBuilder) Build(ctx context.Context) (*RecordedVideo, error) {
	if err := r.recordedVideo.IsValid(ctx); err != nil {
		return nil, fmt.Errorf("invalid recorded video: %v", err)
	}
	return r.recordedVideo, nil
}

// BuildDraft will skip validate data
// only use to load RecordedVideo object from trusted data sources
func (r *RecordedVideoBuilder) BuildDraft() *RecordedVideo {
	return r.recordedVideo
}

func (r *RecordedVideoBuilder) WithID(id string) *RecordedVideoBuilder {
	r.recordedVideo.ID = id
	return r
}

func (r *RecordedVideoBuilder) WithRecordingChannelID(id string) *RecordedVideoBuilder {
	r.recordedVideo.RecordingChannelID = id
	return r
}

func (r *RecordedVideoBuilder) WithDescription(description string) *RecordedVideoBuilder {
	r.recordedVideo.Description = description
	return r
}

func (r *RecordedVideoBuilder) WithDateTimeRecorded(datetimeRecorded time.Time) *RecordedVideoBuilder {
	r.recordedVideo.DateTimeRecorded = datetimeRecorded
	return r
}

func (r *RecordedVideoBuilder) WithCreator(creator string) *RecordedVideoBuilder {
	r.recordedVideo.Creator = creator
	return r
}

func (r *RecordedVideoBuilder) WithModificationTime(createdAt, updatedAt time.Time) *RecordedVideoBuilder {
	r.recordedVideo.CreatedAt = createdAt
	r.recordedVideo.UpdatedAt = updatedAt
	return r
}

func (r *RecordedVideoBuilder) WithMedia(media *media_domain.Media) *RecordedVideoBuilder {
	r.recordedVideo.Media = media
	return r
}

func ToRecordedVideos(ctx context.Context, list []recording.FileInfo, endTime time.Time, recordingChannel, creator, bucketName string, getObject func(context.Context, string, string) (*filestore.StorageObject, error)) ([]*RecordedVideo, error) {
	rs := make([]*RecordedVideo, 0)
	for i := 0; i < len(list); i++ {
		file := getFileByIndex(list, i)
		if file == nil {
			break
		}

		rc, err := getObject(ctx, bucketName, file.Filename)
		if err != nil {
			return nil, err
		}
		size := rc.Size

		recordedVideo := &RecordedVideo{
			RecordingChannelID: recordingChannel,
			DateTimeRecorded:   time.UnixMilli(file.SliceStartTime),
			Creator:            creator,
			Media: &media_domain.Media{
				Resource:      file.Filename,
				FileSizeBytes: size,
				Type:          media_domain.MediaTypeRecordingVideo,
			},
		}
		rs = append(rs, recordedVideo)
	}

	// call calculate duration
	for i := 0; i < len(rs); i++ {
		if i == len(rs)-1 {
			rs[i].Media.Duration = endTime.Sub(rs[i].DateTimeRecorded)
		} else {
			rs[i].Media.Duration = rs[i+1].DateTimeRecorded.Sub(rs[i].DateTimeRecorded)
		}
	}
	return rs, nil
}

func getFileByIndex(files []recording.FileInfo, i int) *recording.FileInfo {
	for _, v := range files {
		if strings.Contains(v.Filename, fmt.Sprintf("_%d.mp4", i)) { // get file mp4 with index
			return &v
		}
	}
	return nil
}

type StorageObject struct {
	Size int64
}
