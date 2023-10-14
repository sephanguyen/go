package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	recording "github.com/manabie-com/backend/internal/golibs/recording"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	lr_queries "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	lr_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_liveroom_repo "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStartRecordingService(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	userId := "user-id"
	resourcePath := "-2147483647"
	ctx = interceptors.ContextWithUserID(ctx, userId)
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRoomStateRepo := &mock_virtual_repo.MockLessonRoomStateRepo{}
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	orgRepo := &mock_virtual_repo.MockOrganizationRepo{}
	lessonId := "lesson-id"
	channelId := "channel-id"
	bucketId := "bucket-id"
	bAccessKey := "bucket-access-key"
	bSecretKey := "bucket-secret-key"
	resourceId := "fake-resource-id"
	sId := "fake-sid"
	orgIDs := []string{
		"-2147483642",
		"-2147483647",
	}
	lrs := &LessonRecordingService{
		Cfg: configurations.Config{
			Agora: configurations.AgoraConfig{
				AppID:           "app-id",
				Cert:            "cert",
				CustomerID:      "customer-id",
				CustomerSecret:  "customer-secret",
				BucketName:      bucketId,
				BucketAccessKey: bAccessKey,
				BucketSecretKey: bSecretKey,
			},
		},
		Logger: ctxzap.Extract(ctx),
		RecordingCommand: commands.RecordingCommand{
			WrapperDBConnection: wrapperConnection,
			LessonmgmtDB:        db,
			LessonRoomStateRepo: lessonRoomStateRepo,
			LiveRoomStateRepo:   liveRoomStateRepo,
		},
		LessonRoomStateQuery: queries.LessonRoomStateQuery{
			WrapperDBConnection: wrapperConnection,
			LessonRoomStateRepo: lessonRoomStateRepo,
		},
		LiveRoomStateQuery: lr_queries.LiveRoomStateQuery{
			LessonmgmtDB:      db,
			LiveRoomStateRepo: liveRoomStateRepo,
		},
		OrganizationQuery: queries.OrganizationQuery{
			WrapperDBConnection: wrapperConnection,
			OrganizationRepo:    orgRepo,
		},
	}
	tr := &vpb.StartRecordingRequest_TranscodingConfig{
		Height:           720,
		Width:            1280,
		Bitrate:          2260,
		Fps:              15,
		MixedVideoLayout: 1,
		BackgroundColor:  "#000000",
	}
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	startReq := &vpb.StartRecordingRequest{
		LessonId:           lessonId,
		SubscribeVideoUids: []string{"1000061831", "1000061832"},
		SubscribeAudioUids: []string{"#allstream#"},
		FileNamePrefix:     []string{lessonId, currentTime},
		TranscodingConfig:  tr,
		ChannelId:          channelId,
	}

	expectedChannel := ""
	acquireRecording := func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var b struct {
			CName string `json:"cname"`
		}
		err := decoder.Decode(&b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		assert.Equal(t, expectedChannel, b.CName)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ResourceID string `json:"resourceId"`
		}{
			ResourceID: resourceId,
		})
	}

	startRecording := func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var b struct {
			CName         string `json:"cname"`
			ClientRequest struct {
				RecordingConfig struct {
					TranscodingConfig struct {
						Height           int32  `json:"height"`
						Width            int32  `json:"width"`
						Bitrate          int32  `json:"bitrate"`
						Fps              int32  `json:"fps"`
						MixedVideoLayout int32  `son:"mixedVideoLayout"`
						BackgroundColor  string `json:"backgroundColor"`
					} `json:"transcodingConfig"`
					SubscribeVideoUIds []string `json:"subscribeVideoUids"`
					SubscribeAudioUids []string `json:"subscribeAudioUids"`
				}
				StorageConfig struct {
					Bucket         string   `json:"bucket"`
					AccessKey      string   `json:"accessKey"`
					SecretKey      string   `json:"secretKey"`
					FileNamePrefix []string `json:"fileNamePrefix"`
				}
			} `json:"clientRequest"`
		}
		err := decoder.Decode(&b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		assert.Equal(t, expectedChannel, b.CName)
		reConfig := b.ClientRequest.RecordingConfig.TranscodingConfig
		assert.Equal(t, tr.Height, reConfig.Height)
		assert.Equal(t, tr.Width, reConfig.Width)
		assert.Equal(t, tr.Bitrate, reConfig.Bitrate)
		assert.Equal(t, tr.Fps, reConfig.Fps)
		assert.Equal(t, tr.BackgroundColor, reConfig.BackgroundColor)
		assert.Equal(t, tr.MixedVideoLayout, reConfig.MixedVideoLayout)

		assert.Equal(t, startReq.SubscribeAudioUids, b.ClientRequest.RecordingConfig.SubscribeAudioUids)
		assert.Equal(t, startReq.SubscribeVideoUids, b.ClientRequest.RecordingConfig.SubscribeVideoUIds)

		assert.Equal(t, bucketId, b.ClientRequest.StorageConfig.Bucket)
		assert.Equal(t, bAccessKey, b.ClientRequest.StorageConfig.AccessKey)
		assert.Equal(t, bSecretKey, b.ClientRequest.StorageConfig.SecretKey)
		assert.Equal(t, startReq.FileNamePrefix, b.ClientRequest.StorageConfig.FileNamePrefix)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ResourceID string `json:"resourceId"`
			SID        string `json:"sid"`
		}{
			ResourceID: resourceId,
			SID:        sId,
		})
	}
	expectedStatus := recording.Status{
		ResourceID: resourceId,
		Sid:        sId,
		ServerResponse: recording.ServerResponse{
			FileListMode: "file-list-mode",
			FileList: []recording.FileInfo{
				{
					Filename:       "filename-1",
					TrackType:      "track-type",
					UID:            "uid-1",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: 352432452,
				},
				{
					Filename:       "filename-2",
					TrackType:      "track-type",
					UID:            "uid-2",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: 352432452,
				},
			},
			Status:         5,
			SliceStartTime: 23543252,
		},
	}

	t.Run("success", func(t *testing.T) {
		startReq.ChannelId = ""
		expectedChannel = lessonId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Times(3)
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "lesson-id").
			Return(nil, vc_domain.ErrLessonRoomStateNotFound).Once()
		queryRecording := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedStatus)
		}

		orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			taskUUID := req.URL.Path
			if strings.Contains(taskUUID, "/acquire") {
				acquireRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/start") {
				startRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/query") {
				queryRecording(w, req)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		lessonRoomStateRepo.On("UpsertRecordingState", ctx, db, "lesson-id", mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
			assert.Equal(t, req.Creator, userId)
			assert.Equal(t, req.IsRecording, true)
			assert.Equal(t, req.SID, sId)
			assert.Equal(t, req.ResourceID, resourceId)
			assert.NotEqual(t, req.UID, 0)
			return true
		})).
			Return(nil).Once()
		_, err := lrs.StartRecording(ctx, startReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
	t.Run("success when call query status 206", func(t *testing.T) {
		startReq.ChannelId = ""
		expectedChannel = lessonId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Times(3)
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "lesson-id").
			Return(nil, vc_domain.ErrLessonRoomStateNotFound).Once()
		queryRecording := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusPartialContent)
			json.NewEncoder(w).Encode(expectedStatus)
		}

		orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			taskUUID := req.URL.Path
			if strings.Contains(taskUUID, "/acquire") {
				acquireRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/start") {
				startRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/query") {
				queryRecording(w, req)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		lessonRoomStateRepo.On("UpsertRecordingState", ctx, db, "lesson-id", mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
			assert.Equal(t, req.Creator, userId)
			assert.Equal(t, req.IsRecording, true)
			assert.Equal(t, req.SID, sId)
			assert.Equal(t, req.ResourceID, resourceId)
			assert.NotEqual(t, req.UID, 0)
			return true
		})).
			Return(nil).Once()
		_, err := lrs.StartRecording(ctx, startReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("success with live room", func(t *testing.T) {
		startReq.ChannelId = channelId
		expectedChannel = channelId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Twice()
		liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, channelId).
			Return(nil, lr_domain.ErrChannelNotFound).Once()
		queryRecording := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedStatus)
		}

		orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			taskUUID := req.URL.Path
			if strings.Contains(taskUUID, "/acquire") {
				acquireRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/start") {
				startRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/query") {
				queryRecording(w, req)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		liveRoomStateRepo.On("UpsertRecordingState", ctx, db, channelId, mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
			assert.Equal(t, req.Creator, userId)
			assert.Equal(t, req.IsRecording, true)
			assert.Equal(t, req.SID, sId)
			assert.Equal(t, req.ResourceID, resourceId)
			assert.NotEqual(t, req.UID, 0)
			return true
		})).Return(nil).Once()

		_, err := lrs.StartRecording(ctx, startReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail when call query status != 200 and != 206", func(t *testing.T) {
		startReq.ChannelId = ""
		expectedChannel = lessonId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Twice()
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "lesson-id").
			Return(nil, vc_domain.ErrLessonRoomStateNotFound).Once()
		queryRecording := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}

		orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			taskUUID := req.URL.Path
			if strings.Contains(taskUUID, "/acquire") {
				acquireRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/start") {
				startRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/query") {
				queryRecording(w, req)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StartRecording(ctx, startReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
	t.Run("fail when recording the live lesson which already recorded", func(t *testing.T) {
		startReq.ChannelId = ""
		expectedChannel = lessonId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
			Return(&vc_domain.LessonRoomState{
				LessonID: lessonId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         123324,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		queryRecording := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}

		orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			taskUUID := req.URL.Path
			if strings.Contains(taskUUID, "/acquire") {
				acquireRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/start") {
				startRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/query") {
				queryRecording(w, req)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StartRecording(ctx, startReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail when recording the live room which already recording", func(t *testing.T) {
		startReq.ChannelId = channelId
		expectedChannel = channelId

		liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, channelId).
			Return(&lr_domain.LiveRoomState{
				ChannelID: channelId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         123324,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		queryRecording := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}

		orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			taskUUID := req.URL.Path
			if strings.Contains(taskUUID, "/acquire") {
				acquireRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/start") {
				startRecording(w, req)
			} else if strings.Contains(taskUUID, "/mode/mix/query") {
				queryRecording(w, req)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StartRecording(ctx, startReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}

func TestStopRecordingService(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	userId := "user-id"
	resourcePath := "resource-path"
	ctx = interceptors.ContextWithUserID(ctx, userId)
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRoomStateRepo := &mock_virtual_repo.MockLessonRoomStateRepo{}
	recordedVideoRepo := &mock_virtual_repo.MockRecordedVideoRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	liveRoomRecordedVideosRepo := &mock_liveroom_repo.MockLiveRoomRecordedVideosRepo{}

	size := int64(20000)
	fileStoreMock := &filestore.Mock{
		GetObjectInfoMock: func(ctx context.Context, bucketName, objectName string) (*filestore.StorageObject, error) {
			return &filestore.StorageObject{Size: size}, nil
		},
	}
	lessonId := "lesson-id"
	channelId := "channel-id"
	bucketId := "bucket-id"
	bAccessKey := "bucket-access-key"
	bSecretKey := "bucket-secret-key"
	resourceId := "fake-resource-id"
	sId := "fake-sid"
	uID := 4321
	lrs := &LessonRecordingService{
		Cfg: configurations.Config{
			Agora: configurations.AgoraConfig{
				AppID:           "app-id",
				Cert:            "cert",
				CustomerID:      "customer-id",
				CustomerSecret:  "customer-secret",
				BucketName:      bucketId,
				BucketAccessKey: bAccessKey,
				BucketSecretKey: bSecretKey,
			},
		},
		Logger: ctxzap.Extract(ctx),
		RecordingCommand: commands.RecordingCommand{
			WrapperDBConnection:        wrapperConnection,
			LessonmgmtDB:               db,
			LessonRoomStateRepo:        lessonRoomStateRepo,
			LiveRoomStateRepo:          liveRoomStateRepo,
			LiveRoomRecordedVideosRepo: liveRoomRecordedVideosRepo,
			RecordedVideoRepo:          recordedVideoRepo,
			MediaModulePort:            mediaModulePort,
		},
		LessonRoomStateQuery: queries.LessonRoomStateQuery{
			WrapperDBConnection: wrapperConnection,
			LessonRoomStateRepo: lessonRoomStateRepo,
		},
		LiveRoomStateQuery: lr_queries.LiveRoomStateQuery{
			LessonmgmtDB:      db,
			LiveRoomStateRepo: liveRoomStateRepo,
		},
		FileStore: fileStoreMock,
	}

	stopReq := &vpb.StopRecordingRequest{
		LessonId:  lessonId,
		ChannelId: channelId,
	}

	expectedStatus := recording.Status{
		ResourceID: resourceId,
		Sid:        sId,
		ServerResponse: recording.ServerResponse{
			FileListMode: "file-list-mode",
			FileList: []recording.FileInfo{
				{
					Filename:       "filename_0.mp4",
					TrackType:      "track-type",
					UID:            "uid-1",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().Add(-time.Hour * 3).UnixMilli(), // three hours ago
				},
				{
					Filename:       "filename_1.mp4",
					TrackType:      "track-type",
					UID:            "uid-2",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().Add(-time.Hour).UnixMilli(), // 1 hours ago
				},
			},
			Status:          5,
			SliceStartTime:  23543252,
			UploadingStatus: "uploaded",
		},
	}

	t.Run("success", func(t *testing.T) {
		stopReq.ChannelId = ""

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Times(3)
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
			Return(&vc_domain.LessonRoomState{
				LessonID: lessonId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         uID,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			decoder := json.NewDecoder(req.Body)
			var b struct {
				Cname         string `json:"cname"`
				UID           string `json:"uid"`
				ClientRequest struct {
				} `json:"clientRequest"`
			}
			err := decoder.Decode(&b)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			assert.Equal(t, lessonId, b.Cname)
			assert.Equal(t, fmt.Sprintf(recording.UIDFormat, uID), b.UID)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		lessonRoomStateRepo.On("UpsertRecordingState", ctx, db, "lesson-id", mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
			var r *vc_domain.CompositeRecordingState
			r = nil
			assert.Equal(t, r, req)
			return true
		})).
			Return(nil).Once()
		fl := expectedStatus.ServerResponse.FileList
		for i := 0; i < len(fl); i++ {
			mediaModulePort.On("CreateMedia", ctx, mock.MatchedBy(func(m *media_domain.Media) bool {
				for i, v := range fl {
					if v.Filename == m.Resource {
						assert.Equal(t, m.Resource, v.Filename)
						assert.Equal(t, m.Type, media_domain.MediaTypeRecordingVideo)
						assert.Equal(t, m.FileSizeBytes, size)
						if i == 0 {
							assert.Equal(t, m.Duration, time.Duration(time.Hour*2))
						}
						if i == 1 {
							assert.Equal(t, int(m.Duration.Minutes()), int(time.Duration(time.Hour*1).Minutes()))
						}

						break
					}
				}
				return true
			})).Return(nil, nil).Once()
		}

		recordedVideoRepo.On("InsertRecordedVideos", ctx, db, mock.MatchedBy(func(rs []*vc_domain.RecordedVideo) bool {
			for _, v := range rs {
				assert.Equal(t, lessonId, v.RecordingChannelID)
				assert.Equal(t, userId, v.Creator)
				for _, c := range fl {
					if c.Filename == v.Media.Resource {
						assert.Equal(t, time.UnixMilli(c.SliceStartTime), v.DateTimeRecorded)
						break
					}
				}
			}
			return true
		})).Return(nil)

		_, err := lrs.StopRecording(ctx, stopReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("success with live room but save in lesson", func(t *testing.T) {
		stopReq.ChannelId = channelId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Twice()
		liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, channelId).
			Return(&lr_domain.LiveRoomState{
				ChannelID: channelId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         uID,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			decoder := json.NewDecoder(req.Body)
			var b struct {
				Cname         string `json:"cname"`
				UID           string `json:"uid"`
				ClientRequest struct {
				} `json:"clientRequest"`
			}
			err := decoder.Decode(&b)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			assert.Equal(t, channelId, b.Cname)
			assert.Equal(t, fmt.Sprintf(recording.UIDFormat, uID), b.UID)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL
		liveRoomStateRepo.On("UpsertRecordingState", ctx, db, channelId, mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
			var r *vc_domain.CompositeRecordingState
			r = nil
			assert.Equal(t, r, req)
			return true
		})).Return(nil).Once()

		fl := expectedStatus.ServerResponse.FileList
		for i := 0; i < len(fl); i++ {
			mediaModulePort.On("CreateMedia", ctx, mock.MatchedBy(func(m *media_domain.Media) bool {
				for i, v := range fl {
					if v.Filename == m.Resource {
						assert.Equal(t, m.Resource, v.Filename)
						assert.Equal(t, m.Type, media_domain.MediaTypeRecordingVideo)
						assert.Equal(t, m.FileSizeBytes, size)
						if i == 0 {
							assert.Equal(t, m.Duration, time.Duration(time.Hour*2))
						}
						if i == 1 {
							assert.Equal(t, int(m.Duration.Minutes()), int(time.Duration(time.Hour*1).Minutes()))
						}

						break
					}
				}
				return true
			})).Return(nil, nil).Once()
		}

		recordedVideoRepo.On("InsertRecordedVideos", ctx, db, mock.MatchedBy(func(rs []*vc_domain.RecordedVideo) bool {
			for _, v := range rs {
				assert.Equal(t, lessonId, v.RecordingChannelID)
				assert.Equal(t, userId, v.Creator)
				for _, c := range fl {
					if c.Filename == v.Media.Resource {
						assert.Equal(t, time.UnixMilli(c.SliceStartTime), v.DateTimeRecorded)
						break
					}
				}
			}
			return true
		})).Return(nil)

		_, err := lrs.StopRecording(ctx, stopReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("success with live room but no save in lesson", func(t *testing.T) {
		stopReq.ChannelId = channelId
		stopReq.LessonId = ""

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Twice()
		liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, channelId).
			Return(&lr_domain.LiveRoomState{
				ChannelID: channelId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         uID,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			decoder := json.NewDecoder(req.Body)
			var b struct {
				Cname         string `json:"cname"`
				UID           string `json:"uid"`
				ClientRequest struct {
				} `json:"clientRequest"`
			}
			err := decoder.Decode(&b)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			assert.Equal(t, channelId, b.Cname)
			assert.Equal(t, fmt.Sprintf(recording.UIDFormat, uID), b.UID)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL
		liveRoomStateRepo.On("UpsertRecordingState", ctx, db, channelId, mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
			var r *vc_domain.CompositeRecordingState
			r = nil
			assert.Equal(t, r, req)
			return true
		})).Return(nil).Once()

		fl := expectedStatus.ServerResponse.FileList
		for i := 0; i < len(fl); i++ {
			mediaModulePort.On("CreateMedia", ctx, mock.MatchedBy(func(m *media_domain.Media) bool {
				for i, v := range fl {
					if v.Filename == m.Resource {
						assert.Equal(t, m.Resource, v.Filename)
						assert.Equal(t, m.Type, media_domain.MediaTypeRecordingVideo)
						assert.Equal(t, m.FileSizeBytes, size)
						if i == 0 {
							assert.Equal(t, m.Duration, time.Duration(time.Hour*2))
						}
						if i == 1 {
							assert.Equal(t, int(m.Duration.Minutes()), int(time.Duration(time.Hour*1).Minutes()))
						}

						break
					}
				}
				return true
			})).Return(nil, nil).Once()
		}

		liveRoomRecordedVideosRepo.On("InsertRecordedVideos", ctx, db, mock.MatchedBy(func(rs []*vc_domain.RecordedVideo) bool {
			for _, v := range rs {
				assert.Equal(t, channelId, v.RecordingChannelID)
				assert.Equal(t, userId, v.Creator)
				for _, c := range fl {
					if c.Filename == v.Media.Resource {
						assert.Equal(t, time.UnixMilli(c.SliceStartTime), v.DateTimeRecorded)
						break
					}
				}
			}
			return true
		})).Return(nil)

		_, err := lrs.StopRecording(ctx, stopReq)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail when call stop recording fail", func(t *testing.T) {
		stopReq.ChannelId = ""
		stopReq.LessonId = lessonId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
			Return(&vc_domain.LessonRoomState{
				LessonID: lessonId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         uID,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StopRecording(ctx, stopReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail when call stop recording fail in live room", func(t *testing.T) {
		stopReq.ChannelId = channelId

		liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, channelId).
			Return(&lr_domain.LiveRoomState{
				ChannelID: channelId,
				Recording: &vc_domain.CompositeRecordingState{
					ResourceID:  resourceId,
					SID:         sId,
					UID:         uID,
					IsRecording: true,
					Creator:     userId,
				},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StopRecording(ctx, stopReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail when stop the live lesson which is not recording", func(t *testing.T) {
		stopReq.ChannelId = ""
		stopReq.LessonId = lessonId

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
			Return(&vc_domain.LessonRoomState{
				LessonID:  lessonId,
				Recording: &vc_domain.CompositeRecordingState{},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StopRecording(ctx, stopReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail when stop the live room which is not recording", func(t *testing.T) {
		stopReq.ChannelId = channelId
		stopReq.LessonId = lessonId

		liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, channelId).
			Return(&lr_domain.LiveRoomState{
				ChannelID: channelId,
				Recording: &vc_domain.CompositeRecordingState{},
			}, nil).Once()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(expectedStatus)
		}))
		defer ts.Close()
		lrs.Cfg.Agora.Endpoint = ts.URL

		_, err := lrs.StopRecording(ctx, stopReq)
		if err == nil {
			t.Errorf("expected error: %v", err)
		}
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}

func TestLessonRoomStateQuery_GetRecordingByLessonID(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	recordedVideoRepo := &mock_virtual_repo.MockRecordedVideoRepo{}
	lessonId := "lesson-1"
	now := time.Now()
	total := uint32(10)

	recs := vc_domain.RecordedVideos{
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
		var limit uint32 = 2
		prePageID := "pre-page-id"
		offsetString := "record-1"
		payload := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			LessonID:        lessonId,
			Limit:           limit,
			RecordedVideoID: offsetString,
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("ListRecordingByLessonIDWithPaging", ctx, db, payload).Return(recs, total, prePageID, uint32(5), nil).Once()

		l := &LessonRecordingService{
			RecordingQuery: queries.RecordedVideoQuery{
				WrapperDBConnection: wrapperConnection,
				RecordedVideoRepo:   recordedVideoRepo,
			},
		}
		res, err := l.GetRecordingByLessonID(ctx, &vpb.GetRecordingByLessonIDRequest{
			LessonId: lessonId,
			Paging: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: offsetString,
				},
			},
		})

		result := &vpb.GetRecordingByLessonIDResponse{
			Items: []*vpb.GetRecordingByLessonIDResponse_RecordingItem{
				{
					Id:        recs[0].ID,
					StartTime: timestamppb.New(recs[0].DateTimeRecorded),
					Duration:  durationpb.New(recs[0].Media.Duration),
					FileSize:  float32(recs[0].Media.FileSizeBytes),
				},
				{
					Id:        recs[1].ID,
					StartTime: timestamppb.New(recs[1].DateTimeRecorded),
					Duration:  durationpb.New(recs[1].Media.Duration),
					FileSize:  float32(recs[1].Media.FileSizeBytes),
				},
			},
			NextPage: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: recs[1].ID,
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: prePageID,
				},
			},
			TotalItems: total,
		}

		assert.Equal(t, result, res)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("success with preTotal <= limit", func(t *testing.T) {
		var limit uint32 = 2
		offsetString := "record-1"
		prePageID := ""
		payload := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			LessonID:        lessonId,
			Limit:           limit,
			RecordedVideoID: offsetString,
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("ListRecordingByLessonIDWithPaging", ctx, db, payload).Return(recs, total, prePageID, uint32(1), nil).Once()

		l := &LessonRecordingService{
			RecordingQuery: queries.RecordedVideoQuery{
				WrapperDBConnection: wrapperConnection,
				RecordedVideoRepo:   recordedVideoRepo,
			},
		}
		res, err := l.GetRecordingByLessonID(ctx, &vpb.GetRecordingByLessonIDRequest{
			LessonId: lessonId,
			Paging: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: offsetString,
				},
			},
		})

		result := &vpb.GetRecordingByLessonIDResponse{
			Items: []*vpb.GetRecordingByLessonIDResponse_RecordingItem{
				{
					Id:        recs[0].ID,
					StartTime: timestamppb.New(recs[0].DateTimeRecorded),
					Duration:  durationpb.New(recs[0].Media.Duration),
					FileSize:  float32(recs[0].Media.FileSizeBytes),
				},
				{
					Id:        recs[1].ID,
					StartTime: timestamppb.New(recs[1].DateTimeRecorded),
					Duration:  durationpb.New(recs[1].Media.Duration),
					FileSize:  float32(recs[1].Media.FileSizeBytes),
				},
			},
			NextPage: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: recs[1].ID,
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: prePageID,
				},
			},
			TotalItems: total,
		}

		assert.Equal(t, result, res)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("fail", func(t *testing.T) {
		var limit uint32 = 2
		offsetString := "record-1"
		payload := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			LessonID:        lessonId,
			Limit:           limit,
			RecordedVideoID: offsetString,
		}
		expectedErr := fmt.Errorf("error when call recordedVideoRepo.ListRecordingByLessonIDWithPaging")
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideoRepo.On("ListRecordingByLessonIDWithPaging", ctx, db, payload).Return(vc_domain.RecordedVideos{}, uint32(0), "", uint32(0), expectedErr).Once()

		l := &LessonRecordingService{
			RecordingQuery: queries.RecordedVideoQuery{
				WrapperDBConnection: wrapperConnection,
				RecordedVideoRepo:   recordedVideoRepo,
			},
		}
		res, err := l.GetRecordingByLessonID(ctx, &vpb.GetRecordingByLessonIDRequest{
			LessonId: lessonId,
			Paging: &cpb.Paging{
				Limit: limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: offsetString,
				},
			},
		})
		var exRes *vpb.GetRecordingByLessonIDResponse = nil
		assert.Equal(t, exRes, res)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}

func TestLessonRecordingService_GetRecordingDownloadLinkByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expectedURL := "https://example.com/manabie/%s?key=12345566"
	expectedObject := "expected-object"

	recordedVideo := &mock_virtual_repo.MockRecordedVideoRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	fileStoreMock := &filestore.Mock{
		GenerateGetObjectURLMock: func(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error) {
			return url.Parse(fmt.Sprintf(expectedURL, objectName))
		},
	}
	recordingID := "recording-id"

	exObject := &vc_domain.RecordedVideo{
		ID: recordingID,
		Media: &media_domain.Media{
			ID:       "media-id",
			Resource: expectedObject,
		}}

	t.Run("get successful", func(t *testing.T) {
		l := &LessonRecordingService{
			RecordingQuery: queries.RecordedVideoQuery{
				WrapperDBConnection: wrapperConnection,
				RecordedVideoRepo:   recordedVideo,
				MediaModulePort:     mediaModulePort,
			},
			FileStore: fileStoreMock,
		}

		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideo.On("GetRecordingByID", ctx, db, &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}).Once().Return(exObject, nil)
		mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{exObject.Media.ID}).Return(media_domain.Medias{
			exObject.Media}, nil).Once()
		res, err := l.GetRecordingDownloadLinkByID(ctx, &vpb.GetRecordingDownloadLinkByIDRequest{
			RecordingId: recordingID,
			Expiry:      durationpb.New(time.Hour),
			FileName:    "es300_20220824_150034.mp4",
		})

		assert.Nil(t, err)
		assert.Equal(t, res.GetUrl(), fmt.Sprintf(expectedURL, expectedObject))
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("get fail because not existed record", func(t *testing.T) {
		l := &LessonRecordingService{
			RecordingQuery: queries.RecordedVideoQuery{
				WrapperDBConnection: wrapperConnection,
				RecordedVideoRepo:   recordedVideo,
			},
		}
		expectedErr := status.Error(codes.Internal, "error when call RecordedVideoRepo.GetRecordingByID: rpc error: code = Internal desc = error message")
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideo.On("GetRecordingByID", ctx, db, &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}).Once().Return(nil, status.Error(codes.Internal, "error message"))
		res, err := l.GetRecordingDownloadLinkByID(ctx, &vpb.GetRecordingDownloadLinkByIDRequest{
			RecordingId: recordingID,
			Expiry:      durationpb.New(time.Hour),
			FileName:    "es300_20220824_150034.mp4",
		})

		assert.Equal(t, err, expectedErr)
		assert.Equal(t, res.GetUrl(), "")
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})

	t.Run("get fail because get url fail", func(t *testing.T) {
		expectedErr := status.Error(codes.Internal, "error when FileStore.GenerateGetObjectURL: rpc error: code = Internal desc = error message")
		fileStoreMock := &filestore.Mock{
			GenerateGetObjectURLMock: func(ctx context.Context, objectName string, fileName string, expiry time.Duration) (*url.URL, error) {
				return nil, status.Error(codes.Internal, "error message")
			},
		}

		l := &LessonRecordingService{
			RecordingQuery: queries.RecordedVideoQuery{
				WrapperDBConnection: wrapperConnection,
				RecordedVideoRepo:   recordedVideo,
				MediaModulePort:     mediaModulePort,
			},
			FileStore: fileStoreMock,
		}
		mockUnleashClient.
			On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil).Once()
		recordedVideo.On("GetRecordingByID", ctx, db, &payloads.GetRecordingByIDPayload{
			RecordedVideoID: recordingID,
		}).Once().Return(exObject, nil)
		mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{exObject.Media.ID}).Return(media_domain.Medias{
			exObject.Media}, nil).Once()
		res, err := l.GetRecordingDownloadLinkByID(ctx, &vpb.GetRecordingDownloadLinkByIDRequest{
			RecordingId: recordingID,
			Expiry:      durationpb.New(time.Hour),
			FileName:    "es300_20220824_150034.mp4",
		})

		assert.Equal(t, err, expectedErr)
		assert.Equal(t, res.GetUrl(), "")
		mock.AssertExpectationsForObjects(t, mockUnleashClient)
	})
}
