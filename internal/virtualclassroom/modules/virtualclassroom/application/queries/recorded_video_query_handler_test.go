package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLessonRoomStateQuery_RetrieveRecordedVideosByLessonID(t *testing.T) {
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	recordedVideoRepo := &mock_virtual_repo.MockRecordedVideoRepo{}
	lessonId := "lesson-1"
	now := time.Now()

	result := domain.RecordedVideos{
		{
			ID:                 "recorded-video-id-1",
			RecordingChannelID: "lesson-id-1",
			Description:        "description 1",
			DateTimeRecorded:   now,
			Creator:            "user-id-1",
			CreatedAt:          now,
			UpdatedAt:          now,
			Media: &media_domain.Media{
				ID:       "media-id-1",
				Resource: "video-id-1",
				Type:     media_domain.MediaTypeRecordingVideo,
			},
		},
		{
			ID:                 "recorded-video-id-2",
			RecordingChannelID: "lesson-id-2",
			Description:        "description 2",
			DateTimeRecorded:   now,
			Creator:            "user-id-2",
			CreatedAt:          now,
			UpdatedAt:          now,
			Media: &media_domain.Media{
				ID:       "media-id-3",
				Resource: "video-id-3",
				Type:     media_domain.MediaTypeRecordingVideo,
			},
		},
	}

	t.Run("success with preTotal > limit", func(t *testing.T) {
		payload := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			LessonID:        lessonId,
			Limit:           2,
			RecordedVideoID: "record-1",
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("ListRecordingByLessonIDWithPaging", ctx, db, payload).Return(result, uint32(10), "pre-page-id", uint32(5), nil).Once()
		q := &RecordedVideoQuery{
			WrapperDBConnection: wrapperConnection,
			RecordedVideoRepo:   recordedVideoRepo,
		}
		res := q.RetrieveRecordedVideosByLessonID(ctx, payload)

		result := &RetrieveRecordedVideosByLessonIDQueryResponse{
			Recs:      result,
			Total:     10,
			PrePageID: "pre-page-id",
			Err:       nil,
		}

		assert.Equal(t, result, res)
	})

	t.Run("success with preTotal <= limit", func(t *testing.T) {
		payload := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			LessonID:        lessonId,
			Limit:           10,
			RecordedVideoID: "record-1",
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("ListRecordingByLessonIDWithPaging", ctx, db, payload).Return(result, uint32(10), "pre-page-id", uint32(5), nil).Once()
		q := &RecordedVideoQuery{
			WrapperDBConnection: wrapperConnection,
			RecordedVideoRepo:   recordedVideoRepo,
		}
		res := q.RetrieveRecordedVideosByLessonID(ctx, payload)

		result := &RetrieveRecordedVideosByLessonIDQueryResponse{
			Recs:      result,
			Total:     10,
			PrePageID: "",
			Err:       nil,
		}

		assert.Equal(t, result, res)
	})

	t.Run("fail", func(t *testing.T) {
		payload := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			LessonID:        lessonId,
			Limit:           10,
			RecordedVideoID: "record-1",
		}
		err := fmt.Errorf("error when call recordedVideoRepo.ListRecordingByLessonIDWithPaging")
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("ListRecordingByLessonIDWithPaging", ctx, db, payload).Return(domain.RecordedVideos{}, uint32(0), "", uint32(0), err).Once()
		q := &RecordedVideoQuery{
			WrapperDBConnection: wrapperConnection,
			RecordedVideoRepo:   recordedVideoRepo,
		}
		res := q.RetrieveRecordedVideosByLessonID(ctx, payload)

		result := &RetrieveRecordedVideosByLessonIDQueryResponse{
			Recs: domain.RecordedVideos{},
			Err:  err,
		}

		assert.Equal(t, result, res)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}

func TestLessonRoomStateQuery_GetRecordingByID(t *testing.T) {
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	recordedVideoRepo := &mock_virtual_repo.MockRecordedVideoRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	recordingID := "recording-id"
	now := time.Now()

	result := &domain.RecordedVideo{
		ID:                 "recorded-video-id-1",
		RecordingChannelID: "lesson-id-1",
		Description:        "description 1",
		DateTimeRecorded:   now,
		Creator:            "user-id-1",
		CreatedAt:          now,
		UpdatedAt:          now,
		Media: &media_domain.Media{
			ID:       "media-id-1",
			Resource: "video-id-1",
			Type:     media_domain.MediaTypeRecordingVideo,
		},
	}

	t.Run("success", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("GetRecordingByID", ctx, db, &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}).Once().Return(result, nil)

		mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{result.Media.ID}).Return(media_domain.Medias{
			result.Media}, nil).Once()

		payload := &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}
		q := &RecordedVideoQuery{
			WrapperDBConnection: wrapperConnection,
			RecordedVideoRepo:   recordedVideoRepo,
			MediaModulePort:     mediaModulePort,
		}
		res, err := q.GetRecordingByID(ctx, payload)

		assert.Nil(t, err)
		assert.Equal(t, result, res)
	})

	t.Run("fail by error when call RecordedVideoRepo.GetRecordingByID", func(t *testing.T) {
		expectedErr := status.Error(codes.Internal, "error when call RecordedVideoRepo.GetRecordingByID: rpc error: code = Internal desc = error message")
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("GetRecordingByID", ctx, db, &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}).Once().Return(nil, status.Error(codes.Internal, "error message"))

		payload := &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}
		q := &RecordedVideoQuery{
			WrapperDBConnection: wrapperConnection,
			RecordedVideoRepo:   recordedVideoRepo,
		}
		res, err := q.GetRecordingByID(ctx, payload)

		assert.Equal(t, err, expectedErr)
		assert.Nil(t, res)
	})

	t.Run("fail by error when call ", func(t *testing.T) {
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("GetRecordingByID", ctx, db, &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}).Once().Return(result, nil)
		expectedErr := status.Error(codes.Internal, "error when call MediaModulePort.RetrieveMediasByIDs: rpc error: code = Internal desc = error message")
		mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{result.Media.ID}).Return(media_domain.Medias{}, status.Error(codes.Internal, "error message")).Once()

		payload := &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}
		q := &RecordedVideoQuery{
			WrapperDBConnection: wrapperConnection,
			RecordedVideoRepo:   recordedVideoRepo,
			MediaModulePort:     mediaModulePort,
		}
		res, err := q.GetRecordingByID(ctx, payload)

		assert.Equal(t, err, expectedErr)
		assert.Nil(t, res)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}
