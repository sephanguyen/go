package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLessonDeletedHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	recordedVideoRepo := &mock_virtual_repo.MockRecordedVideoRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	fileStoreMock := &filestore.Mock{}
	now := time.Now()
	lessonIds := []string{"lesson-id-1"}
	medias := media_domain.Medias{
		{
			ID:       "media-id-1",
			Resource: "video-id-1",
			Type:     media_domain.MediaTypeRecordingVideo,
		},
		{
			ID:       "media-id-2",
			Resource: "video-id-2",
			Type:     media_domain.MediaTypeRecordingVideo,
		},
	}
	rvs := domain.RecordedVideos{
		{
			ID:                 "recorded-video-id-1",
			RecordingChannelID: "lesson-id-1",
			Description:        "description 1",
			DateTimeRecorded:   now,
			Creator:            "user-id-1",
			CreatedAt:          now,
			UpdatedAt:          now,
			Media:              medias[0],
		},
		{
			ID:                 "recorded-video-id-2",
			RecordingChannelID: "lesson-id-1",
			Description:        "description 2",
			DateTimeRecorded:   now,
			Creator:            "user-id-2",
			CreatedAt:          now,
			UpdatedAt:          now,
			Media:              medias[1],
		},
	}

	tcs := []struct {
		name     string
		data     []byte
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "happy case",
			data: func() []byte {
				r := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_DeletedLessons_{
						DeletedLessons: &bpb.EvtLesson_DeletedLessons{
							LessonIds: lessonIds,
						},
					},
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				recordedVideoRepo.On("ListRecordingByLessonIDs", mock.Anything, tx, lessonIds).Return(rvs, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", mock.Anything, rvs.GetMediaIDs()).Once().Return(medias, nil)
				mediaModulePort.On("DeleteMedias", mock.Anything, rvs.GetMediaIDs()).Return(nil).Once()
				recordedVideoRepo.On("DeleteRecording", mock.Anything, tx, rvs.GetRecordIDs()).Return(nil).Once()

				fileStoreMock.GetObjectsWithPrefixMock = func(ctx context.Context, bucketName, prefix, delim string) ([]*filestore.StorageObject, error) {
					return []*filestore.StorageObject{
						{
							Name: rvs[0].Media.Resource,
						},
						{
							Name: rvs[1].Media.Resource,
						},
					}, nil
				}

				fileStoreMock.DeleteObjectMock = func(ctx context.Context, bucketName, objectName string) error {
					return nil
				}
			},
		},
		{
			name: "error ListRecordingByLessonIDs",
			data: func() []byte {
				r := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_DeletedLessons_{
						DeletedLessons: &bpb.EvtLesson_DeletedLessons{
							LessonIds: lessonIds,
						},
					},
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil)
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				recordedVideoRepo.On("ListRecordingByLessonIDs", mock.Anything, tx, lessonIds).Return(domain.RecordedVideos{}, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
		{
			name: "error RetrieveMediasByIDs",
			data: func() []byte {
				r := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_DeletedLessons_{
						DeletedLessons: &bpb.EvtLesson_DeletedLessons{
							LessonIds: lessonIds,
						},
					},
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()

				recordedVideoRepo.On("ListRecordingByLessonIDs", mock.Anything, tx, lessonIds).Return(rvs, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", mock.Anything, rvs.GetMediaIDs()).Once().Return(media_domain.Medias{}, fmt.Errorf("error"))
			},
			hasError: true,
		},
		{
			name: "error DeleteRecording",
			data: func() []byte {
				r := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_DeletedLessons_{
						DeletedLessons: &bpb.EvtLesson_DeletedLessons{
							LessonIds: lessonIds,
						},
					},
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil)
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()

				recordedVideoRepo.On("ListRecordingByLessonIDs", mock.Anything, tx, lessonIds).Return(rvs, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", mock.Anything, rvs.GetMediaIDs()).Once().Return(medias, nil)
				recordedVideoRepo.On("DeleteRecording", mock.Anything, tx, rvs.GetRecordIDs()).Return(fmt.Errorf("error")).Once()
			},
			hasError: true,
		},
		{
			name: "error DeleteMedias",
			data: func() []byte {
				r := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_DeletedLessons_{
						DeletedLessons: &bpb.EvtLesson_DeletedLessons{
							LessonIds: lessonIds,
						},
					},
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()

				recordedVideoRepo.On("ListRecordingByLessonIDs", mock.Anything, tx, lessonIds).Return(rvs, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", mock.Anything, rvs.GetMediaIDs()).Once().Return(medias, nil)
				recordedVideoRepo.On("DeleteRecording", mock.Anything, tx, rvs.GetRecordIDs()).Return(nil).Once()
				mediaModulePort.On("DeleteMedias", mock.Anything, rvs.GetMediaIDs()).Return(fmt.Errorf("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			l := &LessonDeletedHandler{
				Logger:            ctxzap.Extract(ctx),
				WrapperConnection: wrapperConnection,
				JSM:               jsm,
				Cfg: configurations.Config{
					Agora: configurations.AgoraConfig{
						Endpoint: "http://minio-infras.emulator.svc.cluster.local:9000",
					},
				},
				RecordedVideoRepo: recordedVideoRepo,
				MediaModulePort:   mediaModulePort,
				FileStore:         fileStoreMock,
			}

			isLoop, err := l.Handle(ctx, tc.data)
			if tc.hasError {
				require.Error(t, err)
				require.False(t, isLoop)
			} else {
				require.NoError(t, err)
				require.True(t, isLoop)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
