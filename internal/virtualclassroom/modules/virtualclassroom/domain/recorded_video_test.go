package domain

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	recording "github.com/manabie-com/backend/internal/golibs/recording"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordedVideo_IsValid(t *testing.T) {
	t.Parallel()
	lessonId := "lesson-id"
	media := &media_domain.Media{
		Resource:      "resource",
		Type:          media_domain.MediaTypeRecordingVideo,
		FileSizeBytes: 432532,
		Duration:      time.Hour,
	}
	tcs := []struct {
		name          string
		recordedVideo *RecordedVideo
		isValid       bool
	}{
		{
			name: "full fields",
			recordedVideo: &RecordedVideo{
				ID:                 "id-1",
				Description:        "video 1",
				Creator:            "user-id-1",
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
				DateTimeRecorded:   time.Now(),
				RecordingChannelID: lessonId,
				Media:              media,
			},
			isValid: true,
		},
		{
			name: "only required fields",
			recordedVideo: &RecordedVideo{
				ID:                 "id-1",
				Creator:            "user-id-1",
				DateTimeRecorded:   time.Now(),
				RecordingChannelID: lessonId,
				Media:              media,
			},
			isValid: true,
		},
		{
			name: "miss id field",
			recordedVideo: &RecordedVideo{
				Creator:            "user-id-1",
				DateTimeRecorded:   time.Now(),
				RecordingChannelID: lessonId,
				Media:              media,
			},
			isValid: false,
		},
		{
			name: "miss creator field",
			recordedVideo: &RecordedVideo{
				ID:               "id-1",
				DateTimeRecorded: time.Now(),
				Media:            media,
			},
			isValid: false,
		},
		{
			name: "miss date time recorded field",
			recordedVideo: &RecordedVideo{
				ID:      "id-1",
				Creator: "user-id-1",
				Media:   media,
			},
			isValid: false,
		},
		{
			name: "miss media field",
			recordedVideo: &RecordedVideo{
				ID:                 "id-1",
				Creator:            "user-id-1",
				DateTimeRecorded:   time.Now(),
				RecordingChannelID: lessonId,
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.recordedVideo.IsValid(context.Background())
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRecordedVideoBuilder_BuildDraft(t *testing.T) {
	data := RecordedVideo{
		ID:               "id-1",
		Description:      "video 1",
		Creator:          "user-id-1",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now().Add(1 * time.Hour),
		DateTimeRecorded: time.Now().Add(-1 * time.Hour),
		Media: &media_domain.Media{
			Resource:      "resource",
			Type:          media_domain.MediaTypeRecordingVideo,
			FileSizeBytes: 432532,
			Duration:      time.Hour,
		},
		RecordingChannelID: "lesson-id",
	}
	builder := NewRecordedVideoBuilder()
	actual := builder.
		WithID(data.ID).
		WithRecordingChannelID(data.RecordingChannelID).
		WithDescription(data.Description).
		WithCreator(data.Creator).
		WithModificationTime(data.CreatedAt, data.UpdatedAt).
		WithDateTimeRecorded(data.DateTimeRecorded).
		WithMedia(data.Media).BuildDraft()
	assert.EqualValues(t, data, *actual)
}

func TestToRecordedVideos(t *testing.T) {
	list := []recording.FileInfo{
		{
			Filename:       "01GA66MQX3HTSSRZE11MTP0JTB/1664772217/11acdabc654d4e83d68462acd22b72d7_01GA66MQX3HTSSRZE11MTP0JTB_3.mp4",
			SliceStartTime: 1665496800000, // 2022-10-11T14:00:00
		},
		{
			Filename:       "01GA66MQX3HTSSRZE11MTP0JTB/1664772217/11acdabc654d4e83d68462acd22b72d7_01GA66MQX3HTSSRZE11MTP0JTB_0.mp4",
			SliceStartTime: 1665475200000, // 2022-10-11T08:00:00
		},
		{
			Filename:       "01GA66MQX3HTSSRZE11MTP0JTB/1664772217/11acdabc654d4e83d68462acd22b72d7_01GA66MQX3HTSSRZE11MTP0JTB_1.mp4",
			SliceStartTime: 1665482400000, // 2022-10-11T10:00:00
		},
		{
			Filename:       "01GA66MQX3HTSSRZE11MTP0JTB/1664772217/11acdabc654d4e83d68462acd22b72d7_01GA66MQX3HTSSRZE11MTP0JTB_2.ts",
			SliceStartTime: 1664772219854,
		},
		{
			Filename:       "01GA66MQX3HTSSRZE11MTP0JTB/1664772217/11acdabc654d4e83d68462acd22b72d7_01GA66MQX3HTSSRZE11MTP0JTB_2.mp4",
			SliceStartTime: 1665489600000, // 2022-10-11T12:00:00
		},
	}
	ctx := context.Background()
	size := int64(2000000)
	getObject := func(context.Context, string, string) (*filestore.StorageObject, error) {
		ob := &filestore.StorageObject{
			Size: size,
		}
		return ob, nil
	}
	lessonId := "lesson-id"
	creator := "creator"
	endTime := time.UnixMilli(1665498300000) // 2022-10-11T14:25:00

	ts, err := ToRecordedVideos(ctx, list, endTime, lessonId, creator, "bucket-name", getObject)
	assert.Nil(t, err, nil)

	expectedList := []*RecordedVideo{
		{
			Media: &media_domain.Media{
				Resource:      list[1].Filename,
				Type:          media_domain.MediaTypeRecordingVideo,
				FileSizeBytes: size,
				Duration:      time.Duration(2 * time.Hour),
			},
			RecordingChannelID: lessonId,
			Creator:            creator,
			DateTimeRecorded:   time.UnixMilli(list[1].SliceStartTime),
		},
		{
			Media: &media_domain.Media{
				Resource:      list[2].Filename,
				Type:          media_domain.MediaTypeRecordingVideo,
				FileSizeBytes: size,
				Duration:      time.Duration(2 * time.Hour),
			},
			RecordingChannelID: lessonId,
			Creator:            creator,
			DateTimeRecorded:   time.UnixMilli(list[2].SliceStartTime),
		},
		{
			Media: &media_domain.Media{
				Resource:      list[4].Filename,
				Type:          media_domain.MediaTypeRecordingVideo,
				FileSizeBytes: size,
				Duration:      time.Duration(2 * time.Hour),
			},
			RecordingChannelID: lessonId,
			Creator:            creator,
			DateTimeRecorded:   time.UnixMilli(list[4].SliceStartTime),
		},
		{
			Media: &media_domain.Media{
				Resource:      list[0].Filename,
				Type:          media_domain.MediaTypeRecordingVideo,
				FileSizeBytes: size,
				Duration:      time.Duration(25 * time.Minute),
			},
			RecordingChannelID: lessonId,
			Creator:            creator,
			DateTimeRecorded:   time.UnixMilli(list[0].SliceStartTime),
		},
	}
	assert.EqualValues(t, expectedList, ts)
}
