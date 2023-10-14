package lessonrecording

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
)

func TestAcquireRecording(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cfg := Config{
		AppID:           "app-id",
		Cert:            "cert",
		CustomerID:      "customer-id",
		CustomerSecret:  "customer-secret",
		BucketID:        "bucket-id",
		BucketAccessKey: "bucket-access-key",
		BucketSecretKey: "bucket-secret-key",
		MaxIdleTime:     5,
	}
	resourcePath := "-2147483647"
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	lessonId := "lesson-id"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		assert.Equal(t, lessonId, b.CName)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ResourceID string `json:"resourceId"`
		}{
			ResourceID: "fake-resource-id",
		})
	}))
	defer ts.Close()
	cfg.Endpoint = ts.URL
	orgMap := map[string]string{
		"000": "-2147483642",
		"001": "-2147483647",
	}
	rec, err := NewRecorder(ctx, cfg, ctxzap.Extract(context.Background()), lessonId, orgMap)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = rec.Acquire()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, "fake-resource-id", rec.RID)
	assert.True(t, strings.Contains(fmt.Sprint(rec.UID), getIDByResourcePath(orgMap, resourcePath)))
}

func TestStartRecording(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	lessonId := "lesson-id"
	bucketId := "bucket-id"
	bAccessKey := "bucket-access-key"
	bSecretKey := "bucket-secret-key"
	resourcePath := "-2147483647"
	maxIdleTime := 5

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)

	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	subscribeVideoUids := []string{"1000061831", "1000061832"}
	subscribeAudioUids := []string{"#allstream#"}
	fileNamePrefix := []string{"lesson-id", currentTime}

	type TranscodingConfig struct {
		Height           int32  `json:"height"`
		Width            int32  `json:"width"`
		Bitrate          int32  `json:"bitrate"`
		Fps              int32  `json:"fps"`
		MixedVideoLayout int32  `json:"mixedVideoLayout"`
		BackgroundColor  string `json:"backgroundColor"`
	}

	tr := TranscodingConfig{
		Height:           720,
		Width:            1280,
		Bitrate:          2260,
		Fps:              15,
		MixedVideoLayout: 1,
		BackgroundColor:  "#000000",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var b struct {
			CName         string `json:"cname"`
			ClientRequest struct {
				RecordingConfig struct {
					TranscodingConfig  TranscodingConfig `json:"transcodingConfig"`
					SubscribeVideoUIds []string          `json:"subscribeVideoUids"`
					SubscribeAudioUids []string          `json:"subscribeAudioUids"`
					MaxIdleTime        int               `json:"maxIdleTime"`
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
		assert.Equal(t, lessonId, b.CName)
		reConfig := b.ClientRequest.RecordingConfig.TranscodingConfig
		assert.Equal(t, tr.Height, reConfig.Height)
		assert.Equal(t, tr.Width, reConfig.Width)
		assert.Equal(t, tr.Bitrate, reConfig.Bitrate)
		assert.Equal(t, tr.Fps, reConfig.Fps)
		assert.Equal(t, tr.BackgroundColor, reConfig.BackgroundColor)
		assert.Equal(t, tr.MixedVideoLayout, reConfig.MixedVideoLayout)

		assert.Equal(t, subscribeAudioUids, b.ClientRequest.RecordingConfig.SubscribeAudioUids)
		assert.Equal(t, subscribeVideoUids, b.ClientRequest.RecordingConfig.SubscribeVideoUIds)

		assert.Equal(t, maxIdleTime, b.ClientRequest.RecordingConfig.MaxIdleTime)

		assert.Equal(t, bucketId, b.ClientRequest.StorageConfig.Bucket)
		assert.Equal(t, bAccessKey, b.ClientRequest.StorageConfig.AccessKey)
		assert.Equal(t, bSecretKey, b.ClientRequest.StorageConfig.SecretKey)
		assert.Equal(t, fileNamePrefix, b.ClientRequest.StorageConfig.FileNamePrefix)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ResourceID string `json:"resourceId"`
			SID        string `json:"sid"`
		}{
			ResourceID: "fake-resource-id",
			SID:        "fake-sid",
		})
	}))
	defer ts.Close()

	cfg := Config{
		AppID:           "app-id",
		Cert:            "cert",
		CustomerID:      "customer-id",
		CustomerSecret:  "customer-secret",
		BucketID:        "bucket-id",
		BucketAccessKey: "bucket-access-key",
		BucketSecretKey: "bucket-secret-key",
		Endpoint:        ts.URL,
		MaxIdleTime:     5,
	}
	transcodingConfig := fmt.Sprintf(`
		{
			"height": %d,
			"width": %d,
			"bitrate": %d,
			"fps": %d,
			"mixedVideoLayout": %d,
			"backgroundColor": "%s"
		}`, tr.Height, tr.Width, tr.Bitrate, tr.Fps, tr.MixedVideoLayout, tr.BackgroundColor)
	orgMap := map[string]string{
		"000": "-2147483642",
		"001": "-2147483647",
	}
	rec, err := NewRecorder(ctx, cfg, ctxzap.Extract(context.Background()), lessonId, orgMap)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	sc := &StartCall{
		SubscribeVideoUids:    subscribeVideoUids,
		SubscribeAudioUids:    subscribeAudioUids,
		FileNamePrefix:        fileNamePrefix,
		TranscodingConfigJSON: transcodingConfig,
	}
	_, err = rec.Start(sc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, "fake-sid", rec.SID)
}

