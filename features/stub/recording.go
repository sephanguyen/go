package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	"github.com/manabie-com/backend/internal/golibs/configs"
	recording "github.com/manabie-com/backend/internal/golibs/recording"

	"github.com/jackc/fake"
)

func handleRecording(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.URL.Path, "/acquire") {
		acquireRecording(w)
	} else if strings.Contains(req.URL.Path, "/mode/mix/start") {
		startRecording(w)
	} else if strings.Contains(req.URL.Path, "/mode/mix/query") {
		queryRecording(w)
	} else if strings.Contains(req.URL.Path, "/mode/mix/stop") {
		stopRecording(w)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func acquireRecording(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		ResourceID string `json:"resourceId"`
	}{
		ResourceID: "fake-resource-id",
	})
}

func startRecording(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		ResourceID string `json:"resourceId"`
		SID        string `json:"sid"`
	}{
		ResourceID: "fake-resource-id",
		SID:        "fake-sid",
	})
}

func queryRecording(w http.ResponseWriter) {
	expectedStatus := recording.Status{
		ResourceID: "resource-id",
		Sid:        "s-Id",
		ServerResponse: recording.ServerResponse{
			FileListMode: "file-list-mode",
			FileList: []recording.FileInfo{
				{
					Filename:       "filename_0.mp4",
					TrackType:      "track-type",
					UID:            "uid-1",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().AddDate(0, -3, 0).UnixMilli(), // three hours ago
				},
				{
					Filename:       "filename_1.mp4",
					TrackType:      "track-type",
					UID:            "uid-2",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().AddDate(0, -1, 0).UnixMilli(), // 1 hours ago
				},
			},
			Status:         5,
			SliceStartTime: 23543252,
		},
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expectedStatus)
}

func stopRecording(w http.ResponseWriter) {
	randNumer := strconv.Itoa(rand.Int())
	fileNames := []string{
		fmt.Sprintf("recording_%s/filename_%s_0.mp4", randNumer, randNumer),
		fmt.Sprintf("recording_%s/filename_%s_1.mp4", randNumer, randNumer),
		fmt.Sprintf("recording_%s/filename_%s_0.ts", randNumer, randNumer),
		fmt.Sprintf("recording_%s/filename_%s_1.ts", randNumer, randNumer),
	}
	expectedStatus := &recording.Status{
		ResourceID: "resource-id",
		Sid:        "s-Id",
		ServerResponse: recording.ServerResponse{
			FileListMode: "file-list-mode",
			FileList: []recording.FileInfo{
				{
					Filename:       fileNames[0],
					TrackType:      "track-type",
					UID:            "uid-1",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().Add(-time.Hour * 3).UnixMilli(), // three hours ago
				},
				{
					Filename:       fileNames[1],
					TrackType:      "track-type",
					UID:            "uid-2",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().Add(-time.Hour).UnixMilli(), // 1 hours ago
				},
				{
					Filename:       fileNames[2],
					TrackType:      "track-type",
					UID:            "uid-2",
					MixedAllUser:   true,
					IsPlayAble:     true,
					SliceStartTime: time.Now().Add(-time.Hour).UnixMilli(), // 1 hours ago
				},
				{
					Filename:       fileNames[3],
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

	st := &configs.StorageConfig{
		Endpoint:  "http://minio-infras.emulator.svc.cluster.local:9000",
		Bucket:    "manabie",
		AccessKey: "access_key",
		SecretKey: "secret_key",
	}
	ctx := context.Background()
	fs, err := filestore.NewMinIO(st)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to init MinIO file storage %s", err)))
		return
	}
	if err := uploadFile(ctx, fs, fileNames); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expectedStatus)
}

func uploadFile(ctx context.Context, fs *filestore.MinIO, fileNames []string) error {
	for _, v := range fileNames {
		fileData := []string{}
		for i := 0; i < 1000; i++ {
			fileData = append(fileData, fake.Words())
		}

		buf := &bytes.Buffer{}
		gob.NewEncoder(buf).Encode(fileData)
		fileByte := buf.Bytes()
		r := bytes.NewReader(fileByte)
		presignUrl, err := fs.GeneratePresignedPutObjectURL(ctx, v, time.Minute*10)
		if err != nil {
			return err
		}
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPut, presignUrl.String(), r)
		if err != nil {
			return err
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var fileBuf bytes.Buffer
		_, err = io.Copy(&fileBuf, res.Body)
		if err != nil {
			return err
		}

		if !(res.StatusCode >= 200 && res.StatusCode < 300) {
			return fmt.Errorf("expect status code 2xx, got %d", res.StatusCode)
		}
	}
	return nil
}