func TestQueryStatusRecording(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	lessonId := "lesson-id"
	bucketId := "bucket-id"
	bAccessKey := "bucket-access-key"
	bSecretKey := "bucket-secret-key"
	resourceId := "resource-id"
	sId := "s-id"
	expectedStatus := &Status{
		ResourceID: resourceId,
		Sid:        sId,
		ServerResponse: ServerResponse{
			FileListMode: "file-list-mode",
			FileList: []FileInfo{
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
			Status:         200,
			SliceStartTime: 23543252,
		},
	}

	expectURL := ""
	cfg := Config{
		AppID:           "app-id",
		Cert:            "cert",
		CustomerID:      "customer-id",
		CustomerSecret:  "customer-secret",
		BucketID:        bucketId,
		BucketAccessKey: bAccessKey,
		BucketSecretKey: bSecretKey,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actualURL := cfg.Endpoint + req.URL.Path
		assert.Equal(t, expectURL, actualURL)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedStatus)
	}))
	defer ts.Close()

	cfg.Endpoint = ts.URL

	orgMap := map[string]string{
		"000": "-2147483642",
		"001": "-2147483647",
	}

	resourcePath := "-2147483647"

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	rec, err := NewRecorder(ctx, cfg, ctxzap.Extract(context.Background()), lessonId, orgMap)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	rec.RID = resourceId
	rec.SID = sId
	expectURL = rec.Configs.Endpoint + "/v1/apps/" + rec.Configs.AppID + "/cloud_recording/resourceid/" + rec.RID + "/sid/" + rec.SID + "/mode/mix/query"
	st, err := rec.CallStatusAPI()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, expectedStatus, st)
}

func TestStopRecording(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resourcePath := "resource-path"
	lessonId := "lesson-id"
	bucketId := "bucket-id"
	bAccessKey := "bucket-access-key"
	bSecretKey := "bucket-secret-key"
	resourceId := "resource-id"
	sId := "s-id"
	uID := 4321

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)

	expectedStatus := &Status{
		ResourceID: resourceId,
		Sid:        sId,
		ServerResponse: ServerResponse{
			FileListMode: "file-list-mode",
			FileList: []FileInfo{
				{
					Filename:       "filename-1",
					TrackType:      "track-type",
					UID:            "uid-1",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().AddDate(0, -3, 0).UnixMilli(),
				},
				{
					Filename:       "filename-2",
					TrackType:      "track-type",
					UID:            "uid-2",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().AddDate(0, -1, 0).UnixMilli(), // 1 hours ago
				},
			},
			Status:          200,
			SliceStartTime:  23543252,
			UploadingStatus: "uploaded",
		},
	}

	expectURL := ""
	cfg := Config{
		AppID:           "app-id",
		Cert:            "cert",
		CustomerID:      "customer-id",
		CustomerSecret:  "customer-secret",
		BucketID:        bucketId,
		BucketAccessKey: bAccessKey,
		BucketSecretKey: bSecretKey,
	}
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
		assert.Equal(t, fmt.Sprintf(UIDFormat, uID), b.UID)
		actualURL := cfg.Endpoint + req.URL.Path
		assert.Equal(t, expectURL, actualURL)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedStatus)
	}))
	defer ts.Close()

	cfg.Endpoint = ts.URL
	rec := GetExistingRecorder(ctx, cfg, ctxzap.Extract(context.Background()), uID, lessonId, resourceId, sId)

	rec.RID = resourceId
	rec.SID = sId
	expectURL = rec.Configs.Endpoint + "/v1/apps/" + rec.Configs.AppID + "/cloud_recording/resourceid/" + rec.RID + "/sid/" + rec.SID + "/mode/mix/stop"

	st, err := rec.Stop()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	assert.Equal(t, expectedStatus, st)
}
